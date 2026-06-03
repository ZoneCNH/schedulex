package templatex

import (
	"testing"
	"time"
)

func FuzzConfigSanitize(f *testing.F) {
	f.Add("")
	f.Add("plain-text")
	f.Add("secret")
	f.Add("密钥")
	f.Add("line\nbreak")
	f.Add("***")

	f.Fuzz(func(t *testing.T, secret string) {
		sanitized := Config{
			Name:    "templatex",
			Timeout: time.Second,
			Secret:  secret,
		}.Sanitize()

		if secret == "" {
			if sanitized.Secret != "" {
				t.Fatalf("empty secret sanitized to %q", sanitized.Secret)
			}
			return
		}

		if sanitized.Secret != "***" {
			t.Fatalf("secret must sanitize to redaction marker, got %q", sanitized.Secret)
		}
		if secret != "***" && sanitized.Secret == secret {
			t.Fatalf("sanitize leaked secret %q", secret)
		}
	})
}
