package contracts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func TestRepositoryIdentityContract(t *testing.T) {
	mod, err := os.ReadFile(filepath.Join("..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "module github.com/ZoneCNH/schedulex") {
		t.Fatalf("unexpected module identity:\n%s", mod)
	}
	if _, err := os.Stat(filepath.Join("..", "pkg", "templatex")); !os.IsNotExist(err) {
		t.Fatalf("pkg/templatex must not exist for schedulex L1 release")
	}
}

func TestContractsExistAndUseSchedulexIdentity(t *testing.T) {
	paths := []string{
		"public_api.snapshot",
		"trigger_cases/l1_golden.json",
		"trigger_cases/dst_golden.json",
		"trigger_cases/basic.golden.json",
		"timezone_dst_cases/daily_at_dst.golden.json",
		"misfire_cases/l1_golden.json",
		"misfire_cases/basic.golden.json",
		"snapshot.schema.json",
		"lifecycle_event.schema.json",
		"release_manifest.schema.json",
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		text := string(content)
		if strings.Contains(text, "xlib-standard") || strings.Contains(text, "templatex") || strings.Contains(text, "foundationx") {
			t.Fatalf("legacy identity in %s", path)
		}
	}
}

func TestJSONContractsAndGoldensAreValid(t *testing.T) {
	paths := []string{
		"trigger_cases/l1_golden.json",
		"trigger_cases/dst_golden.json",
		"trigger_cases/basic.golden.json",
		"timezone_dst_cases/daily_at_dst.golden.json",
		"misfire_cases/l1_golden.json",
		"misfire_cases/basic.golden.json",
		"snapshot.schema.json",
		"lifecycle_event.schema.json",
		"release_manifest.schema.json",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var doc any
			if err := json.Unmarshal(content, &doc); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
		})
	}
}

func TestPublicAPISnapshotDocumentsExportedSurface(t *testing.T) {
	content, err := os.ReadFile("public_api.snapshot")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	required := []string{
		"func NewScheduler",
		"func Once",
		"func Every",
		"func Cron",
		"func DailyAt",
		"type Scheduler",
		"type Job",
		"type JobFunc",
		"type Trigger",
		"type Clock",
		"type Locker",
		"type Lease",
		"type EventSink",
		"type MisfirePolicy",
		"type OverlapPolicy",
		"const MisfireSkip MisfirePolicy",
		"const MisfireRunOnce MisfirePolicy",
		"const MisfireCatchUp MisfirePolicy",
		"const OverlapSkip OverlapPolicy",
		"const OverlapQueueOne OverlapPolicy",
		"const OverlapAllow OverlapPolicy",
	}
	for _, item := range required {
		if !strings.Contains(text, item) {
			t.Fatalf("public API snapshot missing %q", item)
		}
	}
	if schedulex.ModuleName != "github.com/ZoneCNH/schedulex" || schedulex.Version != "v1.0.0" {
		t.Fatalf("unexpected schedulex identity: %s %s", schedulex.ModuleName, schedulex.Version)
	}
}

func TestReleaseDocumentationAndDownstreamFixtureUseV1Identity(t *testing.T) {
	files := map[string]string{
		"README.md":       filepath.Join("..", "README.md"),
		"docs/release.md": filepath.Join("..", "docs", "release.md"),
		"CONSTITUTION.md": filepath.Join("..", "CONSTITUTION.md"),
		"CHANGELOG.md":    filepath.Join("..", "CHANGELOG.md"),
	}
	for name, path := range files {
		t.Run(name, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text := string(content)
			if strings.Contains(text, "VERSION=v0.1.0") {
				t.Fatalf("%s still documents v0.1.0 release preflight", name)
			}
			if name != "CHANGELOG.md" && (strings.Contains(text, "xlib-standard") || strings.Contains(text, "cmd/goalcli")) {
				t.Fatalf("%s contains legacy governance identity", name)
			}
		})
	}

	for _, path := range []string{
		filepath.Join("..", "README.md"),
		filepath.Join("..", "docs", "release.md"),
		filepath.Join("..", "CHANGELOG.md"),
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "v1.0.0") {
			t.Fatalf("%s does not document v1.0.0", path)
		}
	}

	metadata, err := os.ReadFile(filepath.Join("..", "release", "downstream-adoption", "latest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var downstream struct {
		Module  string `json:"module"`
		Version string `json:"version"`
		Mode    string `json:"mode"`
	}
	if err := json.Unmarshal(metadata, &downstream); err != nil {
		t.Fatal(err)
	}
	if downstream.Module != "github.com/ZoneCNH/schedulex" || downstream.Version != "v1.0.0" || downstream.Mode != "no_local_replace_fixture" {
		t.Fatalf("unexpected downstream adoption metadata: %+v", downstream)
	}

	fixture, err := os.ReadFile(filepath.Join("..", "release", "downstream-adoption", "fixture", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	fixtureText := string(fixture)
	if !strings.Contains(fixtureText, "require github.com/ZoneCNH/schedulex v1.0.0") {
		t.Fatalf("downstream fixture does not require v1.0.0:\n%s", fixtureText)
	}
	if strings.Contains(fixtureText, "replace github.com/ZoneCNH/schedulex") {
		t.Fatalf("downstream fixture must not use local replace:\n%s", fixtureText)
	}
}
