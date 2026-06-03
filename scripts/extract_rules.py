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
    "CORE-001": ("goalcli evidence-check", 8),
    "CORE-002": ("goalcli context-fast-check", 1),
    "CORE-003": ("goalcli acceptance-matrix", 1),
    "CORE-004": ("goalcli traceability-check", 9),
    "CORE-005": ("make governance-check", 1),
    "CORE-006": ("goalcli self-improving-check", 1),
    "WORKTREE": ("goalcli worktree-guard", 5),
    "WORKTREE-AUTO": ("goalcli worktree-guard", 5),
    "WORKTREE-CLEAN": ("goalcli worktree-guard", 5),
    "HARNESS": ("make governance-check", 1),
    "HARNESS-TEST": ("goalcli governance-fixture-test", 1),
    "SECURITY": ("goalcli secrets", 7),
    "SECRET": ("goalcli secrets", 7),
    "SECRET-CHECK": ("goalcli secrets", 7),
    "EVIDENCE": ("goalcli evidence-check", 8),
    "EVIDENCE-ANTI-FAKE": ("goalcli evidence-check", 8),
    "EVID-ALG": ("goalcli evidence-check", 8),
    "EVIDENCE-ALG": ("goalcli evidence-check", 8),
    "TRACE": ("goalcli traceability-check", 9),
    "TRACE-ALG": ("goalcli traceability-check", 9),
    "RELEASE": ("goalcli release-evidence-check", 10),
    "RELEASE-AUTO": ("goalcli release-final-check", 10),
    "RELEASE-CHECK": ("goalcli release-evidence-check", 10),
    "RELEASE-CHANNEL": ("goalcli release-final-check", 10),
    "REL-ARTIFACT": ("goalcli release-evidence-check", 10),
    "MERGE": ("make pr-check", 1),
    "MAIN-SYNC": ("goalcli main-guard", 5),
    "BRANCH": ("goalcli worktree-guard", 5),
    "BRANCH-PROTECTION": ("goalcli github-settings", 1),
    "COMMIT": (".githooks/pre-commit", 1),
    "COMMIT-AUTO": (".githooks/pre-commit", 1),
    "COMMIT-EVID": ("goalcli evidence-check", 8),
    "DOD": ("goalcli done-assertion", 1),
    "DOR": ("goalcli goal-acceptance", 1),
    "PR": ("goalcli pr-template", 1),
    "PR-AUTO": ("goalcli pr-template", 1),
    "PR-CHECK": ("make pr-check", 1),
    "PR-LIFECYCLE": ("goalcli pr-template", 1),
    "PR-SYNC": ("goalcli pr-template", 1),
    "ISSUE": ("goalcli issue-registry", 1),
    "ISSUE-AUTO": ("goalcli issue-registry", 1),
    "ISSUE-LIFECYCLE": ("goalcli issue-registry", 1),
    "ROLLBACK": ("goalcli release-final-check", 10),
    "STATE": ("goalcli goal-runtime", 1),
    "STATE-GATE": ("goalcli goal-runtime", 1),
    "GATE-CONSISTENCY": ("goalcli makefile-baseline", 1),
    "MAKE": ("goalcli makefile-baseline", 1),
    "GHA": ("goalcli github-governance", 1),
    "GHA-WORKFLOW": ("goalcli github-governance", 1),
    "HUMAN": ("", 0),
    "SUPPLY": ("goalcli dependency-check", 1),
    "STOP": ("", 0),
    "WAIVER": ("", 0),
    "VIOLATION": ("", 0),
    # 非 P0 但常用
    "BOUNDARY": ("goalcli boundary", 1),
    "DEBT": ("goalcli debt", 1),
    "DOC": ("goalcli docs-check", 1),
    "DEPENDENCY": ("goalcli dependency-check", 1),
    "DOWNSTREAM": ("goalcli downstream-adoption", 1),
    "DOWNSTREAM-SYNC": ("goalcli downstream-registry", 1),
    "DOWNSTREAM-GATE": ("goalcli downstream-adoption", 1),
    "ADOPTION-GATE": ("goalcli downstream-adoption", 1),
    "AGENT-TEAM": ("goalcli agent-team-contract", 1),
    "GOAL": ("goalcli goal-runtime", 1),
    "GOALPACK": ("goalcli pack-gate", 1),
    "GOALCLI": ("goalcli", 1),
    "GOALCLI-EXIT": ("goalcli", 1),
    "GOALCLI-CONFIG": ("goalcli", 1),
    "GOALCLI-ARCH": ("goalcli", 1),
    "REGISTRY": ("goalcli command-registry", 1),
    "CHECKER": ("goalcli", 1),
    "SCORE": ("goalcli score", 1),
    "SCORE-V14": ("goalcli score", 1),
    "AUDIT": ("goalcli goal-certify", 1),
    "AUDIT-CHECK": ("goalcli goal-certify", 1),
    "RETRO": ("goalcli self-improving-check", 1),
    "RETRO-CHECK": ("goalcli self-improving-check", 1),
    "SI": ("goalcli self-improving-check", 1),
    "SCHEMA": ("goalcli policy-schema", 6),
    "SCHEMA-MIN": ("goalcli policy-schema", 6),
    "CONTRACT": ("goalcli contracts", 1),
    "CONTRACTS": ("goalcli contracts", 1),
    "CLI-CONTRACT": ("goalcli cli-contract", 1),
    "DRIFT": ("goalcli docs-drift", 1),
    "DEPRECATION": ("", 0),
    "TEMPLATE": ("", 0),
    "HOOKS": (".githooks/pre-commit", 1),
    "REPORT": ("", 0),
    "METRIC": ("", 0),
    "CHANGE": ("", 0),
    "CHANGELOG": ("goalcli changelog", 1),
    "FILE": ("", 0),
    "ANTI-CARGO": ("", 0),
    "ANTI-FRAGILE": ("", 0),
    "AR": ("goalcli autoresearch", 1),
    "RESEARCH": ("goalcli autoresearch", 1),
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
    "CONTEXT": ("goalcli context-fast-check", 1),
    "CONTEXT-COMPRESSION": ("", 0),
    "CI": ("goalcli context-fast-check", 1),
    "VERSION": ("", 0),
    "RUNTIME-COMPAT": ("", 0),
    "DEPENDENCY-SCAN": ("goalcli dependency-check", 1),
    "MILESTONE": ("", 0),
    "LABEL": ("", 0),
    "REVIEW": ("goalcli pr-template", 1),
    "REVIEW-BOT": ("", 0),
    "PR-BOT": ("", 0),
    "HANDOFF": ("goalcli goal-handover", 1),
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
    "AUDIT-LEVEL": ("goalcli goal-certify", 1),
    "ADOPTION-SCORE": ("goalcli downstream-adoption", 1),
    "AUTO-SAFETY": ("", 0),
    "ARCHIVE": ("", 0),
    "GOLDEN": ("", 0),
    "SPEC": ("", 0),
    "XSTACK": ("", 0),
    "XGO": ("", 0),
    "TASK-AUTO": ("", 0),
    # === Active promotion batch 1 (2026-06-03)：仅绑定到工具已存在且语义匹配的前缀 ===
    # Goal 对象模型 → goal-runtime
    "OBJECT": ("goalcli goal-runtime", 1),
    "ID": ("goalcli goal-runtime", 1),
    "CONTROL": ("goalcli goal-runtime", 1),
    "SSOT": ("goalcli goal-runtime", 1),
    "ORPHAN": ("goalcli goal-runtime", 1),
    "CONFLICT": ("goalcli goal-runtime", 1),
    "MODE": ("goalcli goal-runtime", 1),
    "MODE-GATE": ("goalcli goal-runtime", 1),
    "CLASS": ("goalcli goal-runtime", 1),
    "PRIORITY": ("goalcli goal-runtime", 1),
    "ORDER": ("goalcli goal-runtime", 1),
    "MILESTONE": ("goalcli goal-runtime", 1),
    # 可追溯性 / 影响 / 验收
    "TASK": ("goalcli traceability-check", 9),
    "CHANGE-TYPE": ("goalcli traceability-check", 9),
    "COVERAGE": ("goalcli traceability-check", 9),
    "FILE": ("goalcli runtime-file-ownership", 1),
    "OWNERSHIP": ("goalcli runtime-file-ownership", 1),
    "SPEC": ("goalcli acceptance-matrix", 1),
    "IMPACT": ("goalcli standard-impact-check", 1),
    # Schema / 兼容 / 迁移 / 版本
    "COMPAT": ("goalcli policy-schema", 6),
    "COMPAT-MATRIX": ("goalcli policy-schema", 6),
    "COMPAT-GUARD": ("goalcli downstream-adoption", 1),
    "SUNSET": ("goalcli policy-schema", 6),
    "MIGRATION": ("goalcli policy-schema", 6),
    "RUNTIME-COMPAT": ("goalcli upgrade-runtime", 1),
    # Agent 平面 → runtime-health / agent-team-contract / self-healing-skeleton
    "AGENT": ("goalcli agent-team-contract", 1),
    "AGENT-AUTH": ("goalcli agent-team-contract", 1),
    "AGENT-MEMORY": ("goalcli agent-team-contract", 1),
    "AUTO-SAFETY": ("goalcli runtime-health", 1),
    "HEARTBEAT": ("goalcli runtime-health", 1),
    "LEASE": ("goalcli runtime-health", 1),
    "DOCTOR": ("goalcli runtime-health", 1),
    "RECONCILE": ("goalcli runtime-health", 1),
    "REPAIR": ("goalcli self-healing-skeleton", 1),
    # Worktree / main 恢复 / freeze
    "CONCURRENCY": ("goalcli worktree-guard", 5),
    "WT-GC": ("goalcli worktree-guard", 5),
    "MAIN-RECOVERY": ("goalcli main-guard", 5),
    "FREEZE": ("goalcli scope-lock", 1),
    "GOAL-FREEZE": ("goalcli scope-lock", 1),
    # 安装 / Profile / 成熟度
    "BOOTSTRAP": ("goalcli install-runtime", 1),
    "PROFILE": ("goalcli conformance-profile", 1),
    "MATURITY": ("goalcli conformance-profile", 1),
    # Evidence 扩展
    "EVIDENCE-HASH": ("goalcli release-evidence-hash", 8),
    "EVIDENCE-COVERAGE": ("goalcli evidence-check", 8),
    "EVIDENCE-RETENTION": ("goalcli evidence-check", 8),
    "EVID-LOSS": ("goalcli evidence-check", 8),
    # GitHub / PR / Issue
    "GITHUB-ISSUE": ("goalcli issue-registry", 1),
    "ISSUE-CANDIDATE": ("goalcli issue-registry", 1),
    "LABEL": ("goalcli github-settings", 1),
    "PERMISSION": ("goalcli github-settings", 1),
    "PR-SIZE": ("goalcli pr-template", 1),
    "PR-BOT": ("goalcli pr-template", 1),
    "REVIEW-BOT": ("goalcli pr-template", 1),
    "MERGE-QUEUE": ("goalcli pr-template", 1),
    # Release 扩展
    "RELEASE-TRAIN": ("goalcli release-final-check", 10),
    "PARTIAL-RELEASE": ("goalcli release-final-check", 10),
    "PROMOTE": ("goalcli downstream-adoption", 1),
    "PROMOTION": ("goalcli release-final-check", 10),
    "ROLLFORWARD": ("goalcli release-final-check", 10),
    # Downstream / xstack
    "XSTACK": ("goalcli attest-conformance", 1),
    "XSTACK-ADMISSION": ("goalcli attest-conformance", 1),
    "DOWNSTREAM-CONTRACT": ("goalcli downstream-registry", 1),
    "MULTIREPO": ("goalcli downstream-registry", 1),
    # Makefile / Gate DAG
    "GATE-DAG": ("goalcli makefile-baseline", 1),
    "PARITY": ("goalcli makefile-baseline", 1),
    # Registry 一致性
    "REGISTRY-CONSISTENCY": ("goalcli command-registry", 1),
    "REGISTRY-LOCK": ("goalcli command-registry", 1),
    # 文档 / 规则维护
    "DOC-DEBT": ("goalcli debt", 1),
    "DRIFT-BUDGET": ("goalcli debt", 1),
    "TEMPLATE": ("goalcli docs-check", 1),
    "RULE-BLOAT": ("goalcli docs-check", 1),
    "RULE-PATCH": ("goalcli docs-check", 1),
    "COMPILER": ("goalcli cli-contract", 1),
    "GLOSSARY": ("make governance-check", 1),
    "CODE": ("make governance-check", 1),
    # Goal 测试 / Golden / 违规样例
    "GOAL-TEST": ("goalcli governance-fixture-test", 1),
    "GOLDEN": ("goalcli governance-fixture-test", 1),
    "GOLDEN-PACK": ("goalcli pack-gate", 1),
    "VIOLATION-FIXTURE": ("goalcli governance-fixture-test", 1),
    # Context 子集
    "CONTEXT-COMPRESSION": ("goalcli execution-context", 1),
    "CONTEXT-WINDOW": ("goalcli execution-context", 1),
    # 仓库布局 / 命名 / 极简
    "REPO-LAYOUT": ("goalcli boundary", 1),
    "ROOT": ("goalcli boundary", 1),
    "XGO": ("goalcli boundary", 1),
    "NAMING": ("goalcli naming", 1),
    "SIMPLICITY": ("goalcli minimal-kernel", 1),
    # P0: WAIVER / VIOLATION / STOP 走 governance-check
    "WAIVER": ("make governance-check", 1),
    "VIOLATION": ("make governance-check", 1),
    "STOP": ("make governance-check", 1),
    "HUMAN": ("goalcli pr-template", 1),
    # Version 同 changelog 已绑过的方向
    "VERSION": ("goalcli changelog", 1),
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
