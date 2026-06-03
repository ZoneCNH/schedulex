package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ZoneCNH/xlib-standard/internal/debtcheck"
)

func runDebt(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "register-update", "trend", "patch-suggest", "lifecycle-check":
			return runDebtHelper(args[0], args[1:], stdout, stderr)
		}
	}
	fs := flag.NewFlagSet("debt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := fs.String("config", debtcheck.DefaultRulesPath, "debt rules path")
	registry := fs.String("registry", debtcheck.DefaultRegistryPath, "debt rule registry path")
	exceptions := fs.String("exceptions", debtcheck.DefaultExceptions, "debt exceptions path")
	purpose := fs.String("dependency-purpose", debtcheck.DefaultPurpose, "dependency purpose path")
	section := fs.String("section", "all", "debt section")
	mode := fs.String("mode", "enforce", "debt mode")
	minScore := fs.Float64("min-score", debtcheck.DefaultMinScore, "minimum score")
	output := fs.String("output", "json", "output format: json or markdown")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	report, err := debtcheck.Run(debtcheck.Options{ConfigPath: *config, RegistryPath: *registry, ExceptionsPath: *exceptions, DependencyPurposePath: *purpose, Section: *section, Mode: *mode, MinScore: *minScore})
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	switch *output {
	case "json":
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			return 2
		}
		_, _ = fmt.Fprintln(stdout, string(encoded))
	case "markdown", "md":
		_, _ = fmt.Fprint(stdout, debtcheck.ToMarkdown(report))
	default:
		_, _ = fmt.Fprintln(stderr, "unsupported debt output format "+strconv.Quote(*output))
		return 2
	}
	return debtcheck.ExitCode(report)
}

type debtHelperArtifact struct {
	SchemaVersion string                      `json:"schema_version"`
	Command       string                      `json:"command"`
	Status        string                      `json:"status"`
	Mode          string                      `json:"mode"`
	ActiveProfile string                      `json:"active_profile"`
	Score         float64                     `json:"score"`
	MinScore      float64                     `json:"min_score"`
	Digests       debtcheck.Digests           `json:"digests"`
	Summary       debtcheck.Summary           `json:"summary"`
	Sections      []debtcheck.SectionEvidence `json:"sections"`
	Details       []string                    `json:"details"`
	Suggestions   []string                    `json:"suggestions,omitempty"`
}

func runDebtHelper(command string, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("debt "+command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := fs.String("config", debtcheck.DefaultRulesPath, "debt rules path")
	registry := fs.String("registry", debtcheck.DefaultRegistryPath, "debt rule registry path")
	exceptions := fs.String("exceptions", debtcheck.DefaultExceptions, "debt exceptions path")
	purpose := fs.String("dependency-purpose", debtcheck.DefaultPurpose, "dependency purpose path")
	mode := fs.String("mode", "enforce", "debt mode")
	minScore := fs.Float64("min-score", debtcheck.DefaultMinScore, "minimum score")
	output := fs.String("output", defaultDebtHelperOutput(command), "artifact output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "debt %s does not accept positional argument %s\n", command, strconv.Quote(fs.Arg(0)))
		return 2
	}

	report, err := debtcheck.Run(debtcheck.Options{
		ConfigPath:            *config,
		RegistryPath:          *registry,
		ExceptionsPath:        *exceptions,
		DependencyPurposePath: *purpose,
		Section:               "all",
		Mode:                  *mode,
		MinScore:              *minScore,
	})
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}

	artifact := buildDebtHelperArtifact(command, report)
	encoded, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	outputPath := filepath.Clean(*output)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o644); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	_, _ = fmt.Fprintf(stdout, "wrote %s\n", filepath.ToSlash(outputPath))
	return debtcheck.ExitCode(report)
}

func defaultDebtHelperOutput(command string) string {
	return filepath.FromSlash("release/debt/" + command + ".json")
}

func buildDebtHelperArtifact(command string, report debtcheck.Report) debtHelperArtifact {
	evidence := debtcheck.EvidenceFromReport(report)
	artifact := debtHelperArtifact{
		SchemaVersion: "debt-helper/v1",
		Command:       command,
		Status:        report.Status,
		Mode:          report.Mode,
		ActiveProfile: report.ActiveProfile,
		Score:         report.Score,
		MinScore:      report.MinScore,
		Digests:       report.Digests,
		Summary:       report.Summary,
		Sections:      evidence.Sections,
	}

	switch command {
	case "register-update":
		artifact.Details = []string{
			"captured debt governance registry state",
			"digest rules=" + report.Digests.Rules,
			"digest rule_registry=" + report.Digests.RuleRegistry,
			"digest exceptions=" + report.Digests.Exceptions,
			"digest dependency_purpose=" + report.Digests.DependencyPurpose,
		}
	case "trend":
		artifact.Details = debtTrendDetails(report)
	case "patch-suggest":
		artifact.Details = []string{"derived patch suggestions from current debt findings"}
		artifact.Suggestions = debtPatchSuggestions(report)
	case "lifecycle-check":
		artifact.Details = []string{
			"validated debt policy inputs and current report lifecycle",
			fmt.Sprintf("score %.2f minimum %.2f", report.Score, report.MinScore),
			"digest report=" + report.Digests.Report,
		}
	}
	return artifact
}

func debtTrendDetails(report debtcheck.Report) []string {
	latestPath := filepath.FromSlash("release/debt/latest.json")
	previousData, err := os.ReadFile(latestPath)
	if err != nil {
		return []string{
			"no prior debt evidence found at " + filepath.ToSlash(latestPath),
			fmt.Sprintf("current status %s score %.2f", report.Status, report.Score),
		}
	}

	var previous debtcheck.Report
	if err := json.Unmarshal(previousData, &previous); err != nil {
		return []string{
			"prior debt evidence at " + filepath.ToSlash(latestPath) + " is not a debt report",
			fmt.Sprintf("current status %s score %.2f", report.Status, report.Score),
		}
	}

	delta := report.Score - previous.Score
	return []string{
		"compared current debt report with " + filepath.ToSlash(latestPath),
		fmt.Sprintf("previous status %s score %.2f", previous.Status, previous.Score),
		fmt.Sprintf("current status %s score %.2f", report.Status, report.Score),
		fmt.Sprintf("score delta %.2f", delta),
	}
}

func debtPatchSuggestions(report debtcheck.Report) []string {
	var suggestions []string
	for _, section := range report.Sections {
		for _, finding := range section.Findings {
			suggestion := strings.TrimSpace(fmt.Sprintf("%s: address %s finding %s at %s: %s", section.Name, finding.Severity, finding.ID, finding.Path, finding.Message))
			if suggestion != "" {
				suggestions = append(suggestions, suggestion)
			}
			if len(suggestions) >= 20 {
				return suggestions
			}
		}
	}
	if len(suggestions) == 0 {
		return []string{"no patch suggestions; current debt report has no findings"}
	}
	return suggestions
}

func runDebtAlias(section, mode string, args []string, stdout, stderr io.Writer) int {
	preset := []string{"--section", section, "--mode", mode}
	preset = append(preset, args...)
	return runDebt(preset, stdout, stderr)
}

func runDebtEvidence(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		_, _ = fmt.Fprintln(stderr, "debt-evidence does not accept arguments")
		return 2
	}

	report, err := debtcheck.Run(debtcheck.Options{
		ConfigPath:            debtcheck.DefaultRulesPath,
		RegistryPath:          debtcheck.DefaultRegistryPath,
		ExceptionsPath:        debtcheck.DefaultExceptions,
		DependencyPurposePath: debtcheck.DefaultPurpose,
		Section:               "all",
		Mode:                  "enforce",
		MinScore:              debtcheck.DefaultMinScore,
	})
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}

	jsonPath := filepath.FromSlash("release/debt/latest.json")
	markdownPath := filepath.FromSlash("release/debt/latest.md")
	checksumPath := filepath.FromSlash("release/debt/latest.json.sha256")
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	jsonData := append(encoded, '\n')
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if err := os.WriteFile(markdownPath, []byte(debtcheck.ToMarkdown(report)), 0o644); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}

	sum := sha256.Sum256(jsonData)
	checksum := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), filepath.ToSlash(jsonPath))
	if err := os.WriteFile(checksumPath, []byte(checksum), 0o644); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}

	_, _ = fmt.Fprintf(stdout, "wrote %s\nwrote %s\nwrote %s\n", filepath.ToSlash(jsonPath), filepath.ToSlash(markdownPath), filepath.ToSlash(checksumPath))
	return debtcheck.ExitCode(report)
}
