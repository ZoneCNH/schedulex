package templatex

import (
	"testing"
	"time"
)

func TestConfigValidateRequiresName(t *testing.T) {
	err := Config{Timeout: time.Second}.Validate()
	if err == nil {
		t.Fatal("expected missing name to fail validation")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
}

func TestConfigValidateRejectsNegativeTimeout(t *testing.T) {
	err := Config{Name: "templatex", Timeout: -time.Second}.Validate()
	if err == nil {
		t.Fatal("expected negative timeout to fail validation")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
}

func TestConfigSanitizeMasksSecret(t *testing.T) {
	sanitized := Config{Name: "templatex", Timeout: time.Second, Secret: "plain-text"}.Sanitize()
	if sanitized.Secret != "***" {
		t.Fatalf("expected masked secret, got %q", sanitized.Secret)
	}
	if sanitized.Name != "templatex" {
		t.Fatalf("expected name to be preserved, got %q", sanitized.Name)
	}
}
