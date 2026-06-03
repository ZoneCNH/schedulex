package templatex

import (
	"testing"
	"testing/quick"
	"time"
)

func TestConfigSanitizeSecretProperty(t *testing.T) {
	property := func(secret string) bool {
		sanitized := Config{
			Name:    "templatex",
			Timeout: time.Second,
			Secret:  secret,
		}.Sanitize()

		if secret == "" {
			return sanitized.Secret == ""
		}
		if sanitized.Secret != "***" {
			return false
		}

		return secret == "***" || sanitized.Secret != secret
	}

	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func TestConfigNegativeTimeoutInvariant(t *testing.T) {
	for _, timeout := range []time.Duration{-1, -time.Nanosecond, -time.Second} {
		err := Config{Name: "templatex", Timeout: timeout}.Validate()
		if err == nil || !IsKind(err, ErrorKindValidation) {
			t.Fatalf("negative timeout must return validation error, got %v", err)
		}
	}
}
