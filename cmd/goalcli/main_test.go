package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ZoneCNH/xlib-standard/internal/goalruntime"
	"github.com/ZoneCNH/xlib-standard/internal/releasequality"
	"github.com/ZoneCNH/xlib-standard/pkg/templatex"
)

func TestMainDispatchesUsageHelpAndUnknownCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "no command",
			wantCode:   2,
			wantStderr: "usage: goalcli <command>",
		},
		{
			name:       "help",
			args:       []string{"help"},
			wantCode:   0,
			wantStdout: "commands:",
		},
		{
			name:       "short help",
			args:       []string{"-h"},
			wantCode:   0,
			wantStdout: "commands:",
		},
		{
			name:       "long help",
			args:       []string{"--help"},
			wantCode:   0,
			wantStdout: "commands:",
		},
		{
			name:       "unknown",
			args:       []string{"missing"},
			wantCode:   2,
			wantStderr: `unknown command "missing"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			got := run(tt.args, strings.NewReader(""), &stdout, &stderr)

			if got != tt.wantCode {
				t.Fatalf("run(%v) = %d; want %d", tt.args, got, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Fatalf("stdout = %q; want containing %q", stdout.String(), tt.wantStdout)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q; want containing %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestUsageDocumentsCommandRegistryRequiredCommands(t *testing.T) {
	for _, command := range commandRegistryRequiredCommands() {
		needle := "\n  " + command
		if !strings.Contains(usage, needle) {
			t.Errorf("usage missing command %q", command)
		}
	}
}

func TestUsageDocumentsDebtEvidenceCommandsAndHelpers(t *testing.T) {
	for _, needle := range []string{
		"\n  debt-evidence",
		"\n  debt-evidence-checksum-check",
		"\n  debt-evidence-hash",
		"\n  debt register-update",
		"\n  debt trend",
		"\n  debt patch-suggest",
		"\n  debt lifecycle-check",
	} {
		if !strings.Contains(usage, needle) {
			t.Errorf("usage missing %q", needle)
		}
	}
}

func TestMainUsesRunExitCode(t *testing.T) {
	originalArgs := os.Args
	originalExit := exit
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	t.Cleanup(func() {
		os.Args = originalArgs
		exit = originalExit
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open os.DevNull: %v", err)
	}
	t.Cleanup(func() { _ = devNull.Close() })
	os.Stdout = devNull
	os.Stderr = devNull
	os.Args = []string{"goalcli", "help"}

	var got int
	exit = func(code int) {
		got = code
	}

	main()

	if got != 0 {
		t.Fatalf("main exit code = %d; want 0", got)
	}
}

func TestEmitReportStatusExitCodes(t *testing.T) {
	tests := []struct {
		status   string
		wantCode int
	}{
		{status: "passed", wantCode: 0},
		{status: "failed", wantCode: 1},
		{status: "planned", wantCode: 1},
		{status: "gap", wantCode: 1},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			var stdout bytes.Buffer
			got := emitReport(&stdout, "test-command", tt.status, nil, nil)
			if got != tt.wantCode {
				t.Fatalf("emitReport status %q exit = %d; want %d", tt.status, got, tt.wantCode)
			}
			if !strings.Contains(stdout.String(), `"status": "`+tt.status+`"`) {
				t.Fatalf("stdout = %q; want status %q", stdout.String(), tt.status)
			}
		})
	}
}

func TestRunDispatchesExternalCommands(t *testing.T) {
	root := t.TempDir()
	writeGateScript(t, root, "scripts/check_boundary.sh")
	writeGateScript(t, root, "scripts/check_contracts.sh")
	writeGateScript(t, root, "scripts/check_dependency_diff.sh")
	writeGateScript(t, root, "scripts/check_docs.sh")
	writeGateScript(t, root, "scripts/run_integration.sh")
	writeGateScript(t, root, "scripts/check_release_evidence.sh")
	writeGateScript(t, root, "scripts/hash_release_evidence.sh")
	writeGateScript(t, root, "scripts/check_secrets.sh")
	writeGateScript(t, root, "scripts/check_standard_impact.sh")
	writeGateScript(t, root, "scripts/check_rendered_template.sh")
	writePathTool(t, root, "govulncheck")
	writePathTool(t, root, "go")
	writePathTool(t, root, "make")
	writePathTool(t, root, "python3")
	chdir(t, root)
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))

	tests := []struct {
		name       string
		args       []string
		wantStdout string
	}{
		{name: "boundary", args: []string{"boundary"}, wantStdout: "check_boundary.sh"},
		{name: "contracts", args: []string{"contracts"}, wantStdout: "check_contracts.sh"},
		{name: "dependency-check", args: []string{"dependency-check"}, wantStdout: "check_dependency_diff.sh"},
		{name: "docs-check", args: []string{"docs-check"}, wantStdout: "check_docs.sh"},
		{name: "evidence", args: []string{"evidence"}, wantStdout: "go run ./internal/tools/releasemanifest --out release/manifest/latest.json"},
		{name: "manifest", args: []string{"manifest"}, wantStdout: "go run ./internal/tools/releasemanifest --out release/manifest/latest.json"},
		{name: "integration", args: []string{"integration"}, wantStdout: "run_integration.sh"},
		{name: "manifest", args: []string{"manifest"}, wantStdout: "go run ./internal/tools/releasemanifest --out release/manifest/latest.json"},
		{name: "debt-evidence-checksum-check", args: []string{"debt-evidence-checksum-check"}, wantStdout: "hash_release_evidence.sh --check release/debt/latest.json release/debt/latest.json.sha256"},
		{name: "debt-evidence-hash", args: []string{"debt-evidence-hash"}, wantStdout: "hash_release_evidence.sh release/debt/latest.json release/debt/latest.json.sha256"},
		{name: "release-evidence-check", args: []string{"release-evidence-check"}, wantStdout: "check_release_evidence.sh"},
		{name: "release-evidence-checksum-check", args: []string{"release-evidence-checksum-check"}, wantStdout: "hash_release_evidence.sh --check"},
		{name: "release-evidence-hash", args: []string{"release-evidence-hash"}, wantStdout: "hash_release_evidence.sh"},
		{name: "release-final-check", args: []string{"release-final-check"}, wantStdout: "make release-final-check"},
		{name: "render-check", args: []string{"render-check", "rendered"}, wantStdout: "check_rendered_template.sh rendered"},
		{name: "rules-verify", args: []string{"rules-verify"}, wantStdout: "python3 scripts/verify_rules.py"},
		{name: "secrets", args: []string{"secrets"}, wantStdout: "check_secrets.sh"},
		{name: "security", args: []string{"security"}, wantStdout: "check_secrets.sh"},
		{name: "standard-impact-check", args: []string{"standard-impact-check"}, wantStdout: "check_standard_impact.sh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			got := run(tt.args, strings.NewReader("input"), &stdout, &stderr)

			if got != 0 {
				t.Fatalf("run(%v) = %d, stderr %q; want 0", tt.args, got, stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Fatalf("stdout = %q; want containing %q", stdout.String(), tt.wantStdout)
			}
		})
	}
}

func TestDebtHelpersWriteArtifacts(t *testing.T) {
	chdir(t, repoRoot(t))

	for _, helper := range []string{"register-update", "trend", "patch-suggest", "lifecycle-check"} {
		t.Run(helper, func(t *testing.T) {
			outPath := filepath.Join(t.TempDir(), helper+".json")
			var stdout, stderr bytes.Buffer

			got := run([]string{"debt", helper, "--output", outPath}, strings.NewReader(""), &stdout, &stderr)

			if got != 0 {
				t.Fatalf("run(debt %s) = %d, stderr %q, stdout %q; want 0", helper, got, stderr.String(), stdout.String())
			}
			if !strings.Contains(stdout.String(), filepath.ToSlash(outPath)) {
				t.Fatalf("stdout = %q; want output path %q", stdout.String(), filepath.ToSlash(outPath))
			}

			data, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read helper artifact: %v", err)
			}
			var artifact struct {
				SchemaVersion string   `json:"schema_version"`
				Command       string   `json:"command"`
				Status        string   `json:"status"`
				Details       []string `json:"details"`
			}
			if err := json.Unmarshal(data, &artifact); err != nil {
				t.Fatalf("helper artifact is not JSON: %v; data %q", err, string(data))
			}
			if artifact.SchemaVersion != "debt-helper/v1" || artifact.Command != helper || artifact.Status != "passed" || len(artifact.Details) == 0 {
				t.Fatalf("artifact = %#v; want helper command, passed status, and details", artifact)
			}
		})
	}
}

func TestDownstreamDebtAliasUsesSupportedDebtSurface(t *testing.T) {
	chdir(t, repoRoot(t))
	var stdout, stderr bytes.Buffer

	got := run([]string{"downstream-debt", "--mode", "enforce"}, strings.NewReader(""), &stdout, &stderr)

	if got != 0 {
		t.Fatalf("downstream-debt exit = %d, stderr %q, stdout %q; want 0", got, stderr.String(), stdout.String())
	}
	if strings.Contains(stderr.String(), "unsupported debt section") {
		t.Fatalf("stderr = %q; downstream-debt must not use unsupported debt section", stderr.String())
	}
	if !strings.Contains(stdout.String(), `"name": "downstream"`) {
		t.Fatalf("stdout = %q; want downstream debt report", stdout.String())
	}
	if strings.Contains(stdout.String(), `"name": "architecture"`) {
		t.Fatalf("stdout = %q; downstream-debt must only report downstream section", stdout.String())
	}
}

func TestGoalGovernanceCommandSurface(t *testing.T) {
	chdir(t, repoRoot(t))
	required := []struct {
		command    string
		args       []string
		wantCode   int
		wantStatus string
	}{
		{command: "version"},
		{command: "doctor"},
		{command: "minimal-kernel"},
		{command: "main-guard", args: []string{"--context", "local_readonly"}},
		{command: "worktree-guard", args: []string{"--context", "local_readonly"}},
		{command: "evidence-check"},
		{command: "done-assertion"},
		{command: "cli-contract"},
		{command: "issue-registry"},
		{command: "command-registry"},
		{command: "makefile-baseline"},
		{command: "self-improving-check"},
		{command: "context-profile"},
		{command: "context-profile-check"},
		{command: "context-schema-check"},
		{command: "context-lite"},
		{command: "context-standard"},
		{command: "context-full"},
		{command: "context-release"},
		{command: "context-fast-check"},
		{command: "context-standard-check"},
		{command: "context-full-check"},
		{command: "agent-team-contract"},
		{command: "scope-lock"},
		{command: "pr-template"},
		{command: "acceptance-matrix"},
		{command: "runtime-health"},
		{command: "goal-runtime"},
		{command: "goal-acceptance", args: []string{"--json"}},
		{command: "goal-delivery", args: []string{"--json"}},
		{command: "goal-handover", args: []string{"--json"}},
		{command: "goal-downstream-adoption", args: []string{"--json"}},
		{command: "goal-certify", args: []string{"--json"}},
		{command: "goal-runtime-final", args: []string{"--json"}},
		{command: "github-settings"},
		{command: "github-governance"},
		{command: "governance-fixture-test"},
		{command: "policy-schema"},
		{command: "toolchain"},
		{command: "evidence-artifacts"},
		{command: "supply-chain"},
		{command: "autoresearch"},
		{command: "changelog"},
		{command: "naming"},
		{command: "upgrade-standard", args: []string{"--repo", "kernel/configx", "--mode", "patch-only"}, wantCode: 1, wantStatus: "gap"},
		{command: "conformance-profile"},
		{command: "downstream-registry"},
		{command: "self-healing-skeleton"},
		{command: "install-runtime"},
		{command: "upgrade-runtime"},
		{command: "release-ready"},
		{command: "evidence-replay"},
		{command: "attest-conformance", args: []string{"--profile", "standard-source"}},
		{command: "pack-standard"},
		{command: "pack-gate"},
		{command: "pack-evidence"},
		{command: "downstream-baseline", args: []string{"--repo", "kernel/configx", "--mode", "patch-only"}, wantCode: 1, wantStatus: "gap"},
		{command: "downstream-adoption", args: []string{"--repo", "kernel/configx", "--mode", "patch-only"}, wantCode: 1, wantStatus: "gap"},
		{command: "runtime-file-ownership"},
		{command: "execution-context"},
	}

	for _, tt := range required {
		t.Run(tt.command, func(t *testing.T) {
			command := tt.command
			if !strings.Contains(usage, command) {
				t.Fatalf("usage does not include %q", command)
			}

			var stdout, stderr bytes.Buffer
			args := append([]string{command}, tt.args...)

			got := run(args, strings.NewReader(""), &stdout, &stderr)
			if got != tt.wantCode {
				t.Fatalf("run(%v) = %d, stderr %q, stdout %q; want %d", args, got, stderr.String(), stdout.String(), tt.wantCode)
			}

			var report gateReport
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
			}
			if report.Command != command {
				t.Fatalf("report command = %q; want %q", report.Command, command)
			}
			wantStatus := tt.wantStatus
			if wantStatus == "" {
				wantStatus = "passed"
			}
			if report.Status != wantStatus {
				t.Fatalf("report status = %q; want %s; report %#v", report.Status, wantStatus, report)
			}
		})
	}
}

func TestGoalcliMVACommandSurfaceRequiresG12ThroughG16Equivalents(t *testing.T) {
	chdir(t, repoRoot(t))
	const fixtureGoalID = "GOAL-20260603-XLIB-GOALCLI-001"
	tests := []struct {
		command   string
		wantGates int
	}{
		{command: "goal-acceptance", wantGates: 1},
		{command: "goal-delivery", wantGates: 1},
		{command: "goal-handover", wantGates: 1},
		{command: "goal-downstream-adoption", wantGates: 1},
		{command: "goal-certify", wantGates: 1},
		{command: "goal-runtime-final", wantGates: 5},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			root := t.TempDir()
			writeGoalcliAuthorityFixture(t, root)
			if tt.command == "goal-runtime-final" {
				writeGoalcliPrerequisiteLedgerFixture(t, root, fixtureGoalID)
			}
			chdir(t, root)

			var stdout, stderr bytes.Buffer
			got := run([]string{
				tt.command,
				"--goal-id", fixtureGoalID,
				"--mode", "FULL",
				"--json",
			}, strings.NewReader(""), &stdout, &stderr)
			if got != 0 {
				t.Fatalf("run(%s) = %d, stderr %q, stdout %q; want 0", tt.command, got, stderr.String(), stdout.String())
			}

			var report struct {
				Command          string   `json:"command"`
				Status           string   `json:"status"`
				GoalID           string   `json:"goal_id"`
				Executor         string   `json:"executor"`
				ControlPlane     string   `json:"control_plane"`
				Blocking         bool     `json:"blocking"`
				MVAStatus        string   `json:"mva_status"`
				LedgerPath       string   `json:"ledger_path"`
				EvidencePackPath string   `json:"evidence_pack_path"`
				Evidence         []string `json:"evidence"`
				Gates            []struct {
					ID       string `json:"id"`
					Command  string `json:"command"`
					Status   string `json:"status"`
					Blocking bool   `json:"blocking"`
				} `json:"gates"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("stdout is not goalcli report JSON: %v; stdout %q", err, stdout.String())
			}
			if report.Command != tt.command || report.Status != "passed" {
				t.Fatalf("report = %#v; want command %q with passed status", report, tt.command)
			}
			if report.GoalID != fixtureGoalID {
				t.Fatalf("goal_id = %q; want fixture goal id", report.GoalID)
			}
			if report.Executor != "goalcli" || report.ControlPlane != "Harness Runtime" {
				t.Fatalf("executor/control_plane = %q/%q; want goalcli/Harness Runtime", report.Executor, report.ControlPlane)
			}
			if !report.Blocking {
				t.Fatalf("blocking = false; want %s to be goalcli MVA-blocking", tt.command)
			}
			if report.MVAStatus != "complete" {
				t.Fatalf("mva_status = %q; want complete for %s", report.MVAStatus, tt.command)
			}
			if !slicesContain(report.Evidence, "source_evidence_ledger="+goalruntime.SourceLedgerPath) {
				t.Fatalf("evidence = %#v; want source evidence ledger path", report.Evidence)
			}
			if !slicesContain(report.Evidence, "generated_evidence_pack="+goalruntime.EvidenceLedgerPath) {
				t.Fatalf("evidence = %#v; want generated evidence pack path", report.Evidence)
			}
			if report.LedgerPath != ".agent/evidence/ledger.jsonl" || !strings.HasPrefix(report.EvidencePackPath, "release/evidence/goalcli") {
				t.Fatalf("ledger/evidence paths = %q/%q; want source ledger and generated pack split", report.LedgerPath, report.EvidencePackPath)
			}
			if len(report.Gates) != tt.wantGates {
				t.Fatalf("gates = %#v; want %d gates", report.Gates, tt.wantGates)
			}
			for _, gate := range report.Gates {
				if gate.Status != "passed" || !gate.Blocking {
					t.Fatalf("gate = %#v; want passed blocking gate", gate)
				}
			}
		})
	}
}

func TestGoalcliRuntimeFinalWritesEvidenceWhenRequested(t *testing.T) {
	root := t.TempDir()
	writeGoalcliAuthorityFixture(t, root)
	writeGoalcliPrerequisiteLedgerFixture(t, root, "GOAL-20260603-XLIB-GOALCLI-001")
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	got := run([]string{
		"goal-runtime-final",
		"--goal-id", "GOAL-20260603-XLIB-GOALCLI-001",
		"--mode", "FULL",
		"--json",
		"--write-evidence",
	}, strings.NewReader(""), &stdout, &stderr)
	if got != 0 {
		t.Fatalf("run(goal-runtime-final) = %d, stderr %q, stdout %q; want 0", got, stderr.String(), stdout.String())
	}

	var report struct {
		EvidencePackPath string `json:"evidence_pack_path"`
		MVAStatus        string `json:"mva_status"`
		Blocking         bool   `json:"blocking"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not goalcli report JSON: %v; stdout %q", err, stdout.String())
	}
	if report.MVAStatus != "complete" || !report.Blocking {
		t.Fatalf("report = %#v; want complete blocking report", report)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(report.EvidencePackPath))); err != nil {
		t.Fatalf("evidence pack was not written: %v", err)
	}
	ledger, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(goalruntime.SourceLedgerPath)))
	if err != nil {
		t.Fatalf("source evidence ledger was not written: %v", err)
	}
	if !strings.Contains(string(ledger), report.EvidencePackPath) || !strings.Contains(string(ledger), `"mva_status":"complete"`) {
		t.Fatalf("ledger = %s; want complete entry for evidence pack %s", ledger, report.EvidencePackPath)
	}
}

func TestGoalcliRuntimeTargetsRouteThroughGoalcli(t *testing.T) {
	chdir(t, repoRoot(t))
	makefile := readText(t, "Makefile")
	if !strings.Contains(makefile, "GOALCLI ?= go run ./cmd/goalcli") {
		t.Fatalf("Makefile must define GOALCLI as the cmd/goalcli execution surface")
	}
	for _, command := range []string{
		"goal-acceptance",
		"goal-delivery",
		"goal-handover",
		"goal-downstream-adoption",
		"goal-certify",
		"goal-runtime-final",
	} {
		if !strings.Contains(makefile, ".PHONY: "+command) {
			t.Fatalf("Makefile missing .PHONY for %s", command)
		}
		if !strings.Contains(makefile, command+": require-gowork-off") {
			t.Fatalf("Makefile target %s must require GOWORK=off", command)
		}
		if !strings.Contains(makefile, "$(GOALCLI) $@ --goal-id") {
			t.Fatalf("Makefile target %s must route through goalcli", command)
		}
	}
	legacyRuntimeNames := []string{
		"$(" + "XLIB" + "GATE)",
		"$(" + "GOAL" + "KIT)",
		"go run ./cmd/" + "xlib" + "gate",
		"cmd/" + "xlib" + "gate",
		"go run ./cmd/" + "goal" + "kit",
		"cmd/" + "goal" + "kit",
		"release/evidence/" + "goal" + "kit",
	}
	for _, legacy := range legacyRuntimeNames {
		if strings.Contains(makefile, legacy) {
			t.Fatalf("Makefile must not route goalcli v0.1.0 through legacy authority %q", legacy)
		}
	}
	if !strings.Contains(makefile, "$(GOALCLI) $@ --goal-id \"$(GOAL_ID)\" --mode \"$(GOAL_RUNTIME_MODE)\" --json --write-evidence") {
		t.Fatalf("goal-runtime-final must explicitly request evidence writing")
	}
}

func TestGoalcliControlPlaneDocumentsEvidenceLedgerAndCompleteMVA(t *testing.T) {
	root := repoRoot(t)
	chdir(t, root)
	files := map[string][]string{
		".agent/harness.yaml": {
			"goalcli_v0_1_0",
			"goalcli_mva_gates:",
			"control_plane: Harness Runtime",
			"executor: goalcli",
			"source_evidence_ledger: .agent/evidence/ledger.jsonl",
			"generated_evidence_pack: release/evidence/goalcli/",
			"blocking: true",
		},
		".agent/registry/runtime.yaml": {
			"mva_status: complete",
			"control_plane: Harness Runtime",
			"executor: goalcli",
			"source_evidence_ledger: .agent/evidence/ledger.jsonl",
			"generated_evidence_pack: release/evidence/goalcli/",
		},
		".agent/registry/commands.yaml": {
			"goal-acceptance",
			"goal-downstream-adoption",
			"goal-runtime-final",
			"mva_status: complete",
			"blocking: true",
		},
		".agent/command-implementation-status.yaml": {
			"goalcli_v0_1_0_mva_blocking",
			"mva_status: complete",
			"evidence_strength: full_mva_evidence",
		},
		".agent/evidence/README.md": {
			".agent/evidence/ledger.jsonl",
			"release/evidence/goalcli/",
			"mva_status: complete",
		},
		"docs/standard/goalcli-runtime.md": {
			"不引入第二套并列执行面",
			"cmd/goalcli",
			"Harness Runtime",
			".agent/evidence/ledger.jsonl",
			"mva_status: complete",
		},
		"docs/plans/goalcli-v0.1.0-roadmap.md": {
			"PR-4",
			"G12-G16",
			"MVA-blocking",
			"completion rollup",
			"mva_status: complete",
		},
	}
	if isStandardSourceRepo(t, root) {
		files["docs/adr/ADR-20260603-001-goalcli-runtime.md"] = []string{
			"Accepted",
			"拒绝第二套并列执行面",
			"cmd/goalcli",
			"Harness Runtime",
		}
	}

	for path, needles := range files {
		text := readText(t, path)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				t.Fatalf("%s missing %q", path, needle)
			}
		}
	}
	if _, err := os.Stat(".agent/evidence/ledger.jsonl"); err != nil {
		t.Fatalf("source evidence ledger missing: %v", err)
	}
}

func TestMakefileBaselineDetectsMissingTargets(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	if err := os.WriteFile("Makefile", []byte("score-check:\n"), 0o644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	var stdout, stderr bytes.Buffer
	got := runMakefileBaseline(nil, &stdout, &stderr)
	if got == 0 {
		t.Fatal("runMakefileBaseline succeeded; want missing target failure")
	}
	if !strings.Contains(stderr.String(), "makefile-baseline found") {
		t.Fatalf("stderr = %q; want makefile-baseline gaps", stderr.String())
	}
	if !strings.Contains(stdout.String(), "main-guard") {
		t.Fatalf("stdout = %q; want main-guard missing", stdout.String())
	}
	if !strings.Contains(stdout.String(), "execution-context") {
		t.Fatalf("stdout = %q; want execution-context missing", stdout.String())
	}
}

func TestMakefileBaselineRequiresExecutionContext(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agent"), 0o755); err != nil {
		t.Fatalf("mkdir .agent: %v", err)
	}

	staleTargets := []string{
		"fmt", "vet", "lint", "test", "race", "boundary", "security", "contracts", "docs-check", "evidence", "score-check",
		"main-guard", "worktree-guard", "evidence-check", "cli-contract", "issue-registry", "command-registry", "makefile-baseline",
		"governance-check", "p1-governance-check", "p2-runtime-check", "release-check", "release-final-check",
	}
	var makefile strings.Builder
	makefile.WriteString(".PHONY:")
	for _, target := range staleTargets {
		makefile.WriteString(" " + target)
	}
	makefile.WriteString("\n")
	for _, target := range staleTargets {
		makefile.WriteString(target + ":\n")
	}
	if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte(makefile.String()), 0o644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	registry := "schema_version: \"2.9.3\"\nmodule: xlib-standard\ntargets:\n  - " + strings.Join(staleTargets, "\n  - ") + "\n"
	if err := os.WriteFile(filepath.Join(root, ".agent", "makefile-target-registry.yaml"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write makefile target registry: %v", err)
	}
	baseline := "schema_version: \"2.9.3\"\nmodule: xlib-standard\nbaseline_targets:\n"
	for _, target := range staleTargets {
		baseline += "  " + target + ": fixture\n"
	}
	if err := os.WriteFile(filepath.Join(root, ".agent", "makefile-baseline.yaml"), []byte(baseline), 0o644); err != nil {
		t.Fatalf("write makefile baseline: %v", err)
	}

	chdir(t, root)
	var stdout, stderr bytes.Buffer
	got := runMakefileBaseline(nil, &stdout, &stderr)
	if got == 0 {
		t.Fatal("runMakefileBaseline accepted a stale fixture without execution-context; want missing target failure")
	}
	for _, want := range []string{
		"Makefile missing .PHONY: execution-context",
		"Makefile missing execution-context:",
		".agent/makefile-target-registry.yaml missing execution-context",
		".agent/makefile-baseline.yaml missing execution-context",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q; want %q", stdout.String(), want)
		}
	}
}

func TestCommandRegistryRequiresCompleteGovernanceSurface(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agent"), 0o755); err != nil {
		t.Fatalf("mkdir .agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agent", "command-registry.yaml"), []byte("commands:\n  - name: version\n"), 0o644); err != nil {
		t.Fatalf("write command registry: %v", err)
	}
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	got := run([]string{"command-registry"}, strings.NewReader(""), &stdout, &stderr)
	if got != 1 {
		t.Fatalf("command-registry exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
	}
	if !strings.Contains(stderr.String(), "command-registry found") {
		t.Fatalf("stderr = %q; want command-registry gaps", stderr.String())
	}

	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if report.Command != "command-registry" || report.Status != "failed" {
		t.Fatalf("report = %#v; want failed command-registry report", report)
	}
	if !slicesContain(report.Gaps, ".agent/command-registry.yaml missing name: execution-context") {
		t.Fatalf("gaps = %#v; want missing execution-context registry entry", report.Gaps)
	}
}

func TestRunScore(t *testing.T) {
	resetReleaseQuality := func() {
		computeReleaseQuality = releasequality.Compute
		marshalReleaseQuality = releasequality.Marshal
		verifyReleaseQuality = releasequality.Verify
	}
	t.Cleanup(resetReleaseQuality)

	t.Run("success", func(t *testing.T) {
		resetReleaseQuality()
		var gotMinimum float64
		computeReleaseQuality = func(minimum float64) releasequality.Report {
			gotMinimum = minimum
			return releasequality.Report{Value: 9.9, Threshold: minimum, Status: "passed"}
		}
		marshalReleaseQuality = func(report releasequality.Report) ([]byte, error) {
			if report.Value != 9.9 {
				t.Fatalf("marshal report Value = %.1f; want 9.9", report.Value)
			}
			return []byte(`{"status":"passed"}`), nil
		}
		verifyReleaseQuality = func(report releasequality.Report, minimum float64) error {
			if minimum != 9.7 {
				t.Fatalf("verify minimum = %.1f; want 9.7", minimum)
			}
			return nil
		}

		var stdout, stderr bytes.Buffer
		got := run([]string{"score", "--min", "9.7"}, strings.NewReader(""), &stdout, &stderr)

		if got != 0 {
			t.Fatalf("score exit = %d, stderr %q; want 0", got, stderr.String())
		}
		if gotMinimum != 9.7 {
			t.Fatalf("compute minimum = %.1f; want 9.7", gotMinimum)
		}
		if strings.TrimSpace(stdout.String()) != `{"status":"passed"}` {
			t.Fatalf("stdout = %q; want JSON", stdout.String())
		}
	})

	t.Run("flag parse error", func(t *testing.T) {
		resetReleaseQuality()
		var stdout, stderr bytes.Buffer
		got := run([]string{"score", "--min", "nope"}, strings.NewReader(""), &stdout, &stderr)
		if got != 2 {
			t.Fatalf("score parse exit = %d; want 2", got)
		}
		if !strings.Contains(stderr.String(), "invalid value") {
			t.Fatalf("stderr = %q; want invalid value", stderr.String())
		}
	})

	t.Run("flag help", func(t *testing.T) {
		resetReleaseQuality()
		var stdout, stderr bytes.Buffer
		got := run([]string{"score", "-h"}, strings.NewReader(""), &stdout, &stderr)
		if got != 0 {
			t.Fatalf("score help exit = %d; want 0", got)
		}
		if !strings.Contains(stderr.String(), "minimum acceptable release score") {
			t.Fatalf("stderr = %q; want help", stderr.String())
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		resetReleaseQuality()
		computeReleaseQuality = func(minimum float64) releasequality.Report {
			return releasequality.Report{Value: 10, Threshold: minimum, Status: "passed"}
		}
		marshalReleaseQuality = func(report releasequality.Report) ([]byte, error) {
			return nil, errors.New("marshal boom")
		}

		var stdout, stderr bytes.Buffer
		got := run([]string{"score"}, strings.NewReader(""), &stdout, &stderr)

		if got != 1 {
			t.Fatalf("score marshal exit = %d; want 1", got)
		}
		if !strings.Contains(stderr.String(), "marshal boom") {
			t.Fatalf("stderr = %q; want marshal error", stderr.String())
		}
	})

	t.Run("verify error", func(t *testing.T) {
		resetReleaseQuality()
		computeReleaseQuality = func(minimum float64) releasequality.Report {
			return releasequality.Report{Value: 1, Threshold: minimum, Status: "failed"}
		}
		marshalReleaseQuality = func(report releasequality.Report) ([]byte, error) {
			return []byte(`{"status":"failed"}`), nil
		}
		verifyReleaseQuality = func(report releasequality.Report, minimum float64) error {
			return fmt.Errorf("score too low")
		}

		var stdout, stderr bytes.Buffer
		got := run([]string{"score"}, strings.NewReader(""), &stdout, &stderr)

		if got != 1 {
			t.Fatalf("score verify exit = %d; want 1", got)
		}
		if !strings.Contains(stdout.String(), `"failed"`) {
			t.Fatalf("stdout = %q; want failed report", stdout.String())
		}
		if !strings.Contains(stderr.String(), "score too low") {
			t.Fatalf("stderr = %q; want verify error", stderr.String())
		}
	})
}

func TestRunGovernanceCommands(t *testing.T) {
	t.Run("version json", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		got := run([]string{"version", "--json"}, strings.NewReader(""), &stdout, &stderr)
		if got != 0 {
			t.Fatalf("version exit = %d, stderr %q; want 0", got, stderr.String())
		}
		var report gateReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
		}
		if report.Command != "version" ||
			report.Status != "passed" ||
			!slicesContain(report.Details, "xlib-standard release v0.4.6") ||
			!slicesContain(report.Details, "goalcli governance runtime v2.9.3") {
			t.Fatalf("report = %#v; want version gate report", report)
		}
	})

	t.Run("artifact gate passes when required files exist", func(t *testing.T) {
		root := t.TempDir()
		commandSurface := strings.Join(commandRegistryRequiredCommands(), "\n")
		registrySurface := strings.Join(requiredCommandRegistryNeedles(), "\n")
		files := map[string]string{
			"docs/standard/goalcli-cli-contract.md": "goalcli\n" + commandSurface + "\n",
			"contracts/goalcli-report.schema.json":  "command status details gaps\n",
			".agent/command-registry.yaml":          registrySurface + "\n",
		}
		for rel, content := range files {
			path := filepath.Join(root, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
		}
		chdir(t, root)
		var stdout, stderr bytes.Buffer
		got := run([]string{"cli-contract", "--json"}, strings.NewReader(""), &stdout, &stderr)
		if got != 0 {
			t.Fatalf("cli-contract exit = %d, stderr %q; want 0", got, stderr.String())
		}
		var report gateReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
		}
		if report.Status != "passed" {
			t.Fatalf("report status = %q; want passed; report %#v", report.Status, report)
		}
	})

	t.Run("cli contract requires full command surface in docs and registry", func(t *testing.T) {
		root := t.TempDir()
		fullRegistry := strings.Join(commandRegistryRequiredCommands(), "\n") + "\n"
		fullCommandRegistry := strings.Join(requiredCommandRegistryNeedles(), "\n") + "\n"
		files := map[string]string{
			"docs/standard/goalcli-cli-contract.md": strings.Replace(fullRegistry, "execution-context\n", "", 1),
			"contracts/goalcli-report.schema.json":  "command status details gaps\n",
			".agent/command-registry.yaml":          fullCommandRegistry,
		}
		for rel, content := range files {
			path := filepath.Join(root, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
		}
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"cli-contract"}, strings.NewReader(""), &stdout, &stderr)
		if got != 1 {
			t.Fatalf("cli-contract incomplete docs exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
		}
		if !strings.Contains(stdout.String(), "docs/standard/goalcli-cli-contract.md missing execution-context") {
			t.Fatalf("stdout = %q; want missing execution-context gap", stdout.String())
		}
		if !strings.Contains(stderr.String(), "cli-contract found") {
			t.Fatalf("stderr = %q; want cli-contract gap summary", stderr.String())
		}
	})

	t.Run("artifact gate reports missing files", func(t *testing.T) {
		root := t.TempDir()
		chdir(t, root)
		var stdout, stderr bytes.Buffer
		got := run([]string{"issue-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 1 {
			t.Fatalf("issue-registry exit = %d; want 1", got)
		}
		if !strings.Contains(stderr.String(), "issue-registry found") {
			t.Fatalf("stderr = %q; want issue-registry gaps", stderr.String())
		}
	})
}

func TestRunExternalErrorPaths(t *testing.T) {
	t.Run("exit error returns command status", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		got := runExternal(strings.NewReader(""), &stdout, &stderr, shellPath(t), "-c", "exit 7")
		if got != 7 {
			t.Fatalf("runExternal exit status = %d; want 7", got)
		}
	})

	t.Run("non exit error is reported", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		got := runExternal(strings.NewReader(""), &stdout, &stderr, filepath.Join(t.TempDir(), "missing-command"))
		if got != 1 {
			t.Fatalf("runExternal missing command = %d; want 1", got)
		}
		if !strings.Contains(stderr.String(), "ERROR:") {
			t.Fatalf("stderr = %q; want ERROR", stderr.String())
		}
	})
}

func TestRunDoctorAllowsRenderedDownstreamWithoutSourceGoal(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                                "module github.com/ZoneCNH/kernel\n\nreplace github.com/ZoneCNH/xlib-standard => ../xlib-standard\n",
		".agent/harness.yaml":                   "checks: [version, doctor]\n",
		".agent/issue-registry.yaml":            issueRegistryFixture("P0-001", "P1-001", "P2-001", "CTX-001"),
		".agent/command-registry.yaml":          "commands: [version, doctor]\n",
		".agent/makefile-target-registry.yaml":  "targets: []\n",
		".agent/makefile-baseline.yaml":         "targets: []\n",
		"docs/standard/goalcli-cli-contract.md": "goalcli doctor\n",
		"contracts/goalcli-report.schema.json":  "{\"type\":\"object\"}\n",
		"Makefile":                              "doctor:\n\tgo run ./cmd/goalcli doctor\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	chdir(t, root)
	var stdout, stderr bytes.Buffer
	got := run([]string{"doctor"}, strings.NewReader(""), &stdout, &stderr)
	if got != 0 {
		t.Fatalf("doctor exit = %d, stderr %q, stdout %q; want 0", got, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), `"status": "passed"`) {
		t.Fatalf("stdout = %q; want passed", stdout.String())
	}
}

func TestCommandRegistryRequiresFullCommandSurface(t *testing.T) {
	t.Run("accepts complete registry surface", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, ".agent", "command-registry.yaml")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(strings.Join(requiredCommandRegistryNeedles(), "\n")+"\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"command-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 0 {
			t.Fatalf("command-registry complete fixture exit = %d, stderr %q, stdout %q; want 0", got, stderr.String(), stdout.String())
		}
		if !strings.Contains(stdout.String(), `"status": "passed"`) {
			t.Fatalf("stdout = %q; want passed report", stdout.String())
		}
	})

	t.Run("rejects incomplete registry surface", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, ".agent", "command-registry.yaml")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		content := strings.Replace(strings.Join(requiredCommandRegistryNeedles(), "\n")+"\n", "name: goal-certify\n", "", 1)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"command-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 1 {
			t.Fatalf("command-registry incomplete fixture exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
		}
		if !strings.Contains(stdout.String(), ".agent/command-registry.yaml missing name: goal-certify") {
			t.Fatalf("stdout = %q; want missing goal-certify gap", stdout.String())
		}
	})
}

func TestIssueRegistryRequiresDynamicContract(t *testing.T) {
	t.Run("accepts dynamic counts", func(t *testing.T) {
		root := t.TempDir()
		writeTestFiles(t, root, map[string]string{
			".agent/issue-registry.yaml": issueRegistryFixture("P0-001", "P0-002", "P1-001", "P2-001", "CTX-001", "CTX-002"),
		})
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"issue-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 0 {
			t.Fatalf("issue-registry dynamic fixture exit = %d, stderr %q, stdout %q; want 0", got, stderr.String(), stdout.String())
		}
		if !strings.Contains(stdout.String(), `"status": "passed"`) {
			t.Fatalf("stdout = %q; want passed report", stdout.String())
		}
	})

	t.Run("rejects non-contiguous ids", func(t *testing.T) {
		root := t.TempDir()
		writeTestFiles(t, root, map[string]string{
			".agent/issue-registry.yaml": issueRegistryFixture("P0-001", "P0-003", "P1-001", "P2-001", "CTX-001"),
		})
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"issue-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 1 {
			t.Fatalf("issue-registry gap fixture exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
		}
		var report gateReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
		}
		if !gapsContainSubstring(report.Gaps, ".agent/issue-registry.yaml P0 ids must be contiguous; missing P0-002") {
			t.Fatalf("gaps = %#v; want missing P0-002 gap", report.Gaps)
		}
	})

	t.Run("rejects duplicate ids", func(t *testing.T) {
		root := t.TempDir()
		writeTestFiles(t, root, map[string]string{
			".agent/issue-registry.yaml": issueRegistryFixture("P0-001", "P0-001", "P1-001", "P2-001", "CTX-001"),
		})
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"issue-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 1 {
			t.Fatalf("issue-registry duplicate fixture exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
		}
		var report gateReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
		}
		if !gapsContainSubstring(report.Gaps, ".agent/issue-registry.yaml duplicate issue id P0-001") {
			t.Fatalf("gaps = %#v; want duplicate id gap", report.Gaps)
		}
	})

	t.Run("rejects missing implemented evidence", func(t *testing.T) {
		root := t.TempDir()
		writeTestFiles(t, root, map[string]string{
			".agent/issue-registry.yaml": `schema_version: "2.9.3"
issues:
  - id: P0-001
    title: planned issue
    status: planned
    command: issue-registry
  - id: P1-001
    title: implemented issue
    status: implemented
    command: issue-registry
    evidence:
      - go test ./cmd/goalcli
  - id: P2-001
    title: implemented issue
    status: implemented
    command: issue-registry
    evidence:
      - go test ./cmd/goalcli
  - id: CTX-001
    title: implemented issue
    status: implemented
    command: issue-registry
    evidence:
      - go test ./cmd/goalcli
`,
		})
		chdir(t, root)

		var stdout, stderr bytes.Buffer
		got := run([]string{"issue-registry"}, strings.NewReader(""), &stdout, &stderr)
		if got != 1 {
			t.Fatalf("issue-registry invalid fixture exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
		}
		var report gateReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
		}
		if !gapsContainSubstring(report.Gaps, ".agent/issue-registry.yaml P0-001 status must be implemented") {
			t.Fatalf("gaps = %#v; want status gap", report.Gaps)
		}
		if !gapsContainSubstring(report.Gaps, ".agent/issue-registry.yaml P0-001 missing evidence") {
			t.Fatalf("gaps = %#v; want evidence gap", report.Gaps)
		}
	})
}

func TestRunInternalGovernanceCommands(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
	}{
		{name: "version", args: []string{"version"}, wantStdout: `"command": "version"`},
		{name: "doctor", args: []string{"doctor"}, wantStdout: `"status": "passed"`},
		{name: "main guard", args: []string{"main-guard", "--context", "local_readonly"}, wantStdout: `"command": "main-guard"`},
		{name: "worktree guard", args: []string{"worktree-guard", "--context", "local_readonly"}, wantStdout: `"command": "worktree-guard"`},
		{name: "worktree check", args: []string{"worktree-check", "--context", "local_readonly"}, wantStdout: `"command": "worktree-check"`},
		{name: "context check", args: []string{"context-check"}, wantStdout: `"status": "passed"`},
		{name: "spec check", args: []string{"spec-check"}, wantStdout: `"command": "spec-check"`},
		{name: "design check", args: []string{"design-check"}, wantStdout: `"status": "passed"`},
		{name: "task check", args: []string{"task-check"}, wantStdout: `"status": "passed"`},
		{name: "pr check dry-run", args: []string{"pr-check", "--context", "local_readonly", "--dry-run"}, wantStdout: `"command": "pr-check"`},
		{name: "evidence check", args: []string{"evidence-check"}, wantStdout: `"status": "passed"`},
		{name: "cli contract", args: []string{"cli-contract"}, wantStdout: `"command": "cli-contract"`},
		{name: "issue registry", args: []string{"issue-registry"}, wantStdout: `"command": "issue-registry"`},
		{name: "command registry", args: []string{"command-registry"}, wantStdout: `"command": "command-registry"`},
		{name: "makefile baseline", args: []string{"makefile-baseline"}, wantStdout: `"command": "makefile-baseline"`},
		{name: "p1 planned", args: []string{"agent-team-contract"}, wantStdout: `"status": "passed"`},
		{name: "p2 downstream gap explicit", args: []string{"downstream-adoption", "--repo", "kernel/configx", "--mode", "patch-only"}, wantCode: 1, wantStdout: `"status": "gap"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tt.args, strings.NewReader(""), &stdout, &stderr)
			if got != tt.wantCode {
				t.Fatalf("run(%v) = %d, stderr %q; want %d", tt.args, got, stderr.String(), tt.wantCode)
			}
			if !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Fatalf("stdout = %q; want containing %q", stdout.String(), tt.wantStdout)
			}
		})
	}
}

func TestRunInternalGovernanceCommandsRejectInvalidArgs(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	tests := []struct {
		name string
		args []string
	}{
		{name: "version unknown flag", args: []string{"version", "--bogus"}},
		{name: "doctor unknown flag", args: []string{"doctor", "--bogus"}},
		{name: "evidence-check unknown flag", args: []string{"evidence-check", "--bogus"}},
		{name: "cli-contract unknown flag", args: []string{"cli-contract", "--bogus"}},
		{name: "issue-registry unknown flag", args: []string{"issue-registry", "--bogus"}},
		{name: "command-registry unknown flag", args: []string{"command-registry", "--bogus"}},
		{name: "makefile-baseline unknown flag", args: []string{"makefile-baseline", "--bogus"}},
		{name: "cli-contract positional", args: []string{"cli-contract", "unexpected"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tt.args, strings.NewReader(""), &stdout, &stderr)
			if got != 2 {
				t.Fatalf("run(%v) = %d, stderr %q, stdout %q; want 2", tt.args, got, stderr.String(), stdout.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q; want empty output for invalid args", stdout.String())
			}
			if !strings.Contains(stderr.String(), "invalid arguments") {
				t.Fatalf("stderr = %q; want invalid arguments", stderr.String())
			}
		})
	}
}

func TestContextProfileRejectsUnknownProfileFlag(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	var stdout, stderr bytes.Buffer
	got := run([]string{"context-profile", "--profile", "missing"}, strings.NewReader(""), &stdout, &stderr)

	if got != 2 {
		t.Fatalf("context-profile unknown profile exit = %d, stderr %q; want 2", got, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q; want empty output for invalid profile", stdout.String())
	}
	if !strings.Contains(stderr.String(), `invalid context profile "missing"`) {
		t.Fatalf("stderr = %q; want invalid profile message", stderr.String())
	}
}

func TestContextProfileCheckAcceptsCurrentMakefileDAG(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	makefile := readText(t, "Makefile")
	var gaps []string
	appendContextProfileDAGGaps(makefile, &gaps)
	appendReleaseFinalDelegationGaps(makefile, &gaps)

	if len(gaps) > 0 {
		t.Fatalf("current context profile DAG gaps = %#v; want none", gaps)
	}
}

func TestContextProfileCheckAcceptsContinuedMakefileDependencies(t *testing.T) {
	makefile := contextProfileMakefileFixture(map[string]string{
		"context-standard": "require-gowork-off governance-check \\\n  p1-governance-check docs-check",
	}, "$(MAKE) context-release")
	var gaps []string

	appendMakefileTargetDependencyGaps(
		makefile,
		"context-standard",
		[]string{"require-gowork-off", "governance-check", "p1-governance-check", "docs-check"},
		[]string{"context-lite", "context-profile-check", "release-check", "release-final-check"},
		&gaps,
	)
	appendContextProfileDAGGaps(makefile, &gaps)

	if len(gaps) > 0 {
		t.Fatalf("continued context-standard dependency gaps = %#v; want none", gaps)
	}
}

func TestContextProfileCheckRejectsUnknownProfileGate(t *testing.T) {
	makefile := contextProfileMakefileFixture(map[string]string{}, "$(MAKE) context-release")
	var gaps []string

	appendContextProfileDAGGaps(strings.Replace(makefile, "docs-check", "missing-gate", 1), &gaps)

	if !gapsContainSubstring(gaps, "context-standard references unknown context gate missing-gate") {
		t.Fatalf("gaps = %#v; want unknown gate gap", gaps)
	}
}

func TestContextProfileCheckRejectsCyclicProfileDAG(t *testing.T) {
	makefile := contextProfileMakefileFixture(map[string]string{
		"context-full":    "require-gowork-off governance-check p1-governance-check p2-runtime-check context-release",
		"context-release": "require-gowork-off context-full integration dependency-check standard-impact-check score-check",
	}, "$(MAKE) context-release")
	var gaps []string

	appendContextProfileDAGGaps(makefile, &gaps)

	if !gapsContainSubstring(gaps, "Makefile context profile DAG cycle:") {
		t.Fatalf("gaps = %#v; want profile DAG cycle gap", gaps)
	}
}

func TestContextProfileCheckRejectsReleaseFinalSelfRecursion(t *testing.T) {
	makefile := contextProfileMakefileFixture(map[string]string{}, "$(MAKE) release-final-check")
	var gaps []string

	appendReleaseFinalDelegationGaps(makefile, &gaps)

	if !gapsContainSubstring(gaps, "release-final-check must not call itself") {
		t.Fatalf("gaps = %#v; want self-recursion gap", gaps)
	}
	if !gapsContainSubstring(gaps, "release-final-check must call context-release") {
		t.Fatalf("gaps = %#v; want context-release delegation gap", gaps)
	}
}

func TestContextProfileCheckRejectsTransitiveReleaseFinalRecursion(t *testing.T) {
	makefile := contextProfileMakefileFixture(map[string]string{
		"context-full": "require-gowork-off governance-check p1-governance-check p2-runtime-check release-final-check",
	}, "$(MAKE) context-release")
	var gaps []string

	appendContextProfileDAGGaps(makefile, &gaps)

	if !gapsContainSubstring(gaps, "Makefile context-release must not reach release-final-check") {
		t.Fatalf("gaps = %#v; want transitive release-final gap", gaps)
	}
}

func TestPlannedCommandVerifyPassesWithManifestCoverage(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	var stdout, stderr bytes.Buffer
	got := run([]string{"agent-team-contract", "--dry-run", "--verify"}, strings.NewReader(""), &stdout, &stderr)
	if got != 0 {
		t.Fatalf("verified planned command exit = %d, stderr %q, stdout %q; want 0", got, stderr.String(), stdout.String())
	}

	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if report.Status != "passed" {
		t.Fatalf("report status = %q; want passed; report %#v", report.Status, report)
	}
	if !slicesContain(report.Details, "local dry-run verifier satisfied manifest coverage") {
		t.Fatalf("details = %#v; want manifest verifier detail", report.Details)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q; want empty stderr", stderr.String())
	}
}

func TestDownstreamGapVerifyIsBlocking(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	var stdout, stderr bytes.Buffer
	got := run([]string{"downstream-adoption", "--repo", "kernel/configx", "--mode", "patch-only", "--verify"}, strings.NewReader(""), &stdout, &stderr)
	if got != 1 {
		t.Fatalf("verified downstream gap exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
	}

	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if report.Status != "gap" {
		t.Fatalf("report status = %q; want gap; report %#v", report.Status, report)
	}
	if !slicesContain(report.Gaps, "downstream repo unavailable in worker workspace: kernel/configx") {
		t.Fatalf("gaps = %#v; want downstream repo gap", report.Gaps)
	}
	if !slicesContain(report.Details, "dry_run=true") {
		t.Fatalf("details = %#v; want dry_run=true", report.Details)
	}
}

func TestDownstreamGapWithoutVerifyIsBlocking(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	var stdout, stderr bytes.Buffer
	got := run([]string{"downstream-adoption", "--repo", "kernel/configx", "--mode", "patch-only"}, strings.NewReader(""), &stdout, &stderr)
	if got != 1 {
		t.Fatalf("downstream gap exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
	}

	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if report.Status != "gap" {
		t.Fatalf("report status = %q; want gap; report %#v", report.Status, report)
	}
	if !strings.Contains(stderr.String(), "cannot satisfy a release gate") {
		t.Fatalf("stderr = %q; want release gate blocker", stderr.String())
	}
}

func TestPlannedCommandRequiresManifestCoverage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	got := runPlannedCommand("missing-planned-command", []string{"--dry-run"}, &stdout, &stderr)
	if got != 1 {
		t.Fatalf("missing planned coverage exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
	}

	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if report.Status != "failed" {
		t.Fatalf("report status = %q; want failed; report %#v", report.Status, report)
	}
	if !slicesContain(report.Gaps, "planned command has no manifest coverage: missing-planned-command") {
		t.Fatalf("gaps = %#v; want manifest coverage gap", report.Gaps)
	}
}

func TestPlannedCommandRequiresSemanticManifestContent(t *testing.T) {
	root := t.TempDir()
	writeTestFiles(t, root, map[string]string{
		".agent/team-contract.yaml": `schema_version: "2.9.3"
roles:
  - leader
`,
	})
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	got := run([]string{"agent-team-contract", "--dry-run", "--verify"}, strings.NewReader(""), &stdout, &stderr)
	if got != 1 {
		t.Fatalf("semantic planned command exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
	}
	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if !gapsContainSubstring(report.Gaps, ".agent/team-contract.yaml missing semantic marker rule:") {
		t.Fatalf("gaps = %#v; want semantic marker gap", report.Gaps)
	}
}

func TestGoalcliMVAGoalCommandsRequireHarnessMarkers(t *testing.T) {
	tests := []struct {
		command string
		marker  string
	}{
		{command: "goal-acceptance", marker: "G12_ACCEPTANCE"},
		{command: "goal-delivery", marker: "G13_DELIVERY"},
		{command: "goal-handover", marker: "G14_HANDOVER"},
		{command: "goal-downstream-adoption", marker: "G15_DOWNSTREAM_ADOPTION"},
		{command: "goal-certify", marker: "G16_CERTIFY"},
		{command: "goal-runtime-final", marker: "G12_G16_FINAL"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			root := t.TempDir()
			writeTestFiles(t, root, map[string]string{
				".agent/harness.yaml": "goalcli_mva_gates:\n  command: " + tt.command + "\n  status: dry_run_ready\n",
			})
			chdir(t, root)

			var stdout, stderr bytes.Buffer
			got := run([]string{tt.command, "--dry-run", "--verify"}, strings.NewReader(""), &stdout, &stderr)
			if got != 1 {
				t.Fatalf("%s missing marker exit = %d, stderr %q, stdout %q; want 1", tt.command, got, stderr.String(), stdout.String())
			}
			var report gateReport
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
			}
			if !gapsContainSubstring(report.Gaps, ".agent/harness.yaml missing semantic marker "+tt.marker) {
				t.Fatalf("gaps = %#v; want missing %s marker gap", report.Gaps, tt.marker)
			}
		})
	}
}

func TestPlannedCommandRejectsDirectoryManifest(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agent", "team-contract.yaml"), 0o755); err != nil {
		t.Fatalf("mkdir fixture directory: %v", err)
	}
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	got := run([]string{"agent-team-contract", "--dry-run", "--verify"}, strings.NewReader(""), &stdout, &stderr)
	if got != 1 {
		t.Fatalf("directory planned command exit = %d, stderr %q, stdout %q; want 1", got, stderr.String(), stdout.String())
	}
	var report gateReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
	}
	if !gapsContainSubstring(report.Gaps, ".agent/team-contract.yaml must be a file") {
		t.Fatalf("gaps = %#v; want file requirement gap", report.Gaps)
	}
}

func TestPlannedCommandRejectsInvalidArgs(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{name: "unknown flag", args: []string{"agent-team-contract", "--bogus"}, wantStderr: "invalid arguments"},
		{name: "invalid context", args: []string{"agent-team-contract", "--context", "bad_context"}, wantStderr: `invalid context "bad_context"`},
		{name: "positional arg", args: []string{"agent-team-contract", "extra"}, wantStderr: `unexpected positional argument "extra"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tt.args, strings.NewReader(""), &stdout, &stderr)
			if got != 2 {
				t.Fatalf("invalid planned command arg exit = %d, stderr %q, stdout %q; want 2", got, stderr.String(), stdout.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q; want empty output for invalid args", stdout.String())
			}
			if !strings.Contains(stderr.String(), "invalid arguments") || !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q; want invalid arguments and %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestInternalGovernanceCommandsRejectUnknownArgs(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	for _, command := range []string{"cli-contract", "issue-registry", "command-registry", "makefile-baseline"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run([]string{command, "--bogus"}, strings.NewReader(""), &stdout, &stderr)
			if got != 2 {
				t.Fatalf("%s invalid arg exit = %d, stderr %q, stdout %q; want 2", command, got, stderr.String(), stdout.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q; want empty output for invalid args", stdout.String())
			}
			if !strings.Contains(stderr.String(), command+" invalid arguments") {
				t.Fatalf("stderr = %q; want command-specific invalid arguments", stderr.String())
			}
		})
	}
}

func TestGuardContextsIncludePullRequest(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	for _, context := range []string{"local_write", "local_readonly", "ci_pull_request", "ci_main_verify", "release_verify"} {
		if !validContext(context) {
			t.Fatalf("validContext(%q) = false; want true", context)
		}
	}

	for _, command := range []string{"main-guard", "worktree-guard"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run([]string{command, "--context", "ci_pull_request"}, strings.NewReader(""), &stdout, &stderr)
			if got != 0 {
				t.Fatalf("run(%s ci_pull_request) = %d, stderr %q, stdout %q; want 0", command, got, stderr.String(), stdout.String())
			}
			var report gateReport
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("stdout is not gateReport JSON: %v; stdout %q", err, stdout.String())
			}
			if report.Status != "passed" || !slicesContain(report.Details, "context=ci_pull_request") {
				t.Fatalf("report = %#v; want passed ci_pull_request context", report)
			}
		})
	}
}

func TestMakefileReleaseGuardsUseContextVariable(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repoRoot(t), "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	text := string(content)

	for _, want := range []string{
		"$(GOALCLI) main-guard --context $(XLIB_CONTEXT)",
		"$(GOALCLI) worktree-guard --context $(XLIB_CONTEXT)",
		"XLIB_CONTEXT=release_verify GOWORK=off $(MAKE) context-release",
		"$(MAKE) debt-evidence-checksum-check",
		"GOWORK=off XLIB_CONTEXT=release_verify VERSION=\"$(VERSION)\" $(MAKE) release-final-check",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("Makefile missing %q", want)
		}
	}
}

func TestContextProfileCheckRejectsUnknownProfile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	got := run([]string{"context-profile-check", "--profile", "missing"}, strings.NewReader(""), &stdout, &stderr)
	if got != 2 {
		t.Fatalf("unknown context-profile-check profile exit = %d, stderr %q, stdout %q; want 2", got, stderr.String(), stdout.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q; want empty output for invalid profile", stdout.String())
	}
	if !strings.Contains(stderr.String(), `invalid context profile "missing"`) {
		t.Fatalf("stderr = %q; want invalid profile message", stderr.String())
	}
}

func TestContextProfileContractRejectsUnknownGateAndCycles(t *testing.T) {
	original := contextProfileGates
	t.Cleanup(func() {
		contextProfileGates = original
	})

	contextProfileGates = map[string][]string{
		"lite":     {"governance-check"},
		"standard": {"context-lite", "missing-gate"},
	}
	var gaps []string
	appendContextProfileContractGaps("governance-check:\ncontext-lite:\n", &gaps)
	if !slicesContain(gaps, "context profile standard references unknown Makefile gate missing-gate") {
		t.Fatalf("gaps = %#v; want unknown gate gap", gaps)
	}

	contextProfileGates = map[string][]string{
		"lite":     {"context-standard"},
		"standard": {"context-lite"},
	}
	gaps = nil
	appendContextProfileContractGaps("context-lite:\ncontext-standard:\n", &gaps)
	if !gapsContainSubstring(gaps, "context profile DAG cycle:") {
		t.Fatalf("gaps = %#v; want context profile DAG cycle gap", gaps)
	}
}

func TestReleaseFinalDelegationRejectsSelfRecursion(t *testing.T) {
	makefile := "release-final-check: context-release release-final-check\n\t$(MAKE) release-final-check\n"
	var gaps []string
	appendReleaseFinalDelegationGaps(makefile, &gaps)
	if !slicesContain(gaps, "release-final-check must not call itself") {
		t.Fatalf("gaps = %#v; want self-recursion gap", gaps)
	}
}

func TestContextReleaseRejectsForbiddenReleaseReferences(t *testing.T) {
	makefile := "context-release:\n\t$(MAKE) release-final-check\n"
	var gaps []string
	appendMakefileTargetForbiddenReferenceGaps(makefile, "context-release", []string{"release-check", "release-final-check"}, &gaps)
	if !slicesContain(gaps, "Makefile context-release must not reference release-final-check") {
		t.Fatalf("gaps = %#v; want forbidden release-final-check gap", gaps)
	}
}

func TestCIWorkflowReleaseCheckUsesExplicitContext(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read ci workflow: %v", err)
	}
	text := string(content)

	want := "GOWORK=off XLIB_CONTEXT=ci_pull_request make release-check"
	if !strings.Contains(text, want) {
		t.Fatalf("ci workflow missing %q", want)
	}

	for _, target := range []string{"governance-check", "p1-governance-check", "p2-runtime-check"} {
		for _, line := range strings.Split(text, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(trimmed, "make "+target) {
				t.Fatalf("ci workflow runs duplicate %s outside release-check: %q", target, trimmed)
			}
		}
	}
}

func TestRunGuardsRejectInvalidContext(t *testing.T) {
	for _, command := range []string{"main-guard", "worktree-guard"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run([]string{command, "--context", "invalid"}, strings.NewReader(""), &stdout, &stderr)
			if got != 2 {
				t.Fatalf("%s invalid context = %d; want 2", command, got)
			}
			if !strings.Contains(stderr.String(), "invalid context") {
				t.Fatalf("stderr = %q; want invalid context", stderr.String())
			}
		})
	}
}

func TestRunGuardsAcceptPullRequestContext(t *testing.T) {
	chdir(t, filepath.Join("..", ".."))

	for _, command := range []string{"main-guard", "worktree-guard"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run([]string{command, "--context", "ci_pull_request"}, strings.NewReader(""), &stdout, &stderr)
			if got != 0 {
				t.Fatalf("%s ci_pull_request exit = %d, stderr %q; want 0", command, got, stderr.String())
			}
			if !strings.Contains(stdout.String(), "context=ci_pull_request") {
				t.Fatalf("stdout = %q; want ci_pull_request context detail", stdout.String())
			}
		})
	}
}

func TestVersionConstantsTrackChangelogRelease(t *testing.T) {
	root := repoRoot(t)
	latest := latestChangelogVersion(t, readText(t, filepath.Join(root, "CHANGELOG.md")))
	_, standardMarkerErr := os.Stat(filepath.Join(root, "docs", "goal", "goal.md"))
	if standardMarkerErr != nil && !errors.Is(standardMarkerErr, os.ErrNotExist) {
		t.Fatalf("stat docs/goal/goal.md: %v", standardMarkerErr)
	}
	isStandardSource := standardMarkerErr == nil
	if projectReleaseVersion != latest {
		t.Fatalf("projectReleaseVersion = %q; want latest changelog version %q", projectReleaseVersion, latest)
	}
	if templatex.Version != latest {
		t.Fatalf("templatex.Version = %q; want latest changelog version %q", templatex.Version, latest)
	}

	for _, rel := range []string{
		"release/manifest/template.json",
		"internal/tools/releasemanifest/main.go",
		".agent/harness.yaml",
		"README.md",
		"docs/release.md",
		"AGENTS.md",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		text, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) && !isStandardSource {
			continue
		}
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(text), latest) {
			t.Fatalf("%s does not contain latest release version %s", rel, latest)
		}
	}
}

func writeGateScript(t *testing.T, root string, relative string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	body := "#!/bin/sh\nprintf '%s' \"$(basename \"$0\")\"\nfor arg in \"$@\"; do printf ' %s' \"$arg\"; done\nprintf '\\n'\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writePathTool(t *testing.T, root string, name string) {
	t.Helper()
	path := filepath.Join(root, name)
	body := "#!/bin/sh\nprintf '%s' \"$(basename \"$0\")\"\nfor arg in \"$@\"; do printf ' %s' \"$arg\"; done\nprintf '\\n'\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeTestFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

func issueRegistryFixture(ids ...string) string {
	var b strings.Builder
	b.WriteString("schema_version: \"2.9.3\"\nissues:\n")
	for _, id := range ids {
		b.WriteString("  - id: ")
		b.WriteString(id)
		b.WriteString("\n")
		b.WriteString("    title: implemented issue ")
		b.WriteString(id)
		b.WriteString("\n")
		b.WriteString("    status: implemented\n")
		b.WriteString("    command: issue-registry\n")
		b.WriteString("    evidence:\n")
		b.WriteString("      - go test ./cmd/goalcli\n")
	}
	return b.String()
}

func gapsContainSubstring(gaps []string, want string) bool {
	for _, gap := range gaps {
		if strings.Contains(gap, want) {
			return true
		}
	}
	return false
}

func contextProfileMakefileFixture(overrides map[string]string, releaseFinalBody string) string {
	dependencies := map[string]string{
		"context-lite":     "require-gowork-off governance-check",
		"context-standard": "require-gowork-off governance-check p1-governance-check docs-check",
		"context-full":     "require-gowork-off governance-check p1-governance-check p2-runtime-check",
		"context-release":  "require-gowork-off context-full integration dependency-check standard-impact-check score-check",
	}
	for target, deps := range overrides {
		dependencies[target] = deps
	}
	targets := []string{"context-lite", "context-standard", "context-full", "context-release"}
	var b strings.Builder
	for _, target := range targets {
		b.WriteString(target)
		b.WriteString(": ")
		b.WriteString(dependencies[target])
		b.WriteString("\n\t@true\n")
	}
	b.WriteString("release-final-check:\n\t")
	b.WriteString(releaseFinalBody)
	b.WriteString("\n")
	return b.String()
}

func writeGoalcliAuthorityFixture(t *testing.T, root string) {
	t.Helper()
	for _, path := range []string{
		".worktree/goalcli-v0.1.0-plan.md",
		".omx/context/goalcli-v0.1.0-team-20260603T005302Z.md",
		"docs/standard/goalcli-cli-contract.md",
		".agent/harness.yaml",
		".agent/command-registry.yaml",
		".agent/registry/runtime.yaml",
		".agent/registry/commands.yaml",
		".agent/command-implementation-status.yaml",
		".agent/evidence/README.md",
		"docs/standard/goalcli-runtime.md",
		"docs/plans/goalcli-v0.1.0-roadmap.md",
		"docs/adr/ADR-20260603-001-goalcli-runtime.md",
		"Makefile",
	} {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir fixture path %s: %v", path, err)
		}
		if err := os.WriteFile(full, []byte("fixture\n"), 0o644); err != nil {
			t.Fatalf("write fixture path %s: %v", path, err)
		}
	}
}

func writeGoalcliPrerequisiteLedgerFixture(t *testing.T, root string, goalID string) {
	t.Helper()
	for _, command := range []string{
		"goal-acceptance",
		"goal-delivery",
		"goal-handover",
		"goal-downstream-adoption",
		"goal-certify",
	} {
		report, err := goalruntime.Evaluate(command, goalruntime.Options{
			Root:   root,
			GoalID: goalID,
		})
		if err != nil {
			t.Fatalf("Evaluate prerequisite %s returned error: %v", command, err)
		}
		if err := goalruntime.WriteEvidence(root, report); err != nil {
			t.Fatalf("WriteEvidence prerequisite %s returned error: %v", command, err)
		}
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatalf("restore cwd %s: %v", old, err)
		}
	})
}

func shellPath(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	for _, candidate := range []string{"/bin/sh", "/usr/bin/sh"} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Fatal("no POSIX shell found")
	return ""
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func isStandardSourceRepo(t *testing.T, root string) bool {
	t.Helper()
	data := readText(t, filepath.Join(root, "go.mod"))
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1] == "github.com/ZoneCNH/"+strings.Join([]string{"xlib", "standard"}, "-")
		}
	}
	t.Fatalf("module path not found in %s", filepath.Join(root, "go.mod"))
	return false
}

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func latestChangelogVersion(t *testing.T, text string) string {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## v") {
			fields := strings.Fields(strings.TrimPrefix(trimmed, "## "))
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}
	t.Fatal("latest changelog version not found")
	return ""
}

func slicesContain(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestRulesConsistencyCheckPassesOnConsistentFixture(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	if err := os.MkdirAll(".agent/standard", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".agent/rules", 0o755); err != nil {
		t.Fatal(err)
	}
	canonical := "# Canonical\n\n## 1. 八条铁律\n\n" +
		"| ID | 铁律 | 机器化实现 |\n|---|---|---|\n" +
		"| RULE-CORE-001 | x | y |\n" +
		"| RULE-WORKTREE-001 | a | b |\n\n## 2. 其他\n"
	iron := "# Iron\n\n## 七律\n\n" +
		"1. evidence (RULE-CORE-001 / RULE-EVIDENCE-001).\n" +
		"5. worktree (RULE-WORKTREE-001 / RULE-MERGE-001).\n"
	registry := "rules:\n" +
		"  - id: RULE-CORE-001\n    level: P0\n" +
		"  - id: RULE-WORKTREE-001\n    level: P0\n" +
		"  - id: RULE-EVIDENCE-001\n    level: P0\n" +
		"  - id: RULE-MERGE-001\n    level: P0\n"
	for path, content := range map[string]string{
		".agent/standard/goal-runtime-canonical.md": canonical,
		".agent/rules/iron-rules.md":                iron,
		".agent/rules/registry.yaml":                registry,
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	got := runRulesConsistencyCheck(nil, &stdout, &stderr)
	if got != 0 {
		t.Fatalf("got exit=%d stderr=%q stdout=%q", got, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), `"passed"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRulesConsistencyCheckDetectsDriftAndUnregistered(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	if err := os.MkdirAll(".agent/standard", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".agent/rules", 0o755); err != nil {
		t.Fatal(err)
	}
	canonical := "## 1. 八条铁律\n| ID | 铁律 | 实现 |\n|---|---|---|\n" +
		"| RULE-CORE-001 | x | y |\n" +
		"| RULE-MISSING-001 | a | b |\n\n## 2. 其他\n"
	iron := "## 七律\n1. (RULE-CORE-001).\n"
	registry := "rules:\n  - id: RULE-CORE-001\n"
	for path, content := range map[string]string{
		".agent/standard/goal-runtime-canonical.md": canonical,
		".agent/rules/iron-rules.md":                iron,
		".agent/rules/registry.yaml":                registry,
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	got := runRulesConsistencyCheck(nil, &stdout, &stderr)
	if got == 0 {
		t.Fatalf("expected failure; stdout=%q", stdout.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "RULE-MISSING-001 未在") {
		t.Errorf("want unregistered gap; stdout=%q", out)
	}
	if !strings.Contains(out, "漂移：") {
		t.Errorf("want drift gap; stdout=%q", out)
	}
}
