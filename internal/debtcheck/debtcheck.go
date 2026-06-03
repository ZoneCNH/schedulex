package debtcheck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	SchemaVersion       = "debt-report/v1"
	ManifestSchema      = "debt-evidence/v1"
	DefaultRulesPath    = ".agent/debt/rules.yaml"
	DefaultRegistryPath = ".agent/debt/rule-registry.yaml"
	DefaultExceptions   = ".agent/debt/exceptions.yaml"
	DefaultPurpose      = ".agent/debt/dependency-purpose.yaml"
	DefaultMinScore     = 9.8
	privateKeyPrefix    = "-----BEGIN " + "PRIVATE KEY-----"
)

type Options struct {
	Root                  string
	ConfigPath            string
	RegistryPath          string
	ExceptionsPath        string
	DependencyPurposePath string
	Section               string
	Mode                  string
	MinScore              float64
}

type Report struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	Mode          string          `json:"mode"`
	ActiveProfile string          `json:"active_profile"`
	Score         float64         `json:"score"`
	MinScore      float64         `json:"min_score"`
	Digests       Digests         `json:"digests"`
	Summary       Summary         `json:"summary"`
	Sections      []SectionReport `json:"sections"`
}

type Digests struct {
	Rules             string `json:"rules"`
	RuleRegistry      string `json:"rule_registry"`
	Exceptions        string `json:"exceptions"`
	DependencyPurpose string `json:"dependency_purpose"`
	Report            string `json:"report"`
}

type Summary struct {
	P0 int `json:"p0"`
	P1 int `json:"p1"`
	P2 int `json:"p2"`
}

type SectionReport struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	P0       int       `json:"p0"`
	P1       int       `json:"p1"`
	P2       int       `json:"p2"`
	Findings []Finding `json:"findings"`
}

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
}

func Run(opts Options) (Report, error) {
	opts = normalize(opts)
	if err := validateMode(opts.Mode); err != nil {
		return Report{}, err
	}
	if err := validateSection(opts.Section); err != nil {
		return Report{}, err
	}
	report := Report{
		SchemaVersion: SchemaVersion,
		Mode:          opts.Mode,
		ActiveProfile: "xlib-standard-debt-v1",
		MinScore:      opts.MinScore,
		Digests: Digests{
			Rules:             digestFile(opts.Root, opts.ConfigPath),
			RuleRegistry:      digestFile(opts.Root, opts.RegistryPath),
			Exceptions:        digestFile(opts.Root, opts.ExceptionsPath),
			DependencyPurpose: digestFile(opts.Root, opts.DependencyPurposePath),
		},
	}

	missing := missingPolicyFindings(opts)
	for _, section := range selectedSections(opts.Section) {
		findings := append([]Finding{}, missing...)
		findings = append(findings, scanSection(opts.Root, section)...)
		report.Sections = append(report.Sections, buildSection(section, findings))
	}
	for _, section := range report.Sections {
		report.Summary.P0 += section.P0
		report.Summary.P1 += section.P1
		report.Summary.P2 += section.P2
	}
	report.Score = score(report.Summary)
	report.Status = status(report.Summary, report.Score, report.MinScore, opts.Mode)
	report.Digests.Report = ReportDigest(report)
	return report, nil
}

func ExitCode(report Report) int {
	if report.Mode == "observe" || report.Mode == "warn" {
		return 0
	}
	if report.Status != "passed" {
		return 1
	}
	return 0
}

func ToMarkdown(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Debt Governance Report\n\n")
	fmt.Fprintf(&b, "- Status: %s\n", report.Status)
	fmt.Fprintf(&b, "- Mode: %s\n", report.Mode)
	fmt.Fprintf(&b, "- Score: %.2f (minimum %.2f)\n", report.Score, report.MinScore)
	fmt.Fprintf(&b, "- Active profile: %s\n", report.ActiveProfile)
	fmt.Fprintf(&b, "- P0/P1/P2: %d/%d/%d\n\n", report.Summary.P0, report.Summary.P1, report.Summary.P2)
	fmt.Fprintf(&b, "## Sections\n\n")
	for _, section := range report.Sections {
		fmt.Fprintf(&b, "### %s\n\n", section.Name)
		fmt.Fprintf(&b, "Status: %s; P0/P1/P2: %d/%d/%d\n\n", section.Status, section.P0, section.P1, section.P2)
		if len(section.Findings) == 0 {
			fmt.Fprintf(&b, "No findings.\n\n")
			continue
		}
		for _, finding := range section.Findings {
			path := finding.Path
			if path == "" {
				path = "policy"
			}
			fmt.Fprintf(&b, "- [%s] %s %s: %s\n", finding.Severity, finding.ID, path, finding.Message)
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func ReportDigest(report Report) string {
	reportCopy := report
	reportCopy.Digests.Report = ""
	encoded, _ := json.Marshal(reportCopy)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func EvidenceFromReport(report Report) Evidence {
	sections := make([]SectionEvidence, 0, len(report.Sections))
	for _, section := range report.Sections {
		sections = append(sections, SectionEvidence{Name: section.Name, Status: section.Status, P0: section.P0, P1: section.P1, P2: section.P2})
	}
	return Evidence{
		SchemaVersion:       ManifestSchema,
		ReportSchemaVersion: report.SchemaVersion,
		Status:              report.Status,
		Score:               report.Score,
		MinScore:            report.MinScore,
		Mode:                report.Mode,
		ActiveProfile:       report.ActiveProfile,
		Digests:             report.Digests,
		Sections:            sections,
	}
}

type Evidence struct {
	SchemaVersion       string            `json:"schema_version"`
	ReportSchemaVersion string            `json:"report_schema_version"`
	Status              string            `json:"status"`
	Score               float64           `json:"score"`
	MinScore            float64           `json:"min_score"`
	Mode                string            `json:"mode"`
	ActiveProfile       string            `json:"active_profile"`
	Digests             Digests           `json:"digests"`
	Sections            []SectionEvidence `json:"sections"`
}

type SectionEvidence struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	P0     int    `json:"p0"`
	P1     int    `json:"p1"`
	P2     int    `json:"p2"`
}

func ValidateEvidence(e Evidence, minScore float64) []string {
	var problems []string
	if e.SchemaVersion != ManifestSchema {
		problems = append(problems, "debt schema version mismatch")
	}
	if e.ReportSchemaVersion != SchemaVersion {
		problems = append(problems, "debt report schema version mismatch")
	}
	if e.Status != "passed" {
		problems = append(problems, fmt.Sprintf("debt status is %s", e.Status))
	}
	if e.Score < minScore {
		problems = append(problems, fmt.Sprintf("debt score %.2f below %.2f", e.Score, minScore))
	}
	for _, section := range e.Sections {
		if section.P0 != 0 {
			problems = append(problems, fmt.Sprintf("debt section %s has %d P0 findings", section.Name, section.P0))
		}
		if section.Status != "passed" {
			problems = append(problems, fmt.Sprintf("debt section %s status is %s", section.Name, section.Status))
		}
	}
	return problems
}

func normalize(opts Options) Options {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = DefaultRulesPath
	}
	if opts.RegistryPath == "" {
		opts.RegistryPath = DefaultRegistryPath
	}
	if opts.ExceptionsPath == "" {
		opts.ExceptionsPath = DefaultExceptions
	}
	if opts.DependencyPurposePath == "" {
		opts.DependencyPurposePath = DefaultPurpose
	}
	if opts.Section == "" {
		opts.Section = "all"
	}
	if opts.Mode == "" {
		opts.Mode = "enforce"
	}
	if opts.MinScore == 0 {
		opts.MinScore = DefaultMinScore
	}
	return opts
}

func validateMode(mode string) error {
	switch mode {
	case "enforce", "warn", "observe":
		return nil
	default:
		return fmt.Errorf("unsupported debt mode %q", mode)
	}
}

func validateSection(section string) error {
	for _, allowed := range append([]string{"all"}, allSections()...) {
		if section == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported debt section %q", section)
}

func allSections() []string {
	return []string{"architecture", "domain", "docs", "dependency", "testing", "implementation", "security", "downstream"}
}

func selectedSections(section string) []string {
	if section != "all" {
		return []string{section}
	}
	return allSections()
}

func missingPolicyFindings(opts Options) []Finding {
	paths := []struct{ id, path string }{
		{"debt.rules.missing", opts.ConfigPath},
		{"debt.registry.missing", opts.RegistryPath},
		{"debt.exceptions.missing", opts.ExceptionsPath},
		{"debt.dependency-purpose.missing", opts.DependencyPurposePath},
	}
	var findings []Finding
	for _, item := range paths {
		if _, err := os.Stat(filepath.Join(opts.Root, item.path)); err != nil {
			findings = append(findings, Finding{ID: item.id, Severity: "P0", Path: item.path, Message: "required debt governance policy file is missing"})
		}
	}
	return findings
}

func scanSection(root, section string) []Finding {
	switch section {
	case "architecture":
		return scanGoImports(root)
	case "domain":
		return scanTextMarker(root, "xlib-domain-forbidden", "debt.domain.marker", "domain debt marker is present")
	case "docs":
		return scanTextMarker(root, "xlib-docs-drift", "debt.docs.marker", "documentation drift marker is present")
	case "dependency":
		return scanDependencyDebt(root)
	case "testing":
		return scanTextMarker(root, "xlib-testing-debt", "debt.testing.marker", "testing debt marker is present")
	case "implementation":
		return scanTextMarker(root, "xlib-implementation-debt", "debt.implementation.marker", "implementation debt marker is present")
	case "security":
		return scanSecurityDebt(root)
	case "downstream":
		return scanDownstreamDebt(root)
	default:
		return nil
	}
}

var downstreamRequiredFiles = []string{
	".agent/downstream-registry.yaml",
	".agent/downstream-baseline-scan.yaml",
	".agent/downstream-adoption-modes.yaml",
	".agent/downstream-adoption-status.yaml",
	"docs/downstream-matrix.md",
	"docs/standard/downstream-compatibility.md",
	"scripts/run_integration.sh",
	"scripts/render_template.sh",
}

var downstreamRepresentativeRepos = []string{
	"kernel/configx",
	"kernel/redisx",
	"corekit",
}

var downstreamTargetLibraries = []string{
	"kernel",
	"configx",
	"observex",
	"testkitx",
	"postgresx",
	"redisx",
	"kafkax",
	"taosx",
	"ossx",
	"clickhousex",
}

var downstreamIntegrationTokens = []string{
	"kernel|github.com/ZoneCNH/kernel|kernel",
	"configx|github.com/ZoneCNH/configx|configx",
	"redisx|github.com/ZoneCNH/redisx|redisx",
	"GOWORK=off make debt",
	"GOWORK=off make debt-evidence",
	"GOWORK=off make debt-evidence-checksum-check",
}

var downstreamRenderTemplateExclusions = []string{
	"release/debt/latest.json",
	"release/debt/latest.md",
	"release/debt/latest.json.sha256",
}

func scanDownstreamDebt(root string) []Finding {
	files := make(map[string]string, len(downstreamRequiredFiles))
	var findings []Finding

	for _, relPath := range downstreamRequiredFiles {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
		if err != nil {
			findings = append(findings, Finding{
				ID:       "debt.downstream.file-missing",
				Severity: "P0",
				Path:     relPath,
				Message:  "required downstream governance file is missing",
			})
			continue
		}

		text := string(content)
		files[relPath] = text
		if strings.HasSuffix(relPath, ".yaml") && !strings.Contains(text, "schema_version:") {
			findings = append(findings, Finding{
				ID:       "debt.downstream.schema-missing",
				Severity: "P0",
				Path:     relPath,
				Message:  "downstream governance YAML must declare schema_version",
			})
		}
		if containsDownstreamPlaceholder(text) {
			findings = append(findings, Finding{
				ID:       "debt.downstream.placeholder",
				Severity: "P1",
				Path:     relPath,
				Message:  "downstream governance file still contains placeholder text",
			})
		}
	}

	registry := files[".agent/downstream-registry.yaml"]
	for _, repo := range downstreamRepresentativeRepos {
		requireDownstreamToken(&findings, registry, ".agent/downstream-registry.yaml", "repo: "+repo, "debt.downstream.registry-missing-repo", "downstream registry is missing required representative repo")
	}

	baseline := files[".agent/downstream-baseline-scan.yaml"]
	requireDownstreamToken(&findings, baseline, ".agent/downstream-baseline-scan.yaml", "repo: kernel/configx", "debt.downstream.baseline-missing-repo", "downstream baseline must name the reference repo")
	requireDownstreamToken(&findings, baseline, ".agent/downstream-baseline-scan.yaml", "mode: patch-only", "debt.downstream.baseline-missing-mode", "downstream baseline must preserve patch-only mode")
	if baseline != "" && !strings.Contains(baseline, "gap") {
		findings = append(findings, Finding{
			ID:       "debt.downstream.baseline-missing-gap-status",
			Severity: "P0",
			Path:     ".agent/downstream-baseline-scan.yaml",
			Message:  "downstream baseline must explicitly record the repo-missing gap status",
		})
	}

	modes := files[".agent/downstream-adoption-modes.yaml"]
	requireDownstreamToken(&findings, modes, ".agent/downstream-adoption-modes.yaml", "patch-only", "debt.downstream.mode-missing-patch-only", "downstream adoption modes must include patch-only")
	requireDownstreamToken(&findings, modes, ".agent/downstream-adoption-modes.yaml", "direct_downstream_write_without_repo", "debt.downstream.mode-missing-write-guard", "downstream adoption modes must forbid direct downstream writes without a repo")

	status := files[".agent/downstream-adoption-status.yaml"]
	for _, name := range downstreamTargetLibraries {
		requireDownstreamToken(&findings, status, ".agent/downstream-adoption-status.yaml", "name: "+name, "debt.downstream.status-missing-target", "downstream adoption status must include every standard target library")
	}
	if hasYAMLScalarLine(status, "adoption_status", "adopted") {
		findings = append(findings, Finding{
			ID:       "debt.downstream.false-adoption-claim",
			Severity: "P0",
			Path:     ".agent/downstream-adoption-status.yaml",
			Message:  "downstream adoption status claims adopted without proof-based adoption evidence",
		})
	}
	if hasYAMLScalarLine(status, "proof_based_adoption", "true") {
		findings = append(findings, Finding{
			ID:       "debt.downstream.false-proof-claim",
			Severity: "P0",
			Path:     ".agent/downstream-adoption-status.yaml",
			Message:  "downstream adoption status claims proof-based adoption without downstream proof gate evidence",
		})
	}

	matrix := files["docs/downstream-matrix.md"]
	for _, name := range downstreamTargetLibraries {
		requireDownstreamToken(&findings, matrix, "docs/downstream-matrix.md", "`"+name+"`", "debt.downstream.matrix-missing-target", "downstream matrix must cover every standard target library")
	}

	compatibility := files["docs/standard/downstream-compatibility.md"]
	for _, token := range []string{"`kernel`", "`corekit`", "GOWORK=off make integration"} {
		requireDownstreamToken(&findings, compatibility, "docs/standard/downstream-compatibility.md", token, "debt.downstream.compatibility-missing-contract", "downstream compatibility standard must preserve default downstream and verification contract")
	}

	integration := files["scripts/run_integration.sh"]
	for _, token := range downstreamIntegrationTokens {
		requireDownstreamToken(&findings, integration, "scripts/run_integration.sh", token, "debt.downstream.integration-missing-contract", "downstream integration script must render required targets and preserve debt evidence gates")
	}

	renderTemplate := files["scripts/render_template.sh"]
	for _, token := range downstreamRenderTemplateExclusions {
		requireDownstreamToken(&findings, renderTemplate, "scripts/render_template.sh", token, "debt.downstream.render-template-missing-exclusion", "rendered template must exclude generated debt evidence artifacts")
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Path == findings[j].Path {
			return findings[i].ID < findings[j].ID
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

func requireDownstreamToken(findings *[]Finding, text, path, token, id, message string) {
	if text == "" || strings.Contains(text, token) {
		return
	}
	*findings = append(*findings, Finding{
		ID:       id,
		Severity: "P0",
		Path:     path,
		Message:  message,
	})
}

func containsDownstreamPlaceholder(text string) bool {
	upper := strings.ToUpper(text)
	return strings.Contains(upper, "TODO") ||
		strings.Contains(upper, "TBD") ||
		strings.Contains(upper, "PLACEHOLDER")
}

func hasYAMLScalarLine(text, key, value string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, key+":") {
			continue
		}
		if strings.TrimSpace(strings.TrimPrefix(trimmed, key+":")) == value {
			return true
		}
	}
	return false
}

func scanGoImports(root string) []Finding {
	var findings []Finding
	_ = walkFiles(root, func(path string) error {
		if skipPath(root, path) || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
		if parseErr != nil {
			findings = append(findings, Finding{ID: "debt.architecture.parse", Severity: "P0", Path: rel(root, path), Message: "production Go file could not be parsed for imports"})
		} else {
			for _, imported := range file.Imports {
				importPath := strings.Trim(imported.Path.Value, `"`)
				if importPath == "github.com/ZoneCNH/x.go" || strings.HasPrefix(importPath, "github.com/ZoneCNH/x/") {
					findings = append(findings, Finding{ID: "debt.architecture.legacy-import", Severity: "P0", Path: rel(root, path), Message: "production code imports legacy ZoneCNH x module"})
					break
				}
			}
		}
		return nil
	})
	return findings
}

func scanDependencyDebt(root string) []Finding {
	return scanTrackedText(root, func(path, text string) []Finding {
		var findings []Finding
		if strings.Contains(text, "@latest") && !strings.Contains(path, ".md") {
			findings = append(findings, Finding{ID: "debt.dependency.unpinned-latest", Severity: "P1", Path: rel(root, path), Message: "non-documentation file references @latest"})
		}
		if strings.Contains(text, "curl ") && strings.Contains(text, "| bash") {
			findings = append(findings, Finding{ID: "debt.dependency.curl-pipe-bash", Severity: "P1", Path: rel(root, path), Message: "curl pipe bash pattern needs dependency-purpose justification"})
		}
		return findings
	})
}

func scanSecurityDebt(root string) []Finding {
	return scanTrackedText(root, func(path, text string) []Finding {
		var findings []Finding
		if strings.Contains(text, privateKeyPrefix) {
			findings = append(findings, Finding{ID: "debt.security.private-key", Severity: "P0", Path: rel(root, path), Message: "private key material marker is present"})
		}
		if strings.Contains(text, "xlib-security-debt") {
			findings = append(findings, Finding{ID: "debt.security.marker", Severity: "P1", Path: rel(root, path), Message: "security debt marker is present"})
		}
		return findings
	})
}

func scanTextMarker(root, marker, id, message string) []Finding {
	return scanTrackedText(root, func(path, text string) []Finding {
		if strings.Contains(text, marker) {
			finding := Finding{ID: id, Severity: "P1", Path: rel(root, path), Message: message}
			return []Finding{finding}
		}
		return nil
	})
}

func scanTrackedText(root string, inspect func(path, text string) []Finding) []Finding {
	var findings []Finding
	_ = walkFiles(root, func(path string) error {
		if skipPath(root, path) || skipFile(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytesLookBinary(data) {
			return nil
		}
		findings = append(findings, inspect(path, string(data))...)
		return nil
	})
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path == findings[j].Path {
			return findings[i].ID < findings[j].ID
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

func walkFiles(root string, visit func(path string) error) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return visit(root)
	}
	return walkDir(root, visit)
}

func walkDir(dir string, visit func(path string) error) error {
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	names, readErr := file.Readdirnames(-1)
	closeErr := file.Close()
	if readErr != nil {
		return readErr
	}
	if closeErr != nil {
		return closeErr
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if skipDir(name) {
				continue
			}
			if err := walkDir(path, visit); err != nil {
				return err
			}
			continue
		}
		if err := visit(path); err != nil {
			return err
		}
	}
	return nil
}

func buildSection(name string, findings []Finding) SectionReport {
	section := SectionReport{Name: name, Findings: findings}
	for _, finding := range findings {
		switch finding.Severity {
		case "P0":
			section.P0++
		case "P1":
			section.P1++
		default:
			section.P2++
		}
	}
	if section.P0 == 0 && section.P1 == 0 && section.P2 == 0 {
		section.Status = "passed"
	} else if section.P0 == 0 {
		section.Status = "warning"
	} else {
		section.Status = "failed"
	}
	return section
}

func score(summary Summary) float64 {
	value := 10 - float64(summary.P1)*0.1 - float64(summary.P2)*0.05
	if summary.P0 > 0 {
		value -= float64(summary.P0)
	}
	return math.Max(0, math.Round(value*100)/100)
}

func status(summary Summary, score, minScore float64, mode string) string {
	if summary.P0 > 0 {
		return "failed"
	}
	if score < minScore {
		if mode == "observe" || mode == "warn" {
			return "warning"
		}
		return "failed"
	}
	if summary.P1 > 0 || summary.P2 > 0 {
		if mode == "enforce" {
			return "passed"
		}
		return "warning"
	}
	return "passed"
}

func digestFile(root, path string) string {
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return "missing"
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func skipDir(name string) bool {
	switch name {
	case ".git", ".omx", ".worktree", "vendor", "node_modules", "release", "tmp", ".cache":
		return true
	default:
		return false
	}
}

func skipPath(root, path string) bool {
	relPath := rel(root, path)
	return strings.HasPrefix(relPath, ".agent/debt/") ||
		strings.HasPrefix(relPath, ".agent/inbox/") ||
		strings.HasPrefix(relPath, "internal/debtcheck/")
}

func skipFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != ".gitignore" {
		return true
	}
	if strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".gif") || strings.HasSuffix(path, ".pdf") {
		return true
	}
	return false
}

func bytesLookBinary(data []byte) bool {
	limit := len(data)
	if limit > 4096 {
		limit = 4096
	}
	for _, b := range data[:limit] {
		if b == 0 {
			return true
		}
	}
	return false
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(r)
}

func ReadReport(path string) (Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return Report{}, err
	}
	if report.SchemaVersion != SchemaVersion {
		return Report{}, errors.New("unsupported debt report schema")
	}
	return report, nil
}
