#!/usr/bin/env python3
"""Extract all RULE-* definitions from goal-patch.md into a machine-readable registry.yaml.

源文档约定:
- 一级章节: `# N. Title`
- 规则定义: `## RULE-<DOMAIN>-NNN：<title>` 或 `### RULE-<DOMAIN>-NNN：<title>`
  (中文冒号或英文冒号; v1.8 §237 RULE-FREEZE-RULE-001/002 是 ### 级别)

输出: .agent/rules/registry.yaml (字段见 schema 注释)
"""
from __future__ import annotations

import os
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def _find_source() -> Path:
    """定位 goal-patch.md 源文件。

    顺序:
      1. 环境变量 XLIB_GOAL_PATCH_PATH (绝对路径)
      2. 从脚本所在目录向上搜寻包含 `.worktree/goal-patch.md` 的祖先目录
         (使得脚本既能从主仓库根目录运行, 也能从任意 worktree 工作区运行)
    避免在源码中出现仓库目录名字面值, 防止下游模板渲染时残留导致 stale module 误报。
    """
    env = os.environ.get("XLIB_GOAL_PATCH_PATH")
    if env:
        return Path(env)
    here = Path(__file__).resolve()
    rel = Path(".worktree") / "goal-patch.md"
    for parent in here.parents:
        candidate = parent / rel
        if candidate.is_file():
            return candidate
    sys.exit(
        "ERROR: goal-patch.md not found; set XLIB_GOAL_PATCH_PATH or run inside a "
        "tree that contains .worktree/goal-patch.md"
    )


SRC = _find_source()
DST = ROOT / ".agent" / "rules" / "registry.yaml"

# ---- 分级策略 (基于 id 前缀关键词) ----
P0_PREFIXES = {
    "CORE", "WORKTREE", "WORKTREE-AUTO", "WORKTREE-CLEAN",
    "HARNESS", "HARNESS-TEST",
    "SECURITY", "SECRET", "SECRET-CHECK",
    "EVIDENCE", "EVIDENCE-ANTI-FAKE", "EVIDENCE-ALG", "EVID-ALG",
    "TRACE", "TRACE-ALG",
    "RELEASE", "RELEASE-AUTO", "RELEASE-CHECK", "RELEASE-CHANNEL", "REL-ARTIFACT",
    "MERGE", "MAIN-SYNC",
    "BRANCH", "BRANCH-PROTECTION",
    "COMMIT", "COMMIT-AUTO", "COMMIT-EVID",
    "DOD", "DOR",
    "PR", "PR-AUTO", "PR-CHECK", "PR-LIFECYCLE", "PR-SYNC",
    "ISSUE", "ISSUE-AUTO", "ISSUE-LIFECYCLE",
    "ROLLBACK",
    "STATE", "STATE-GATE",
    "GATE-CONSISTENCY",
    "HUMAN",
    "GHA", "GHA-WORKFLOW",
    "MAKE",
    "SUPPLY",
    "STOP", "BLOCKER",
    "WAIVER", "VIOLATION",
}

# ---- enforced_by 映射: id 前缀 -> 可执行 gate 命令 ----
ENFORCED_BY = {
    "CORE-001": ("schedulex evidence-check", 8),
    "CORE-002": ("schedulex context-fast-check", 1),
    "CORE-003": ("schedulex acceptance-matrix", 1),
    "CORE-004": ("schedulex traceability-check", 9),
    "CORE-005": ("make governance-check", 1),
    "CORE-006": ("schedulex self-improving-check", 1),
    "WORKTREE": ("schedulex worktree-guard", 5),
    "WORKTREE-AUTO": ("schedulex worktree-guard", 5),
    "WORKTREE-CLEAN": ("schedulex worktree-guard", 5),
    "HARNESS": ("make governance-check", 1),
    "HARNESS-TEST": ("schedulex governance-fixture-test", 1),
    "SECURITY": ("schedulex secrets", 7),
    "SECRET": ("schedulex secrets", 7),
    "SECRET-CHECK": ("schedulex secrets", 7),
    "EVIDENCE": ("schedulex evidence-check", 8),
    "EVIDENCE-ANTI-FAKE": ("schedulex evidence-check", 8),
    "EVID-ALG": ("schedulex evidence-check", 8),
    "EVIDENCE-ALG": ("schedulex evidence-check", 8),
    "TRACE": ("schedulex traceability-check", 9),
    "TRACE-ALG": ("schedulex traceability-check", 9),
    "RELEASE": ("schedulex release-evidence-check", 10),
    "RELEASE-AUTO": ("schedulex release-final-check", 10),
    "RELEASE-CHECK": ("schedulex release-evidence-check", 10),
    "RELEASE-CHANNEL": ("schedulex release-final-check", 10),
    "REL-ARTIFACT": ("schedulex release-evidence-check", 10),
    "MERGE": ("make pr-check", 1),
    "MAIN-SYNC": ("schedulex main-guard", 5),
    "BRANCH": ("schedulex worktree-guard", 5),
    "BRANCH-PROTECTION": ("schedulex github-settings", 1),
    "COMMIT": (".githooks/pre-commit", 1),
    "COMMIT-AUTO": (".githooks/pre-commit", 1),
    "COMMIT-EVID": ("schedulex evidence-check", 8),
    "DOD": ("schedulex done-assertion", 1),
    "DOR": ("schedulex goal-acceptance", 1),
    "PR": ("schedulex pr-template", 1),
    "PR-AUTO": ("schedulex pr-template", 1),
    "PR-CHECK": ("make pr-check", 1),
    "PR-LIFECYCLE": ("schedulex pr-template", 1),
    "PR-SYNC": ("schedulex pr-template", 1),
    "ISSUE": ("schedulex issue-registry", 1),
    "ISSUE-AUTO": ("schedulex issue-registry", 1),
    "ISSUE-LIFECYCLE": ("schedulex issue-registry", 1),
    "ROLLBACK": ("schedulex release-final-check", 10),
    "STATE": ("schedulex goal-runtime", 1),
    "STATE-GATE": ("schedulex goal-runtime", 1),
    "GATE-CONSISTENCY": ("schedulex makefile-baseline", 1),
    "MAKE": ("schedulex makefile-baseline", 1),
    "GHA": ("schedulex github-governance", 1),
    "GHA-WORKFLOW": ("schedulex github-governance", 1),
    "HUMAN": ("", 0),
    "SUPPLY": ("schedulex dependency-check", 1),
    "STOP": ("", 0),
    "WAIVER": ("", 0),
    "VIOLATION": ("", 0),
    # 非 P0 但常用
    "BOUNDARY": ("schedulex boundary", 1),
    "DEBT": ("schedulex debt", 1),
    "DOC": ("schedulex docs-check", 1),
    "DEPENDENCY": ("schedulex dependency-check", 1),
    "DOWNSTREAM": ("schedulex downstream-adoption", 1),
    "DOWNSTREAM-SYNC": ("schedulex downstream-registry", 1),
    "DOWNSTREAM-GATE": ("schedulex downstream-adoption", 1),
    "ADOPTION-GATE": ("schedulex downstream-adoption", 1),
    "AGENT-TEAM": ("schedulex agent-team-contract", 1),
    "GOAL": ("schedulex goal-runtime", 1),
    "GOALPACK": ("schedulex pack-gate", 1),
    "GOALCLI": ("schedulex", 1),
    "GOALCLI-EXIT": ("schedulex", 1),
    "GOALCLI-CONFIG": ("schedulex", 1),
    "GOALCLI-ARCH": ("schedulex", 1),
    "REGISTRY": ("schedulex command-registry", 1),
    "CHECKER": ("schedulex", 1),
    "SCORE": ("schedulex score", 1),
    "SCORE-V14": ("schedulex score", 1),
    "AUDIT": ("schedulex goal-certify", 1),
    "AUDIT-CHECK": ("schedulex goal-certify", 1),
    "RETRO": ("schedulex self-improving-check", 1),
    "RETRO-CHECK": ("schedulex self-improving-check", 1),
    "SI": ("schedulex self-improving-check", 1),
    "SCHEMA": ("schedulex policy-schema", 6),
    "SCHEMA-MIN": ("schedulex policy-schema", 6),
    "CONTRACT": ("schedulex contracts", 1),
    "CONTRACTS": ("schedulex contracts", 1),
    "CLI-CONTRACT": ("schedulex cli-contract", 1),
    "DRIFT": ("schedulex docs-drift", 1),
    "DEPRECATION": ("", 0),
    "TEMPLATE": ("", 0),
    "HOOKS": (".githooks/pre-commit", 1),
    "REPORT": ("", 0),
    "METRIC": ("", 0),
    "CHANGE": ("", 0),
    "CHANGELOG": ("schedulex changelog", 1),
    "FILE": ("", 0),
    "ANTI-CARGO": ("", 0),
    "ANTI-FRAGILE": ("", 0),
    "AR": ("schedulex autoresearch", 1),
    "RESEARCH": ("schedulex autoresearch", 1),
    "FACTORY": ("", 0),
    "ID": ("", 0),
    "OBJECT": ("", 0),
    "MODE": ("", 0),
    "MODE-GATE": ("", 0),
    "CLASS": ("", 0),
    "PRIORITY": ("", 0),
    "CODE": ("", 0),
    "SIMPLICITY": ("", 0),
    "ORDER": ("", 0),
    "DESIGN": ("", 0),
    "TASK": ("", 0),
    "RISK": ("", 0),
    "DECISION": ("", 0),
    "CONTEXT": ("schedulex context-fast-check", 1),
    "CONTEXT-COMPRESSION": ("", 0),
    "CI": ("schedulex context-fast-check", 1),
    "VERSION": ("", 0),
    "RUNTIME-COMPAT": ("", 0),
    "DEPENDENCY-SCAN": ("schedulex dependency-check", 1),
    "MILESTONE": ("", 0),
    "LABEL": ("", 0),
    "REVIEW": ("schedulex pr-template", 1),
    "REVIEW-BOT": ("", 0),
    "PR-BOT": ("", 0),
    "HANDOFF": ("schedulex goal-handover", 1),
    "REPO-LAYOUT": ("", 0),
    "ROOT": ("", 0),
    "RUNBOOK": ("", 0),
    "REPAIR": ("", 0),
    "AGENT": ("", 0),
    "AGENT-AUTH": ("", 0),
    "AGENT-MEMORY": ("", 0),
    "CONCURRENCY": ("", 0),
    "FAILURE": ("", 0),
    "BATCH": ("", 0),
    "AUDIT-LEVEL": ("schedulex goal-certify", 1),
    "ADOPTION-SCORE": ("schedulex downstream-adoption", 1),
    "AUTO-SAFETY": ("", 0),
    "ARCHIVE": ("", 0),
    "GOLDEN": ("", 0),
    "SPEC": ("", 0),
    "XSTACK": ("", 0),
    "XGO": ("", 0),
    "TASK-AUTO": ("", 0),
    # === Active promotion batch 1 (2026-06-03)：仅绑定到工具已存在且语义匹配的前缀 ===
    # Goal 对象模型 → goal-runtime
    "OBJECT": ("schedulex goal-runtime", 1),
    "ID": ("schedulex goal-runtime", 1),
    "CONTROL": ("schedulex goal-runtime", 1),
    "SSOT": ("schedulex goal-runtime", 1),
    "ORPHAN": ("schedulex goal-runtime", 1),
    "CONFLICT": ("schedulex goal-runtime", 1),
    "MODE": ("schedulex goal-runtime", 1),
    "MODE-GATE": ("schedulex goal-runtime", 1),
    "CLASS": ("schedulex goal-runtime", 1),
    "PRIORITY": ("schedulex goal-runtime", 1),
    "ORDER": ("schedulex goal-runtime", 1),
    "MILESTONE": ("schedulex goal-runtime", 1),
    # 可追溯性 / 影响 / 验收
    "TASK": ("schedulex traceability-check", 9),
    "CHANGE-TYPE": ("schedulex traceability-check", 9),
    "COVERAGE": ("schedulex traceability-check", 9),
    "FILE": ("schedulex runtime-file-ownership", 1),
    "OWNERSHIP": ("schedulex runtime-file-ownership", 1),
    "SPEC": ("schedulex acceptance-matrix", 1),
    "IMPACT": ("schedulex standard-impact-check", 1),
    # Schema / 兼容 / 迁移 / 版本
    "COMPAT": ("schedulex policy-schema", 6),
    "COMPAT-MATRIX": ("schedulex policy-schema", 6),
    "COMPAT-GUARD": ("schedulex downstream-adoption", 1),
    "SUNSET": ("schedulex policy-schema", 6),
    "MIGRATION": ("schedulex policy-schema", 6),
    "RUNTIME-COMPAT": ("schedulex upgrade-runtime", 1),
    # Agent 平面 → runtime-health / agent-team-contract / self-healing-skeleton
    "AGENT": ("schedulex agent-team-contract", 1),
    "AGENT-AUTH": ("schedulex agent-team-contract", 1),
    "AGENT-MEMORY": ("schedulex agent-team-contract", 1),
    "AUTO-SAFETY": ("schedulex runtime-health", 1),
    "HEARTBEAT": ("schedulex runtime-health", 1),
    "LEASE": ("schedulex runtime-health", 1),
    "DOCTOR": ("schedulex runtime-health", 1),
    "RECONCILE": ("schedulex runtime-health", 1),
    "REPAIR": ("schedulex self-healing-skeleton", 1),
    # Worktree / main 恢复 / freeze
    "CONCURRENCY": ("schedulex worktree-guard", 5),
    "WT-GC": ("schedulex worktree-guard", 5),
    "MAIN-RECOVERY": ("schedulex main-guard", 5),
    "FREEZE": ("schedulex scope-lock", 1),
    "GOAL-FREEZE": ("schedulex scope-lock", 1),
    # 安装 / Profile / 成熟度
    "BOOTSTRAP": ("schedulex install-runtime", 1),
    "PROFILE": ("schedulex conformance-profile", 1),
    "MATURITY": ("schedulex conformance-profile", 1),
    # Evidence 扩展
    "EVIDENCE-HASH": ("schedulex release-evidence-hash", 8),
    "EVIDENCE-COVERAGE": ("schedulex evidence-check", 8),
    "EVIDENCE-RETENTION": ("schedulex evidence-check", 8),
    "EVID-LOSS": ("schedulex evidence-check", 8),
    # GitHub / PR / Issue
    "GITHUB-ISSUE": ("schedulex issue-registry", 1),
    "ISSUE-CANDIDATE": ("schedulex issue-registry", 1),
    "LABEL": ("schedulex github-settings", 1),
    "PERMISSION": ("schedulex github-settings", 1),
    "PR-SIZE": ("schedulex pr-template", 1),
    "PR-BOT": ("schedulex pr-template", 1),
    "REVIEW-BOT": ("schedulex pr-template", 1),
    "MERGE-QUEUE": ("schedulex pr-template", 1),
    # Release 扩展
    "RELEASE-TRAIN": ("schedulex release-final-check", 10),
    "PARTIAL-RELEASE": ("schedulex release-final-check", 10),
    "PROMOTE": ("schedulex downstream-adoption", 1),
    "PROMOTION": ("schedulex release-final-check", 10),
    "ROLLFORWARD": ("schedulex release-final-check", 10),
    # Downstream / xstack
    "XSTACK": ("schedulex attest-conformance", 1),
    "XSTACK-ADMISSION": ("schedulex attest-conformance", 1),
    "DOWNSTREAM-CONTRACT": ("schedulex downstream-registry", 1),
    "MULTIREPO": ("schedulex downstream-registry", 1),
    # Makefile / Gate DAG
    "GATE-DAG": ("schedulex makefile-baseline", 1),
    "PARITY": ("schedulex makefile-baseline", 1),
    # Registry 一致性
    "REGISTRY-CONSISTENCY": ("schedulex command-registry", 1),
    "REGISTRY-LOCK": ("schedulex command-registry", 1),
    # 文档 / 规则维护
    "DOC-DEBT": ("schedulex debt", 1),
    "DRIFT-BUDGET": ("schedulex debt", 1),
    "TEMPLATE": ("schedulex docs-check", 1),
    "RULE-BLOAT": ("schedulex docs-check", 1),
    "RULE-PATCH": ("schedulex docs-check", 1),
    "COMPILER": ("schedulex cli-contract", 1),
    "GLOSSARY": ("make governance-check", 1),
    "CODE": ("make governance-check", 1),
    # Goal 测试 / Golden / 违规样例
    "GOAL-TEST": ("schedulex governance-fixture-test", 1),
    "GOLDEN": ("schedulex governance-fixture-test", 1),
    "GOLDEN-PACK": ("schedulex pack-gate", 1),
    "VIOLATION-FIXTURE": ("schedulex governance-fixture-test", 1),
    # Context 子集
    "CONTEXT-COMPRESSION": ("schedulex execution-context", 1),
    "CONTEXT-WINDOW": ("schedulex execution-context", 1),
    # 仓库布局 / 命名 / 极简
    "REPO-LAYOUT": ("schedulex boundary", 1),
    "ROOT": ("schedulex boundary", 1),
    "XGO": ("schedulex boundary", 1),
    "NAMING": ("schedulex naming", 1),
    "SIMPLICITY": ("schedulex minimal-kernel", 1),
    # P0: WAIVER / VIOLATION / STOP 走 governance-check
    "WAIVER": ("make governance-check", 1),
    "VIOLATION": ("make governance-check", 1),
    "STOP": ("make governance-check", 1),
    "HUMAN": ("schedulex pr-template", 1),
    # Version 同 changelog 已绑过的方向
    "VERSION": ("schedulex changelog", 1),
}

DEFINITION_RE = re.compile(r"^#{2,3}\s+(RULE-[A-Z-]+-\d+)[：:]\s*(.+)$")
SECTION_RE = re.compile(r"^#{1,2} (\d+)\. (.+)$")


def split_prefix(rid: str) -> str:
    """RULE-FOO-BAR-001 -> FOO-BAR"""
    parts = rid.split("-")
    return "-".join(parts[1:-1])


def classify(rid: str) -> str:
    prefix = split_prefix(rid)
    return "P0" if prefix in P0_PREFIXES else "P1"


def lookup_enforced(rid: str) -> tuple[str, int]:
    """完整 id 优先 (RULE-CORE-001)，否则 fallback 到前缀"""
    short = rid.removeprefix("RULE-")
    if short in ENFORCED_BY:
        return ENFORCED_BY[short]
    prefix = split_prefix(rid)
    return ENFORCED_BY.get(prefix, ("", 0))


def main() -> int:
    text = SRC.read_text(encoding="utf-8")
    lines = text.split("\n")
    rules: dict[str, dict] = {}
    cur_section_num = ""
    cur_section_title = ""
    for i, line in enumerate(lines, 1):
        m_sec = SECTION_RE.match(line)
        if m_sec:
            cur_section_num = m_sec.group(1)
            cur_section_title = m_sec.group(2).strip()
            continue
        m = DEFINITION_RE.match(line)
        if not m:
            continue
        rid = m.group(1)
        title = m.group(2).strip()
        if rid in rules:
            # 重复定义: 记录到 duplicate_at
            rules[rid].setdefault("duplicate_at", []).append(i)
            continue
        enforced_by, exit_code = lookup_enforced(rid)
        status = "active" if enforced_by else "indexed"
        rules[rid] = {
            "id": rid,
            "level": classify(rid),
            "title": title,
            "source_section": cur_section_num,
            "source_section_title": cur_section_title,
            "source_line": i,
            "enforced_by": enforced_by,
            "exit_code": exit_code,
            "status": status,
        }

    # ---- 序列化为 YAML (手写以保证字段顺序 + 注释) ----
    out: list[str] = []
    out.append("# .agent/rules/registry.yaml")
    out.append("# 自动从 .worktree/goal-patch.md 提取生成。")
    out.append("# 重新生成: python3 scripts/extract_rules.py")
    out.append("# 字段语义:")
    out.append("#   id                   - 唯一规则 ID")
    out.append("#   level                - P0 (release-blocking) / P1 (governance) / P2 (advisory)")
    out.append("#   title                - 规则一句话标题 (来源: ## 行)")
    out.append("#   source_section       - goal-patch.md 一级章节编号")
    out.append("#   source_section_title - 章节标题")
    out.append("#   source_line          - 定义所在行号")
    out.append("#   enforced_by          - 实际执行该规则的命令 (空=尚未机器化)")
    out.append("#   exit_code            - 违规时该命令应返回的标准退出码")
    out.append("#   status               - active (有 enforced_by) / indexed (仅登记) / deprecated")
    out.append("#   duplicate_at         - 该 id 在源文档中被二次定义的行号 (若有)")
    out.append("")
    out.append("version: 1")
    out.append(f"generated_from: .worktree/goal-patch.md")
    out.append(f"generated_at: {date.today().isoformat()}")
    out.append(f"total_rules: {len(rules)}")
    p0 = sum(1 for r in rules.values() if r["level"] == "P0")
    p1 = sum(1 for r in rules.values() if r["level"] == "P1")
    active = sum(1 for r in rules.values() if r["status"] == "active")
    out.append(f"p0_count: {p0}")
    out.append(f"p1_count: {p1}")
    out.append(f"active_count: {active}")
    out.append(f"indexed_count: {len(rules) - active}")
    out.append("")
    out.append("rules:")
    # 排序: P0 优先, 同级按 id
    for r in sorted(rules.values(), key=lambda x: (0 if x["level"] == "P0" else 1, x["id"])):
        out.append(f"  - id: {r['id']}")
        out.append(f"    level: {r['level']}")
        # title 用 YAML block scalar 避免转义中文冒号
        title_escaped = r["title"].replace('"', '\\"')
        out.append(f'    title: "{title_escaped}"')
        out.append(f"    source_section: {r['source_section']}")
        sec_title_escaped = r["source_section_title"].replace('"', '\\"')
        out.append(f'    source_section_title: "{sec_title_escaped}"')
        out.append(f"    source_line: {r['source_line']}")
        out.append(f'    enforced_by: "{r["enforced_by"]}"')
        out.append(f"    exit_code: {r['exit_code']}")
        out.append(f"    status: {r['status']}")
        if "duplicate_at" in r:
            out.append(f"    duplicate_at: {r['duplicate_at']}")

    DST.write_text("\n".join(out) + "\n", encoding="utf-8")
    print(f"wrote {DST} ({len(rules)} rules, P0={p0}, P1={p1}, active={active})")
    return 0


if __name__ == "__main__":
    sys.exit(main())
