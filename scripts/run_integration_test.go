package scripts_test

import (
	"os"
	"strings"
	"testing"
)

func TestRunIntegrationGeneratesStandardImpactBeforeReleaseEvidence(t *testing.T) {
	contents, err := os.ReadFile("run_integration.sh")
	if err != nil {
		t.Fatalf("read run_integration.sh: %v", err)
	}

	script := string(contents)
	standardImpactIndex := strings.Index(script, "GOWORK=off make standard-impact-check")
	if standardImpactIndex < 0 {
		t.Fatal("run_integration.sh does not generate standard impact evidence")
	}
	debtIndex := strings.Index(script, "\n    GOWORK=off make debt\n")
	if debtIndex < 0 {
		t.Fatal("run_integration.sh does not run downstream debt gate")
	}
	debtEvidenceIndex := strings.Index(script, "GOWORK=off make debt-evidence")
	if debtEvidenceIndex < 0 {
		t.Fatal("run_integration.sh does not generate downstream debt evidence")
	}
	debtChecksumIndex := strings.Index(script, "GOWORK=off make debt-evidence-checksum-check")
	if debtChecksumIndex < 0 {
		t.Fatal("run_integration.sh does not verify downstream debt evidence checksum")
	}
	evidenceIndex := strings.Index(script, "CHECK_STATUS=passed GOWORK=off make evidence")
	if evidenceIndex < 0 {
		t.Fatal("run_integration.sh does not generate release evidence")
	}
	checkIndex := strings.Index(script, "RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check")
	if checkIndex < 0 {
		t.Fatal("run_integration.sh does not verify release evidence")
	}

	if standardImpactIndex > debtIndex || debtIndex > debtEvidenceIndex || debtEvidenceIndex > debtChecksumIndex || debtChecksumIndex > evidenceIndex || evidenceIndex > checkIndex {
		t.Fatalf(
			"integration evidence order is wrong: standard-impact=%d debt=%d debt-evidence=%d debt-checksum=%d evidence=%d release-check=%d",
			standardImpactIndex,
			debtIndex,
			debtEvidenceIndex,
			debtChecksumIndex,
			evidenceIndex,
			checkIndex,
		)
	}
}

func TestRunIntegrationCoversRequiredDownstreams(t *testing.T) {
	contents, err := os.ReadFile("run_integration.sh")
	if err != nil {
		t.Fatalf("read run_integration.sh: %v", err)
	}

	script := string(contents)
	for _, target := range []string{
		"kernel|github.com/ZoneCNH/kernel|kernel",
		"configx|github.com/ZoneCNH/configx|configx",
		"redisx|github.com/ZoneCNH/redisx|redisx",
	} {
		if !strings.Contains(script, target) {
			t.Fatalf("run_integration.sh missing downstream target %q", target)
		}
	}

	if strings.Contains(script, "corekit|example.com/acme/corekit|corekit") {
		t.Fatal("run_integration.sh still includes legacy corekit integration target")
	}
}
