#!/usr/bin/env python3
"""验证 .agent/rules/registry.yaml 中每条 active 规则的 enforced_by 命令真实存在。

校验规则:
  - "goalcli <sub>"   → <sub> 必须在 cmd/goalcli/main.go 的 case 分支中
  - "make <target>"    → <target> 必须在 Makefile 中声明
  - ".githooks/<x>"    → 文件必须存在
  - "scripts/<x>"      → 文件必须存在
  - 空字符串 + status=indexed → 跳过 (合法待办)

退出码:
  0 - 全部 active 规则均可执行
  1 - 存在不可执行的 active 规则 (registry 漂移)
  2 - registry.yaml 解析失败

用法:
  python3 scripts/verify_rules.py
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
REGISTRY = ROOT / ".agent" / "rules" / "registry.yaml"
MAIN_GO = ROOT / "cmd" / "goalcli" / "main.go"
MAKEFILE = ROOT / "Makefile"


def load_goalcli_commands() -> set[str]:
    text = MAIN_GO.read_text(encoding="utf-8")
    cmds: set[str] = set()
    for line in text.splitlines():
        m = re.search(r"case\s+(.+?):\s*$", line)
        if not m:
            continue
        for tok in re.findall(r'"([^"]+)"', m.group(1)):
            cmds.add(tok)
    return cmds


def load_make_targets() -> set[str]:
    text = MAKEFILE.read_text(encoding="utf-8")
    targets: set[str] = set()
    for line in text.splitlines():
        m = re.match(r"^([a-zA-Z][a-zA-Z0-9_-]*)\s*:", line)
        if m:
            targets.add(m.group(1))
    return targets


def resolve(enforced_by: str, xcmds: set[str], mtargets: set[str]) -> str | None:
    """返回 None 表示合法; 返回字符串描述表示问题"""
    if not enforced_by:
        return None  # 应配合 status=indexed 检查
    parts = enforced_by.split()
    head = parts[0]
    if head == "goalcli":
        if len(parts) == 1:
            return None  # 裸 "goalcli" 视为命令本身存在
        sub = parts[1]
        if sub in xcmds:
            return None
        return f"unknown goalcli subcommand: {sub}"
    if head == "make":
        if len(parts) < 2:
            return "make without target"
        if parts[1] in mtargets:
            return None
        return f"unknown make target: {parts[1]}"
    if head.startswith(".githooks/") or head.startswith("scripts/"):
        if (ROOT / head).exists():
            return None
        return f"file not found: {head}"
    return f"unrecognized enforced_by format: {enforced_by}"


def main() -> int:
    try:
        data = yaml.safe_load(REGISTRY.read_text(encoding="utf-8"))
    except Exception as exc:
        print(f"ERROR: failed to load {REGISTRY}: {exc}", file=sys.stderr)
        return 2

    xcmds = load_goalcli_commands()
    mtargets = load_make_targets()

    problems: list[str] = []
    inconsistent_status: list[str] = []
    for rule in data["rules"]:
        rid = rule["id"]
        enforced = rule.get("enforced_by", "")
        status = rule.get("status", "")
        if status == "active":
            err = resolve(enforced, xcmds, mtargets)
            if err is not None:
                problems.append(f"{rid} (active, enforced_by={enforced!r}): {err}")
            elif not enforced:
                inconsistent_status.append(f"{rid}: status=active but enforced_by is empty")
        elif status == "indexed":
            if enforced:
                # 有 enforced_by 但 status 仍标 indexed → 应该是 active
                inconsistent_status.append(
                    f"{rid}: status=indexed but enforced_by={enforced!r} (should be active)"
                )

    total = len(data["rules"])
    active = sum(1 for r in data["rules"] if r.get("status") == "active")
    print(f"rules total: {total}")
    print(f"rules active: {active}")
    print(f"goalcli subcommands available: {len(xcmds)}")
    print(f"makefile targets available: {len(mtargets)}")

    if problems:
        print(f"\n=== {len(problems)} active rules reference unknown commands ===", file=sys.stderr)
        for p in problems:
            print(f"  {p}", file=sys.stderr)
    if inconsistent_status:
        print(f"\n=== {len(inconsistent_status)} status/enforced_by inconsistencies ===", file=sys.stderr)
        for p in inconsistent_status:
            print(f"  {p}", file=sys.stderr)

    if problems or inconsistent_status:
        return 1
    print("\nall active rules have valid enforced_by commands")
    return 0


if __name__ == "__main__":
    sys.exit(main())
