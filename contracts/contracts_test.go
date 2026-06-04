package contracts_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestContractsExistAndUseSchedulexIdentity(t *testing.T) {
	paths := []string{"public_api.snapshot", "trigger_cases/l1_golden.json", "trigger_cases/dst_golden.json", "misfire_cases/l1_golden.json", "lifecycle_event.schema.json", "release_manifest.schema.json"}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "xlib-standard") || strings.Contains(string(b), "templatex") {
			t.Fatalf("legacy identity in %s", p)
		}
	}
}

func TestGoldenJSONValid(t *testing.T) {
	for _, p := range []string{"trigger_cases/l1_golden.json", "trigger_cases/dst_golden.json", "misfire_cases/l1_golden.json"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var v any
		if err := json.Unmarshal(b, &v); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
	}
}
