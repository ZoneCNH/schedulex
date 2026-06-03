package releasequality

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const DefaultMinimum = 9.8

// Report is the executable release quality score recorded in release evidence.
type Report struct {
	Value      float64     `json:"value"`
	Threshold  float64     `json:"threshold"`
	Status     string      `json:"status"`
	Dimensions []Dimension `json:"dimensions"`
}

type Dimension struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Passed bool    `json:"passed"`
	Detail string  `json:"detail"`
}

func Compute(threshold float64) Report {
	if threshold <= 0 {
		threshold = DefaultMinimum
	}
	dimensions := []Dimension{
		fileDimension("scorecard_doc", 1, "docs/scorecard.md", "scorecard rubric is documented"),
		textDimension("manifest_score_schema", 1, "release/manifest/template.json", []string{"\"score\"", "\"workflow_run_id\"", "\"artifact_url\""}, "manifest records score and workflow evidence"),
		textDimension("score_cli", 1, "cmd/goalcli/main.go", []string{"score", "--min"}, "goalcli score command is runnable"),
		scoreGateDimension("Makefile"),
		textDimension("manifest_min_score_verify", 1, "scripts/check_release_evidence.sh", []string{"RELEASE_EVIDENCE_MIN_SCORE", "--min-score"}, "release evidence verification passes score threshold"),
		textDimension("security_gate", 1, "scripts/check_secrets.sh", []string{"github_pat_", "ghp_[A-Za-z0-9_]{36,}", "PRIVATE KEY-----"}, "secret scanner covers provider tokens and private keys"),
		textDimension("release_docs", 1, "docs/release.md", []string{"go run ./cmd/goalcli score --min 9.8", "workflow_run_id", "artifact_url"}, "release docs bind score and CI artifact evidence"),
		textDimension("supply_chain_docs", 1, "docs/supply-chain.md", []string{"score", "workflow_run_id", "artifact_url"}, "supply-chain docs include score/workflow evidence"),
		textDimension("retrospective_template", 1, ".agent/retrospective-template.md", []string{"Score", "Gate", "Patch"}, "retrospectives capture gate score and patch rationale"),
		textDimension("release_template", 1, ".agent/release-template.md", []string{"go run ./cmd/goalcli score --min 9.8", "CI artifact", "score"}, "release template requires score and artifact evidence"),
	}

	var total, passed float64
	for _, dimension := range dimensions {
		total += dimension.Weight
		if dimension.Passed {
			passed += dimension.Weight
		}
	}
	value := 0.0
	if total > 0 {
		value = math.Round((passed/total)*100) / 10
	}
	status := "failed"
	if value >= threshold {
		status = "passed"
	}
	return Report{Value: value, Threshold: threshold, Status: status, Dimensions: dimensions}
}

func Verify(report Report, minimum float64) error {
	if minimum <= 0 {
		minimum = report.Threshold
	}
	if report.Value < minimum {
		return fmt.Errorf("release score %.1f is below minimum %.1f", report.Value, minimum)
	}
	if report.Status != "passed" && report.Value < report.Threshold {
		return fmt.Errorf("release score status is %q at %.1f below threshold %.1f", report.Status, report.Value, report.Threshold)
	}
	return nil
}

func Marshal(report Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func fileDimension(name string, weight float64, path string, detail string) Dimension {
	_, err := os.Stat(path)
	return Dimension{Name: name, Weight: weight, Passed: err == nil, Detail: detail}
}

func textDimension(name string, weight float64, path string, needles []string, detail string) Dimension {
	data, err := os.ReadFile(path)
	if err != nil {
		return Dimension{Name: name, Weight: weight, Passed: false, Detail: detail + ": missing " + path}
	}
	text := string(data)
	var missing []string
	for _, needle := range needles {
		if !strings.Contains(text, needle) {
			missing = append(missing, needle)
		}
	}
	passed := len(missing) == 0
	if !passed {
		detail = detail + ": missing " + strings.Join(missing, ", ")
	}
	return Dimension{Name: name, Weight: weight, Passed: passed, Detail: detail}
}

func scoreGateDimension(path string) Dimension {
	const detail = "release targets enforce score thresholds"
	data, err := os.ReadFile(path)
	if err != nil {
		return Dimension{Name: "score_gate", Weight: 1, Passed: false, Detail: detail + ": missing " + path}
	}
	text := string(data)
	var missing []string
	for _, needle := range []string{"score-check", "release-final-check"} {
		if !strings.Contains(text, needle) {
			missing = append(missing, needle)
		}
	}
	if !hasScoreMinimumAtLeast(text, 9.8) {
		missing = append(missing, "score --min >= 9.8")
	}
	passed := len(missing) == 0
	gateDetail := detail
	if !passed {
		gateDetail = gateDetail + ": missing " + strings.Join(missing, ", ")
	}
	return Dimension{Name: "score_gate", Weight: 1, Passed: passed, Detail: gateDetail}
}

var scoreMinPattern = regexp.MustCompile(`\bscore\s+--min\s+([0-9]+(?:\.[0-9]+)?)\b`)

func hasScoreMinimumAtLeast(text string, minimum float64) bool {
	for _, match := range scoreMinPattern.FindAllStringSubmatch(text, -1) {
		value, err := strconv.ParseFloat(match[1], 64)
		if err == nil && value >= minimum {
			return true
		}
	}
	return false
}
