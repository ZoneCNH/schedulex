package testkit

import (
	"os"
	"path/filepath"
	"testing"
)

func RequireGolden(t testing.TB, goldenPath string, actual []byte) {
	requireGolden(t, os.ReadFile, goldenPath, actual)
}

func requireGolden(t fatalHelper, readFile func(string) ([]byte, error), goldenPath string, actual []byte) {
	t.Helper()

	expected, err := readFile(filepath.Clean(goldenPath))
	if err != nil {
		t.Fatalf("read golden file %s: %v", goldenPath, err)
	}

	if string(expected) != string(actual) {
		t.Fatalf(
			"golden mismatch for %s\nexpected:\n%s\nactual:\n%s",
			goldenPath,
			expected,
			actual,
		)
	}
}
