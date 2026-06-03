package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ZoneCNH/xlib-standard/internal/releasequality"
)

func main() {
	exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

var exit = os.Exit

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		write(stderr, usage)
		return 2
	}
	switch args[0] {
	case "version":
		return runVersion(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "main-guard":
		return runMainGuard(args[1:], stdout, stderr)
	case "worktree-guard":
		return runWorktreeGuard(args[1:], stdout, stderr)
	case "worktree-check":
		return runWorktreeCheck(args[1:], stdout, stderr)
	case "context-check":
		return runContextCheck(args[1:], stdout, stderr)
	case "spec-check":
		return runSpecCheck(args[1:], stdout, stderr)
	case "design-check":
		return runDesignCheck(args[1:], stdout, stderr)
	case "task-check":
		return runTaskCheck(args[1:], stdout, stderr)
	case "pr-check":
		return runPRCheck(args[1:], stdin, stdout, stderr)
	case "evidence-check":
		return runEvidenceCheck(args[1:], stdout, stderr)
	case "cli-contract":
		return runCLIContract(args[1:], stdout, stderr)
	case "issue-registry":
		return runIssueRegistry(args[1:], stdout, stderr)
	case "command-registry":
		return runCommandRegistry(args[1:], stdout, stderr)
	case "makefile-baseline":
		return runMakefileBaseline(args[1:], stdout, stderr)
	case "context-profile":
		return runContextProfile(args[1:], stdout, stderr)
	case "context-profile-check":
		return runContextProfileCheck("context-profile-check", args[1:], stdout, stderr)
	case "context-schema-check":
		return runContextProfileCheck("context-schema-check", args[1:], stdout, stderr)
	case "context-lite", "context-standard", "context-full", "context-release", "context-fast-check", "context-standard-check", "context-full-check":
		return runContextProfileAlias(args[0], args[1:], stdout, stderr)
	case "debt":
		return runDebt(args[1:], stdout, stderr)
	case "architecture":
		return runDebtAlias("architecture", "enforce", args[1:], stdout, stderr)
	case "domain":
		return runDebtAlias("domain", "enforce", args[1:], stdout, stderr)
	case "docs-drift":
		return runDebtAlias("docs", "warn", args[1:], stdout, stderr)
	case "dependency-debt":
		return runDebtAlias("dependency", "warn", args[1:], stdout, stderr)
	case "testing-debt":
		return runDebtAlias("testing", "warn", args[1:], stdout, stderr)
	case "implementation-debt":
		return runDebtAlias("implementation", "observe", args[1:], stdout, stderr)
	case "security-debt":
		return runDebtAlias("security", "warn", args[1:], stdout, stderr)
	case "downstream-debt":
		return runDebtAlias("downstream", "warn", args[1:], stdout, stderr)
	case "goal-acceptance", "goal-delivery", "goal-handover", "goal-downstream-adoption", "goal-certify", "goal-runtime-final":
		return runGoalRuntimeCommand(args[0], args[1:], stdout, stderr)
	case "minimal-kernel", "done-assertion", "agent-team-contract", "scope-lock", "pr-template", "acceptance-matrix", "runtime-health", "goal-runtime", "naming", "upgrade-standard", "conformance-profile", "downstream-registry", "self-healing-skeleton", "policy-schema", "github-settings", "toolchain", "evidence-artifacts", "install-runtime", "upgrade-runtime", "release-ready", "evidence-replay", "attest-conformance", "pack-standard", "pack-gate", "pack-evidence", "runtime-file-ownership", "downstream-baseline", "downstream-adoption", "autoresearch", "changelog", "github-governance", "governance-fixture-test", "supply-chain", "execution-context":
		return runPlannedCommand(args[0], args[1:], stdout, stderr)
	case "boundary":
		return runExternal(stdin, stdout, stderr, "./scripts/check_boundary.sh")
	case "contracts":
		return runExternal(stdin, stdout, stderr, "./scripts/check_contracts.sh")
	case "debt-evidence":
		return runDebtEvidence(args[1:], stdout, stderr)
	case "debt-evidence-checksum-check":
		return runExternal(stdin, stdout, stderr, "./scripts/hash_release_evidence.sh", "--check", "release/debt/latest.json", "release/debt/latest.json.sha256")
	case "debt-evidence-hash":
		return runExternal(stdin, stdout, stderr, "./scripts/hash_release_evidence.sh", "release/debt/latest.json", "release/debt/latest.json.sha256")
	case "dependency-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_dependency_diff.sh")
	case "docs-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_docs.sh")
	case "evidence", "manifest":
		return runExternal(stdin, stdout, stderr, "go", "run", "./internal/tools/releasemanifest", "--out", "release/manifest/latest.json")
	case "integration":
		return runExternal(stdin, stdout, stderr, "./scripts/run_integration.sh")
	case "release-evidence-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_release_evidence.sh")
	case "release-evidence-checksum-check":
		return runExternal(stdin, stdout, stderr, "./scripts/hash_release_evidence.sh", "--check")
	case "release-evidence-hash":
		return runExternal(stdin, stdout, stderr, "./scripts/hash_release_evidence.sh")
	case "release-final-check":
		return runExternal(stdin, stdout, stderr, "make", "release-final-check")
	case "render-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_rendered_template.sh", args[1:]...)
	case "rules-consistency-check":
		return runRulesConsistencyCheck(args[1:], stdout, stderr)
	case "rules-verify":
		return runExternal(stdin, stdout, stderr, "python3", "scripts/verify_rules.py")
	case "score":
		return runScore(args[1:], stdout, stderr)
	case "secrets":
		return runExternal(stdin, stdout, stderr, "./scripts/check_secrets.sh")
	case "security":
		return runExternalSequence(stdin, stdout, stderr,
			externalCommand{name: "govulncheck", args: []string{"./..."}},
			externalCommand{name: "./scripts/check_secrets.sh"},
		)
	case "standard-impact-check":
		return runExternal(stdin, stdout, stderr, "./scripts/check_standard_impact.sh")
	case "self-improving-check", "retro-check":
		return runSelfImprovingCheck(args[0], args[1:], stdout, stderr)
	case "traceability-check":
		return runTraceabilityCheck(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		write(stdout, usage)
		return 0
	default:
		write(stderr, "unknown command %q\n", args[0])
		return 2
	}
}

func runScore(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli score", flag.ContinueOnError)
	flags.SetOutput(stderr)
	minimum := flags.Float64("min", releasequality.DefaultMinimum, "minimum acceptable release score")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	report := computeReleaseQuality(*minimum)
	data, err := marshalReleaseQuality(report)
	if err != nil {
		write(stderr, "ERROR: %v\n", err)
		return 1
	}
	write(stdout, "%s\n", data)
	if err := verifyReleaseQuality(report, *minimum); err != nil {
		write(stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}

var (
	computeReleaseQuality = releasequality.Compute
	marshalReleaseQuality = releasequality.Marshal
	verifyReleaseQuality  = releasequality.Verify
)

func runExternal(stdin io.Reader, stdout io.Writer, stderr io.Writer, name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		write(stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}

type externalCommand struct {
	name string
	args []string
}

func runExternalSequence(stdin io.Reader, stdout io.Writer, stderr io.Writer, commands ...externalCommand) int {
	for _, command := range commands {
		if code := runExternal(stdin, stdout, stderr, command.name, command.args...); code != 0 {
			return code
		}
	}
	return 0
}

func write(writer io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(writer, format, args...)
}

const usage = `usage: goalcli <command> [args]

commands:
  agent-team-contract [--dry-run]
  acceptance-matrix
  architecture [debt args]
  attest-conformance [--profile <name>]
  autoresearch
  boundary
  changelog
  cli-contract [--json|--output <path>|--explain]
  command-registry
  conformance-profile [--profile <name>]
  context-check [--json]
  context-fast-check
  context-full
  context-full-check
  context-lite
  context-profile [--profile <name>] [--json]
  context-profile-check [--json]
  context-release
  context-schema-check [--json]
  context-standard
  context-standard-check
  contracts
  debt [--config <path>] [--section <name>] [--mode <enforce|warn|observe>] [--min-score <score>] [--output json|markdown]
  debt lifecycle-check [--output <path>]
  debt patch-suggest [--output <path>]
  debt register-update [--output <path>]
  debt trend [--output <path>]
  debt-evidence
  debt-evidence-checksum-check
  debt-evidence-hash
  dependency-debt [debt args]
  dependency-check
  docs-drift [debt args]
  domain [debt args]
  doctor [--json]
  docs-check
  design-check [--json]
  downstream-adoption
  downstream-baseline
  downstream-registry
  downstream-debt [debt args]
  evidence
  evidence-artifacts
  evidence-check
  evidence-replay
  execution-context
  github-governance
  github-settings [--verify]
  goal-acceptance [--goal-id <id>] [--json]
  goal-delivery [--goal-id <id>] [--json]
  goal-handover [--goal-id <id>] [--json]
  goal-downstream-adoption [--goal-id <id>] [--json]
  goal-certify [--goal-id <id>] [--json]
  goal-runtime
  goal-runtime-final [--goal-id <id>] [--json] [--write-evidence]
  governance-fixture-test
  install-runtime [--dry-run]
  integration
  implementation-debt [debt args]
  issue-registry
  main-guard [--context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify]
  makefile-baseline
  manifest
  minimal-kernel
  done-assertion
  naming
  pack-evidence
  pack-gate
  pack-standard
  policy-schema
  pr-check [--context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify] [--dry-run] [--json]
  pr-template
  release-evidence-check
  release-evidence-checksum-check
  release-evidence-hash
  release-final-check
  release-ready
  render-check <rendered-dir>
  retro-check [--root <path>] [--strict]
  rules-consistency-check
  rules-verify
  runtime-file-ownership
  runtime-health
  scope-lock
  score [--min <score>]
  secrets
  security
  security-debt [debt args]
  self-improving-check [--root <path>] [--strict]
  self-healing-skeleton
  spec-check [--json]
  standard-impact-check
  supply-chain
  task-check [--json]
  toolchain
  testing-debt [debt args]
  traceability-check [--matrix .agent/traceability-matrix.md] [--json]
  upgrade-runtime [--dry-run]
  upgrade-standard [--dry-run]
  version [--json]
  worktree-check [--context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify]
  worktree-guard [--context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify]
`
