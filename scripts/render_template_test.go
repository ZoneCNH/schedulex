package scripts_test

import (
	"os"
	"strings"
	"testing"
)

func TestRenderTemplateExcludesGeneratedDebtArtifacts(t *testing.T) {
	contents, err := os.ReadFile("render_template.sh")
	if err != nil {
		t.Fatalf("read render_template.sh: %v", err)
	}

	script := string(contents)
	for _, exclude := range []string{
		"--exclude='./release/debt/latest.json'",
		"--exclude='./release/debt/latest.md'",
		"--exclude='./release/debt/latest.json.sha256'",
	} {
		if !strings.Contains(script, exclude) {
			t.Fatalf("render_template.sh missing generated debt artifact exclude %q", exclude)
		}
	}
}
