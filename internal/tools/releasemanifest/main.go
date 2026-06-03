package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ZoneCNH/xlib-standard/internal/debtcheck"
	"github.com/ZoneCNH/xlib-standard/internal/releasequality"
)

var checkNames = []string{
	"fmt",
	"vet",
	"lint",
	"unit_test",
	"race_test",
	"boundary",
	"secret_scan",
	"security",
	"contract",
	"integration",
	"dependency_check",
	"standard_impact",
	"docs_check",
	"property",
	"golden",
	"fuzz_smoke",
	"debt",
	"architecture",
	"domain",
	"docs_drift",
	"dependency_debt",
	"security_debt",
	"testing_debt",
	"implementation_debt",
	"downstream_debt",
}

var checkEnvNames = map[string]string{
	"fmt":                 "FMT_STATUS",
	"vet":                 "VET_STATUS",
	"lint":                "LINT_STATUS",
	"unit_test":           "UNIT_TEST_STATUS",
	"race_test":           "RACE_TEST_STATUS",
	"boundary":            "BOUNDARY_STATUS",
	"secret_scan":         "SECRET_SCAN_STATUS",
	"security":            "SECURITY_STATUS",
	"contract":            "CONTRACT_STATUS",
	"integration":         "INTEGRATION_STATUS",
	"dependency_check":    "DEPENDENCY_CHECK_STATUS",
	"standard_impact":     "STANDARD_IMPACT_STATUS",
	"docs_check":          "DOCS_CHECK_STATUS",
	"property":            "PROPERTY_STATUS",
	"golden":              "GOLDEN_STATUS",
	"fuzz_smoke":          "FUZZ_SMOKE_STATUS",
	"debt":                "DEBT_STATUS",
	"architecture":        "ARCHITECTURE_STATUS",
	"domain":              "DOMAIN_STATUS",
	"docs_drift":          "DOCS_DRIFT_STATUS",
	"dependency_debt":     "DEPENDENCY_DEBT_STATUS",
	"security_debt":       "SECURITY_DEBT_STATUS",
	"testing_debt":        "TESTING_DEBT_STATUS",
	"implementation_debt": "IMPLEMENTATION_DEBT_STATUS",
	"downstream_debt":     "DOWNSTREAM_DEBT_STATUS",
}

var contractFiles = []string{
	"contracts/config.schema.json",
	"contracts/error.schema.json",
	"contracts/health.schema.json",
	"contracts/metrics.md",
}

var requiredArtifacts = []string{
	"release/manifest/latest.json",
	"release/manifest/latest.json.sha256",
	"release/debt/latest.json",
	"release/debt/latest.md",
	"release/debt/latest.json.sha256",
}

const standardImpactReportPath = "release/standard-impact/latest.md"
const debtReportPath = "release/debt/latest.json"
const debtMarkdownPath = "release/debt/latest.md"
const debtChecksumPath = "release/debt/latest.json.sha256"
const governanceRuntimeVersion = "v2.9.3"

var downstreamReleaseDecisionValues = []string{
	"required",
	"not_required",
}

var repositoryRulesReleaseDecisionValues = []string{
	"audit_required",
	"not_required",
}

var generatorEvidenceTargets = []GeneratorTarget{
	{Name: "kernel", ModulePath: "github.com/ZoneCNH/kernel", PackageName: "kernel"},
	{Name: "configx", ModulePath: "github.com/ZoneCNH/configx", PackageName: "configx"},
	{Name: "redisx", ModulePath: "github.com/ZoneCNH/redisx", PackageName: "redisx"},
}

var governanceRuntimeGateStatuses = map[string]string{
	"governance": "passed",
}

var governanceRuntimeProfileStatuses = map[string]string{
	"p1_governance": "passed",
	"p2_runtime":    "passed",
}

type Manifest struct {
	Module                 string                 `json:"module"`
	Version                string                 `json:"version"`
	Commit                 string                 `json:"commit"`
	TreeSHA                string                 `json:"tree_sha"`
	SourceDigest           string                 `json:"source_digest"`
	TrackedFileCount       int                    `json:"tracked_file_count"`
	GoVersion              string                 `json:"go_version"`
	GeneratedAt            string                 `json:"generated_at"`
	GeneratedBy            string                 `json:"generated_by"`
	TreeState              string                 `json:"tree_state"`
	Checks                 map[string]string      `json:"checks"`
	Workflow               WorkflowEvidence       `json:"workflow"`
	Score                  releasequality.Report  `json:"score"`
	Contracts              []FileDigest           `json:"contracts"`
	Dependencies           []ModuleDigest         `json:"dependencies"`
	StandardImpact         StandardImpactEvidence `json:"standard_impact"`
	Debt                   DebtEvidence           `json:"debt"`
	GovernanceRuntime      GovernanceRuntime      `json:"governance_runtime"`
	DownstreamSyncRequired bool                   `json:"downstream_sync_required"`
	GeneratorEvidence      GeneratorEvidence      `json:"generator_evidence"`
	Tools                  map[string]string      `json:"tools"`
	Artifacts              []string               `json:"artifacts"`
	Notes                  Notes                  `json:"notes"`
}

type WorkflowEvidence struct {
	WorkflowRunID string `json:"workflow_run_id"`
	ArtifactName  string `json:"artifact_name"`
	ArtifactURL   string `json:"artifact_url"`
}

type FileDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ModuleDigest struct {
	Path    string         `json:"path"`
	Version string         `json:"version,omitempty"`
	Main    bool           `json:"main,omitempty"`
	Replace *ModuleReplace `json:"replace,omitempty"`
}

type ModuleReplace struct {
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}

type StandardImpactEvidence struct {
	ReportPath                     string `json:"report_path"`
	ReportSHA256                   string `json:"report_sha256"`
	Status                         string `json:"status"`
	DownstreamSyncRequired         bool   `json:"downstream_sync_required"`
	ContextRuntimeChange           bool   `json:"context_runtime_change"`
	GovernanceRegistryChange       bool   `json:"governance_registry_change"`
	DownstreamReleaseDecision      string `json:"downstream_release_decision"`
	RepositoryRulesReleaseDecision string `json:"repository_rules_release_decision"`
	PrimaryDownstream              string `json:"primary_downstream"`
}

type DebtEvidence struct {
	ReportPath   string  `json:"report_path"`
	MarkdownPath string  `json:"markdown_path"`
	ChecksumPath string  `json:"checksum_path"`
	ReportSHA256 string  `json:"report_sha256"`
	Status       string  `json:"status"`
	Score        float64 `json:"score"`
	MinScore     float64 `json:"min_score"`
	CheckCount   int     `json:"check_count"`
}

type GovernanceRuntime struct {
	Runtime         string            `json:"runtime"`
	SchemaVersion   string            `json:"schema_version"`
	RuntimeVersion  string            `json:"runtime_version"`
	Status          string            `json:"status"`
	Profiles        []string          `json:"profiles"`
	ProfileCheck    string            `json:"profile_check"`
	ReleaseTarget   string            `json:"release_target"`
	LegacyAliases   []string          `json:"legacy_aliases"`
	GateStatuses    map[string]string `json:"gate_statuses"`
	ProfileStatuses map[string]string `json:"profile_statuses"`
}

type GovernanceRuntimeEvidence = GovernanceRuntime

type GeneratorEvidence struct {
	Command  string            `json:"command"`
	Required bool              `json:"required"`
	Targets  []GeneratorTarget `json:"targets"`
}

type GeneratorTarget struct {
	Name        string `json:"name"`
	ModulePath  string `json:"module_path"`
	PackageName string `json:"package_name"`
}

type Notes struct {
	BreakingChanges string   `json:"breaking_changes"`
	KnownRisks      []string `json:"known_risks"`
}

var exit = os.Exit

func main() {
	exit(runCLI(os.Args[0], os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(name string, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "release/manifest/latest.json", "release manifest output path")
	verify := flags.String("verify", "", "verify an existing release manifest instead of generating one")
	requirePassed := flags.Bool("require-passed", false, "require all release checks to be passed during verification")
	requireClean := flags.Bool("require-clean", false, "require a clean git tree during verification")
	expectVersion := flags.String("expect-version", "", "require the manifest version to match this release version during verification")
	minScore := flags.Float64("min-score", 0, "require the release score to be at least this value during verification")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if *verify != "" {
		if err := verifyManifest(*verify, *requirePassed, *requireClean, *expectVersion, *minScore); err != nil {
			return printCLIError(stderr, err)
		}
		return printCLIStatus(stdout, "release evidence verified: %s\n", *verify)
	}

	manifest, err := buildManifest()
	if err != nil {
		return printCLIError(stderr, err)
	}
	if err := writeManifest(*out, manifest); err != nil {
		return printCLIError(stderr, err)
	}
	return printCLIStatus(stdout, "generated %s\n", *out)
}

func printCLIError(w io.Writer, err error) int {
	return printCLIMessage(w, 1, "ERROR: %v\n", err)
}

func printCLIStatus(w io.Writer, format string, args ...any) int {
	return printCLIMessage(w, 0, format, args...)
}

func printCLIMessage(w io.Writer, exitCode int, format string, args ...any) int {
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		return 1
	}
	return exitCode
}

func buildManifest() (Manifest, error) {
	module, err := runTrimmed("go", "list", "-m")
	if err != nil {
		return Manifest{}, err
	}

	sourceDigest, trackedFileCount, err := sourceDigest()
	if err != nil {
		return Manifest{}, err
	}
	contracts, err := contractDigests()
	if err != nil {
		return Manifest{}, err
	}
	dependencies, err := moduleDigests()
	if err != nil {
		return Manifest{}, err
	}
	standardImpact, err := buildStandardImpactEvidence()
	if err != nil {
		return Manifest{}, err
	}
	debtEvidence, err := buildDebtEvidence()
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		Module:                 module,
		Version:                envDefault("VERSION", "v0.4.6"),
		Commit:                 runTrimmedDefault("unknown", "git", "rev-parse", "HEAD"),
		TreeSHA:                runTrimmedDefault("unknown", "git", "rev-parse", "HEAD^{tree}"),
		SourceDigest:           sourceDigest,
		TrackedFileCount:       trackedFileCount,
		GoVersion:              runtime.Version(),
		GeneratedAt:            time.Now().UTC().Format(time.RFC3339),
		GeneratedBy:            envDefault("GENERATED_BY", "scripts/generate_manifest.sh"),
		TreeState:              treeState(),
		Checks:                 buildChecks(),
		Workflow:               buildWorkflowEvidence(),
		Score:                  releasequality.Compute(releasequality.DefaultMinimum),
		Contracts:              contracts,
		Dependencies:           dependencies,
		StandardImpact:         standardImpact,
		Debt:                   debtEvidence,
		GovernanceRuntime:      buildGovernanceRuntime(),
		DownstreamSyncRequired: standardImpact.DownstreamSyncRequired,
		GeneratorEvidence:      buildGeneratorEvidence(),
		Tools: map[string]string{
			"go":            firstLine(runTrimmedDefault(runtime.Version(), "go", "version")),
			"golangci-lint": toolVersion("golangci-lint", "--version"),
			"govulncheck":   toolVersion("govulncheck", "-version"),
		},
		Artifacts: append([]string(nil), requiredArtifacts...),
		Notes: Notes{
			BreakingChanges: "none",
			KnownRisks:      []string{},
		},
	}, nil
}

func verifyManifest(path string, requirePassed bool, requireClean bool, expectVersion string, minScore float64) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		return err
	}

	current, err := buildManifest()
	if err != nil {
		return err
	}

	var failures []string
	requireNonEmpty(&failures, "module", got.Module)
	requireNonEmpty(&failures, "version", got.Version)
	requireNonEmpty(&failures, "commit", got.Commit)
	requireNonEmpty(&failures, "tree_sha", got.TreeSHA)
	requireNonEmpty(&failures, "source_digest", got.SourceDigest)
	requireNonEmpty(&failures, "go_version", got.GoVersion)
	requireNonEmpty(&failures, "generated_at", got.GeneratedAt)
	requireNonEmpty(&failures, "generated_by", got.GeneratedBy)
	requireNonEmpty(&failures, "tree_state", got.TreeState)
	requireNonEmpty(&failures, "workflow.workflow_run_id", got.Workflow.WorkflowRunID)
	requireNonEmpty(&failures, "workflow.artifact_name", got.Workflow.ArtifactName)
	requireNonEmpty(&failures, "workflow.artifact_url", got.Workflow.ArtifactURL)

	expectVersion = strings.TrimSpace(expectVersion)
	if _, err := time.Parse(time.RFC3339, got.GeneratedAt); err != nil {
		failures = append(failures, "generated_at must be RFC3339")
	}
	if expectVersion != "" && got.Version != expectVersion {
		failures = append(failures, fmt.Sprintf("version mismatch: got %q, want %q", got.Version, expectVersion))
	}
	if got.Module != current.Module {
		failures = append(failures, fmt.Sprintf("module mismatch: got %q, want %q", got.Module, current.Module))
	}
	if got.Commit != current.Commit {
		failures = append(failures, fmt.Sprintf("commit mismatch: got %q, want %q", got.Commit, current.Commit))
	}
	if got.TreeSHA != current.TreeSHA {
		failures = append(failures, fmt.Sprintf("tree_sha mismatch: got %q, want %q", got.TreeSHA, current.TreeSHA))
	}
	if got.SourceDigest != current.SourceDigest {
		failures = append(failures, "source_digest does not match current tracked file contents")
	}
	if got.TrackedFileCount != current.TrackedFileCount {
		failures = append(failures, fmt.Sprintf("tracked_file_count mismatch: got %d, want %d", got.TrackedFileCount, current.TrackedFileCount))
	}
	if got.TreeState != current.TreeState {
		failures = append(failures, fmt.Sprintf("tree_state mismatch: got %q, want %q", got.TreeState, current.TreeState))
	}
	if got.Score.Value != current.Score.Value {
		failures = append(failures, fmt.Sprintf("score.value mismatch: got %.1f, want %.1f", got.Score.Value, current.Score.Value))
	}
	if got.Score.Status == "" {
		failures = append(failures, "score.status is required")
	}
	if got.Score.Threshold == 0 {
		failures = append(failures, "score.threshold is required")
	}
	if minScore > 0 {
		if err := releasequality.Verify(got.Score, minScore); err != nil {
			failures = append(failures, err.Error())
		}
	}
	if requireClean && got.TreeState != "clean" {
		failures = append(failures, fmt.Sprintf("tree_state must be clean, got %q", got.TreeState))
	}
	if !reflect.DeepEqual(got.Contracts, current.Contracts) {
		failures = append(failures, "contract fingerprints do not match current contract files")
	}
	if !reflect.DeepEqual(got.Dependencies, current.Dependencies) {
		failures = append(failures, "dependency inventory does not match go list -m -json all")
	}
	if !reflect.DeepEqual(got.StandardImpact, current.StandardImpact) {
		failures = append(failures, "standard_impact does not match current standard impact evidence")
	}
	if !reflect.DeepEqual(got.Debt, current.Debt) {
		failures = append(failures, "debt does not match current debt evidence")
	}
	if !reflect.DeepEqual(got.GovernanceRuntime, current.GovernanceRuntime) {
		failures = append(failures, "governance_runtime does not match current context runtime evidence")
	}
	if governanceFailures := validateGovernanceRuntimeEvidence(got.GovernanceRuntime); len(governanceFailures) > 0 {
		failures = append(failures, "governance_runtime does not match current governance runtime evidence")
		failures = append(failures, governanceFailures...)
	}
	if got.DownstreamSyncRequired != current.DownstreamSyncRequired {
		failures = append(failures, fmt.Sprintf("downstream_sync_required mismatch: got %t, want %t", got.DownstreamSyncRequired, current.DownstreamSyncRequired))
	}
	if got.DownstreamSyncRequired != got.StandardImpact.DownstreamSyncRequired {
		failures = append(failures, "downstream_sync_required must match standard_impact.downstream_sync_required")
	}
	requireNonEmpty(&failures, "standard_impact.report_path", got.StandardImpact.ReportPath)
	requireNonEmpty(&failures, "standard_impact.status", got.StandardImpact.Status)
	requireEnumValue(&failures, "standard_impact.downstream_release_decision", got.StandardImpact.DownstreamReleaseDecision, downstreamReleaseDecisionValues)
	requireEnumValue(&failures, "standard_impact.repository_rules_release_decision", got.StandardImpact.RepositoryRulesReleaseDecision, repositoryRulesReleaseDecisionValues)
	if requirePassed {
		if got.StandardImpact.Status != "present" {
			failures = append(failures, fmt.Sprintf("standard_impact.status must be present, got %q", got.StandardImpact.Status))
		}
		requireNonEmpty(&failures, "standard_impact.report_sha256", got.StandardImpact.ReportSHA256)
		requireNonEmpty(&failures, "standard_impact.primary_downstream", got.StandardImpact.PrimaryDownstream)
	}
	requireNonEmpty(&failures, "debt.report_path", got.Debt.ReportPath)
	requireNonEmpty(&failures, "debt.markdown_path", got.Debt.MarkdownPath)
	requireNonEmpty(&failures, "debt.checksum_path", got.Debt.ChecksumPath)
	requireNonEmpty(&failures, "debt.status", got.Debt.Status)
	if requirePassed {
		if got.Debt.Status != "passed" {
			failures = append(failures, fmt.Sprintf("debt.status must be passed, got %q", got.Debt.Status))
		}
		requireNonEmpty(&failures, "debt.report_sha256", got.Debt.ReportSHA256)
		if got.Debt.Score < got.Debt.MinScore {
			failures = append(failures, fmt.Sprintf("debt.score %.1f is below minimum %.1f", got.Debt.Score, got.Debt.MinScore))
		}
		if got.Debt.CheckCount == 0 {
			failures = append(failures, "debt.check_count must be greater than zero")
		}
	}
	requireNonEmpty(&failures, "governance_runtime.runtime", got.GovernanceRuntime.Runtime)
	requireNonEmpty(&failures, "governance_runtime.schema_version", got.GovernanceRuntime.SchemaVersion)
	requireNonEmpty(&failures, "governance_runtime.status", got.GovernanceRuntime.Status)
	requireNonEmpty(&failures, "governance_runtime.profile_check", got.GovernanceRuntime.ProfileCheck)
	requireNonEmpty(&failures, "governance_runtime.release_target", got.GovernanceRuntime.ReleaseTarget)
	if len(got.GovernanceRuntime.Profiles) == 0 {
		failures = append(failures, "governance_runtime.profiles is required")
	}
	if len(got.GovernanceRuntime.LegacyAliases) == 0 {
		failures = append(failures, "governance_runtime.legacy_aliases is required")
	}
	if !reflect.DeepEqual(got.GeneratorEvidence, current.GeneratorEvidence) {
		failures = append(failures, "generator_evidence does not match current integration evidence")
	}
	requireNonEmpty(&failures, "generator_evidence.command", got.GeneratorEvidence.Command)
	if !got.GeneratorEvidence.Required {
		failures = append(failures, "generator_evidence.required must be true")
	}
	if len(got.GeneratorEvidence.Targets) == 0 {
		failures = append(failures, "generator_evidence.targets is required")
	}
	for _, artifact := range requiredArtifacts {
		if !contains(got.Artifacts, artifact) {
			failures = append(failures, fmt.Sprintf("artifacts must include %s", artifact))
		}
	}
	if got.Tools["go"] == "" {
		failures = append(failures, "tools.go must be recorded")
	}
	failures = append(failures, validateChecks(got.Checks, requirePassed)...)

	if len(failures) > 0 {
		return errors.New("release evidence verification failed:\n - " + strings.Join(failures, "\n - "))
	}
	return nil
}

func writeManifest(path string, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := encodeManifestFunc(&buf, manifest); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func encodeManifest(w io.Writer, manifest Manifest) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

var encodeManifestFunc = encodeManifest

func buildChecks() map[string]string {
	defaultStatus := envDefault("CHECK_STATUS", "unknown")
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = envDefault(checkEnvNames[name], defaultStatus)
	}
	return checks
}

func buildWorkflowEvidence() WorkflowEvidence {
	runID := envDefault("WORKFLOW_RUN_ID", envDefault("GITHUB_RUN_ID", "local"))
	artifactName := envDefault("ARTIFACT_NAME", "release-manifest-"+runID)
	artifactURL := envDefault("ARTIFACT_URL", "")
	if artifactURL == "" {
		server := strings.TrimRight(envDefault("GITHUB_SERVER_URL", ""), "/")
		repo := strings.Trim(os.Getenv("GITHUB_REPOSITORY"), "/")
		if server != "" && repo != "" && runID != "local" {
			artifactURL = server + "/" + repo + "/actions/runs/" + runID
		} else {
			artifactURL = "local:" + artifactName
		}
	}
	return WorkflowEvidence{
		WorkflowRunID: runID,
		ArtifactName:  artifactName,
		ArtifactURL:   artifactURL,
	}
}

func buildStandardImpactEvidence() (StandardImpactEvidence, error) {
	evidence := StandardImpactEvidence{
		ReportPath:                     standardImpactReportPath,
		Status:                         "missing",
		DownstreamReleaseDecision:      "not_required",
		RepositoryRulesReleaseDecision: "not_required",
	}

	data, err := os.ReadFile(standardImpactReportPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return evidence, nil
		}
		return StandardImpactEvidence{}, err
	}

	sum := sha256.Sum256(data)
	report := string(data)
	evidence.ReportSHA256 = "sha256:" + hex.EncodeToString(sum[:])
	evidence.Status = "present"
	evidence.DownstreamSyncRequired = strings.EqualFold(parseReportValue(report, "downstream_sync_required"), "true")
	evidence.ContextRuntimeChange = strings.EqualFold(parseReportValue(report, "context_runtime_change"), "true")
	evidence.GovernanceRegistryChange = strings.EqualFold(parseReportValue(report, "governance_registry_change"), "true")
	evidence.PrimaryDownstream = parseReportValue(report, "primary_downstream")
	evidence.DownstreamReleaseDecision = reportValueDefault(report, "downstream_release_decision", "not_required")
	evidence.RepositoryRulesReleaseDecision = reportValueDefault(report, "repository_rules_release_decision", "not_required")
	return evidence, nil
}

func buildDebtEvidence() (DebtEvidence, error) {
	evidence := DebtEvidence{
		ReportPath:   debtReportPath,
		MarkdownPath: debtMarkdownPath,
		ChecksumPath: debtChecksumPath,
		Status:       "missing",
		MinScore:     debtcheck.DefaultMinScore,
	}

	data, err := os.ReadFile(debtReportPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return evidence, nil
		}
		return DebtEvidence{}, err
	}

	sum := sha256.Sum256(data)
	evidence.ReportSHA256 = "sha256:" + hex.EncodeToString(sum[:])

	var report struct {
		Status   string            `json:"status"`
		Score    float64           `json:"score"`
		MinScore float64           `json:"min_score"`
		Checks   []json.RawMessage `json:"checks"`
		Sections []json.RawMessage `json:"sections"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return DebtEvidence{}, err
	}

	evidence.Status = report.Status
	evidence.Score = report.Score
	if report.MinScore != 0 {
		evidence.MinScore = report.MinScore
	}
	evidence.CheckCount = len(report.Checks)
	if evidence.CheckCount == 0 {
		evidence.CheckCount = len(report.Sections)
	}
	return evidence, nil
}

func buildGovernanceRuntime() GovernanceRuntime {
	evidence := buildGovernanceRuntimeEvidence()
	evidence.Runtime = "context-runtime-v4.0"
	evidence.Status = "present"
	evidence.Profiles = []string{
		"context-lite",
		"context-standard",
		"context-full",
		"context-release",
	}
	evidence.ProfileCheck = "context-profile-check"
	evidence.ReleaseTarget = "context-release"
	evidence.LegacyAliases = []string{
		"context-fast-check",
		"context-standard-check",
		"context-full-check",
	}
	return evidence
}

func buildGovernanceRuntimeEvidence() GovernanceRuntimeEvidence {
	return GovernanceRuntimeEvidence{
		SchemaVersion:   governanceRuntimeVersion,
		RuntimeVersion:  governanceRuntimeVersion,
		GateStatuses:    copyStatusMap(governanceRuntimeGateStatuses),
		ProfileStatuses: copyStatusMap(governanceRuntimeProfileStatuses),
	}
}

func copyStatusMap(statuses map[string]string) map[string]string {
	copied := make(map[string]string, len(statuses))
	for name, status := range statuses {
		copied[name] = status
	}
	return copied
}

func parseReportValue(report string, key string) string {
	prefix := key + ":"
	for _, line := range strings.Split(report, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if strings.HasPrefix(line, prefix) {
			value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			value = strings.Trim(value, "`")
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func reportValueDefault(report string, key string, fallback string) string {
	value := parseReportValue(report, key)
	if value == "" {
		return fallback
	}
	return value
}

func buildGeneratorEvidence() GeneratorEvidence {
	return GeneratorEvidence{
		Command:  "GOWORK=off make integration",
		Required: true,
		Targets:  append([]GeneratorTarget(nil), generatorEvidenceTargets...),
	}
}

func validateChecks(checks map[string]string, requirePassed bool) []string {
	var failures []string
	for _, name := range checkNames {
		status := strings.TrimSpace(checks[name])
		if status == "" {
			failures = append(failures, "checks."+name+" is required")
			continue
		}
		if requirePassed && status != "passed" {
			failures = append(failures, fmt.Sprintf("checks.%s must be passed, got %q", name, status))
		}
	}
	return failures
}

func validateGovernanceRuntimeEvidence(evidence GovernanceRuntimeEvidence) []string {
	var failures []string
	if evidence.SchemaVersion != governanceRuntimeVersion {
		failures = append(failures, fmt.Sprintf("governance_runtime.schema_version must be %q, got %q", governanceRuntimeVersion, evidence.SchemaVersion))
	}
	if evidence.RuntimeVersion != governanceRuntimeVersion {
		failures = append(failures, fmt.Sprintf("governance_runtime.runtime_version must be %q, got %q", governanceRuntimeVersion, evidence.RuntimeVersion))
	}
	appendRequiredPassedStatuses(&failures, "governance_runtime.gate_statuses", evidence.GateStatuses, governanceRuntimeGateStatuses)
	appendRequiredPassedStatuses(&failures, "governance_runtime.profile_statuses", evidence.ProfileStatuses, governanceRuntimeProfileStatuses)
	return failures
}

func appendRequiredPassedStatuses(failures *[]string, field string, got map[string]string, required map[string]string) {
	for name, want := range required {
		status := strings.TrimSpace(got[name])
		if status == "" {
			*failures = append(*failures, fmt.Sprintf("%s.%s is required", field, name))
			continue
		}
		if status != want {
			*failures = append(*failures, fmt.Sprintf("%s.%s must be %s, got %q", field, name, want, status))
		}
	}
}

func sourceDigest() (string, int, error) {
	raw, err := runRawCommand("git", "ls-files", "-z")
	if err != nil {
		return "", 0, err
	}
	parts := strings.Split(string(raw), "\x00")
	files := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			files = append(files, part)
		}
	}
	sort.Strings(files)

	digest := sha256.New()
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", 0, err
		}
		fileSum := sha256.Sum256(data)
		digest.Write([]byte(path))
		digest.Write([]byte{0})
		digest.Write([]byte(hex.EncodeToString(fileSum[:])))
		digest.Write([]byte{0})
	}

	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), len(files), nil
}

func contractDigests() ([]FileDigest, error) {
	digests := make([]FileDigest, 0, len(contractFiles))
	for _, path := range contractFiles {
		digest, err := fileDigest(path)
		if err != nil {
			return nil, err
		}
		digests = append(digests, digest)
	}
	return digests, nil
}

func fileDigest(path string) (FileDigest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileDigest{}, err
	}
	sum := sha256.Sum256(data)
	return FileDigest{
		Path:   path,
		SHA256: "sha256:" + hex.EncodeToString(sum[:]),
	}, nil
}

func moduleDigests() ([]ModuleDigest, error) {
	raw, err := runRawCommand("go", "list", "-m", "-json", "all")
	if err != nil {
		return nil, err
	}

	type goModule struct {
		Path    string
		Version string
		Main    bool
		Replace *struct {
			Path    string
			Version string
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	var modules []ModuleDigest
	for {
		var module goModule
		if err := decoder.Decode(&module); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		digest := ModuleDigest{
			Path:    module.Path,
			Version: module.Version,
			Main:    module.Main,
		}
		if module.Replace != nil {
			digest.Replace = &ModuleReplace{
				Path:    module.Replace.Path,
				Version: module.Replace.Version,
			}
		}
		modules = append(modules, digest)
	}
	return modules, nil
}

func treeState() string {
	status, err := runTrimmed("git", "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return "unknown"
	}
	if status == "" {
		return "clean"
	}
	return "dirty"
}

func toolVersion(name string, args ...string) string {
	if _, err := exec.LookPath(name); err != nil {
		return "missing"
	}
	output, err := runTrimmed(name, args...)
	if err != nil {
		return "error: " + firstLine(err.Error())
	}
	return firstLine(output)
}

func runTrimmedDefault(fallback string, name string, args ...string) string {
	output, err := runTrimmed(name, args...)
	if err != nil {
		return fallback
	}
	return output
}

func runTrimmed(name string, args ...string) (string, error) {
	output, err := runRawCommand(name, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func runRaw(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

var runRawCommand = runRaw

func envDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		return value[:idx]
	}
	return value
}

func requireNonEmpty(failures *[]string, field string, value string) {
	if strings.TrimSpace(value) == "" {
		*failures = append(*failures, field+" is required")
	}
}

func requireEnumValue(failures *[]string, field string, value string, allowed []string) {
	if strings.TrimSpace(value) == "" {
		*failures = append(*failures, field+" is required")
		return
	}
	if contains(allowed, value) {
		return
	}
	*failures = append(*failures, fmt.Sprintf("%s must be one of %s, got %q", field, strings.Join(allowed, ", "), value))
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
