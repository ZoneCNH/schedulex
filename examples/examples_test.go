package examples_test

import (
	"os/exec"
	"testing"
)

func TestExamplesCompile(t *testing.T) {
	for _, dir := range []string{"once", "interval", "cron_like", "daily_at", "misfire", "shutdown", "lock_interface"} {
		t.Run(dir, func(t *testing.T) {
			cmd := exec.Command("go", "run", "./"+dir)
			cmd.Dir = "."
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("example %s failed: %v\n%s", dir, err, out)
			}
		})
	}
}
