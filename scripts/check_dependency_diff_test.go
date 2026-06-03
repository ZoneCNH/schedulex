package scripts_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDependencyAutomationPolicyAccepted(t *testing.T) {
	repo := newDependencyGateRepo(t)

	result := runDependencyGate(t, repo)
	if result.code != 0 {
		t.Fatalf("expected dependency gate to pass, got exit %d\noutput:\n%s", result.code, result.output)
	}

	for _, want := range []string{
		"checking dependency automation...",
		"Go module dependencies:",
		"GitHub Actions dependencies:",
		"dependency_surface_changed=false",
		"standard_contract_generator_review_required=false",
		"dependency automation check passed",
	} {
		if !strings.Contains(result.output, want) {
			t.Fatalf("expected output to contain %q\noutput:\n%s", want, result.output)
		}
	}
}

func TestDependencyAutomationPolicyRejectsMissingHumanReview(t *testing.T) {
	repo := newDependencyGateRepo(t)
	renovatePath := filepath.Join(repo, "renovate.json")
	renovate := readFile(t, renovatePath)
	renovate = strings.Replace(renovate, `      "dependencyDashboardApproval": true,`+"\n", "", 1)
	writeFile(t, renovatePath, renovate)

	result := runDependencyGate(t, repo)
	if result.code == 0 {
		t.Fatalf("expected dependency gate to fail when major updates lack dashboard approval\noutput:\n%s", result.output)
	}

	want := `ERROR: renovate.json must mention: "dependencyDashboardApproval": true`
	if !strings.Contains(result.output, want) {
		t.Fatalf("expected output to contain %q\noutput:\n%s", want, result.output)
	}
}

func TestDependencySurfaceChangeTriggersImpactOutput(t *testing.T) {
	repo := newDependencyGateRepo(t)
	writeFile(t, filepath.Join(repo, "renovate.json"), readFile(t, filepath.Join(repo, "renovate.json"))+"\n")

	result := runDependencyGate(t, repo)
	if result.code != 0 {
		t.Fatalf("expected dependency gate to pass with dependency file change, got exit %d\noutput:\n%s", result.code, result.output)
	}

	for _, want := range []string{
		"Changed files considered for dependency governance:",
		"  - renovate.json",
		"dependency_surface_changed=true",
	} {
		if !strings.Contains(result.output, want) {
			t.Fatalf("expected output to contain %q\noutput:\n%s", want, result.output)
		}
	}
}

type commandResult struct {
	code   int
	output string
}

func newDependencyGateRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "scripts"))
	mkdirAll(t, filepath.Join(root, ".github", "workflows"))
	copyFile(t, filepath.Join("check_dependency_diff.sh"), filepath.Join(root, "scripts", "check_dependency_diff.sh"), 0o755)
	copyFile(t, filepath.Join("..", "renovate.json"), filepath.Join(root, "renovate.json"), 0o644)
	copyFile(t, filepath.Join("..", ".github", "dependabot.yml"), filepath.Join(root, ".github", "dependabot.yml"), 0o644)
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/dependency-gate\n\ngo 1.23\n")
	writeFile(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "name: ci\non: [push]\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n")

	runGit(t, root, "init", "-b", "main")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Dependency Gate Test")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "initial dependency gate fixture")

	return root
}

func runDependencyGate(t *testing.T, repo string) commandResult {
	t.Helper()

	cmd := exec.Command("bash", "scripts/check_dependency_diff.sh")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return commandResult{code: 0, output: string(output)}
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("dependency gate failed without exit status: %v\noutput:\n%s", err, output)
	}
	return commandResult{code: exitErr.ExitCode(), output: string(output)}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\noutput:\n%s", strings.Join(args, " "), err, output)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create directory %s: %v", path, err)
	}
}

func copyFile(t *testing.T, from, to string, mode os.FileMode) {
	t.Helper()

	contents := readFile(t, from)
	if err := os.WriteFile(to, []byte(contents), mode); err != nil {
		t.Fatalf("write %s: %v", to, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
