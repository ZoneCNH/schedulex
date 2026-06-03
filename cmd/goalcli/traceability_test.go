package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTraceFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", full, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

const traceabilityValidMatrix = `# Traceability

| REQ | 需求摘要 | 主要产物 | 验证/Evidence | 收敛 owner |
| --- | --- | --- | --- | --- |
| REQ-001 | 示例需求 A | docs/standard/foo.md; .agent/state.yaml | docs/release.md; cmd/goalcli/main.go | agent-runtime |
| REQ-002 | 示例需求 B 含说明文字 | docs/spec.md; object-model; harness yaml | docs/release.md | agent-runtime |
`

func TestTraceabilityCheckPasses(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeTraceFile(t, dir, ".agent/traceability-matrix.md", traceabilityValidMatrix)
	writeTraceFile(t, dir, "docs/standard/foo.md", "x")
	writeTraceFile(t, dir, ".agent/state.yaml", "x")
	writeTraceFile(t, dir, "docs/release.md", "x")
	writeTraceFile(t, dir, "cmd/goalcli/main.go", "x")
	writeTraceFile(t, dir, "docs/spec.md", "x")

	var stdout, stderr bytes.Buffer
	code := runTraceabilityCheck(nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit=%d want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"status": "passed"`) {
		t.Fatalf("expected passed JSON, got: %s", stdout.String())
	}
}

func TestTraceabilityCheckGapExits9(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeTraceFile(t, dir, ".agent/traceability-matrix.md", traceabilityValidMatrix)
	// 故意只创建部分产物, 其余缺失
	writeTraceFile(t, dir, "docs/standard/foo.md", "x")

	var stdout, stderr bytes.Buffer
	code := runTraceabilityCheck(nil, &stdout, &stderr)
	if code != 9 {
		t.Fatalf("exit=%d want 9 (gap); stdout=%s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), `"status": "failed"`) {
		t.Fatalf("expected failed status JSON, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "missing artifact") {
		t.Fatalf("expected gap detail, got: %s", stdout.String())
	}
}

func TestTraceabilityCheckMissingMatrixExits2(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	var stdout, stderr bytes.Buffer
	code := runTraceabilityCheck(nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit=%d want 2 (parse error); stderr=%s", code, stderr.String())
	}
}

func TestSplitCellTokensPreservesLeadingDot(t *testing.T) {
	tokens := splitCellTokens(".agent/foo.yaml; docs/goal.md v2.9.3 Complete; .github/workflows/ci.yml")
	want := []string{".agent/foo.yaml", "docs/goal.md", ".github/workflows/ci.yml"}
	if len(tokens) != len(want) {
		t.Fatalf("len=%d want %d; tokens=%#v", len(tokens), len(want), tokens)
	}
	for i := range tokens {
		if tokens[i] != want[i] {
			t.Errorf("tokens[%d]=%q want %q", i, tokens[i], want[i])
		}
	}
}

func TestLooksLikePath(t *testing.T) {
	cases := map[string]bool{
		".agent/foo.yaml": true,
		"docs/spec.md":    true,
		"cmd/goalcli":     true,
		"Makefile":        true,
		"README.md":       true,
		"object-model":    false,
		"render worker":   false,
		"harness yaml":    false,
		"foo/bar.unknown": false,
	}
	for in, want := range cases {
		if got := looksLikePath(in); got != want {
			t.Errorf("looksLikePath(%q)=%v want %v", in, got, want)
		}
	}
}
