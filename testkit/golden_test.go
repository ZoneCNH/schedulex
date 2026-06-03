package testkit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRequireGoldenAcceptsMatchingContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.golden")

	if err := os.WriteFile(path, []byte("ok\n"), 0o600); err != nil {
		t.Fatalf("write golden: %v", err)
	}

	RequireGolden(t, path, []byte("ok\n"))
}

func TestRequireGoldenReportsReadError(t *testing.T) {
	tb := newRecordingTB()

	expectFatal(t, func() {
		requireGolden(tb, func(string) ([]byte, error) {
			return nil, errors.New("missing")
		}, "missing.golden", []byte("actual"))
	})

	if !tb.helperCalled {
		t.Fatal("expected Helper to be called")
	}
	if tb.message != "read golden file missing.golden: missing" {
		t.Fatalf("unexpected fatal message: %q", tb.message)
	}
}

func TestRequireGoldenReportsMismatch(t *testing.T) {
	tb := newRecordingTB()

	expectFatal(t, func() {
		requireGolden(tb, func(string) ([]byte, error) {
			return []byte("expected\n"), nil
		}, "sample.golden", []byte("actual\n"))
	})

	if !tb.helperCalled {
		t.Fatal("expected Helper to be called")
	}
	want := "golden mismatch for sample.golden\nexpected:\nexpected\n\nactual:\nactual\n"
	if tb.message != want {
		t.Fatalf("unexpected fatal message:\n%s", tb.message)
	}
}
