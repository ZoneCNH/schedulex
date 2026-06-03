package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	projectReleaseVersion    = "v0.4.6"
	governanceRuntimeVersion = "v2.9.3"
)

type gateReport struct {
	Command string   `json:"command"`
	Status  string   `json:"status"`
	Details []string `json:"details,omitempty"`
	Gaps    []string `json:"gaps,omitempty"`
}

func emitReport(stdout io.Writer, command, status string, details []string, gaps []string) int {
	report := gateReport{Command: command, Status: status, Details: details, Gaps: gaps}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		write(stdout, "{\"command\":%q,\"status\":%q}\n", command, status)
	} else {
		write(stdout, "%s\n", data)
	}
	if status == "passed" {
		return 0
	}
	return 1
}

func runVersion(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("version", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("version", err, stderr)
	}
	return emitReport(stdout, "version", "passed", []string{"xlib-standard release " + projectReleaseVersion, "goalcli governance runtime " + governanceRuntimeVersion, "goalcli governance CLI available"}, nil)
}

func runDoctor(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("doctor", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("doctor", err, stderr)
	}
	required := []string{
		".agent/harness.yaml",
		".agent/issue-registry.yaml",
		".agent/command-registry.yaml",
		".agent/makefile-target-registry.yaml",
		".agent/makefile-baseline.yaml",
		"docs/standard/goalcli-cli-contract.md",
		"contracts/goalcli-report.schema.json",
		"Makefile",
	}
	if isXlibStandardSourceModule() {
		required = append([]string{"docs/goal/goal.md"}, required...)
	}
	var gaps []string
	for _, path := range required {
		if !fileExists(path) {
			gaps = append(gaps, "missing "+path)
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: doctor found %d gap(s)\n", len(gaps))
		return emitReport(stdout, "doctor", "failed", nil, gaps)
	}
	details := []string{"required governance files are present"}
	details = append(details, hooksStatusDetail())
	return emitReport(stdout, "doctor", "passed", details, nil)
}

// hooksStatusDetail 返回 git hooks 启用状态作为 informational details。
// 不影响 doctor 的 pass/fail：CI 环境无须本地 hooks；本地环境若未启用，
// 提示运行 make install-hooks。对应 .agent/standard/goal-runtime-canonical.md
// 中的 RULE-WORKTREE-001 / RULE-SECRET-001 本地防线。
func hooksStatusDetail() string {
	if !fileExists(".githooks/pre-commit") {
		return "hooks: .githooks/pre-commit 不存在（仓库可能未初始化 hooks 目录）"
	}
	current := strings.TrimSpace(gitOutput("config", "--get", "core.hooksPath"))
	if current == ".githooks" {
		return "hooks: ✅ core.hooksPath=.githooks 已启用"
	}
	if current == "" {
		return "hooks: ⚠️  core.hooksPath 未设置，运行 make install-hooks 启用本地 P0 防线"
	}
	return "hooks: ⚠️  core.hooksPath=" + current + "（非 .githooks），本地 P0 防线未启用"
}

func runMainGuard(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli main-guard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	context := flags.String("context", envDefault("XLIB_CONTEXT", "local_write"), "execution context")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if !validContext(*context) {
		write(stderr, "ERROR: invalid context %q\n", *context)
		return 2
	}
	branch := gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	if *context == "local_write" && (branch == "main" || branch == "master") {
		return emitReport(stdout, "main-guard", "failed", nil, []string{"local_write is forbidden on " + branch})
	}
	return emitReport(stdout, "main-guard", "passed", []string{"context=" + *context, "branch=" + fallback(branch, "unknown")}, nil)
}

func runWorktreeGuard(args []string, stdout io.Writer, stderr io.Writer) int {
	return runWorktreeGate("worktree-guard", args, stdout, stderr)
}

func runWorktreeCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	return runWorktreeGate("worktree-check", args, stdout, stderr)
}

func runWorktreeGate(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	context := flags.String("context", envDefault("XLIB_CONTEXT", "local_write"), "execution context")
	flags.Bool("json", false, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		write(stderr, "ERROR: %s invalid arguments: unexpected positional argument %q\n", command, flags.Arg(0))
		return 2
	}
	if !validContext(*context) {
		write(stderr, "ERROR: invalid context %q\n", *context)
		return 2
	}
	details, gaps := evaluateWorktreeGate(*context)
	if len(gaps) > 0 {
		return emitReport(stdout, command, "failed", details, gaps)
	}
	return emitReport(stdout, command, "passed", details, nil)
}

func evaluateWorktreeGate(context string) ([]string, []string) {
	top := gitOutput("rev-parse", "--show-toplevel")
	common := gitOutput("rev-parse", "--path-format=absolute", "--git-common-dir")
	isWorkerTree := strings.Contains(top, string(filepath.Separator)+".worktree"+string(filepath.Separator)) || strings.Contains(top, string(filepath.Separator)+".worktrees"+string(filepath.Separator)) || strings.Contains(common, string(filepath.Separator)+"worktrees"+string(filepath.Separator))
	details := []string{"context=" + context, "top=" + fallback(top, "unknown")}
	if context == "local_write" && !isWorkerTree {
		return details, []string{"local_write requires a worker worktree"}
	}
	return details, nil
}

func runContextCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("context-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("context-check", err, stderr)
	}
	required := []string{"docs/goal", "docs/goal/goal.md"}
	var gaps []string
	for _, path := range required {
		if !fileExists(path) {
			gaps = append(gaps, "missing "+path)
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: context-check found %d gap(s)\n", len(gaps))
		return emitReport(stdout, "context-check", "failed", nil, gaps)
	}
	return emitReport(stdout, "context-check", "passed", []string{"docs/goal context is present"}, nil)
}

func runSpecCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("spec-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("spec-check", err, stderr)
	}
	if !fileExists("docs") {
		write(stderr, "ERROR: spec-check found 1 gap(s)\n")
		return emitReport(stdout, "spec-check", "failed", nil, []string{"missing docs"})
	}
	found := false
	var gaps []string
	paths, err := trackedDocsMarkdownFiles()
	if err != nil {
		gaps = append(gaps, "scan docs: "+err.Error())
	}
	for _, path := range paths {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			gaps = append(gaps, "read "+path+": "+readErr.Error())
			continue
		}
		if strings.Contains(string(data), "REQ-") {
			found = true
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: spec-check found %d gap(s)\n", len(gaps))
		return emitReport(stdout, "spec-check", "failed", nil, gaps)
	}
	details := []string{fmt.Sprintf("scanned_markdown=%d", len(paths))}
	if !found {
		details = append(details, "warning: no docs markdown file contains REQ-")
	}
	return emitReport(stdout, "spec-check", "passed", details, nil)
}

func trackedDocsMarkdownFiles() ([]string, error) {
	out, err := exec.Command("git", "ls-files", "-z", "--", "docs").Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, path := range strings.Split(string(out), "\x00") {
		if path == "" || filepath.Ext(path) != ".md" {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func runDesignCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("design-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("design-check", err, stderr)
	}
	if !fileExists("docs/adr") {
		return emitReport(stdout, "design-check", "passed", []string{"warning: optional docs/adr not present"}, nil)
	}
	return emitReport(stdout, "design-check", "passed", []string{"docs/adr is present"}, nil)
}

func runTaskCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("task-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("task-check", err, stderr)
	}
	switch {
	case fileExists(".agent/command-registry.yaml"):
		return emitReport(stdout, "task-check", "passed", []string{".agent/command-registry.yaml is present"}, nil)
	case fileExists(".agent/registry/commands.yaml"):
		return emitReport(stdout, "task-check", "passed", []string{"legacy .agent/registry/commands.yaml is present"}, nil)
	default:
		return emitReport(stdout, "task-check", "passed", []string{"warning: command registry not present"}, nil)
	}
}

func runPRCheck(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli pr-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	context := flags.String("context", envDefault("XLIB_CONTEXT", "local_write"), "execution context")
	dryRun := flags.Bool("dry-run", false, "")
	flags.Bool("json", false, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		write(stderr, "ERROR: pr-check invalid arguments: unexpected positional argument %q\n", flags.Arg(0))
		return 2
	}
	if !validContext(*context) {
		write(stderr, "ERROR: invalid context %q\n", *context)
		return 2
	}
	if *dryRun {
		return emitReport(stdout, "pr-check", "passed", []string{"mode=dry-run", "context=" + *context, "delegates=worktree-check,lint,test"}, nil)
	}
	if details, gaps := evaluateWorktreeGate(*context); len(gaps) > 0 {
		return emitReport(stdout, "pr-check", "failed", details, gaps)
	}
	if code := runExternal(stdin, stderr, stderr, "make", "lint"); code != 0 {
		return emitReport(stdout, "pr-check", "failed", nil, []string{fmt.Sprintf("make lint exited %d", code)})
	}
	if code := runExternal(stdin, stderr, stderr, "make", "test"); code != 0 {
		return emitReport(stdout, "pr-check", "failed", nil, []string{fmt.Sprintf("make test exited %d", code)})
	}
	return emitReport(stdout, "pr-check", "passed", []string{"context=" + *context, "make lint passed", "make test passed"}, nil)
}

func runRegistryCheck(command string, required map[string][]string, stdout io.Writer, stderr io.Writer) int {
	var gaps []string
	for path, needles := range required {
		content, err := os.ReadFile(path)
		if err != nil {
			gaps = append(gaps, "missing "+path)
			continue
		}
		text := string(content)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				gaps = append(gaps, path+" missing "+needle)
			}
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", nil, gaps)
	}
	return emitReport(stdout, command, "passed", []string{"registry contract satisfied"}, nil)
}

func runEvidenceCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("evidence-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("evidence-check", err, stderr)
	}
	return runRegistryCheck("evidence-check", map[string][]string{
		".agent/done-assertion.yaml":               {"DONE with evidence", "commit", "gates"},
		".agent/evidence-artifact-policy.yaml":     {"redaction", "sha256", "release/manifest/latest.json"},
		".agent/harness.yaml":                      {"manifest", "checksum", "required_fields"},
		".agent/evidence-artifacts.yaml":           {"release_evidence", "execution_evidence", "schema:", "contracts/execution-evidence.schema.json"},
		"contracts/execution-evidence.schema.json": {"evidence_id", "stdout_sha256", "commit", "exit_code", "artifact_path"},
	}, stdout, stderr)
}

func runCLIContract(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("cli-contract", args, internalCommandFlagSpec{boolFlags: []string{"json", "explain"}, stringFlags: []string{"output"}}); err != nil {
		return invalidInternalArgsExit("cli-contract", err, stderr)
	}
	return runRegistryCheck("cli-contract", map[string][]string{
		"docs/standard/goalcli-cli-contract.md": commandRegistryRequiredCommands(),
		"contracts/goalcli-report.schema.json":  {"command", "status", "details", "gaps"},
		".agent/command-registry.yaml":          requiredCommandRegistryNeedles(),
	}, stdout, stderr)
}

func runIssueRegistry(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("issue-registry", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("issue-registry", err, stderr)
	}
	var gaps []string
	appendIssueRegistryGaps(".agent/issue-registry.yaml", &gaps)
	if len(gaps) > 0 {
		write(stderr, "ERROR: issue-registry found %d gap(s)\n", len(gaps))
		return emitReport(stdout, "issue-registry", "failed", nil, gaps)
	}
	return emitReport(stdout, "issue-registry", "passed", []string{"issue registry entries are implemented, unique, and contiguous"}, nil)
}

func runCommandRegistry(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("command-registry", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("command-registry", err, stderr)
	}
	return runRegistryCheck("command-registry", map[string][]string{
		".agent/command-registry.yaml": requiredCommandRegistryNeedles(),
	}, stdout, stderr)
}

func runMakefileBaseline(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("makefile-baseline", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("makefile-baseline", err, stderr)
	}
	requiredTargets := append([]string{"fmt", "vet", "lint", "test", "race", "boundary", "security", "contracts", "docs-check", "rules-verify", "evidence", "score-check", "main-guard", "worktree-guard", "worktree-check", "context-check", "spec-check", "design-check", "task-check", "pr-check", "evidence-check", "cli-contract", "issue-registry", "command-registry", "makefile-baseline", "governance-check", "p1-governance-check", "execution-context", "p2-runtime-check", "release-check", "release-final-check"}, contextRuntimeTargets()...)
	requiredTargets = append(requiredTargets, goalcliMakefileTargets()...)
	required := map[string][]string{"Makefile": {}, ".agent/makefile-target-registry.yaml": requiredTargets, ".agent/makefile-baseline.yaml": requiredTargets}
	for _, target := range requiredTargets {
		required["Makefile"] = append(required["Makefile"], ".PHONY: "+target, target+":")
	}
	return runRegistryCheck("makefile-baseline", required, stdout, stderr)
}

func goalcliMakefileTargets() []string {
	return []string{
		"goal-acceptance",
		"goal-delivery",
		"goal-handover",
		"goal-downstream-adoption",
		"goal-certify",
		"goal-runtime-final",
	}
}

var contextProfileGates = map[string][]string{
	"lite":     {"governance-check"},
	"standard": {"governance-check", "p1-governance-check", "docs-check"},
	"full":     {"governance-check", "p1-governance-check", "p2-runtime-check"},
	"release":  {"context-full", "integration", "dependency-check", "standard-impact-check", "score-check", "debt-evidence", "evidence", "release-evidence-hash", "release-evidence-check", "release-evidence-checksum-check"},
}

func runContextProfile(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli context-profile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profile := flags.String("profile", "standard", "context runtime profile")
	flags.Bool("json", false, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		write(stderr, "ERROR: context-profile invalid arguments: unexpected positional argument %q\n", flags.Arg(0))
		return 2
	}
	normalized := normalizeContextProfile(*profile)
	gates, ok := contextProfileGates[normalized]
	if !ok {
		write(stderr, "ERROR: invalid context profile %q\n", *profile)
		return 2
	}
	return emitReport(stdout, "context-profile", "passed", contextProfileDetails(normalized, gates), nil)
}

func runContextProfileAlias(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs(command, args, internalCommandFlagSpec{boolFlags: []string{"json", "strict"}}); err != nil {
		return invalidInternalArgsExit(command, err, stderr)
	}
	profile := mapContextAliasToProfile(command)
	gates := contextProfileGates[profile]
	return emitReport(stdout, command, "passed", contextProfileDetails(profile, gates), nil)
}

func runContextProfileCheck(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	profile, err := parseContextProfileCheckProfile(command, args)
	if err != nil {
		return invalidInternalArgsExit(command, err, stderr)
	}
	if profile != "" {
		normalized := normalizeContextProfile(profile)
		if _, ok := contextProfileGates[normalized]; !ok {
			write(stderr, "ERROR: invalid context profile %q\n", profile)
			return 2
		}
	}
	contextTargets := contextRuntimeTargets()
	required := map[string][]string{
		".agent/command-registry.yaml":          requiredCommandRegistryNeedles(),
		".agent/makefile-target-registry.yaml":  contextTargets,
		".agent/makefile-baseline.yaml":         contextTargets,
		"docs/standard/goalcli-cli-contract.md": commandRegistryRequiredCommands(),
		"Makefile":                              {"release-final-check:", "$(MAKE) context-release"},
	}
	for _, target := range contextTargets {
		required["Makefile"] = append(required["Makefile"], ".PHONY: "+target, target+":")
	}
	var gaps []string
	for path, needles := range required {
		content, err := os.ReadFile(path)
		if err != nil {
			gaps = append(gaps, "missing "+path)
			continue
		}
		text := string(content)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				gaps = append(gaps, path+" missing "+needle)
			}
		}
	}
	appendIssueRegistryGaps(".agent/issue-registry.yaml", &gaps)
	if makefile, err := os.ReadFile("Makefile"); err == nil {
		makefileText := string(makefile)
		appendMakefileDuplicateGaps(makefileText, contextTargets, &gaps)
		appendContextProfileContractGaps(makefileText, &gaps)
		appendMakefileTargetDependencyGaps(makefileText, "context-lite", []string{"require-gowork-off", "governance-check"}, []string{"context-profile-check", "main-guard", "worktree-guard", "release-check", "release-final-check"}, &gaps)
		appendMakefileTargetDependencyGaps(makefileText, "context-standard", []string{"require-gowork-off", "governance-check", "p1-governance-check", "docs-check"}, []string{"context-lite", "context-profile-check", "release-check", "release-final-check"}, &gaps)
		appendMakefileTargetDependencyGaps(makefileText, "context-full", []string{"require-gowork-off", "governance-check", "p1-governance-check", "p2-runtime-check"}, []string{"context-standard", "docs-check", "context-profile-check", "release-check", "release-final-check"}, &gaps)
		appendMakefileTargetDependencyGaps(makefileText, "context-release", []string{"require-gowork-off", "context-full", "integration", "dependency-check", "standard-impact-check", "score-check", "debt-evidence"}, []string{"context-standard", "release-check", "release-final-check"}, &gaps)
		appendMakefileTargetForbiddenReferenceGaps(makefileText, "context-release", []string{"release-check", "release-final-check"}, &gaps)
		appendContextProfileDAGGaps(makefileText, &gaps)
		appendReleaseFinalDelegationGaps(makefileText, &gaps)
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", nil, gaps)
	}
	return emitReport(stdout, command, "passed", []string{"context runtime v4.0 profile DAG and registry contract satisfied", ".agent/context not required or claimed", "context profiles reject unknown gates", "context profile Makefile dependencies parse continuations", "context-release excludes release-check and release-final-check", "release-final-check delegates to context-release without self-recursion"}, nil)
}

func parseContextProfileCheckProfile(command string, args []string) (string, error) {
	flags := flag.NewFlagSet("goalcli "+command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Bool("json", false, "")
	flags.Bool("strict", false, "")
	profile := flags.String("profile", "", "")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	if flags.NArg() > 0 {
		return "", fmt.Errorf("unexpected positional argument %q", flags.Arg(0))
	}
	return *profile, nil
}

func contextRuntimeTargets() []string {
	return []string{
		"context-profile",
		"context-profile-check",
		"context-schema-check",
		"context-lite",
		"context-standard",
		"context-full",
		"context-release",
		"context-fast-check",
		"context-standard-check",
		"context-full-check",
	}
}

func contextProfileDetails(profile string, gates []string) []string {
	return []string{
		"context_runtime=v4.0",
		"profile=" + profile,
		"gates=" + strings.Join(gates, ","),
		"legacy_aliases=context-fast-check,context-standard-check,context-full-check",
		"release_final_delegates=context-release",
	}
}

func normalizeContextProfile(profile string) string {
	switch profile {
	case "fast":
		return "lite"
	default:
		return profile
	}
}

func mapContextAliasToProfile(command string) string {
	switch command {
	case "context-lite", "context-fast-check":
		return "lite"
	case "context-full", "context-full-check":
		return "full"
	case "context-release":
		return "release"
	default:
		return "standard"
	}
}

func makefileTargetBlock(content, target string) string {
	lines := strings.Split(content, "\n")
	var block []string
	inBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, target+":") {
			inBlock = true
			block = append(block, line)
			continue
		}
		if inBlock {
			if line != "" && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") && strings.Contains(line, ":") {
				break
			}
			block = append(block, line)
		}
	}
	return strings.Join(block, "\n")
}

func appendMakefileDuplicateGaps(content string, targets []string, gaps *[]string) {
	for _, target := range targets {
		if count := makefileTargetDefinitionCount(content, target); count != 1 {
			*gaps = append(*gaps, fmt.Sprintf("Makefile target %s must be defined exactly once, found %d", target, count))
		}
	}
}

func makefileTargetDefinitionCount(content, target string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, target+":") {
			count++
		}
	}
	return count
}

func makefileTargetNames(content string) map[string]bool {
	targets := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "#") || strings.Contains(line, ":=") {
			continue
		}
		header := strings.SplitN(line, ":", 2)[0]
		for _, target := range strings.Fields(header) {
			if target != ".PHONY" {
				targets[target] = true
			}
		}
	}
	return targets
}

func appendContextProfileContractGaps(makefileText string, gaps *[]string) {
	makefileTargets := makefileTargetNames(makefileText)
	for profile, gates := range contextProfileGates {
		if profile == "" {
			*gaps = append(*gaps, "context profile name must not be empty")
		}
		if !validContextProfileName(profile) {
			*gaps = append(*gaps, "unknown context profile "+profile)
		}
		for _, gate := range gates {
			if gate == "release-check" || gate == "release-final-check" {
				*gaps = append(*gaps, "context profile "+profile+" must not include "+gate)
			}
			if !makefileTargets[gate] {
				*gaps = append(*gaps, "context profile "+profile+" references unknown Makefile gate "+gate)
			}
		}
	}
	appendContextProfileCycleGaps(gaps)
}

func validContextProfileName(profile string) bool {
	switch profile {
	case "lite", "standard", "full", "release":
		return true
	default:
		return false
	}
}

func appendContextProfileCycleGaps(gaps *[]string) {
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(profile string, path []string)
	visit = func(profile string, path []string) {
		if visiting[profile] {
			*gaps = append(*gaps, "context profile DAG cycle: "+strings.Join(append(path, profile), " -> "))
			return
		}
		if visited[profile] {
			return
		}
		visiting[profile] = true
		for _, gate := range contextProfileGates[profile] {
			if next, ok := contextGateProfile(gate); ok {
				visit(next, append(path, profile))
			}
		}
		visiting[profile] = false
		visited[profile] = true
	}
	for profile := range contextProfileGates {
		visit(profile, nil)
	}
}

func appendContextProfileDAGGaps(content string, gaps *[]string) {
	profileTargets := []string{"context-lite", "context-standard", "context-full", "context-release"}
	profileTargetSet := map[string]bool{}
	for _, target := range profileTargets {
		profileTargetSet[target] = true
	}
	allowedLeaf := map[string]bool{
		"require-gowork-off":              true,
		"governance-check":                true,
		"p1-governance-check":             true,
		"docs-check":                      true,
		"p2-runtime-check":                true,
		"integration":                     true,
		"dependency-check":                true,
		"standard-impact-check":           true,
		"score-check":                     true,
		"debt-evidence":                   true,
		"evidence":                        true,
		"release-evidence-hash":           true,
		"release-evidence-check":          true,
		"release-evidence-checksum-check": true,
		"context-profile-check":           true,
		"context-schema-check":            true,
		"context-profile":                 true,
		"context-fast-check":              true,
		"context-standard-check":          true,
		"context-full-check":              true,
	}
	graph := map[string][]string{}
	for _, target := range profileTargets {
		for _, dep := range makefileTargetDependencies(content, target) {
			switch {
			case profileTargetSet[dep] || dep == "release-final-check":
				graph[target] = append(graph[target], dep)
			case allowedLeaf[dep]:
				continue
			default:
				*gaps = append(*gaps, "Makefile "+target+" references unknown context gate "+dep)
			}
		}
	}
	appendMakefileProfileCycleGaps(graph, profileTargets, gaps)
	if makefileGraphReaches(graph, "context-release", "release-final-check", map[string]bool{}) {
		*gaps = append(*gaps, "Makefile context-release must not reach release-final-check")
	}
}

func makefileTargetDependencies(content, target string) []string {
	block := makefileTargetBlock(content, target)
	if block == "" {
		return nil
	}
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return nil
	}
	var dependencyLines []string
	for i, line := range lines {
		if i == 0 {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return nil
			}
			line = parts[1]
		} else {
			if strings.HasPrefix(line, "\t") {
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
		}
		trimmed := strings.TrimSpace(line)
		continued := strings.HasSuffix(trimmed, "\\")
		trimmed = strings.TrimSpace(strings.TrimSuffix(trimmed, "\\"))
		if trimmed != "" {
			dependencyLines = append(dependencyLines, trimmed)
		}
		if !continued {
			break
		}
	}
	return strings.Fields(strings.Join(dependencyLines, " "))
}

func appendMakefileProfileCycleGaps(graph map[string][]string, roots []string, gaps *[]string) {
	visiting := map[string]bool{}
	visited := map[string]bool{}
	reported := map[string]bool{}
	var visit func(target string, path []string)
	visit = func(target string, path []string) {
		if visiting[target] {
			cycle := append(path, target)
			key := strings.Join(cycle, " -> ")
			if !reported[key] {
				*gaps = append(*gaps, "Makefile context profile DAG cycle: "+key)
				reported[key] = true
			}
			return
		}
		if visited[target] {
			return
		}
		visiting[target] = true
		for _, dep := range graph[target] {
			if strings.HasPrefix(dep, "context-") {
				visit(dep, append(path, target))
			}
		}
		visiting[target] = false
		visited[target] = true
	}
	for _, root := range roots {
		visit(root, nil)
	}
}

func makefileGraphReaches(graph map[string][]string, start, target string, seen map[string]bool) bool {
	if start == target {
		return true
	}
	if seen[start] {
		return false
	}
	seen[start] = true
	for _, dep := range graph[start] {
		if makefileGraphReaches(graph, dep, target, seen) {
			return true
		}
	}
	return false
}

func contextGateProfile(gate string) (string, bool) {
	switch gate {
	case "context-lite", "context-fast-check":
		return "lite", true
	case "context-standard", "context-standard-check":
		return "standard", true
	case "context-full", "context-full-check":
		return "full", true
	case "context-release":
		return "release", true
	default:
		return "", false
	}
}

func appendMakefileTargetDependencyGaps(content, target string, required []string, forbidden []string, gaps *[]string) {
	block := makefileTargetBlock(content, target)
	if block == "" {
		*gaps = append(*gaps, "Makefile missing target block "+target)
		return
	}
	dependencies := makefileTargetDependencies(content, target)
	for _, token := range required {
		if !makefileDependencyHasToken(dependencies, token) {
			*gaps = append(*gaps, "Makefile "+target+" missing dependency "+token)
		}
	}
	for _, token := range forbidden {
		if makefileDependencyHasToken(dependencies, token) {
			*gaps = append(*gaps, "Makefile "+target+" must not depend on "+token)
		}
	}
}

func appendMakefileTargetForbiddenReferenceGaps(content, target string, forbidden []string, gaps *[]string) {
	block := makefileTargetBlock(content, target)
	if block == "" {
		*gaps = append(*gaps, "Makefile missing target block "+target)
		return
	}
	for _, token := range forbidden {
		if strings.Contains(block, token) {
			*gaps = append(*gaps, "Makefile "+target+" must not reference "+token)
		}
	}
}

func appendReleaseFinalDelegationGaps(content string, gaps *[]string) {
	block := makefileTargetBlock(content, "release-final-check")
	if block == "" {
		*gaps = append(*gaps, "Makefile missing target block release-final-check")
		return
	}
	if makefileDependencyHasToken(makefileTargetDependencies(content, "release-final-check"), "release-final-check") || strings.Contains(block, "$(MAKE) release-final-check") || strings.Contains(block, "make release-final-check") || strings.Contains(block, "$(GOALCLI) release-final-check") {
		*gaps = append(*gaps, "release-final-check must not call itself")
	}
	if !strings.Contains(block, "$(MAKE) context-release") {
		*gaps = append(*gaps, "release-final-check must call context-release")
	}
}

func makefileDependencyHasToken(dependencies []string, token string) bool {
	for _, field := range dependencies {
		if field == token {
			return true
		}
	}
	return false
}

var plannedCommandFiles = map[string][]string{
	"minimal-kernel":           {".agent/minimal-kernel.yaml"},
	"done-assertion":           {".agent/done-assertion.yaml"},
	"agent-team-contract":      {".agent/team-contract.yaml"},
	"scope-lock":               {".agent/scope-locks.yaml"},
	"pr-template":              {".agent/pr-template-contract.yaml", ".github/pull_request_template.md"},
	"acceptance-matrix":        {".agent/acceptance-matrix.yaml"},
	"runtime-health":           {".agent/runtime-health.yaml"},
	"goal-runtime":             {".agent/goal-runtime.md", ".agent/harness.yaml"},
	"goal-acceptance":          {".agent/harness.yaml"},
	"goal-delivery":            {".agent/harness.yaml"},
	"goal-handover":            {".agent/harness.yaml"},
	"goal-downstream-adoption": {".agent/harness.yaml"},
	"goal-certify":             {".agent/harness.yaml"},
	"goal-runtime-final":       {".agent/harness.yaml"},
	"naming":                   {"docs/standard/repository-roles.md", "docs/standard/module-boundary.md"},
	"upgrade-standard":         {".agent/downstream-registry.yaml"},
	"conformance-profile":      {".agent/conformance-profiles.yaml"},
	"downstream-registry":      {".agent/downstream-registry.yaml"},
	"self-healing-skeleton":    {".agent/failure-taxonomy.yaml", ".agent/root-cause.yaml", ".agent/regression-memory.yaml"},
	"policy-schema":            {".agent/policy-schema.yaml"},
	"github-settings":          {".agent/github-settings.yaml"},
	"github-governance":        {".agent/github-governance.yaml"},
	"governance-fixture-test":  {".agent/governance-fixture-test.yaml"},
	"toolchain":                {".agent/toolchain.yaml"},
	"evidence-artifacts":       {".agent/evidence-artifact-policy.yaml"},
	"install-runtime":          {".agent/runtime-install.yaml"},
	"upgrade-runtime":          {".agent/runtime-upgrade.yaml"},
	"release-ready":            {".agent/release-readiness-formula.yaml"},
	"evidence-replay":          {".agent/evidence-replay.yaml"},
	"attest-conformance":       {".agent/conformance-profiles.yaml"},
	"pack-standard":            {".agent/standard-pack.yaml"},
	"pack-gate":                {".agent/gate-pack.yaml"},
	"pack-evidence":            {".agent/evidence-pack.yaml"},
	"runtime-file-ownership":   {".agent/runtime-file-ownership.yaml"},
	"downstream-baseline":      {".agent/downstream-baseline-scan.yaml", ".agent/downstream-registry.yaml"},
	"downstream-adoption":      {".agent/downstream-adoption-modes.yaml", ".agent/downstream-registry.yaml"},
	"autoresearch":             {".agent/autoresearch.yaml"},
	"changelog":                {".agent/changelog.yaml"},
	"supply-chain":             {"docs/supply-chain.md"},
	"execution-context":        {".agent/execution-context.yaml", "contracts/execution-context.schema.json"},
}

var plannedCommandSemanticMarkers = map[string]map[string][]string{
	"agent-team-contract": {
		".agent/team-contract.yaml": {"schema_version:", "roles:", "rule:"},
	},
	"acceptance-matrix": {
		".agent/acceptance-matrix.yaml": {"schema_version:", "acceptance:"},
	},
	"runtime-health": {
		".agent/runtime-health.yaml": {"schema_version:", "checks:", "toolchain"},
	},
	"goal-acceptance": {
		".agent/harness.yaml": {"goalcli_mva_gates:", "G12_ACCEPTANCE", "goal-acceptance"},
	},
	"goal-delivery": {
		".agent/harness.yaml": {"goalcli_mva_gates:", "G13_DELIVERY", "goal-delivery"},
	},
	"goal-handover": {
		".agent/harness.yaml": {"goalcli_mva_gates:", "G14_HANDOVER", "goal-handover"},
	},
	"goal-downstream-adoption": {
		".agent/harness.yaml": {"goalcli_mva_gates:", "G15_DOWNSTREAM_ADOPTION", "goal-downstream-adoption"},
	},
	"goal-certify": {
		".agent/harness.yaml": {"goalcli_mva_gates:", "G16_CERTIFY", "goal-certify"},
	},
	"goal-runtime-final": {
		".agent/harness.yaml": {"goalcli_mva_gates:", "G12_G16_FINAL", "goal-runtime-final"},
	},
	"execution-context": {
		".agent/execution-context.yaml": {"schema_version:", "contexts:", "local_write", "ci_pull_request", "release_verify"},
	},
	"runtime-file-ownership": {
		".agent/runtime-file-ownership.yaml": {"schema_version:", "owners:", "owner:", "review_required:", "rationale:"},
	},
}

func runPlannedCommand(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validatePlannedCommandArgs(command, args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		write(stderr, "ERROR: %s invalid arguments: %v\n", command, err)
		return 2
	}

	var details []string
	var gaps []string
	files, ok := plannedCommandFiles[command]
	if !ok || len(files) == 0 {
		write(stderr, "ERROR: %s has no manifest coverage\n", command)
		return emitReport(stdout, command, "failed", []string{"args=" + strings.Join(args, " ")}, []string{"planned command has no manifest coverage: " + command})
	}
	for _, path := range files {
		content, gap, ok := readPlannedCommandFile(path)
		if !ok {
			gaps = append(gaps, gap)
			continue
		}
		details = append(details, "found "+path)
		gaps = append(gaps, validatePlannedCommandFile(command, path, content)...)
	}
	if command == "downstream-baseline" || command == "downstream-adoption" || command == "upgrade-standard" {
		mode := fallback(flagValue(args, "mode", ""), "patch-only")
		if flagProvided(args, "repo") {
			repo := flagValue(args, "repo", "")
			details = append(details, "repo="+repo, "mode="+mode, "dry_run=true")
			if !fileExists(repo) {
				gaps = append(gaps, "downstream repo unavailable in worker workspace: "+repo)
				return emitPlannedReport(stdout, stderr, command, "gap", details, gaps, args)
			}
		} else {
			details = append(details, "repo=manifest-only", "mode="+mode, "dry_run=true")
		}
	}
	if len(gaps) > 0 {
		write(stderr, "ERROR: %s found %d gap(s)\n", command, len(gaps))
		return emitReport(stdout, command, "failed", details, gaps)
	}
	details = append(details, "args="+strings.Join(args, " "))
	if plannedCommandVerifyRequested(args) {
		details = append(details, "local dry-run verifier satisfied manifest coverage")
	}
	return emitReport(stdout, command, "passed", details, nil)
}

func readPlannedCommandFile(path string) ([]byte, string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "missing " + path, false
	}
	if info.IsDir() {
		return nil, path + " must be a file", false
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, path + " unreadable: " + err.Error(), false
	}
	return content, "", true
}

func validatePlannedCommandFile(command string, path string, content []byte) []string {
	var gaps []string
	text := string(content)
	if strings.TrimSpace(text) == "" {
		gaps = append(gaps, path+" must not be empty")
	}
	if filepath.Ext(path) == ".json" && !json.Valid(content) {
		gaps = append(gaps, path+" must be valid JSON")
	}
	for _, marker := range plannedCommandMarkers(command, path) {
		if !strings.Contains(text, marker) {
			gaps = append(gaps, path+" missing semantic marker "+marker)
		}
	}
	return gaps
}

func plannedCommandMarkers(command string, path string) []string {
	files, ok := plannedCommandSemanticMarkers[command]
	if !ok {
		return nil
	}
	return files[path]
}

func flagProvided(args []string, name string) bool {
	for _, arg := range args {
		if arg == "--"+name || strings.HasPrefix(arg, "--"+name+"=") {
			return true
		}
	}
	return false
}

func validatePlannedCommandArgs(command string, args []string) error {
	flags := flag.NewFlagSet("goalcli "+command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Bool("dry-run", false, "")
	flags.Bool("verify", false, "")
	flags.Bool("strict", false, "")
	flags.Bool("json", false, "")
	flags.String("repo", "", "")
	flags.String("mode", "", "")
	context := flags.String("context", "", "")
	flags.String("profile", "", "")
	flags.String("output", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected positional argument %q", flags.Arg(0))
	}
	if *context != "" && !validContext(*context) {
		return fmt.Errorf("invalid context %q", *context)
	}
	return nil
}

type internalCommandFlagSpec struct {
	boolFlags   []string
	stringFlags []string
}

func validateInternalCommandArgs(command string, args []string, spec internalCommandFlagSpec) error {
	flags := flag.NewFlagSet("goalcli "+command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	for _, name := range spec.boolFlags {
		flags.Bool(name, false, "")
	}
	for _, name := range spec.stringFlags {
		flags.String(name, "", "")
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected positional argument %q", flags.Arg(0))
	}
	return nil
}

func invalidInternalArgsExit(command string, err error, stderr io.Writer) int {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	write(stderr, "ERROR: %s invalid arguments: %v\n", command, err)
	return 2
}

func isXlibStandardSourceModule() bool {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return true
	}
	sourceModule := strings.Join([]string{"github.com", "ZoneCNH", "xlib" + "-standard"}, "/")
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1] == sourceModule
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func envDefault(name, fallbackValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallbackValue
}

func fallback(value, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}
	return value
}

func validContext(value string) bool {
	for _, context := range validExecutionContexts {
		if value == context {
			return true
		}
	}
	return false
}

var validExecutionContexts = []string{"local_write", "local_readonly", "ci_pull_request", "ci_main_verify", "release_verify"}

var commandRegistryCommands = []string{
	"version",
	"doctor",
	"minimal-kernel",
	"main-guard",
	"worktree-guard",
	"worktree-check",
	"context-check",
	"spec-check",
	"design-check",
	"task-check",
	"pr-check",
	"evidence-check",
	"done-assertion",
	"cli-contract",
	"issue-registry",
	"command-registry",
	"makefile-baseline",
	"context-profile",
	"context-profile-check",
	"context-schema-check",
	"context-lite",
	"context-standard",
	"context-full",
	"context-release",
	"context-fast-check",
	"context-standard-check",
	"context-full-check",
	"agent-team-contract",
	"scope-lock",
	"pr-template",
	"acceptance-matrix",
	"runtime-health",
	"goal-runtime",
	"goal-acceptance",
	"goal-delivery",
	"goal-handover",
	"goal-downstream-adoption",
	"goal-certify",
	"goal-runtime-final",
	"naming",
	"upgrade-standard",
	"conformance-profile",
	"downstream-registry",
	"self-healing-skeleton",
	"policy-schema",
	"github-settings",
	"github-governance",
	"governance-fixture-test",
	"toolchain",
	"evidence-artifacts",
	"install-runtime",
	"upgrade-runtime",
	"release-ready",
	"evidence-replay",
	"attest-conformance",
	"pack-standard",
	"pack-gate",
	"pack-evidence",
	"runtime-file-ownership",
	"downstream-baseline",
	"downstream-adoption",
	"autoresearch",
	"changelog",
	"supply-chain",
	"execution-context",
	"boundary",
	"contracts",
	"dependency-check",
	"docs-check",
	"evidence",
	"manifest",
	"integration",
	"release-evidence-check",
	"release-evidence-checksum-check",
	"release-evidence-hash",
	"release-final-check",
	"render-check",
	"rules-verify",
	"score",
	"secrets",
	"security",
	"standard-impact-check",
}

func commandRegistryRequiredCommands() []string {
	return append([]string(nil), commandRegistryCommands...)
}

func requiredCommandRegistryNeedles() []string {
	needles := make([]string, 0, len(commandRegistryCommands))
	for _, command := range commandRegistryCommands {
		needles = append(needles, "name: "+command)
	}
	return needles
}

type issueRegistryEntry struct {
	id    string
	block string
}

var issueRegistryIDPattern = regexp.MustCompile(`^(P0|P1|P2|CTX)-([0-9]{3})$`)

func appendIssueRegistryGaps(path string, gaps *[]string) {
	content, err := os.ReadFile(path)
	if err != nil {
		*gaps = append(*gaps, "missing "+path)
		return
	}
	*gaps = append(*gaps, validateIssueRegistryEntries(path, parseIssueRegistryEntries(string(content)))...)
}

func parseIssueRegistryEntries(text string) []issueRegistryEntry {
	var entries []issueRegistryEntry
	var currentID string
	var currentLines []string
	flush := func() {
		if currentID != "" {
			entries = append(entries, issueRegistryEntry{id: currentID, block: strings.Join(currentLines, "\n")})
		}
	}
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- id:") {
			flush()
			currentID = trimYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- id:")))
			currentLines = []string{line}
			continue
		}
		if currentID != "" {
			currentLines = append(currentLines, line)
		}
	}
	flush()
	return entries
}

func validateIssueRegistryEntries(path string, entries []issueRegistryEntry) []string {
	if len(entries) == 0 {
		return []string{path + " must contain issue entries"}
	}
	var gaps []string
	seen := map[string]bool{}
	numsByPrefix := map[string][]int{
		"P0":  {},
		"P1":  {},
		"P2":  {},
		"CTX": {},
	}
	for _, entry := range entries {
		if seen[entry.id] {
			gaps = append(gaps, path+" duplicate issue id "+entry.id)
		}
		seen[entry.id] = true
		match := issueRegistryIDPattern.FindStringSubmatch(entry.id)
		if match == nil {
			gaps = append(gaps, path+" invalid issue id "+entry.id)
			continue
		}
		num, _ := strconv.Atoi(match[2])
		numsByPrefix[match[1]] = append(numsByPrefix[match[1]], num)
		if !blockHasNonEmptyYAMLValue(entry.block, "title") {
			gaps = append(gaps, path+" "+entry.id+" missing title")
		}
		status, ok := blockYAMLValue(entry.block, "status")
		if !ok || status != "implemented" {
			gaps = append(gaps, path+" "+entry.id+" status must be implemented")
		}
		if !blockHasNonEmptyYAMLValue(entry.block, "command") {
			gaps = append(gaps, path+" "+entry.id+" missing command")
		}
		if !blockHasEvidence(entry.block) {
			gaps = append(gaps, path+" "+entry.id+" missing evidence")
		}
	}
	for _, prefix := range []string{"P0", "P1", "P2", "CTX"} {
		nums := numsByPrefix[prefix]
		if len(nums) == 0 {
			gaps = append(gaps, path+" missing "+prefix+"-001")
			continue
		}
		sort.Ints(nums)
		last := 0
		for _, num := range nums {
			if num == last {
				continue
			}
			if num != last+1 {
				gaps = append(gaps, fmt.Sprintf("%s %s ids must be contiguous; missing %s-%03d", path, prefix, prefix, last+1))
				break
			}
			last = num
		}
	}
	return gaps
}

func blockHasEvidence(block string) bool {
	value, ok := blockYAMLValue(block, "evidence")
	if ok && value != "" && value != "[]" {
		return true
	}
	return blockHasYAMLListItem(block, "evidence")
}

func blockHasNonEmptyYAMLValue(block string, key string) bool {
	value, ok := blockYAMLValue(block, key)
	return ok && value != "" && value != "[]"
}

func blockYAMLValue(block string, key string) (string, bool) {
	prefix := key + ":"
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return trimYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))), true
		}
	}
	return "", false
}

func blockHasYAMLListItem(block string, key string) bool {
	lines := strings.Split(block, "\n")
	inList := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !inList {
			if strings.HasPrefix(trimmed, key+":") {
				inList = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			return true
		}
		if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "- ") {
			return false
		}
	}
	return false
}

func trimYAMLScalar(value string) string {
	if i := strings.Index(value, "#"); i >= 0 {
		value = value[:i]
	}
	value = strings.TrimSpace(value)
	return strings.Trim(value, `"'`)
}

func gitOutput(args ...string) string {
	cmd := exec.Command("git", args...)
	data, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func flagValue(args []string, name string, fallbackValue string) string {
	for i, arg := range args {
		if arg == "--"+name && i+1 < len(args) {
			return args[i+1]
		}
		prefix := "--" + name + "="
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return fallbackValue
}

func plannedCommandVerifyRequested(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "--verify", "--strict", "--context=release_verify":
			return true
		}
		if arg == "--context" {
			continue
		}
	}
	for i, arg := range args {
		if arg == "--context" && i+1 < len(args) && args[i+1] == "release_verify" {
			return true
		}
	}
	return false
}

func emitPlannedReport(stdout io.Writer, stderr io.Writer, command, status string, details []string, gaps []string, args []string) int {
	exitCode := emitReport(stdout, command, status, details, gaps)
	if status == "planned" || status == "gap" {
		if plannedCommandVerifyRequested(args) {
			write(stderr, "ERROR: %s is %s under --verify/strict context\n", command, status)
		} else {
			write(stderr, "ERROR: %s is %s and cannot satisfy a release gate\n", command, status)
		}
	}
	return exitCode
}

// runRulesConsistencyCheck 校验 .agent/standard/goal-runtime-canonical.md（叙事层）
// 与 .agent/rules/iron-rules.md（机器层）引用的 RULE-* 编号集合一致，
// 并要求两侧引用的所有 RULE-* 都在 .agent/rules/registry.yaml 中登记。
//
// 用途：PR #36 引入双 SSOT 后，防止两份文档的铁律编号映射悄然漂移。
func runRulesConsistencyCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := validateInternalCommandArgs("rules-consistency-check", args, internalCommandFlagSpec{boolFlags: []string{"json"}}); err != nil {
		return invalidInternalArgsExit("rules-consistency-check", err, stderr)
	}

	canonicalPath := ".agent/standard/goal-runtime-canonical.md"
	ironPath := ".agent/rules/iron-rules.md"
	registryPath := ".agent/rules/registry.yaml"

	canonical, err := os.ReadFile(canonicalPath)
	if err != nil {
		write(stderr, "ERROR: read %s: %v\n", canonicalPath, err)
		return 1
	}
	iron, err := os.ReadFile(ironPath)
	if err != nil {
		write(stderr, "ERROR: read %s: %v\n", ironPath, err)
		return 1
	}
	registry, err := os.ReadFile(registryPath)
	if err != nil {
		write(stderr, "ERROR: read %s: %v\n", registryPath, err)
		return 1
	}

	// 从 canonical 抽"八条铁律"段表格的 RULE-* ID
	canonRules := extractCanonicalIronRuleIDs(string(canonical))
	// 从 iron-rules 抽"七律"段括号内的 RULE-* ID
	ironRules := extractIronRulesIDs(string(iron))
	// 从 registry.yaml 抽所有 - id: RULE-* 行
	registryRules := extractRegistryRuleIDs(string(registry))

	var gaps []string

	if len(canonRules) == 0 {
		gaps = append(gaps, fmt.Sprintf("%s: 未发现八条铁律段的 RULE-* 引用", canonicalPath))
	}
	if len(ironRules) == 0 {
		gaps = append(gaps, fmt.Sprintf("%s: 未发现七律段的 RULE-* 引用", ironPath))
	}
	if len(registryRules) == 0 {
		gaps = append(gaps, fmt.Sprintf("%s: 未发现 RULE-* 登记", registryPath))
	}

	// 两侧引用集合（去重并集）必须各自完全包含在 registry 中
	for id := range canonRules {
		if !registryRules[id] {
			gaps = append(gaps, fmt.Sprintf("%s 引用 %s 未在 %s 登记", canonicalPath, id, registryPath))
		}
	}
	for id := range ironRules {
		if !registryRules[id] {
			gaps = append(gaps, fmt.Sprintf("%s 引用 %s 未在 %s 登记", ironPath, id, registryPath))
		}
	}

	// canonical 的核心铁律 RULE-* 必须全部出现在 iron-rules 中（canonical ⊆ iron），
	// 但反向不强求：iron-rules 每条会附带多个关联 RULE-*（主+关联），
	// 这些不必都出现在 canonical 表格里。
	for id := range canonRules {
		if !ironRules[id] {
			gaps = append(gaps, fmt.Sprintf("漂移：%s 引用 %s 但 %s 未引用", canonicalPath, id, ironPath))
		}
	}

	if len(gaps) > 0 {
		write(stderr, "ERROR: rules-consistency-check found %d gap(s)\n", len(gaps))
		return emitReport(stdout, "rules-consistency-check", "failed", nil, gaps)
	}
	details := []string{
		fmt.Sprintf("canonical=%d iron=%d registry=%d 引用集合一致", len(canonRules), len(ironRules), len(registryRules)),
	}
	return emitReport(stdout, "rules-consistency-check", "passed", details, nil)
}

// extractCanonicalIronRuleIDs 抓取 canonical 的"八条铁律"段表格中
// 形如 `| RULE-XXX-NNN |` 的 ID。仅在该段内（首个 `## 1.` 之后到下一个 `##` 之前）。
func extractCanonicalIronRuleIDs(text string) map[string]bool {
	out := map[string]bool{}
	startIdx := strings.Index(text, "## 1.")
	if startIdx < 0 {
		return out
	}
	section := text[startIdx:]
	if nextIdx := strings.Index(section[5:], "\n## "); nextIdx >= 0 {
		section = section[:nextIdx+5]
	}
	re := regexp.MustCompile(`\|\s*(RULE-[A-Z]+(?:-[A-Z]+)*-\d+)\s*\|`)
	for _, m := range re.FindAllStringSubmatch(section, -1) {
		out[m[1]] = true
	}
	return out
}

// extractIronRulesIDs 抓取 iron-rules 的"七律"段中括号内引用的 RULE-* ID。
// 仅在 `## 七律` 段内。
func extractIronRulesIDs(text string) map[string]bool {
	out := map[string]bool{}
	startIdx := strings.Index(text, "## 七律")
	if startIdx < 0 {
		return out
	}
	section := text[startIdx:]
	if nextIdx := strings.Index(section[6:], "\n## "); nextIdx >= 0 {
		section = section[:nextIdx+6]
	}
	re := regexp.MustCompile(`RULE-[A-Z]+(?:-[A-Z]+)*-\d+`)
	for _, m := range re.FindAllString(section, -1) {
		out[m] = true
	}
	return out
}

// extractRegistryRuleIDs 抓取 registry.yaml 中所有 `- id: RULE-XXX-NNN` 行的 ID。
func extractRegistryRuleIDs(text string) map[string]bool {
	out := map[string]bool{}
	re := regexp.MustCompile(`(?m)^\s*-\s*id:\s*(RULE-[A-Z]+(?:-[A-Z]+)*-\d+)`)
	for _, m := range re.FindAllStringSubmatch(text, -1) {
		out[m[1]] = true
	}
	return out
}
