package releasequality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeWithEmptyInputUsesDefaultThresholdAndFails(t *testing.T) {
	chdir(t, t.TempDir())

	got := Compute(0)

	if got.Threshold != DefaultMinimum {
		t.Fatalf("Compute(0).Threshold = %.1f; want %.1f", got.Threshold, DefaultMinimum)
	}
	if got.Value != 0 {
		t.Fatalf("Compute(0).Value = %.1f; want 0.0", got.Value)
	}
	if got.Status != "failed" {
		t.Fatalf("Compute(0).Status = %q; want failed", got.Status)
	}
	if len(got.Dimensions) == 0 {
		t.Fatal("Compute(0).Dimensions is empty")
	}
	for _, dimension := range got.Dimensions {
		if dimension.Passed {
			t.Fatalf("dimension %q passed in an empty working directory", dimension.Name)
		}
	}
}

func TestComputeWithCompleteInputPasses(t *testing.T) {
	root := t.TempDir()
	writeReleaseQualityFixture(t, root)
	chdir(t, root)

	got := Compute(9.5)

	if got.Threshold != 9.5 {
		t.Fatalf("Compute(9.5).Threshold = %.1f; want 9.5", got.Threshold)
	}
	if got.Value != 10 {
		t.Fatalf("Compute(9.5).Value = %.1f; want 10.0", got.Value)
	}
	if got.Status != "passed" {
		t.Fatalf("Compute(9.5).Status = %q; want passed", got.Status)
	}
	for _, dimension := range got.Dimensions {
		if !dimension.Passed {
			t.Fatalf("dimension %q failed with detail %q", dimension.Name, dimension.Detail)
		}
	}
}

func TestVerifyPassAndFailurePaths(t *testing.T) {
	tests := []struct {
		name    string
		report  Report
		minimum float64
		wantErr string
	}{
		{
			name: "passes at default threshold",
			report: Report{
				Value:     9.8,
				Threshold: 9.8,
				Status:    "passed",
			},
		},
		{
			name: "fails below explicit minimum",
			report: Report{
				Value:     9.7,
				Threshold: 9.5,
				Status:    "passed",
			},
			minimum: 9.8,
			wantErr: "below minimum",
		},
		{
			name: "fails non-passed status below threshold",
			report: Report{
				Value:     9.7,
				Threshold: 9.8,
				Status:    "failed",
			},
			minimum: 9.0,
			wantErr: "status is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Verify(tt.report, tt.minimum)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Verify() returned unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Verify() error = nil; want containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Verify() error = %q; want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestMarshalStableJSON(t *testing.T) {
	report := Report{
		Value:     9.8,
		Threshold: 9.5,
		Status:    "passed",
		Dimensions: []Dimension{
			{
				Name:   "scorecard_doc",
				Weight: 1,
				Passed: true,
				Detail: "scorecard rubric is documented",
			},
		},
	}

	got, err := Marshal(report)
	if err != nil {
		t.Fatalf("Marshal() returned unexpected error: %v", err)
	}

	const want = `{
  "value": 9.8,
  "threshold": 9.5,
  "status": "passed",
  "dimensions": [
    {
      "name": "scorecard_doc",
      "weight": 1,
      "passed": true,
      "detail": "scorecard rubric is documented"
    }
  ]
}`
	if string(got) != want {
		t.Fatalf("Marshal() = %s; want %s", got, want)
	}
}

func TestScoreGateDimensionAcceptsStricterScoreMinimum(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Makefile")
	content := `score-check:
	score --min 9.8
release-final-check: score-check
	score --min 9.8
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write Makefile fixture: %v", err)
	}

	dimension := scoreGateDimension(path)
	if !dimension.Passed {
		t.Fatalf("dimension failed with detail %q", dimension.Detail)
	}
}

func TestScoreGateDimensionRejectsWeakScoreMinimum(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Makefile")
	content := `score-check:
	score --min 9.4
release-final-check: score-check
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write Makefile fixture: %v", err)
	}

	dimension := scoreGateDimension(path)
	if dimension.Passed {
		t.Fatal("expected weak score minimum to fail")
	}
	if !strings.Contains(dimension.Detail, "score --min >= 9.8") {
		t.Fatalf("dimension detail = %q; want threshold explanation", dimension.Detail)
	}
}

func TestTextDimensionReportsMissingNeedles(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "doc.txt")
	if err := os.WriteFile(path, []byte("present"), 0o644); err != nil {
		t.Fatalf("write doc fixture: %v", err)
	}

	dimension := textDimension("doc", 1, path, []string{"present", "missing"}, "doc has required text")
	if dimension.Passed {
		t.Fatal("expected text dimension to fail")
	}
	if !strings.Contains(dimension.Detail, "missing missing") {
		t.Fatalf("dimension detail = %q; want missing needle", dimension.Detail)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func writeReleaseQualityFixture(t *testing.T, root string) {
	t.Helper()

	files := map[string]string{
		"docs/scorecard.md": "scorecard rubric",
		"release/manifest/template.json": `{
  "score": 10,
  "workflow_run_id": "123",
  "artifact_url": "https://example.test/artifact"
}`,
		"cmd/goalcli/main.go": "score --min",
		"Makefile": `score-check:
	score --min 9.8
release-final-check: score-check
`,
		"scripts/check_release_evidence.sh": "RELEASE_EVIDENCE_MIN_SCORE --min-score",
		"scripts/check_secrets.sh":          "github_pat_ ghp_[A-Za-z0-9_]{36,} PRIVATE KEY-----",
		"docs/release.md":                   "go run ./cmd/goalcli score --min 9.8 workflow_run_id artifact_url",
		"docs/supply-chain.md":              "score workflow_run_id artifact_url",
		".agent/retrospective-template.md":  "Score Gate Patch",
		".agent/release-template.md":        "go run ./cmd/goalcli score --min 9.8 CI artifact score",
	}

	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create fixture directory for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
}
