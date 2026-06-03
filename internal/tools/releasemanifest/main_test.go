package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestBuildChecksUsesGlobalAndSpecificStatus(t *testing.T) {
	t.Setenv("CHECK_STATUS", "passed")
	t.Setenv("LINT_STATUS", "failed")
	t.Setenv("DOCS_CHECK_STATUS", "skipped")

	checks := buildChecks()

	if checks["fmt"] != "passed" {
		t.Fatalf("fmt status = %q, want passed", checks["fmt"])
	}
	if checks["lint"] != "failed" {
		t.Fatalf("lint status = %q, want failed", checks["lint"])
	}
	if checks["docs_check"] != "skipped" {
		t.Fatalf("docs_check status = %q, want skipped", checks["docs_check"])
	}
}

func TestCheckNamesCoverReleaseTargets(t *testing.T) {
	wantNames := []string{
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

	if strings.Join(checkNames, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("checkNames = %v, want %v", checkNames, wantNames)
	}
	for _, name := range wantNames {
		if checkEnvNames[name] == "" {
			t.Fatalf("checkEnvNames[%q] is empty", name)
		}
	}
}

func TestValidateChecksRequiresPassedStatuses(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = "passed"
	}
	checks["security"] = "unknown"

	failures := validateChecks(checks, true)

	if len(failures) != 1 {
		t.Fatalf("len(failures) = %d, want 1: %v", len(failures), failures)
	}
	if !strings.Contains(failures[0], "checks.security") {
		t.Fatalf("failure = %q, want security check failure", failures[0])
	}
}

func TestFileDigestRecordsPathAndSHA256(t *testing.T) {
	path := t.TempDir() + "/contract.json"
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}

	digest, err := fileDigest(path)
	if err != nil {
		t.Fatal(err)
	}

	if digest.Path != path {
		t.Fatalf("path = %q, want %q", digest.Path, path)
	}
	const want = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if digest.SHA256 != want {
		t.Fatalf("sha256 = %q, want %q", digest.SHA256, want)
	}
}

func TestRunCLIGeneratesManifestToOut(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v1.2.3-cli")
	t.Setenv("GENERATED_BY", "releasemanifest-cli-test")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "custom", "latest.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-out", outPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if want := "generated " + outPath; !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("generated manifest is invalid JSON: %s", data)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Module != "example.com/releasefixture" {
		t.Fatalf("module = %q, want fixture module", manifest.Module)
	}
	if manifest.Version != "v1.2.3-cli" {
		t.Fatalf("version = %q, want v1.2.3-cli", manifest.Version)
	}
	if manifest.GeneratedBy != "releasemanifest-cli-test" {
		t.Fatalf("generated_by = %q, want releasemanifest-cli-test", manifest.GeneratedBy)
	}
	if manifest.Workflow.WorkflowRunID == "" || manifest.Workflow.ArtifactName == "" || manifest.Workflow.ArtifactURL == "" {
		t.Fatalf("workflow evidence is incomplete: %+v", manifest.Workflow)
	}
	if manifest.Score.Threshold != 9.8 || manifest.Score.Status == "" || len(manifest.Score.Dimensions) == 0 {
		t.Fatalf("score report is incomplete: %+v", manifest.Score)
	}
	if manifest.StandardImpact.ReportPath != standardImpactReportPath || manifest.StandardImpact.Status == "" {
		t.Fatalf("standard impact evidence is incomplete: %+v", manifest.StandardImpact)
	}
	if manifest.StandardImpact.Status == "present" && (manifest.StandardImpact.DownstreamReleaseDecision == "" || manifest.StandardImpact.RepositoryRulesReleaseDecision == "") {
		t.Fatalf("standard impact release decisions are incomplete: %+v", manifest.StandardImpact)
	}
	if manifest.Debt.ReportPath != debtReportPath || manifest.Debt.Status != "passed" || manifest.Debt.CheckCount == 0 {
		t.Fatalf("debt evidence is incomplete: %+v", manifest.Debt)
	}
	assertGovernanceRuntimeEvidence(t, manifest.GovernanceRuntime)
	if manifest.DownstreamSyncRequired != manifest.StandardImpact.DownstreamSyncRequired {
		t.Fatalf("downstream_sync_required = %t, want standard impact value %t", manifest.DownstreamSyncRequired, manifest.StandardImpact.DownstreamSyncRequired)
	}
	if !manifest.GeneratorEvidence.Required || manifest.GeneratorEvidence.Command == "" {
		t.Fatalf("generator evidence is incomplete: %+v", manifest.GeneratorEvidence)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "kernel", "github.com/ZoneCNH/kernel", "kernel") {
		t.Fatalf("generator targets = %+v, want kernel target", manifest.GeneratorEvidence.Targets)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "configx", "github.com/ZoneCNH/configx", "configx") {
		t.Fatalf("generator targets = %+v, want configx target", manifest.GeneratorEvidence.Targets)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "redisx", "github.com/ZoneCNH/redisx", "redisx") {
		t.Fatalf("generator targets = %+v, want redisx target", manifest.GeneratorEvidence.Targets)
	}
	for _, name := range checkNames {
		if manifest.Checks[name] != "passed" {
			t.Fatalf("checks[%q] = %q, want passed", name, manifest.Checks[name])
		}
	}
}

func TestRunCLIGenerateReportsBuildManifestFailure(t *testing.T) {
	t.Setenv("GOWORK", "off")
	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/brokenmanifest\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repo, "git", "add", ".")
	chdir(t, repo)

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-out", outPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI generate exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	message := stderr.String()
	for _, want := range []string{"ERROR:", "contracts/config.schema.json"} {
		if !strings.Contains(message, want) {
			t.Fatalf("stderr = %q, want substring %q", message, want)
		}
	}
	if _, err := os.Stat(outPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("generated manifest exists after failed build: %v", err)
	}
}

func TestRunCLIGenerateReportsWriteManifestFailure(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-out", outPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI generate exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := "ERROR:"; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIVerifiesManifestWithRequirePassed(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v1.2.3")
	t.Setenv("CHECK_STATUS", "passed")
	repo := releaseManifestFixtureRepo(t)
	writeStandardImpactReportFixture(t, repo)
	chdir(t, repo)

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-require-passed", "-expect-version", "v1.2.3"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI verify exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if want := "release evidence verified: " + outPath; !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
	}
}

func TestRunCLIVerifyRejectsScoreBelowMinimum(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-min-score", "9.8"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{"release score", "below minimum"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
}

func TestBuildWorkflowEvidencePrefersExplicitEnvironment(t *testing.T) {
	t.Setenv("WORKFLOW_RUN_ID", "12345")
	t.Setenv("ARTIFACT_NAME", "manifest-artifact")
	t.Setenv("ARTIFACT_URL", "https://example.invalid/artifacts/manifest")

	got := buildWorkflowEvidence()

	if got.WorkflowRunID != "12345" || got.ArtifactName != "manifest-artifact" || got.ArtifactURL != "https://example.invalid/artifacts/manifest" {
		t.Fatalf("workflow evidence = %+v, want explicit env values", got)
	}
}

func TestRunCLIVerifyRejectsExpectedVersionMismatch(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v1.2.3")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-expect-version", "v9.9.9"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := `version mismatch: got "v1.2.3", want "v9.9.9"`; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIVerifyReportsDrift(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.SourceDigest = "sha256:stale"
	manifest.Checks["lint"] = "failed"
	manifest.StandardImpact.Status = "stale"
	manifest.Debt.Status = "stale"
	manifest.GovernanceRuntime.Status = "stale"
	manifest.DownstreamSyncRequired = !manifest.StandardImpact.DownstreamSyncRequired
	manifest.GovernanceRuntime.RuntimeVersion = "v2.9.2"
	manifest.GovernanceRuntime.ProfileStatuses["p2_runtime"] = "failed"
	manifest.GeneratorEvidence.Targets = nil
	if err := writeManifest(outPath, manifest); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-require-passed"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	message := stderr.String()
	for _, want := range []string{
		"ERROR: release evidence verification failed",
		"source_digest does not match current tracked file contents",
		"standard_impact does not match current standard impact evidence",
		"debt does not match current debt evidence",
		`debt.status must be passed, got "stale"`,
		"governance_runtime does not match current context runtime evidence",
		"generator_evidence does not match current integration evidence",
		`checks.lint must be passed, got "failed"`,
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("stderr = %q, want substring %q", message, want)
		}
	}
}

func TestRunCLIVerifyRequiresCleanTree(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	outPath := filepath.Join(t.TempDir(), "latest.json")
	var generateStdout bytes.Buffer
	var generateStderr bytes.Buffer
	if code := runCLI("releasemanifest", []string{"-out", outPath}, &generateStdout, &generateStderr); code != 0 {
		t.Fatalf("runCLI generate exit code = %d, want 0; stderr: %s", code, generateStderr.String())
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.TreeState = "dirty"
	if err := writeManifest(outPath, manifest); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runCLI("releasemanifest", []string{"-verify", outPath, "-require-clean"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runCLI verify exit code = %d, want 1; stdout: %s; stderr: %s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := `tree_state must be clean, got "dirty"`; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIHelpReturnsSuccess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-h"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runCLI help exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := "Usage of releasemanifest"; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestRunCLIRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runCLI("releasemanifest", []string{"-unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("runCLI unknown flag exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if want := "flag provided but not defined"; !strings.Contains(stderr.String(), want) {
		t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
	}
}

func TestPrintCLIMessageReportsWriterFailure(t *testing.T) {
	if code := printCLIStatus(errorWriter{}, "ok\n"); code != 1 {
		t.Fatalf("printCLIStatus exit code = %d, want 1", code)
	}
	if code := printCLIError(errorWriter{}, errors.New("boom")); code != 1 {
		t.Fatalf("printCLIError exit code = %d, want 1", code)
	}
}

func TestMainDelegatesToRunCLIAndExit(t *testing.T) {
	previousArgs := os.Args
	previousExit := exit
	t.Cleanup(func() {
		os.Args = previousArgs
		exit = previousExit
	})

	var gotCode *int
	exit = func(code int) {
		gotCode = &code
	}
	os.Args = []string{"releasemanifest", "-h"}

	main()

	if gotCode == nil {
		t.Fatal("main did not call exit")
	}
	if *gotCode != 0 {
		t.Fatalf("main exit code = %d, want 0", *gotCode)
	}
}

func TestBuildManifestRecordsFixtureRepositoryFacts(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("VERSION", "v9.9.9-test")
	t.Setenv("GENERATED_BY", "releasemanifest-test")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	if manifest.Module != "example.com/releasefixture" {
		t.Fatalf("module = %q, want example.com/releasefixture", manifest.Module)
	}
	if manifest.Version != "v9.9.9-test" {
		t.Fatalf("version = %q, want v9.9.9-test", manifest.Version)
	}
	if manifest.GeneratedBy != "releasemanifest-test" {
		t.Fatalf("generated_by = %q, want releasemanifest-test", manifest.GeneratedBy)
	}
	if _, err := time.Parse(time.RFC3339, manifest.GeneratedAt); err != nil {
		t.Fatalf("generated_at = %q, want RFC3339: %v", manifest.GeneratedAt, err)
	}
	if !strings.HasPrefix(manifest.SourceDigest, "sha256:") {
		t.Fatalf("source_digest = %q, want sha256 prefix", manifest.SourceDigest)
	}
	if manifest.TrackedFileCount == 0 {
		t.Fatal("tracked_file_count = 0, want tracked files")
	}
	if len(manifest.Contracts) != len(contractFiles) {
		t.Fatalf("len(contracts) = %d, want %d", len(manifest.Contracts), len(contractFiles))
	}
	if len(manifest.Dependencies) == 0 || manifest.Dependencies[0].Path != manifest.Module || !manifest.Dependencies[0].Main {
		t.Fatalf("dependencies[0] = %+v, want main module %q", manifest.Dependencies, manifest.Module)
	}
	if manifest.Tools["go"] == "" {
		t.Fatal("tools.go is empty")
	}
	if manifest.StandardImpact.ReportPath != standardImpactReportPath {
		t.Fatalf("standard_impact.report_path = %q, want %q", manifest.StandardImpact.ReportPath, standardImpactReportPath)
	}
	if manifest.StandardImpact.Status != "missing" && manifest.StandardImpact.Status != "present" {
		t.Fatalf("standard_impact.status = %q, want missing or present", manifest.StandardImpact.Status)
	}
	if manifest.GovernanceRuntime.Runtime != "context-runtime-v4.0" {
		t.Fatalf("governance_runtime.runtime = %q, want context-runtime-v4.0", manifest.GovernanceRuntime.Runtime)
	}
	if manifest.GovernanceRuntime.SchemaVersion != governanceRuntimeVersion || manifest.GovernanceRuntime.Status != "present" {
		t.Fatalf("governance_runtime = %+v, want schema %s and present", manifest.GovernanceRuntime, governanceRuntimeVersion)
	}
	for _, profile := range []string{"context-lite", "context-standard", "context-full", "context-release"} {
		if !contains(manifest.GovernanceRuntime.Profiles, profile) {
			t.Fatalf("governance_runtime.profiles = %v, want %s", manifest.GovernanceRuntime.Profiles, profile)
		}
	}
	if manifest.GovernanceRuntime.ProfileCheck != "context-profile-check" {
		t.Fatalf("governance_runtime.profile_check = %q, want context-profile-check", manifest.GovernanceRuntime.ProfileCheck)
	}
	if manifest.GovernanceRuntime.ReleaseTarget != "context-release" {
		t.Fatalf("governance_runtime.release_target = %q, want context-release", manifest.GovernanceRuntime.ReleaseTarget)
	}
	for _, alias := range []string{"context-fast-check", "context-standard-check", "context-full-check"} {
		if !contains(manifest.GovernanceRuntime.LegacyAliases, alias) {
			t.Fatalf("governance_runtime.legacy_aliases = %v, want %s", manifest.GovernanceRuntime.LegacyAliases, alias)
		}
	}
	if manifest.DownstreamSyncRequired != manifest.StandardImpact.DownstreamSyncRequired {
		t.Fatalf("downstream_sync_required = %t, want standard impact value %t", manifest.DownstreamSyncRequired, manifest.StandardImpact.DownstreamSyncRequired)
	}
	assertGovernanceRuntimeEvidence(t, manifest.GovernanceRuntime)
	if manifest.GeneratorEvidence.Command != "GOWORK=off make integration" || !manifest.GeneratorEvidence.Required {
		t.Fatalf("generator_evidence = %+v, want integration command and required=true", manifest.GeneratorEvidence)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "kernel", "github.com/ZoneCNH/kernel", "kernel") {
		t.Fatalf("generator targets = %+v, want kernel target", manifest.GeneratorEvidence.Targets)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "configx", "github.com/ZoneCNH/configx", "configx") {
		t.Fatalf("generator targets = %+v, want configx target", manifest.GeneratorEvidence.Targets)
	}
	if !hasGeneratorTarget(manifest.GeneratorEvidence.Targets, "redisx", "github.com/ZoneCNH/redisx", "redisx") {
		t.Fatalf("generator targets = %+v, want redisx target", manifest.GeneratorEvidence.Targets)
	}
	for _, artifact := range requiredArtifacts {
		if !contains(manifest.Artifacts, artifact) {
			t.Fatalf("artifacts = %v, want %s", manifest.Artifacts, artifact)
		}
	}
	for _, name := range checkNames {
		if manifest.Checks[name] != "passed" {
			t.Fatalf("checks[%q] = %q, want passed", name, manifest.Checks[name])
		}
	}
	if manifest.TreeState != "clean" && manifest.TreeState != "dirty" {
		t.Fatalf("tree_state = %q, want clean or dirty", manifest.TreeState)
	}
}

func TestBuildManifestReportsBuilderFailures(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(t *testing.T) string
		mock      func(name string, args ...string) ([]byte, error)
		wantError string
	}{
		{
			name: "module name",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			mock: func(name string, args ...string) ([]byte, error) {
				if name == "go" && strings.Join(args, " ") == "list -m" {
					return nil, errors.New("module command failed")
				}
				return runRaw(name, args...)
			},
			wantError: "module command failed",
		},
		{
			name: "source digest",
			setup: func(t *testing.T) string {
				t.Setenv("GOWORK", "off")
				return releaseManifestFixtureRepo(t)
			},
			mock: func(name string, args ...string) ([]byte, error) {
				if name == "git" && strings.Join(args, " ") == "ls-files -z" {
					return nil, errors.New("source command failed")
				}
				return runRaw(name, args...)
			},
			wantError: "source command failed",
		},
		{
			name: "module digests",
			setup: func(t *testing.T) string {
				t.Setenv("GOWORK", "off")
				return releaseManifestFixtureRepo(t)
			},
			mock: func(name string, args ...string) ([]byte, error) {
				if name == "go" && strings.Join(args, " ") == "list -m -json all" {
					return nil, errors.New("dependency command failed")
				}
				return runRaw(name, args...)
			},
			wantError: "dependency command failed",
		},
		{
			name: "standard impact",
			setup: func(t *testing.T) string {
				t.Setenv("GOWORK", "off")
				repo := releaseManifestFixtureRepo(t)
				reportPath := filepath.Join(repo, filepath.FromSlash(standardImpactReportPath))
				if err := os.MkdirAll(reportPath, 0o755); err != nil {
					t.Fatal(err)
				}
				return repo
			},
			wantError: "is a directory",
		},
		{
			name: "debt evidence",
			setup: func(t *testing.T) string {
				t.Setenv("GOWORK", "off")
				repo := releaseManifestFixtureRepo(t)
				reportPath := filepath.Join(repo, filepath.FromSlash(debtReportPath))
				if err := os.Remove(reportPath); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(reportPath, 0o755); err != nil {
					t.Fatal(err)
				}
				return repo
			},
			wantError: "is a directory",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chdir(t, tc.setup(t))
			if tc.mock != nil {
				withRunRawCommand(t, tc.mock)
			}

			_, err := buildManifest()
			if err == nil {
				t.Fatal("buildManifest succeeded, want error")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tc.wantError)
			}
		})
	}
}

func TestVerifyManifestAcceptsFreshManifestAndRejectsDrift(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	repo := releaseManifestFixtureRepo(t)
	writeStandardImpactReportFixture(t, repo)
	chdir(t, repo)

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	goodPath := filepath.Join(t.TempDir(), "latest.json")
	if err := writeManifest(goodPath, manifest); err != nil {
		t.Fatal(err)
	}
	if err := verifyManifest(goodPath, true, false, "", 0); err != nil {
		t.Fatalf("verify fresh manifest: %v", err)
	}

	manifest.SourceDigest = "sha256:bad"
	manifest.Checks["lint"] = "unknown"
	manifest.Artifacts = []string{"release/manifest/latest.json"}
	manifest.StandardImpact.Status = "stale"
	manifest.Debt.Status = "stale"
	manifest.GovernanceRuntime.Status = "stale"
	manifest.DownstreamSyncRequired = !manifest.StandardImpact.DownstreamSyncRequired
	manifest.GovernanceRuntime.GateStatuses["governance"] = "failed"
	manifest.GeneratorEvidence.Command = "make old-integration"
	badPath := filepath.Join(t.TempDir(), "stale.json")
	if err := writeManifest(badPath, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(badPath, true, false, "", 0)
	if err == nil {
		t.Fatal("verify stale manifest succeeded, want error")
	}
	message := err.Error()
	for _, want := range []string{
		"source_digest does not match current tracked file contents",
		`checks.lint must be passed, got "unknown"`,
		"artifacts must include release/manifest/latest.json.sha256",
		"standard_impact does not match current standard impact evidence",
		"debt does not match current debt evidence",
		"governance_runtime does not match current context runtime evidence",
		"downstream_sync_required must match standard_impact.downstream_sync_required",
		"governance_runtime does not match current governance runtime evidence",
		`governance_runtime.gate_statuses.governance must be passed, got "failed"`,
		"generator_evidence does not match current integration evidence",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error = %q, want substring %q", message, want)
		}
	}
}

func TestVerifyManifestRejectsInvalidStandardImpactReleaseDecisionEnums(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	repo := releaseManifestFixtureRepo(t)
	writeStandardImpactReportFixture(t, repo)
	chdir(t, repo)

	base, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*Manifest)
		want   string
	}{
		{
			name: "downstream_release_decision",
			mutate: func(manifest *Manifest) {
				manifest.StandardImpact.DownstreamReleaseDecision = "later"
			},
			want: `standard_impact.downstream_release_decision must be one of required, not_required, got "later"`,
		},
		{
			name: "repository_rules_release_decision",
			mutate: func(manifest *Manifest) {
				manifest.StandardImpact.RepositoryRulesReleaseDecision = "skip_audit"
			},
			want: `standard_impact.repository_rules_release_decision must be one of audit_required, not_required, got "skip_audit"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := base
			tt.mutate(&manifest)

			path := filepath.Join(t.TempDir(), "manifest.json")
			if err := writeManifest(path, manifest); err != nil {
				t.Fatal(err)
			}

			err := verifyManifest(path, false, false, "", 0)
			if err == nil {
				t.Fatal("verify invalid standard impact release decision succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func TestVerifyManifestRequiresStandardImpactEvidenceWhenChecksRequired(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if manifest.StandardImpact.Status != "missing" {
		t.Fatalf("fixture standard_impact.status = %q, want missing", manifest.StandardImpact.Status)
	}

	path := filepath.Join(t.TempDir(), "latest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, true, false, "", 0)
	if err == nil {
		t.Fatal("verify manifest without standard impact report succeeded, want error")
	}
	message := err.Error()
	for _, want := range []string{
		`standard_impact.status must be present, got "missing"`,
		"standard_impact.report_sha256 is required",
		"standard_impact.primary_downstream is required",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error = %q, want substring %q", message, want)
		}
	}
}

func TestVerifyManifestRejectsCorruptedManifestFields(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, releaseManifestFixtureRepo(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.GeneratedAt = "not-rfc3339"
	manifest.Module = "example.com/wrong"
	manifest.Commit = "wrong-commit"
	manifest.TreeSHA = "wrong-tree"
	manifest.TrackedFileCount++
	manifest.TreeState = "wrong-state"
	manifest.Score.Value++
	manifest.Score.Status = ""
	manifest.Score.Threshold = 0
	manifest.Contracts = nil
	manifest.Dependencies = nil
	manifest.GovernanceRuntime = GovernanceRuntime{}
	manifest.Debt = DebtEvidence{}
	manifest.GeneratorEvidence.Required = false
	manifest.Tools = map[string]string{}

	path := filepath.Join(t.TempDir(), "corrupt.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "", 0)
	if err == nil {
		t.Fatal("verify corrupt manifest succeeded, want error")
	}
	message := err.Error()
	for _, want := range []string{
		"generated_at must be RFC3339",
		"module mismatch:",
		"commit mismatch:",
		"tree_sha mismatch:",
		"tracked_file_count mismatch:",
		"tree_state mismatch:",
		"score.value mismatch:",
		"score.status is required",
		"score.threshold is required",
		"contract fingerprints do not match current contract files",
		"dependency inventory does not match go list -m -json all",
		"governance_runtime.runtime is required",
		"debt.report_path is required",
		"debt.status is required",
		"governance_runtime.profiles is required",
		"governance_runtime.legacy_aliases is required",
		"generator_evidence.required must be true",
		"tools.go must be recorded",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error = %q, want substring %q", message, want)
		}
	}
}

func TestVerifyManifestReportsFileAndDecodeAndCurrentBuildErrors(t *testing.T) {
	if err := verifyManifest(filepath.Join(t.TempDir(), "missing.json"), false, false, "", 0); err == nil {
		t.Fatal("verify missing file succeeded, want error")
	}

	invalidPath := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifyManifest(invalidPath, false, false, "", 0); err == nil {
		t.Fatal("verify invalid JSON succeeded, want error")
	}

	root := t.TempDir()
	validPath := filepath.Join(root, "manifest.json")
	if err := writeManifest(validPath, Manifest{GeneratedAt: "2026-01-02T03:04:05Z"}); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)
	if err := verifyManifest(validPath, false, false, "", 0); err == nil {
		t.Fatal("verify without current build context succeeded, want error")
	}
}

func TestBuildStandardImpactEvidenceAllowsMissingReport(t *testing.T) {
	chdir(t, t.TempDir())

	got, err := buildStandardImpactEvidence()
	if err != nil {
		t.Fatal(err)
	}

	if got.ReportPath != standardImpactReportPath {
		t.Fatalf("report_path = %q, want %q", got.ReportPath, standardImpactReportPath)
	}
	if got.Status != "missing" {
		t.Fatalf("status = %q, want missing", got.Status)
	}
	if got.ReportSHA256 != "" {
		t.Fatalf("report_sha256 = %q, want empty", got.ReportSHA256)
	}
	if got.DownstreamSyncRequired {
		t.Fatal("downstream_sync_required = true, want false")
	}
	if got.ContextRuntimeChange {
		t.Fatal("context_runtime_change = true, want false")
	}
	if got.GovernanceRegistryChange {
		t.Fatal("governance_registry_change = true, want false")
	}
	if got.DownstreamReleaseDecision != "not_required" {
		t.Fatalf("downstream_release_decision = %q, want not_required", got.DownstreamReleaseDecision)
	}
	if got.RepositoryRulesReleaseDecision != "not_required" {
		t.Fatalf("repository_rules_release_decision = %q, want not_required", got.RepositoryRulesReleaseDecision)
	}
}

func TestBuildStandardImpactEvidenceReadsReport(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, filepath.FromSlash(standardImpactReportPath))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := strings.Join([]string{
		"# Standard Impact",
		"",
		"- downstream_sync_required: `true`",
		"- context_runtime_change: `true`",
		"- governance_registry_change: `true`",
		"- downstream_release_decision: `required`",
		"- repository_rules_release_decision: `audit_required`",
		"- primary_downstream: `github.com/ZoneCNH/kernel`",
		"",
	}, "\n")
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)

	got, err := buildStandardImpactEvidence()
	if err != nil {
		t.Fatal(err)
	}

	if got.Status != "present" {
		t.Fatalf("status = %q, want present", got.Status)
	}
	if !strings.HasPrefix(got.ReportSHA256, "sha256:") {
		t.Fatalf("report_sha256 = %q, want sha256 prefix", got.ReportSHA256)
	}
	if !got.DownstreamSyncRequired {
		t.Fatal("downstream_sync_required = false, want true")
	}
	if !got.ContextRuntimeChange {
		t.Fatal("context_runtime_change = false, want true")
	}
	if !got.GovernanceRegistryChange {
		t.Fatal("governance_registry_change = false, want true")
	}
	if got.DownstreamReleaseDecision != "required" {
		t.Fatalf("downstream_release_decision = %q, want required", got.DownstreamReleaseDecision)
	}
	if got.RepositoryRulesReleaseDecision != "audit_required" {
		t.Fatalf("repository_rules_release_decision = %q, want audit_required", got.RepositoryRulesReleaseDecision)
	}
	if got.PrimaryDownstream != "github.com/ZoneCNH/kernel" {
		t.Fatalf("primary_downstream = %q, want github.com/ZoneCNH/kernel", got.PrimaryDownstream)
	}
}

func TestBuildStandardImpactEvidenceReportsReadError(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, filepath.FromSlash(standardImpactReportPath))
	if err := os.MkdirAll(reportPath, 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)

	if _, err := buildStandardImpactEvidence(); err == nil {
		t.Fatal("buildStandardImpactEvidence succeeded for directory report, want error")
	}
}

func TestBuildDebtEvidenceAllowsMissingReport(t *testing.T) {
	chdir(t, t.TempDir())

	got, err := buildDebtEvidence()
	if err != nil {
		t.Fatal(err)
	}

	if got.ReportPath != debtReportPath || got.MarkdownPath != debtMarkdownPath || got.ChecksumPath != debtChecksumPath {
		t.Fatalf("paths = %+v, want debt evidence paths", got)
	}
	if got.Status != "missing" {
		t.Fatalf("status = %q, want missing", got.Status)
	}
	if got.MinScore != 9.8 {
		t.Fatalf("min_score = %.1f, want 9.8", got.MinScore)
	}
}

func TestBuildDebtEvidenceReadsReport(t *testing.T) {
	root := t.TempDir()
	writeDebtReportFixture(t, root)
	chdir(t, root)

	got, err := buildDebtEvidence()
	if err != nil {
		t.Fatal(err)
	}

	if got.Status != "passed" || got.Score != 9.8 || got.MinScore != 9.8 || got.CheckCount != 1 {
		t.Fatalf("debt evidence = %+v, want passed score and one check", got)
	}
	if !strings.HasPrefix(got.ReportSHA256, "sha256:") {
		t.Fatalf("report_sha256 = %q, want sha256 prefix", got.ReportSHA256)
	}
}

func TestBuildDebtEvidenceReportsReadError(t *testing.T) {
	root := t.TempDir()
	reportPath := filepath.Join(root, filepath.FromSlash(debtReportPath))
	if err := os.MkdirAll(reportPath, 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, root)

	if _, err := buildDebtEvidence(); err == nil {
		t.Fatal("buildDebtEvidence succeeded for directory report, want error")
	}
}

func TestParseReportValueHandlesMissingAndFirstMatch(t *testing.T) {
	report := "- other: `ignored`\n- primary_downstream: ` first value `\n- primary_downstream: `second value`\n"
	if got := parseReportValue(report, "primary_downstream"); got != "first value" {
		t.Fatalf("parseReportValue first match = %q, want first value", got)
	}
	if got := parseReportValue(report, "missing"); got != "" {
		t.Fatalf("parseReportValue missing = %q, want empty", got)
	}
}

func TestBuildGeneratorEvidenceRecordsRepresentativeDownstreams(t *testing.T) {
	got := buildGeneratorEvidence()

	if got.Command != "GOWORK=off make integration" {
		t.Fatalf("command = %q, want integration command", got.Command)
	}
	if !got.Required {
		t.Fatal("required = false, want true")
	}
	if !hasGeneratorTarget(got.Targets, "kernel", "github.com/ZoneCNH/kernel", "kernel") {
		t.Fatalf("targets = %+v, want kernel target", got.Targets)
	}
	if !hasGeneratorTarget(got.Targets, "configx", "github.com/ZoneCNH/configx", "configx") {
		t.Fatalf("targets = %+v, want configx target", got.Targets)
	}
	if !hasGeneratorTarget(got.Targets, "redisx", "github.com/ZoneCNH/redisx", "redisx") {
		t.Fatalf("targets = %+v, want redisx target", got.Targets)
	}
}

func TestBuildGovernanceRuntimeEvidenceRecordsVersionsAndPassedStatuses(t *testing.T) {
	got := buildGovernanceRuntimeEvidence()

	assertGovernanceRuntimeEvidence(t, got)

	got.GateStatuses["governance"] = "failed"
	got.ProfileStatuses["p1_governance"] = "failed"
	if governanceRuntimeGateStatuses["governance"] != "passed" {
		t.Fatalf("gate status fixture was mutated: %v", governanceRuntimeGateStatuses)
	}
	if governanceRuntimeProfileStatuses["p1_governance"] != "passed" {
		t.Fatalf("profile status fixture was mutated: %v", governanceRuntimeProfileStatuses)
	}
}

func TestValidateGovernanceRuntimeEvidenceRequiresVersionsAndPassedStatuses(t *testing.T) {
	got := buildGovernanceRuntimeEvidence()
	got.SchemaVersion = "v2.9.2"
	got.RuntimeVersion = ""
	got.GateStatuses["governance"] = "failed"
	delete(got.ProfileStatuses, "p2_runtime")

	failures := validateGovernanceRuntimeEvidence(got)
	for _, want := range []string{
		`governance_runtime.schema_version must be "v2.9.3", got "v2.9.2"`,
		`governance_runtime.runtime_version must be "v2.9.3", got ""`,
		`governance_runtime.gate_statuses.governance must be passed, got "failed"`,
		"governance_runtime.profile_statuses.p2_runtime is required",
	} {
		if !contains(failures, want) {
			t.Fatalf("failures = %v, want %q", failures, want)
		}
	}
}

func TestValidateChecksRequiresStatusButOnlyRequiresPassedWhenRequested(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = "failed"
	}
	checks["fmt"] = " "

	failures := validateChecks(checks, false)
	if len(failures) != 1 || !strings.Contains(failures[0], "checks.fmt is required") {
		t.Fatalf("failures = %v, want only missing fmt", failures)
	}
}

func TestVerifyManifestRequiresCleanTree(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	repo := releaseManifestFixtureRepo(t)
	writeStandardImpactReportFixture(t, repo)
	chdir(t, repo)

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.TreeState = "dirty"

	path := filepath.Join(t.TempDir(), "dirty.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, true, true, "", 0)
	if err == nil {
		t.Fatal("verify dirty manifest with requireClean succeeded, want error")
	}
	if !strings.Contains(err.Error(), `tree_state must be clean, got "dirty"`) {
		t.Fatalf("error = %q, want require-clean failure", err)
	}
}

func TestSourceDigestUsesTrackedFileNamesAndContents(t *testing.T) {
	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")

	files := map[string]string{
		"a.txt":          "alpha\n",
		"nested/b.txt":   "bravo\n",
		"nested/cfg.yml": "name: charlie\n",
	}
	for path, content := range files {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTestCommand(t, repo, "git", "add", ".")
	chdir(t, repo)

	gotDigest, gotCount, err := sourceDigest()
	if err != nil {
		t.Fatal(err)
	}

	if gotCount != len(files) {
		t.Fatalf("tracked file count = %d, want %d", gotCount, len(files))
	}
	if want := expectedSourceDigest(files); gotDigest != want {
		t.Fatalf("source digest = %q, want %q", gotDigest, want)
	}
}

func TestSourceDigestReportsGitAndTrackedFileReadErrors(t *testing.T) {
	withRunRawCommand(t, func(name string, args ...string) ([]byte, error) {
		if name == "git" && strings.Join(args, " ") == "ls-files -z" {
			return nil, errors.New("git ls-files failed")
		}
		return runRaw(name, args...)
	})
	if _, _, err := sourceDigest(); err == nil || !strings.Contains(err.Error(), "git ls-files failed") {
		t.Fatalf("sourceDigest git error = %v, want git ls-files failed", err)
	}

	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	path := filepath.Join(repo, "tracked.txt")
	if err := os.WriteFile(path, []byte("tracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repo, "git", "add", ".")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	chdir(t, repo)

	withRunRawCommand(t, runRaw)
	if _, _, err := sourceDigest(); err == nil {
		t.Fatal("sourceDigest succeeded for missing tracked file, want error")
	}
}

func TestModuleDigestsIncludesReplaceMetadata(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(`module example.com/root

go 1.23

require example.com/dep v0.0.0

replace example.com/dep => ./dep
`), 0o644); err != nil {
		t.Fatal(err)
	}
	depDir := filepath.Join(root, "dep")
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(depDir, "go.mod"), []byte("module example.com/dep\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOWORK", "off")
	chdir(t, root)

	modules, err := moduleDigests()
	if err != nil {
		t.Fatal(err)
	}

	var foundMain bool
	var foundReplace bool
	for _, module := range modules {
		if module.Path == "example.com/root" && module.Main {
			foundMain = true
		}
		if module.Path == "example.com/dep" && module.Replace != nil && module.Replace.Path == "./dep" {
			foundReplace = true
		}
	}
	if !foundMain {
		t.Fatalf("modules = %+v, want main module", modules)
	}
	if !foundReplace {
		t.Fatalf("modules = %+v, want replace metadata for example.com/dep", modules)
	}
}

func TestModuleDigestsReportsCommandAndDecodeErrors(t *testing.T) {
	withRunRawCommand(t, func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("go list modules failed")
	})
	if _, err := moduleDigests(); err == nil || !strings.Contains(err.Error(), "go list modules failed") {
		t.Fatalf("moduleDigests command error = %v, want go list modules failed", err)
	}

	withRunRawCommand(t, func(name string, args ...string) ([]byte, error) {
		return []byte("{"), nil
	})
	if _, err := moduleDigests(); err == nil {
		t.Fatal("moduleDigests succeeded for malformed JSON, want error")
	}
}

func TestWriteManifestCreatesParentAndWritesIndentedJSON(t *testing.T) {
	manifest := Manifest{
		Module:           "example.com/lib",
		Version:          "v1.2.3",
		Commit:           "abc123",
		TreeSHA:          "tree123",
		SourceDigest:     "sha256:source",
		TrackedFileCount: 1,
		GoVersion:        "go1.23.0",
		GeneratedAt:      "2026-01-02T03:04:05Z",
		GeneratedBy:      "test",
		TreeState:        "clean",
		Checks:           map[string]string{"fmt": "passed"},
		Tools:            map[string]string{"go": "go version"},
		Artifacts:        append([]string(nil), requiredArtifacts...),
	}
	path := filepath.Join(t.TempDir(), "release", "manifest", "latest.json")

	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatalf("manifest JSON is invalid: %s", data)
	}
	if !strings.Contains(string(data), "\n  ") {
		t.Fatalf("manifest JSON is not indented: %s", data)
	}

	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Module != manifest.Module || got.Version != manifest.Version {
		t.Fatalf("round-trip manifest = %+v, want %+v", got, manifest)
	}
}

func TestWriteManifestReportsMkdirAndEncodeErrors(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeManifest(filepath.Join(blocker, "latest.json"), Manifest{}); err == nil {
		t.Fatal("writeManifest succeeded with file parent, want error")
	}
	if err := encodeManifest(errorWriter{}, Manifest{}); err == nil {
		t.Fatal("encodeManifest succeeded with failing writer, want error")
	}

	previousEncode := encodeManifestFunc
	encodeManifestFunc = func(io.Writer, Manifest) error {
		return errors.New("encode failed")
	}
	t.Cleanup(func() {
		encodeManifestFunc = previousEncode
	})
	if err := writeManifest(filepath.Join(t.TempDir(), "latest.json"), Manifest{}); err == nil || !strings.Contains(err.Error(), "encode failed") {
		t.Fatalf("writeManifest encode error = %v, want encode failed", err)
	}
}

func TestToolVersionReportsMissingBinary(t *testing.T) {
	got := toolVersion("definitely-missing-releasemanifest-test-binary")
	if got != "missing" {
		t.Fatalf("toolVersion missing binary = %q, want missing", got)
	}
}

func TestWorkflowEvidenceBuildsGitHubAndLocalURLs(t *testing.T) {
	server := "https://github.com/"
	repo := "/ZoneCNH/xlib-standard/"
	t.Setenv("WORKFLOW_RUN_ID", "123")
	t.Setenv("GITHUB_SERVER_URL", server)
	t.Setenv("GITHUB_REPOSITORY", repo)
	t.Setenv("ARTIFACT_NAME", "")
	t.Setenv("ARTIFACT_URL", "")

	got := buildWorkflowEvidence()
	if got.WorkflowRunID != "123" {
		t.Fatalf("workflow_run_id = %q, want 123", got.WorkflowRunID)
	}
	if got.ArtifactName != "release-manifest-123" {
		t.Fatalf("artifact_name = %q, want release-manifest-123", got.ArtifactName)
	}
	wantArtifactURL := strings.TrimRight(server, "/") + "/" + strings.Trim(repo, "/") + "/actions/runs/123"
	if got.ArtifactURL != wantArtifactURL {
		t.Fatalf("artifact_url = %q, want GitHub Actions URL", got.ArtifactURL)
	}

	t.Setenv("WORKFLOW_RUN_ID", "")
	t.Setenv("GITHUB_RUN_ID", "")
	t.Setenv("GITHUB_SERVER_URL", "")
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("ARTIFACT_NAME", "")
	t.Setenv("ARTIFACT_URL", "")
	got = buildWorkflowEvidence()
	if got.WorkflowRunID != "local" || got.ArtifactURL != "local:release-manifest-local" {
		t.Fatalf("local workflow evidence = %+v, want local fallback", got)
	}
}

func TestTreeStateBranches(t *testing.T) {
	withRunRawCommand(t, func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("git status failed")
	})
	if got := treeState(); got != "unknown" {
		t.Fatalf("treeState error = %q, want unknown", got)
	}

	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	runTestCommand(t, repo, "git", "config", "user.email", "test@example.com")
	runTestCommand(t, repo, "git", "config", "user.name", "Release Manifest Test")
	if err := os.WriteFile(filepath.Join(repo, "clean.txt"), []byte("clean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repo, "git", "add", ".")
	runTestCommand(t, repo, "git", "commit", "-m", "fixture")
	chdir(t, repo)

	withRunRawCommand(t, runRaw)
	if got := treeState(); got != "clean" {
		t.Fatalf("treeState clean = %q, want clean", got)
	}
}

func TestToolVersionReportsCommandError(t *testing.T) {
	dir := t.TempDir()
	name := "releasemanifest-failing-tool"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho first line\necho second line\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	got := toolVersion(name)
	if !strings.HasPrefix(got, "error: ") {
		t.Fatalf("toolVersion failing binary = %q, want error prefix", got)
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("toolVersion failing binary = %q, want first line only", got)
	}
}

func TestRequireNonEmptyAppendsOnlyForBlankValues(t *testing.T) {
	failures := []string{"existing"}
	requireNonEmpty(&failures, "blank", " \t")
	requireNonEmpty(&failures, "filled", "value")

	if strings.Join(failures, "|") != "existing|blank is required" {
		t.Fatalf("failures = %v, want blank appended only", failures)
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

func withRunRawCommand(t *testing.T, runner func(name string, args ...string) ([]byte, error)) {
	t.Helper()

	previous := runRawCommand
	runRawCommand = runner
	t.Cleanup(func() {
		runRawCommand = previous
	})
}

func runTestCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
}

func releaseManifestFixtureRepo(t *testing.T) string {
	t.Helper()

	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/releasefixture\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range contractFiles {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		content := "{}\n"
		if strings.HasSuffix(path, ".md") {
			content = "# Fixture Contract\n"
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFixtureDebtPolicy(t, repo)
	writeFixtureOMCState(t, repo)
	writeDebtReportFixture(t, repo)
	runTestCommand(t, repo, "git", "add", ".")
	return repo
}

func writeFixtureDebtPolicy(t *testing.T, repo string) {
	t.Helper()
	files := map[string]string{
		".agent/debt/rules.yaml":              "schema_version: debt-rules/v1\nprofile: fixture\n",
		".agent/debt/rule-registry.yaml":      "schema_version: debt-rule-registry/v1\nrules: []\n",
		".agent/debt/exceptions.yaml":         "schema_version: debt-exceptions/v1\nexceptions: []\n",
		".agent/debt/dependency-purpose.yaml": "schema_version: debt-dependency-purpose/v1\npurposes: []\n",
	}
	for path, content := range files {
		fullPath := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeStandardImpactReportFixture(t *testing.T, repo string) {
	t.Helper()

	reportPath := filepath.Join(repo, filepath.FromSlash(standardImpactReportPath))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := strings.Join([]string{
		"# Standard Impact",
		"",
		"- downstream_sync_required: `true`",
		"- context_runtime_change: `true`",
		"- governance_registry_change: `true`",
		"- downstream_release_decision: `required`",
		"- repository_rules_release_decision: `audit_required`",
		"- primary_downstream: `github.com/ZoneCNH/kernel`",
		"",
	}, "\n")
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeDebtReportFixture(t *testing.T, repo string) {
	t.Helper()

	reportPath := filepath.Join(repo, filepath.FromSlash(debtReportPath))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{
  "schema_version": "1.0",
  "generated_at": "2026-06-02T00:00:00Z",
  "status": "passed",
  "score": 9.8,
  "min_score": 9.8,
  "policy_path": ".agent/debt/rules.yaml",
  "checks": [
    {"id": "policy", "status": "passed"}
  ],
  "downstream_targets": ["kernel/configx", "kernel/redisx", "corekit"]
}` + "\n")
	if err := os.WriteFile(reportPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	markdownPath := filepath.Join(repo, filepath.FromSlash(debtMarkdownPath))
	if err := os.WriteFile(markdownPath, []byte("# Debt Evidence\n\n- status: passed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:]) + "  release/debt/latest.json\n"
	checksumPath := filepath.Join(repo, filepath.FromSlash(debtChecksumPath))
	if err := os.WriteFile(checksumPath, []byte(checksum), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFixtureOMCState(t *testing.T, repo string) {
	t.Helper()

	stateDir := filepath.Join(repo, ".omc", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(stateDir, "agent-replay-fixture.jsonl")
	content := []byte(`{"event":"fixture","source":"releasemanifest-test"}` + "\n")
	if err := os.WriteFile(statePath, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func expectedSourceDigest(files map[string]string) string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	digest := sha256.New()
	for _, path := range paths {
		sum := sha256.Sum256([]byte(files[path]))
		digest.Write([]byte(path))
		digest.Write([]byte{0})
		digest.Write([]byte(hex.EncodeToString(sum[:])))
		digest.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}

func hasGeneratorTarget(targets []GeneratorTarget, name string, modulePath string, packageName string) bool {
	for _, target := range targets {
		if target.Name == name && target.ModulePath == modulePath && target.PackageName == packageName {
			return true
		}
	}
	return false
}

func assertGovernanceRuntimeEvidence(t *testing.T, got GovernanceRuntimeEvidence) {
	t.Helper()

	if got.SchemaVersion != governanceRuntimeVersion {
		t.Fatalf("governance_runtime.schema_version = %q, want %q", got.SchemaVersion, governanceRuntimeVersion)
	}
	if got.RuntimeVersion != governanceRuntimeVersion {
		t.Fatalf("governance_runtime.runtime_version = %q, want %q", got.RuntimeVersion, governanceRuntimeVersion)
	}
	assertMapHas(t, got.GateStatuses, "governance", "passed")
	assertMapHas(t, got.ProfileStatuses, "p1_governance", "passed")
	assertMapHas(t, got.ProfileStatuses, "p2_runtime", "passed")
}

func assertMapHas(t *testing.T, got map[string]string, key string, want string) {
	t.Helper()

	if got[key] != want {
		t.Fatalf("map[%q] = %q, want %q (map: %v)", key, got[key], want, got)
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
