package testkit

import "testing"

type fatalHelper interface {
	Helper()
	Fatalf(format string, args ...any)
}

func RequireNoError(t testing.TB, err error) {
	requireNoError(t, err)
}

func requireNoError(t fatalHelper, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
