package debtcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPassesWithPolicyFilesAndCleanTree(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeDownstreamFiles(t, root)
	writeFile(t, root, "safe.go", "package fixture\n")

	report, err := Run(Options{Root: root, Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "passed" {
		t.Fatalf("status = %q, want passed: %+v", report.Status, report.Summary)
	}
	if code := ExitCode(report); code != 0 {
		t.Fatalf("ExitCode = %d, want 0", code)
	}
	if problems := ValidateEvidence(EvidenceFromReport(report), DefaultMinScore); len(problems) != 0 {
		t.Fatalf("ValidateEvidence problems = %v, want none", problems)
	}
	if report.Digests.Report == "" || report.Digests.Rules == "missing" {
		t.Fatalf("digests = %+v, want populated policy and report digests", report.Digests)
	}
}

func TestRunPassesDownstreamSectionWithGovernanceFiles(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeDownstreamFiles(t, root)

	report, err := Run(Options{Root: root, Section: "downstream", Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "passed" {
		t.Fatalf("status = %q, want passed: %+v", report.Status, report.Summary)
	}
	if len(report.Sections) != 1 || report.Sections[0].Name != "downstream" {
		t.Fatalf("sections = %+v, want only downstream", report.Sections)
	}
	if code := ExitCode(report); code != 0 {
		t.Fatalf("ExitCode = %d, want 0", code)
	}
}

func TestRunFailsDownstreamSectionWithoutRegistryCoverage(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeDownstreamFiles(t, root)
	writeFile(t, root, ".agent/downstream-registry.yaml", `schema_version: "2.9.3"
downstreams:
  - repo: kernel/configx
    mode: patch-only
    status: unavailable_in_worker_workspace_gap_explicit
`)

	report, err := Run(Options{Root: root, Section: "downstream", Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "failed" {
		t.Fatalf("status = %q, want failed", report.Status)
	}
	if report.Summary.P0 != 2 {
		t.Fatalf("P0 = %d, want 2", report.Summary.P0)
	}
	if !strings.Contains(ToMarkdown(report), "debt.downstream.registry-missing-repo") {
		t.Fatalf("markdown missing downstream registry finding: %s", ToMarkdown(report))
	}
}

func TestRunFailsDownstreamSectionOnFalseAdoptionClaim(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeDownstreamFiles(t, root)
	writeFile(t, root, ".agent/downstream-adoption-status.yaml", `schema_version: "2.9.3"
current_registry:
  adoption_status: adopted
  proof_based_adoption: true
standard_target_libraries:
  - name: kernel
  - name: configx
  - name: observex
  - name: testkitx
  - name: postgresx
  - name: redisx
  - name: kafkax
  - name: taosx
  - name: ossx
  - name: clickhousex
`)

	report, err := Run(Options{Root: root, Section: "downstream", Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "failed" {
		t.Fatalf("status = %q, want failed", report.Status)
	}
	markdown := ToMarkdown(report)
	if !strings.Contains(markdown, "debt.downstream.false-adoption-claim") {
		t.Fatalf("markdown missing false adoption finding: %s", markdown)
	}
	if !strings.Contains(markdown, "debt.downstream.false-proof-claim") {
		t.Fatalf("markdown missing false proof finding: %s", markdown)
	}
}

func TestRunFailsDownstreamSectionWithoutIntegrationDebtEvidenceGate(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeDownstreamFiles(t, root)
	writeFile(t, root, "scripts/run_integration.sh", `#!/usr/bin/env bash
TARGETS=(
  "kernel|github.com/ZoneCNH/kernel|kernel"
  "configx|github.com/ZoneCNH/configx|configx"
  "redisx|github.com/ZoneCNH/redisx|redisx"
)
`)

	report, err := Run(Options{Root: root, Section: "downstream", Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "failed" {
		t.Fatalf("status = %q, want failed", report.Status)
	}
	if !strings.Contains(ToMarkdown(report), "debt.downstream.integration-missing-contract") {
		t.Fatalf("markdown missing downstream integration finding: %s", ToMarkdown(report))
	}
}

func TestRunFailsDownstreamSectionWithoutRenderedDebtEvidenceExclusions(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeDownstreamFiles(t, root)
	writeFile(t, root, "scripts/render_template.sh", `#!/usr/bin/env bash
rsync "$@"
`)

	report, err := Run(Options{Root: root, Section: "downstream", Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "failed" {
		t.Fatalf("status = %q, want failed", report.Status)
	}
	if !strings.Contains(ToMarkdown(report), "debt.downstream.render-template-missing-exclusion") {
		t.Fatalf("markdown missing downstream render-template finding: %s", ToMarkdown(report))
	}
}

func TestRunFailsOnLegacyProductionImport(t *testing.T) {
	root := t.TempDir()
	writePolicyFiles(t, root)
	writeFile(t, root, "bad.go", "package fixture\n\nimport _ \"github.com/ZoneCNH/x.go\"\n")

	report, err := Run(Options{Root: root, Section: "architecture", Mode: "enforce", MinScore: DefaultMinScore})
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "failed" {
		t.Fatalf("status = %q, want failed", report.Status)
	}
	if report.Summary.P0 != 1 {
		t.Fatalf("P0 = %d, want 1", report.Summary.P0)
	}
	if code := ExitCode(report); code != 1 {
		t.Fatalf("ExitCode = %d, want 1", code)
	}
	if !strings.Contains(ToMarkdown(report), "legacy ZoneCNH x module") {
		t.Fatalf("markdown missing legacy import finding: %s", ToMarkdown(report))
	}
}

func writePolicyFiles(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		DefaultRulesPath:    "schema_version: debt-rules/v1\nprofile: test\n",
		DefaultRegistryPath: "schema_version: debt-rule-registry/v1\nrules: []\n",
		DefaultExceptions:   "schema_version: debt-exceptions/v1\nexceptions: []\n",
		DefaultPurpose:      "schema_version: debt-dependency-purpose/v1\npurposes: []\n",
	}
	for path, content := range files {
		writeFile(t, root, path, content)
	}
}

func writeDownstreamFiles(t *testing.T, root string) {
	t.Helper()
	files := map[string]string{
		".agent/downstream-registry.yaml": `schema_version: "2.9.3"
downstreams:
  - repo: kernel/configx
    mode: patch-only
    status: unavailable_in_worker_workspace_gap_explicit
  - repo: kernel/redisx
    mode: patch-only
    status: unavailable_in_worker_workspace_gap_explicit
  - repo: corekit
    mode: patch-only
    status: unavailable_in_worker_workspace_gap_explicit
`,
		".agent/downstream-baseline-scan.yaml": `schema_version: "2.9.3"
repo: kernel/configx
mode: patch-only
status: gap_explicit_when_repo_missing
`,
		".agent/downstream-adoption-modes.yaml": `schema_version: "2.9.3"
modes: [patch-only, dry-run]
forbidden: [direct_downstream_write_without_repo]
`,
		".agent/downstream-adoption-status.yaml": `schema_version: "2.9.3"
current_registry:
  adoption_status: not_adopted
  proof_based_adoption: false
first_pr_mva_assertions:
  no_proof_based_adoption: true
standard_target_libraries:
  - name: kernel
  - name: configx
  - name: observex
  - name: testkitx
  - name: postgresx
  - name: redisx
  - name: kafkax
  - name: taosx
  - name: ossx
  - name: clickhousex
`,
		"docs/downstream-matrix.md": `# Downstream Matrix

| Library | Adoption |
| --- | --- |
| ` + "`kernel`" + ` | not_adopted |
| ` + "`configx`" + ` | not_adopted |
| ` + "`observex`" + ` | not_adopted |
| ` + "`testkitx`" + ` | not_adopted |
| ` + "`postgresx`" + ` | not_adopted |
| ` + "`redisx`" + ` | not_adopted |
| ` + "`kafkax`" + ` | not_adopted |
| ` + "`taosx`" + ` | not_adopted |
| ` + "`ossx`" + ` | not_adopted |
| ` + "`clickhousex`" + ` | not_adopted |
`,
		"docs/standard/downstream-compatibility.md": `# Downstream Compatibility

默认 downstream 为 ` + "`kernel`" + ` 与 ` + "`corekit`" + `。
发布验证命令必须包含 GOWORK=off make integration。
`,
		"scripts/run_integration.sh": `#!/usr/bin/env bash
TARGETS=(
  "kernel|github.com/ZoneCNH/kernel|kernel"
  "configx|github.com/ZoneCNH/configx|configx"
  "redisx|github.com/ZoneCNH/redisx|redisx"
)
GOWORK=off make debt
GOWORK=off make debt-evidence
GOWORK=off make debt-evidence-checksum-check
`,
		"scripts/render_template.sh": `#!/usr/bin/env bash
rsync \
  --exclude release/debt/latest.json \
  --exclude release/debt/latest.md \
  --exclude release/debt/latest.json.sha256
`,
	}
	for path, content := range files {
		writeFile(t, root, path, content)
	}
}

func writeFile(t *testing.T, root, path, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
