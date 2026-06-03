package goalruntime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateGoalRuntimeFinalPassesWithAuthority(t *testing.T) {
	root := t.TempDir()
	writeAuthorityFixture(t, root)
	writePrerequisiteLedgerFixture(t, root, DefaultGoalID)

	report, err := Evaluate("goal-runtime-final", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if report.Status != "passed" {
		t.Fatalf("status = %q; gaps %#v", report.Status, report.Gaps)
	}
	if report.GoalID != DefaultGoalID {
		t.Fatalf("goal_id = %q; want %q", report.GoalID, DefaultGoalID)
	}
	if report.Gate != "G12_G16_FINAL" {
		t.Fatalf("gate = %q; want G12_G16_FINAL", report.Gate)
	}
	if !report.Blocking {
		t.Fatalf("blocking = false; want final runtime evidence to be blocking")
	}
	if report.MVAStatus != "complete" {
		t.Fatalf("mva_status = %q; want complete", report.MVAStatus)
	}
	for _, gate := range report.Gates {
		if !gate.Blocking {
			t.Fatalf("gate = %#v; want goalcli MVA gates to be blocking", gate)
		}
	}
	if !contains(report.Evidence, "source_evidence_ledger="+SourceLedgerPath) {
		t.Fatalf("evidence = %#v; want source ledger path", report.Evidence)
	}
	if !contains(report.Evidence, "generated_evidence_pack="+EvidenceLedgerPath) {
		t.Fatalf("evidence = %#v; want generated evidence pack path", report.Evidence)
	}
	if !contains(report.Evidence, "requires=goal-certify") {
		t.Fatalf("evidence = %#v; want final gate dependency", report.Evidence)
	}
	if !containsSubstring(report.Details, "完成状态由本地 authority 校验和 evidence 写入共同证明") {
		t.Fatalf("details = %#v; want completion evidence boundary", report.Details)
	}
	if len(report.AuthorityPaths) == 0 {
		t.Fatalf("authority_paths is empty")
	}
}

func TestEvaluateRejectsUnknownCommand(t *testing.T) {
	if _, err := Evaluate("not-a-goalcli-command", Options{}); err == nil {
		t.Fatalf("Evaluate returned nil error for unknown command")
	}
}

func TestEvaluateReportsMissingAuthorityPaths(t *testing.T) {
	root := t.TempDir()
	report, err := Evaluate("goal-acceptance", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if report.Status != "failed" {
		t.Fatalf("status = %q; want failed", report.Status)
	}
	if len(report.Gaps) == 0 {
		t.Fatalf("gaps is empty; want missing authority paths")
	}
	if !containsSubstring(report.Gaps, ".worktree/goalcli-v0.1.0-plan.md") {
		t.Fatalf("gaps = %#v; want root plan gap", report.Gaps)
	}
	if report.MVAStatus != "not-complete" {
		t.Fatalf("mva_status = %q; want not-complete when authority is missing", report.MVAStatus)
	}
}

func TestEvaluateRenderedDownstreamSkipsSourceOnlyAuthorityPaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/ZoneCNH/kernel\n"), 0o644); err != nil {
		t.Fatalf("write downstream go.mod: %v", err)
	}
	writeAuthorityPaths(t, root, portableAuthorityPaths)

	report, err := Evaluate("goal-acceptance", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if report.Status != "passed" {
		t.Fatalf("status = %q; gaps %#v", report.Status, report.Gaps)
	}
	for _, path := range sourceOnlyAuthorityPaths {
		if contains(report.AuthorityPaths, path) {
			t.Fatalf("authority_paths = %#v; want source-only path %s skipped", report.AuthorityPaths, path)
		}
		if containsSubstring(report.Gaps, path) {
			t.Fatalf("gaps = %#v; want source-only path %s skipped", report.Gaps, path)
		}
	}
	if !containsSubstring(report.Details, "rendered downstream") {
		t.Fatalf("details = %#v; want rendered downstream boundary", report.Details)
	}
}

func TestEvaluateGoalcliGatePassesAsBlockingMVAContract(t *testing.T) {
	root := t.TempDir()
	writeAuthorityFixture(t, root)

	report, err := Evaluate("goal-acceptance", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if report.Status != "passed" || report.MVAStatus != "complete" || !report.Blocking {
		t.Fatalf("report = %#v; want passed complete blocking goalcli gate", report)
	}
	if len(report.Gates) != 1 || !report.Gates[0].Blocking {
		t.Fatalf("gates = %#v; want one blocking gate", report.Gates)
	}
}

func TestEvaluateGoalRuntimeFinalRequiresPrerequisiteLedger(t *testing.T) {
	root := t.TempDir()
	writeAuthorityFixture(t, root)

	report, err := Evaluate("goal-runtime-final", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if report.Status != "failed" || report.MVAStatus != "not-complete" {
		t.Fatalf("report = %#v; want failed not-complete when prerequisite ledger is missing", report)
	}
	if !containsSubstring(report.Gaps, SourceLedgerPath) {
		t.Fatalf("gaps = %#v; want source ledger prerequisite gap", report.Gaps)
	}
}

func TestWriteEvidenceWritesPackAndLedgerIdempotently(t *testing.T) {
	root := t.TempDir()
	writeAuthorityFixture(t, root)
	writePrerequisiteLedgerFixture(t, root, "GOAL-20260603-XLIB-GOALCLI-001")
	report, err := Evaluate("goal-runtime-final", Options{
		Root:   root,
		GoalID: "GOAL-20260603-XLIB-GOALCLI-001",
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if err := WriteEvidence(root, report); err != nil {
		t.Fatalf("WriteEvidence returned error: %v", err)
	}
	if err := WriteEvidence(root, report); err != nil {
		t.Fatalf("second WriteEvidence returned error: %v", err)
	}

	packPath := filepath.Join(root, filepath.FromSlash(report.EvidencePackPath))
	pack, err := os.ReadFile(packPath)
	if err != nil {
		t.Fatalf("read evidence pack: %v", err)
	}
	if !strings.Contains(string(pack), `"mva_status": "complete"`) || !strings.Contains(string(pack), `"blocking": true`) {
		t.Fatalf("evidence pack = %s; want complete blocking report", pack)
	}
	ledgerPath := filepath.Join(root, filepath.FromSlash(SourceLedgerPath))
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read evidence ledger: %v", err)
	}
	if strings.Count(string(ledger), `"command":"goal-runtime-final"`) != 1 {
		t.Fatalf("ledger = %s; want one idempotent final entry for %s", ledger, report.EvidencePackPath)
	}
}

func TestWriteEvidenceRecordsPrerequisiteLedgerEntry(t *testing.T) {
	root := t.TempDir()
	writeAuthorityFixture(t, root)
	report, err := Evaluate("goal-acceptance", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if err := WriteEvidence(root, report); err != nil {
		t.Fatalf("WriteEvidence returned error: %v", err)
	}
	ledger, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(SourceLedgerPath)))
	if err != nil {
		t.Fatalf("read evidence ledger: %v", err)
	}
	if !strings.Contains(string(ledger), `"command":"goal-acceptance"`) || !strings.Contains(string(ledger), `"mva_status":"complete"`) {
		t.Fatalf("ledger = %s; want goal-acceptance complete entry", ledger)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(report.EvidencePackPath))); !os.IsNotExist(err) {
		t.Fatalf("non-final WriteEvidence must not write generated pack, stat err = %v", err)
	}
}

func TestWriteEvidenceRejectsIncompleteReport(t *testing.T) {
	root := t.TempDir()
	writeAuthorityFixture(t, root)
	report, err := Evaluate("goal-acceptance", Options{Root: root})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	report.Status = "failed"
	if err := WriteEvidence(root, report); err == nil {
		t.Fatalf("WriteEvidence returned nil error for incomplete report")
	}
}

func writeAuthorityFixture(t *testing.T, root string) {
	t.Helper()
	writeAuthorityPaths(t, root, requiredAuthorityPaths(true))
}

func writeAuthorityPaths(t *testing.T, root string, paths []string) {
	t.Helper()
	for _, path := range paths {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir fixture path %s: %v", path, err)
		}
		if err := os.WriteFile(full, []byte("fixture\n"), 0o644); err != nil {
			t.Fatalf("write fixture path %s: %v", path, err)
		}
	}
}

func writePrerequisiteLedgerFixture(t *testing.T, root string, goalID string) {
	t.Helper()
	for _, command := range finalPrerequisiteCommands {
		report, err := Evaluate(command, Options{Root: root, GoalID: goalID})
		if err != nil {
			t.Fatalf("Evaluate prerequisite %s returned error: %v", command, err)
		}
		if err := WriteEvidence(root, report); err != nil {
			t.Fatalf("WriteEvidence prerequisite %s returned error: %v", command, err)
		}
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsSubstring(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}
