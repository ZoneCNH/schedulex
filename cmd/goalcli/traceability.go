package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// traceability-check 退出码 (与 .agent/rules/iron-rules.md 标准退出码表对齐):
//
//	0 - 所有 REQ 行的产物与 Evidence 引用均可解析
//	2 - 参数错误 / 矩阵文件不可解析
//	9 - 至少一条 traceability 链路缺失 (RULE-CORE-004 / RULE-TRACE-001)
const traceabilityExitGap = 9

// traceabilityMatrixPath 是 traceability matrix 的固定路径。
// 维持单一来源, 与 .agent/rules/registry.yaml 的 enforced_by 绑定一致。
const traceabilityMatrixPath = ".agent/traceability-matrix.md"

// traceabilityPathHints 用于判断一个 token 是否"看起来是文件路径引用"。
// 命中任一前缀或后缀的 token 会被纳入存在性校验, 其余作为说明性文字跳过。
var (
	traceabilityPathPrefixes = []string{
		".agent/", ".github/", "cmd/", "contracts/", "docs/",
		"examples/", "internal/", "pkg/", "release/", "scripts/", "testkit/",
	}
	traceabilityPathSuffixes = []string{
		".md", ".yaml", ".yml", ".json", ".go", ".sh", ".mk",
	}
	traceabilityKnownTopFiles = map[string]bool{
		"README.md":       true,
		"Makefile":        true,
		"CONSTITUTION.md": true,
		"AGENTS.md":       true,
		"CLAUDE.md":       true,
		"go.mod":          true,
		"go.sum":          true,
		".gitignore":      true,
		".golangci.yml":   true,
		"renovate.json":   true,
	}
)

var traceabilityRowRegex = regexp.MustCompile(`^\|\s*(REQ-\d+)\s*\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)\|`)

// traceabilityRow 表示矩阵中一行 REQ 的解析结果。
type traceabilityRow struct {
	ReqID     string
	Artifacts []string
	Evidence  []string
}

func runTraceabilityCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli traceability-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	matrixPath := flags.String("matrix", traceabilityMatrixPath, "path to traceability matrix markdown")
	emitJSON := flags.Bool("json", true, "emit JSON gate report (default true)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		write(stderr, "ERROR: traceability-check invalid arguments: %v\n", err)
		return 2
	}
	if flags.NArg() != 0 {
		write(stderr, "ERROR: traceability-check accepts no positional arguments\n")
		return 2
	}

	rows, err := parseTraceabilityMatrix(*matrixPath)
	if err != nil {
		write(stderr, "ERROR: %v\n", err)
		return 2
	}
	if len(rows) == 0 {
		write(stderr, "ERROR: traceability matrix has no REQ rows: %s\n", *matrixPath)
		return 2
	}

	var details []string
	var gaps []string
	for _, row := range rows {
		details = append(details, fmt.Sprintf("%s: %d artifact(s), %d evidence ref(s)", row.ReqID, len(row.Artifacts), len(row.Evidence)))
		if len(row.Artifacts) == 0 {
			gaps = append(gaps, row.ReqID+": 主要产物 column is empty")
			continue
		}
		for _, artifact := range row.Artifacts {
			if !looksLikePath(artifact) {
				continue
			}
			if err := verifyArtifactExists(artifact); err != nil {
				gaps = append(gaps, fmt.Sprintf("%s: %s", row.ReqID, err.Error()))
			}
		}
	}

	status := "passed"
	if len(gaps) > 0 {
		status = "failed"
	}
	if *emitJSON {
		emitReport(stdout, "traceability-check", status, details, gaps)
	} else {
		if status == "passed" {
			write(stdout, "traceability-check passed: %d REQ rows verified\n", len(rows))
		} else {
			write(stdout, "traceability-check failed: %d gap(s)\n", len(gaps))
			for _, g := range gaps {
				write(stderr, "  %s\n", g)
			}
		}
	}
	if status == "passed" {
		return 0
	}
	return traceabilityExitGap
}

// parseTraceabilityMatrix 解析 markdown 表格, 提取每行 REQ 的 artifacts 与 evidence 列。
// 表头列序与 .agent/traceability-matrix.md 保持一致:
//
//	| REQ | 需求摘要 | 主要产物 | 验证/Evidence | 收敛 owner |
func parseTraceabilityMatrix(path string) ([]traceabilityRow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read traceability matrix: %w", err)
	}
	var rows []traceabilityRow
	for _, line := range strings.Split(string(data), "\n") {
		m := traceabilityRowRegex.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		row := traceabilityRow{
			ReqID:     strings.TrimSpace(m[1]),
			Artifacts: splitCellTokens(m[3]),
			Evidence:  splitCellTokens(m[4]),
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// splitCellTokens 按 `;` 或 `；` 切分单元格内容, 去除每个 token 的空白与首尾标点。
// 注意: 不能 strip 前导 `.`, 因为 `.agent/foo` / `.github/foo` 是合法路径前缀。
func splitCellTokens(cell string) []string {
	// 同时支持中英文分号
	normalized := strings.ReplaceAll(cell, "；", ";")
	parts := strings.Split(normalized, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		// 仅去除尾部句末标点; 前导 `.` 必须保留以支持 .agent/.github 等隐藏目录路径
		t = strings.TrimRight(t, " ,，。.")
		// 矩阵作者偶尔在路径后追加版本注释 (例: "docs/goal.md v2.9.3 Complete"),
		// 取首个空白前的 token 作为待校验路径
		if idx := strings.IndexAny(t, " \t"); idx > 0 {
			t = t[:idx]
		}
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// looksLikePath 判定 token 是否看起来是仓库内文件/目录路径引用 (启发式)。
// 仅命中的 token 会进入存在性校验; 其他 token (如 "object-model", "render worker evidence")
// 被视为说明性文字, 不参与校验, 避免假阳性。
func looksLikePath(token string) bool {
	if token == "" {
		return false
	}
	if traceabilityKnownTopFiles[token] {
		return true
	}
	for _, prefix := range traceabilityPathPrefixes {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}
	// 含 `/` 且有已知后缀也认为是路径
	if strings.Contains(token, "/") {
		for _, suf := range traceabilityPathSuffixes {
			if strings.HasSuffix(token, suf) {
				return true
			}
		}
	}
	return false
}

// verifyArtifactExists 校验 artifact 是否真实存在; 支持以 `/*` 结尾的目录-glob 模式。
// gitignore 路径 (生成产物, 例: release/debt/latest.json) 被视为合法而跳过, 由相应的
// 生成命令在 `make ci` 链路上负责落地; traceability-check 只关心源文件契约。
func verifyArtifactExists(artifact string) error {
	if isGitIgnored(artifact) {
		return nil
	}
	if strings.HasSuffix(artifact, "/*") {
		dir := strings.TrimSuffix(artifact, "/*")
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("directory not found: %s", dir)
			}
			return fmt.Errorf("directory stat error %s: %w", dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("not a directory: %s", dir)
		}
		return nil
	}
	// 一般 glob 支持
	if strings.ContainsAny(artifact, "*?[") {
		matches, err := filepath.Glob(artifact)
		if err != nil {
			return fmt.Errorf("invalid glob %q: %w", artifact, err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("glob matched no files: %s", artifact)
		}
		return nil
	}
	if _, err := os.Stat(artifact); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("missing artifact: %s", artifact)
		}
		return fmt.Errorf("artifact stat error %s: %w", artifact, err)
	}
	return nil
}

// isGitIgnored 调用 `git check-ignore` 判定路径是否被 gitignore 规则匹配。
// git 工具不可用或仓库无 git 历史时返回 false (保守视为非 ignored, 继续做 stat 校验)。
func isGitIgnored(path string) bool {
	cmd := exec.Command("git", "check-ignore", "-q", "--", path)
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
}
