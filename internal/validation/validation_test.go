package validation

import "testing"

func TestRequireNonEmptyRejectsEmptyValue(t *testing.T) {
	if err := RequireNonEmpty("name", ""); err == nil {
		t.Fatal("expected empty value to fail")
	}
}

func TestRequireNonEmptyAcceptsValue(t *testing.T) {
	if err := RequireNonEmpty("name", "templatex"); err != nil {
		t.Fatalf("expected value to pass: %v", err)
	}
}
