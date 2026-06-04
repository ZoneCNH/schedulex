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
	if schedulex.ModuleName != "github.com/ZoneCNH/schedulex" || schedulex.Version != "v0.1.0" {
		t.Fatalf("unexpected schedulex identity: %s %s", schedulex.ModuleName, schedulex.Version)
	}
}
