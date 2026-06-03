#!/usr/bin/env python3
"""把 registry.yaml + goal-patch.md 中的规则正文渲染到三个域规则文件。

输入:
  - .agent/rules/registry.yaml  (规则 SSOT, 含 id/level/source_section/source_line)
  - .worktree/goal-patch.md     (规则正文来源)

输出:
  - .agent/rules/core-rules.md             (CORE/CONTEXT/STATE/SSOT/ID/MODE 等)
  - .agent/rules/schema-registry-rules.md  (SCHEMA/REGISTRY/GOALPACK/MIGRATION 等)
  - .agent/rules/agent-runtime-rules.md    (AGENT/LEASE/HEARTBEAT/CMD-TXN/goalcli 等)

设计:
  - 不重写 iron-rules.md 已覆盖的 7 条铁律, 但允许域文件中重复引用 RULE-CORE-001..006
    作为锚点（goal-rules.md 中已有先例）。
  - 文件中规则按 source_section 分组排列, 二级标题 = 章节标题, 三级标题 = 规则 ID。
  - 正文从 goal-patch.md 中按"下一个 ### RULE- 或 ## 或 # 之前"截取, 去掉前后空行和
    末尾 `---` 分隔符。
  - 重新生成不破坏既有 *-rules.md, 仅写入三个目标文件。
"""
from __future__ import annotations

import os
import re
import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    sys.exit("ERROR: 需要 PyYAML; pip install pyyaml")

ROOT = Path(__file__).resolve().parents[1]


def _find_goal_patch() -> Path:
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
        "ERROR: goal-patch.md 未找到; 设置 XLIB_GOAL_PATCH_PATH 或在含 "
        ".worktree/goal-patch.md 的工作树内运行"
    )


REGISTRY = ROOT / ".agent" / "rules" / "registry.yaml"
SRC = _find_goal_patch()
OUT_DIR = ROOT / ".agent" / "rules"


# ---- 规则族 → 目标文件 ----
BUCKETS: dict[str, dict] = {
    "core-rules.md": {
        "title": "Core 规则",
        "intro": (
            "本文件覆盖 Goal Runtime **核心控制层**规则：第一性铁律、对象模型、"
            "ID/状态机/模式、Context Recovery、SSOT、规则分级与评分、规则冻结等。\n\n"
            "权威顺序参见 [`README.md`](./README.md)；P0 铁律压缩见 "
            "[`iron-rules.md`](./iron-rules.md)；机器化字段见 "
            "[`registry.yaml`](./registry.yaml)。"
        ),
        "families": {
            "CORE", "CONTEXT", "STATE", "STATE-GATE", "SSOT", "FREEZE",
            "FREEZE-RULE", "GOAL-FREEZE", "MATURITY", "ID", "MODE", "MODE-GATE",
            "FAILURE", "CLASS", "PRIORITY", "SCORE", "OBJECT", "CONFLICT",
            "ORDER", "GLOSSARY", "NAMING", "PROFILE", "RUNTIME-COMPAT", "ROOT",
            "SIMPLICITY",
        },
    },
    "schema-registry-rules.md": {
        "title": "Schema / Registry / Goal Pack 规则",
        "intro": (
            "本文件覆盖 Goal Runtime **机器可读层**规则：Schema 校验、Registry "
            "SSOT 与一致性、Goal Pack 结构、Golden/Violation Fixtures、规则与"
            "文档生命周期管理（archive/sunset/migration）。\n\n"
            "对应 P0 Gate：`schema-check`、`registry-check`、`goalpack-check`、"
            "`fixture-replay`（部分尚未 active，详见 "
            "[`registry.yaml`](./registry.yaml) 中 `status` 字段）。"
        ),
        "families": {
            "SCHEMA", "SCHEMA-MIN", "REGISTRY", "REGISTRY-CONSISTENCY",
            "REGISTRY-LOCK", "GOALPACK", "GOLDEN", "GOLDEN-PACK",
            "VIOLATION-FIXTURE", "COVERAGE", "GOAL-TEST", "ORPHAN", "MIGRATION",
            "ARCHIVE", "DOC", "DOC-DEBT", "DEBT", "RULE-BLOAT", "SUNSET",
            "DEPRECATION", "TEMPLATE", "CHANGE", "CHANGE-TYPE", "IMPACT",
            "COMPAT", "COMPAT-GUARD", "COMPAT-MATRIX", "FILE", "OWNERSHIP",
        },
    },
    "agent-runtime-rules.md": {
        "title": "Agent Runtime / Tooling / 度量规则",
        "intro": (
            "本文件覆盖 Goal Runtime **执行平面层**规则：Agent 协议与边界、"
            "并发与租约、命令事务与 dry-run、Bootstrap/Doctor/Repair、Dashboard "
            "与度量、`goalcli` 工具链架构、控制平面与报告规范。\n\n"
            "对应 Gate：`runtime-doctor`、`runtime-repair`、`dashboard`、"
            "`cmd-txn-check`、`cli-contract`、`gate-dag-check`。"
        ),
        "families": {
            "AGENT", "AGENT-AUTH", "AGENT-MEMORY", "AGENT-TEAM", "HANDOFF",
            "LEASE", "HEARTBEAT", "STOP", "REPAIR", "BOOTSTRAP", "DOCTOR",
            "RUNBOOK", "CONCURRENCY", "CONTEXT-COMPRESSION", "CONTEXT-WINDOW",
            "CMD-TXN", "DRYRUN", "BATCH", "AUTO-SAFETY", "HUMAN", "DASHBOARD",
            "DASHBOARD-HEALTH", "METRIC", "METRIC-GOV", "GOV-CADENCE", "REPORT",
            "CONTROL", "CHECKER", "GOALCLI", "GOALCLI-ARCH", "GOALCLI-CONFIG",
            "GOALCLI-EXIT", "COMPILER", "CODE", "GATE-DAG", "RECONCILE",
            "TRUST", "REPO-LAYOUT",
        },
    },
}


def family(rid: str) -> str:
    parts = rid.split("-")
    return "-".join(parts[1:-1])


def extract_body(lines: list[str], start_line: int) -> str:
    """从 goal-patch.md 中提取规则 body。

    start_line 是 ### RULE-... 或 ## RULE-... 标题所在行（1-based）。
    body 截止条件: 下一行匹配 `^#{1,4} ` 即停止。
    去除前后空行；保留正文中的 fenced code block / 列表 / 表格。
    """
    i = start_line  # lines[i] 是下一行（0-based 索引正好等于 start_line, 因为 enumerate 从 1 开始）
    body: list[str] = []
    while i < len(lines):
        ln = lines[i]
        if re.match(r"^#{1,4}\s+", ln):
            break
        body.append(ln)
        i += 1
    # 去末尾空行 / 单独的 `---` 分隔符
    while body and (body[-1].strip() == "" or body[-1].strip() == "---"):
        body.pop()
    while body and body[0].strip() == "":
        body.pop(0)
    return "\n".join(body)


def render_file(fname: str, meta: dict, rules: list[dict], src_lines: list[str], src_titles: dict[str, str]) -> str:
    out: list[str] = []
    out.append(f"# {meta['title']}")
    out.append("")
    out.append("> 本文件由 `scripts/render_domain_rules.py` 从 [`registry.yaml`](./registry.yaml)")
    out.append("> 与 `.worktree/goal-patch.md` 渲染生成；冲突时以 `iron-rules.md` >")
    out.append("> `registry.yaml` > 本文件 > `.worktree/goal-patch.md` 为序。")
    out.append("")
    out.append(meta["intro"])
    out.append("")
    out.append("---")
    out.append("")

    # 按 source_section 分组
    by_sec: dict[int, list[dict]] = {}
    for r in rules:
        by_sec.setdefault(r["source_section"], []).append(r)

    for sec in sorted(by_sec.keys()):
        rs = sorted(by_sec[sec], key=lambda x: (x["source_line"], x["id"]))
        sec_title = src_titles.get(sec, rs[0].get("source_section_title", ""))
        out.append(f"## §{sec} {sec_title}")
        out.append("")
        for r in rs:
            body = extract_body(src_lines, r["source_line"])
            level_tag = f"**[{r['level']}]**"
            enforced = r.get("enforced_by") or ""
            status = r.get("status", "indexed")
            anchor = (
                f"<sub>level: {r['level']} · status: {status} · "
                f"enforced_by: `{enforced or '（待机器化）'}`"
                + (f" · exit: {r['exit_code']}" if r.get("exit_code") else "")
                + f" · source: §{sec} L{r['source_line']}</sub>"
            )
            out.append(f"### {level_tag} `{r['id']}`：{r['title']}")
            out.append("")
            out.append(anchor)
            out.append("")
            if body:
                out.append(body)
                out.append("")
        out.append("---")
        out.append("")

    # 末尾去掉多余分隔符
    while out and out[-1] in ("", "---"):
        out.pop()
    out.append("")
    return "\n".join(out)


def main() -> int:
    data = yaml.safe_load(REGISTRY.read_text(encoding="utf-8"))
    src_text = SRC.read_text(encoding="utf-8")
    src_lines = src_text.split("\n")

    # 收集源章节标题表（同时支持 # N. 与 ## N.）；
    # 只取首次出现，避免后续 v2.1 §314 中的步骤编号 (# 1. 初始化 goalcli 等) 覆盖真实章节。
    src_titles: dict[int, str] = {}
    sec_re = re.compile(r"^#{1,2} (\d+)\.\s+(.+)$")
    for ln in src_lines:
        m = sec_re.match(ln)
        if m:
            num = int(m.group(1))
            if num not in src_titles:
                src_titles[num] = m.group(2).strip()

    # 分桶
    buckets: dict[str, list[dict]] = {k: [] for k in BUCKETS}
    for r in data["rules"]:
        fam = family(r["id"])
        for fname, meta in BUCKETS.items():
            if fam in meta["families"]:
                buckets[fname].append(r)
                break

    summary: list[str] = []
    for fname, rs in buckets.items():
        meta = BUCKETS[fname]
        content = render_file(fname, meta, rs, src_lines, src_titles)
        (OUT_DIR / fname).write_text(content, encoding="utf-8")
        p0 = sum(1 for x in rs if x["level"] == "P0")
        p1 = sum(1 for x in rs if x["level"] == "P1")
        summary.append(
            f"  {fname:38s}  {len(rs):3d} rules  ({p0} P0 + {p1} P1)  "
            f"{(OUT_DIR / fname).stat().st_size} bytes"
        )

    print("rendered:")
    for line in summary:
        print(line)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
