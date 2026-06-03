package sanitize

import "testing"

func TestSecretMasksNonEmptyValue(t *testing.T) {
	if got := Secret("plain-text"); got != "***" {
		t.Fatalf("expected masked value, got %q", got)
	}
}

func TestSecretPreservesEmptyValue(t *testing.T) {
	if got := Secret(""); got != "" {
		t.Fatalf("expected empty value, got %q", got)
	}
}
