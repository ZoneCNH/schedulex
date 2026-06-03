package goalruntime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultGoalID       = "GOAL-20260603-XLIB-GOALCLI-001"
	EvidenceLedgerPath  = "release/evidence/goalcli/"
	SourceLedgerPath    = ".agent/evidence/ledger.jsonl"
	finalRuntimeCommand = "goal-runtime-final"
)

var standardModulePath = "github.com/ZoneCNH/" + strings.Join([]string{"xlib", "standard"}, "-")

var commandGates = map[string]string{
	"goal-acceptance":          "G12",
	"goal-delivery":            "G13",
	"goal-handover":            "G14",
	"goal-downstream-adoption": "G15A",
	"goal-certify":             "G16",
	"goal-runtime-final":       "G12_G16_FINAL",
}

var commandHarnessGates = map[string]string{
	"goal-acceptance":          "G12_ACCEPTANCE",
	"goal-delivery":            "G13_DELIVERY",
	"goal-handover":            "G14_HANDOVER",
	"goal-downstream-adoption": "G15_DOWNSTREAM_ADOPTION",
	"goal-certify":             "G16_CERTIFY",
}

var gateDescriptions = map[string]string{
	"goal-acceptance":          "验收矩阵和目标 ID contract 已收敛",
	"goal-delivery":            "交付证据路径和 goalcli 执行面已收敛",
	"goal-handover":            "接手材料和边界声明已收敛",
	"goal-downstream-adoption": "下游采用证明保持为本地 contract，不修改 downstream 仓库",
	"goal-certify":             "认证证明保持为本地 contract，不宣称完整 release 完成",
	"goal-runtime-final":       "G12-G16 本地 contract 汇总为 goalcli v0.1.0 MVA final evidence",
}

var sourceOnlyAuthorityPaths = []string{
	".worktree/goalcli-v0.1.0-plan.md",
	".omx/context/goalcli-v0.1.0-team-20260603T005302Z.md",
	"docs/adr/ADR-20260603-001-goalcli-runtime.md",
}

var portableAuthorityPaths = []string{
	"docs/standard/goalcli-cli-contract.md",
	".agent/harness.yaml",
	".agent/command-registry.yaml",
	".agent/registry/runtime.yaml",
	".agent/registry/commands.yaml",
	".agent/command-implementation-status.yaml",
	".agent/evidence/README.md",
	"docs/standard/goalcli-runtime.md",
	"docs/plans/goalcli-v0.1.0-roadmap.md",
	"Makefile",
}

var finalPrerequisiteCommands = []string{
	"goal-acceptance",
	"goal-delivery",
	"goal-handover",
	"goal-downstream-adoption",
	"goal-certify",
}

// Options configures a goalcli MVA contract evaluation.
type Options struct {
	GoalID string
	Mode   string
	Root   string
}

// Report is the machine-readable goalcli MVA evidence returned by goalcli.
type Report struct {
	SchemaVersion    string       `json:"schema_version"`
	Command          string       `json:"command"`
	Status           string       `json:"status"`
	GoalID           string       `json:"goal_id"`
	Gate             string       `json:"gate"`
	Mode             string       `json:"mode"`
	Executor         string       `json:"executor"`
	ControlPlane     string       `json:"control_plane"`
	Blocking         bool         `json:"blocking"`
	MVAStatus        string       `json:"mva_status"`
	LedgerPath       string       `json:"ledger_path"`
	EvidencePackPath string       `json:"evidence_pack_path"`
	Gates            []GateReport `json:"gates,omitempty"`
	Details          []string     `json:"details,omitempty"`
	Evidence         []string     `json:"evidence,omitempty"`
	AuthorityPaths   []string     `json:"authority_paths,omitempty"`
	Gaps             []string     `json:"gaps,omitempty"`
}

// GateReport records one local goalcli evidence gate.
type GateReport struct {
	ID       string `json:"id"`
	Command  string `json:"command"`
	Status   string `json:"status"`
	Blocking bool   `json:"blocking"`
}

// LedgerEntry is the compact JSONL source entry for generated goalcli evidence.
type LedgerEntry struct {
	SchemaVersion    string `json:"schema_version"`
	GoalID           string `json:"goal_id"`
	Command          string `json:"command"`
	Status           string `json:"status"`
	MVAStatus        string `json:"mva_status"`
	Blocking         bool   `json:"blocking"`
	EvidencePackPath string `json:"evidence_pack_path"`
}

// Evaluate verifies the local goalcli v0.1.0 MVA contract for a single command.
func Evaluate(command string, options Options) (Report, error) {
	gate, ok := commandGates[command]
	if !ok {
		return Report{}, fmt.Errorf("unsupported goalcli command %q", command)
	}
	goalID := strings.TrimSpace(options.GoalID)
	if goalID == "" {
		goalID = DefaultGoalID
	}
	mode := strings.TrimSpace(options.Mode)
	if mode == "" {
		mode = "FULL"
	}
	root := options.Root
	if root == "" {
		root = "."
	}
	standardSourceRoot := isStandardSourceRoot(root)
	report := Report{
		SchemaVersion:    "goalcli-mva/v1",
		Command:          command,
		Status:           "passed",
		GoalID:           goalID,
		Gate:             gate,
		Mode:             mode,
		Executor:         "goalcli",
		ControlPlane:     "Harness Runtime",
		Blocking:         true,
		MVAStatus:        "complete",
		LedgerPath:       SourceLedgerPath,
		EvidencePackPath: EvidenceLedgerPath + goalID + ".json",
		Gates:            gatesForCommand(command),
		Details: []string{
			gateDescriptions[command],
			"goalcli v0.1.0 使用 cmd/goalcli 作为唯一执行面；不再保留历史并列入口",
			"G12-G16 是 goalcli MVA evidence gates，不是全局 release blocking gates",
			"root plan 是 authority；完成状态由本地 authority 校验和 evidence 写入共同证明",
		},
		Evidence: []string{
			"fixture_id=" + DefaultGoalID,
			"source_evidence_ledger=" + SourceLedgerPath,
			"generated_evidence_pack=" + EvidenceLedgerPath,
			"runtime_logic=internal/goalruntime",
		},
	}
	if goalID != DefaultGoalID {
		report.Details = append(report.Details, "non-default goal_id accepted for local contract replay")
	}
	if !standardSourceRoot {
		report.Details = append(report.Details, "rendered downstream 按 template-generation-contract 排除 source-only authority，仅校验 portable governance authority")
	}
	for _, path := range requiredAuthorityPaths(standardSourceRoot) {
		full := filepath.Join(root, filepath.FromSlash(path))
		if _, err := os.Stat(full); err != nil {
			report.Gaps = append(report.Gaps, "missing authority path: "+path)
			continue
		}
		report.AuthorityPaths = append(report.AuthorityPaths, path)
	}
	if command == finalRuntimeCommand {
		report.Evidence = append(report.Evidence,
			"requires=goal-acceptance",
			"requires=goal-delivery",
			"requires=goal-handover",
			"requires=goal-downstream-adoption",
			"requires=goal-certify",
		)
		if len(report.Gaps) == 0 {
			report.Gaps = append(report.Gaps, validateFinalPrerequisites(root, goalID)...)
		}
	}
	if len(report.Gaps) > 0 {
		report.Status = "failed"
		report.MVAStatus = "not-complete"
	}
	return report, nil
}

func requiredAuthorityPaths(standardSourceRoot bool) []string {
	paths := make([]string, 0, len(sourceOnlyAuthorityPaths)+len(portableAuthorityPaths))
	if standardSourceRoot {
		paths = append(paths, sourceOnlyAuthorityPaths...)
	}
	paths = append(paths, portableAuthorityPaths...)
	return paths
}

func isStandardSourceRoot(root string) bool {
	modulePath, ok := modulePathForRoot(root)
	if !ok {
		return true
	}
	return modulePath == standardModulePath
}

func modulePathForRoot(root string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], true
		}
	}
	return "", false
}

func gatesForCommand(command string) []GateReport {
	commands := []string{command}
	if command == finalRuntimeCommand {
		commands = finalPrerequisiteCommands
	}
	reports := make([]GateReport, 0, len(commands))
	for _, gateCommand := range commands {
		reports = append(reports, GateReport{
			ID:       commandHarnessGates[gateCommand],
			Command:  gateCommand,
			Status:   "passed",
			Blocking: true,
		})
	}
	return reports
}

// WriteEvidence idempotently records passed goalcli MVA reports in the source
// ledger. Final reports also write the generated evidence pack after the
// prerequisite G12-G16 ledger entries have been reconciled for the same goal.
func WriteEvidence(root string, report Report) error {
	if _, ok := commandGates[report.Command]; !ok {
		return fmt.Errorf("evidence write is not supported for command %s", report.Command)
	}
	if report.Status != "passed" || report.MVAStatus != "complete" || !report.Blocking {
		return fmt.Errorf("refuse to write incomplete goalcli evidence: status=%s mva_status=%s blocking=%t", report.Status, report.MVAStatus, report.Blocking)
	}
	if root == "" {
		root = "."
	}
	if report.Command != finalRuntimeCommand {
		return upsertLedgerEntry(filepath.Join(root, filepath.FromSlash(report.LedgerPath)), ledgerEntryForReport(report))
	}
	if gaps := validateFinalPrerequisites(root, report.GoalID); len(gaps) > 0 {
		return fmt.Errorf("refuse to write final goalcli evidence before prerequisites: %s", strings.Join(gaps, "; "))
	}
	packPath := filepath.Join(root, filepath.FromSlash(report.EvidencePackPath))
	if err := os.MkdirAll(filepath.Dir(packPath), 0o755); err != nil {
		return fmt.Errorf("create evidence pack directory: %w", err)
	}
	pack, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evidence pack: %w", err)
	}
	if err := os.WriteFile(packPath, append(pack, '\n'), 0o644); err != nil {
		return fmt.Errorf("write evidence pack %s: %w", report.EvidencePackPath, err)
	}
	return upsertLedgerEntry(filepath.Join(root, filepath.FromSlash(report.LedgerPath)), ledgerEntryForReport(report))
}

func ledgerEntryForReport(report Report) LedgerEntry {
	return LedgerEntry{
		SchemaVersion:    report.SchemaVersion,
		GoalID:           report.GoalID,
		Command:          report.Command,
		Status:           report.Status,
		MVAStatus:        report.MVAStatus,
		Blocking:         report.Blocking,
		EvidencePackPath: report.EvidencePackPath,
	}
}

func validateFinalPrerequisites(root string, goalID string) []string {
	ledgerPath := filepath.Join(root, filepath.FromSlash(SourceLedgerPath))
	entries, err := readLedgerEntries(ledgerPath)
	if err != nil {
		return []string{"missing prerequisite evidence ledger: " + SourceLedgerPath + " (" + err.Error() + ")"}
	}
	byCommand := make(map[string]LedgerEntry, len(entries))
	for _, entry := range entries {
		if entry.GoalID == goalID {
			byCommand[entry.Command] = entry
		}
	}
	var gaps []string
	for _, command := range finalPrerequisiteCommands {
		entry, ok := byCommand[command]
		if !ok {
			gaps = append(gaps, fmt.Sprintf("missing prerequisite evidence for goal_id %s: %s", goalID, command))
			continue
		}
		if entry.Status != "passed" || entry.MVAStatus != "complete" || !entry.Blocking {
			gaps = append(gaps, fmt.Sprintf("incomplete prerequisite evidence for goal_id %s: %s", goalID, command))
		}
	}
	return gaps
}

func readLedgerEntries(path string) ([]LedgerEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte{'\n'})
	entries := make([]LedgerEntry, 0, len(lines))
	for i, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var entry LedgerEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, fmt.Errorf("invalid ledger entry on line %d: %w", i+1, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func upsertLedgerEntry(path string, entry LedgerEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence ledger directory: %w", err)
	}
	var lines [][]byte
	if data, err := os.ReadFile(path); err == nil {
		lines = bytes.Split(data, []byte{'\n'})
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read evidence ledger: %w", err)
	}
	next := make([]byte, 0)
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var existing LedgerEntry
		if err := json.Unmarshal(line, &existing); err == nil &&
			existing.GoalID == entry.GoalID &&
			existing.Command == entry.Command &&
			existing.EvidencePackPath == entry.EvidencePackPath {
			continue
		}
		next = append(next, line...)
		next = append(next, '\n')
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal evidence ledger entry: %w", err)
	}
	next = append(next, data...)
	next = append(next, '\n')
	if err := os.WriteFile(path, next, 0o644); err != nil {
		return fmt.Errorf("write evidence ledger: %w", err)
	}
	return nil
}

// Commands returns the supported goalcli MVA command names.
func Commands() []string {
	return []string{
		"goal-acceptance",
		"goal-delivery",
		"goal-handover",
		"goal-downstream-adoption",
		"goal-certify",
		"goal-runtime-final",
	}
}
