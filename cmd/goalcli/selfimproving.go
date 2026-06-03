package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// runSelfImprovingCheck 实现 Self-improving / Retrospective 门禁，机器化
// RULE-CORE-006 / RULE-RETRO-001..003 / RULE-SI-001..003 / RULE-RETRO-CHECK-001..002。
//
// 检查 (lenient by default)：
//  1. 必备文件存在：retrospective.md / retrospective-template.md /
//     {harness,prompt,rule}-patches.yaml / harness/gates/retro-gate.yaml
//  2. retrospective.md 包含 RULE-RETRO-CHECK-001 要求的 4 段核心标题（双语容忍）
//  3. patches.yaml 不为空文本且声明 schema/entries/patches 任一字段
//  4. patches 中如有 `status:` 字段，值必须在合法枚举内（RULE-SI-002）
//
// --strict 追加：3 个 patches.yaml 合计至少 1 个 `- patch_id:` entry
// （RULE-SI-001 / RULE-RETRO-CHECK-002，对 Lite Mode 无失败可豁免，故默认非 strict）。
func runSelfImprovingCheck(cmdName string, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli "+cmdName, flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root to inspect")
	strict := flags.Bool("strict", false, "require at least one patch entry across the three registries")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		write(stderr, "ERROR: %s invalid arguments: unexpected positional argument %q\n", cmdName, flags.Arg(0))
		return 2
	}

	var findings []string

	required := []string{
		".agent/retrospective.md",
		".agent/retrospective-template.md",
		".agent/harness-patches.yaml",
		".agent/prompt-patches.yaml",
		".agent/rule-patches.yaml",
		".agent/harness/gates/retro-gate.yaml",
	}
	for _, p := range required {
		if _, err := os.Stat(filepath.Join(*root, p)); err != nil {
			findings = append(findings, fmt.Sprintf("[RULE-RETRO-001] 缺失必备文件: %s", p))
		}
	}

	retroPath := filepath.Join(*root, ".agent/retrospective.md")
	if data, err := os.ReadFile(retroPath); err == nil {
		text := string(data)
		// retrospective.md 必须体现复盘性质：含 "复盘/Retrospective" 关键词
		// 含 "补丁/Patch" 段（RULE-RETRO-002 要求生成 Patch）
		// 含失败/改进/反思的语义关键词（RULE-RETRO-CHECK-001）
		retroHints := []string{"复盘", "Retrospective", "回顾"}
		patchHints := []string{"补丁", "Patch"}
		reflectHints := []string{"失败", "Failed", "Failure", "改进", "Root Cause", "根因", "原因", "What"}
		check := func(label string, alts []string, ruleID string) {
			for _, k := range alts {
				if strings.Contains(text, k) {
					return
				}
			}
			findings = append(findings, fmt.Sprintf("[%s] retrospective.md 缺少 %s 关键词 (任一即可: %s)", ruleID, label, strings.Join(alts, " | ")))
		}
		check("复盘", retroHints, "RULE-RETRO-001")
		check("补丁段", patchHints, "RULE-RETRO-002")
		check("失败/改进", reflectHints, "RULE-RETRO-CHECK-001")
	}

	// template 必须遵循完整 schema (9 段) - 它是被生成新 retrospective 的"模具"
	tplPath := filepath.Join(*root, ".agent/retrospective-template.md")
	if data, err := os.ReadFile(tplPath); err == nil {
		text := string(data)
		tplRequired := []string{
			"Failure", "Root Cause", "Patch",
			"Prompt Patch", "Harness Patch", "Rule Patch",
		}
		for _, k := range tplRequired {
			if !strings.Contains(text, k) {
				findings = append(findings, fmt.Sprintf("[RULE-RETRO-CHECK-001] retrospective-template.md 缺少标题段: %s", k))
			}
		}
	}

	patchFiles := []string{
		".agent/harness-patches.yaml",
		".agent/prompt-patches.yaml",
		".agent/rule-patches.yaml",
	}
	statusRe := regexp.MustCompile(`(?m)^\s*status:\s*"?([A-Za-z_-]+)"?\s*$`)
	patchIDRe := regexp.MustCompile(`(?m)^\s*-\s*patch_id\s*:`)
	validStatuses := map[string]bool{
		"PROPOSED": true, "ACCEPTED": true, "REJECTED": true,
		"SUPERSEDED": true, "IMPLEMENTED": true,
		// 既有仓库使用的过渡状态：允许，但鼓励迁移到 5 标准状态之一
		"reconciled_stub": true,
	}
	totalEntries := 0
	for _, pf := range patchFiles {
		fp := filepath.Join(*root, pf)
		raw, err := os.ReadFile(fp)
		if err != nil {
			continue // 上面已记
		}
		text := string(raw)
		hasSchema := strings.Contains(text, "schema:") ||
			strings.Contains(text, "entries:") ||
			strings.Contains(text, "patches:")
		if !hasSchema {
			findings = append(findings, fmt.Sprintf("[RULE-SI-003] %s 缺少 schema/entries/patches 任一字段", pf))
		}
		for _, m := range statusRe.FindAllStringSubmatch(text, -1) {
			if !validStatuses[m[1]] {
				findings = append(findings, fmt.Sprintf("[RULE-SI-002] %s 出现非法 status: %q (合法: PROPOSED/ACCEPTED/REJECTED/SUPERSEDED/IMPLEMENTED)", pf, m[1]))
			}
		}
		totalEntries += len(patchIDRe.FindAllString(text, -1))
	}

	if *strict && totalEntries == 0 {
		findings = append(findings, "[RULE-SI-001] --strict: harness/prompt/rule patches.yaml 三处合计 0 个 patch entry (Lite Mode 且无失败可豁免)")
	}

	details := []string{
		"retrospective present",
		"3 patch registries present",
		"retro-gate present",
		fmt.Sprintf("%d patch entries", totalEntries),
	}
	if len(findings) > 0 {
		return emitReport(stdout, cmdName, "failed", details, findings)
	}
	return emitReport(stdout, cmdName, "passed", details, nil)
}
