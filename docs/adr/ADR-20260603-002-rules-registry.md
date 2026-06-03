# ADR-20260603-002: Rule Registry 作为规则机器化 SSOT

Status: Accepted
Date: 2026-06-03
Supersedes: 无（与 ADR-20260603-001 互补）

## Context

`.worktree/goal-patch.md`（13856 行, 336 节, 419 个 RULE-* ID）是 Goal Runtime v1.0–v2.2 的完整推导记录。它有三个问题：

1. **不可机器读** —— 规则散落在 Markdown 散文中，无统一字段。
2. **缺 enforced_by 绑定** —— 规则与 `goalcli` 子命令的对应关系靠人工记忆。
3. **严重重复** —— v1.0 / v1.2 / v1.3 / v1.4 / v1.9 各有一份"最终铁律"，违反 RULE-DOC-001。

同时 `.agent/rules/*.md` 11 个文件只覆盖 v1.0 的约 35 条规则（8%），v1.1–v2.2 增量的 380+ 条规则从未进入仓库。

## Decision

建立两个新文件作为规则的 SSOT：

1. **`.agent/rules/iron-rules.md`** — 30 行内压缩 119 条 P0 规则为 7 条第一性铁律 + 标准退出码表。
2. **`.agent/rules/registry.yaml`** — 全部 419 条规则的机器化索引，每条带 `id / level / title / source_section / enforced_by / exit_code / status`。

生成器：`scripts/extract_rules.py`（确定性、可复跑、由 `P0_PREFIXES` 和 `ENFORCED_BY` 字典做人工分级映射）。

权威顺序（与 CONSTITUTION.md 一致）：

```
iron-rules.md  >  registry.yaml  >  *-rules.md  >  ADR-*  >  .worktree/goal-patch.md (考古)
```

## Rationale

- **机器化优先 (RULE-CODE-001)**：YAML 比散文更容易被 `goalcli rules verify`（未来）消费，也更容易被 CI 当成 fixture 比对。
- **不重写 11 个 `*-rules.md`**：它们已经被 governance test 引用，破坏性变更需要单独 ADR。本 ADR 仅"叠加"两个新文件。
- **不新建 `tools/goalcli/`**：遵循 ADR-20260603-001，所有 enforced_by 指向已存在的 `goalcli` 子命令或 `scripts/` 脚本。
- **保留 `.worktree/goal-patch.md`**：作为历史推导记录，不删除，但明确"考古"地位。

## Consequences

- 后续任何规则变更必须：
  1. 修改 `scripts/extract_rules.py` 或源文档
  2. 重跑生成器
  3. 在 commit 中说明 P0/active 计数变化
- P0=119, active=354 (84%) 成为可度量基线。后续 Goal 的 KPI 是把 active 比例提升到 ≥95%。
- 65 条 indexed-only 规则等待机器化，每条都是一个潜在的 Goal 候选。

## Rejected

- 拒绝直接删除 `.worktree/goal-patch.md` —— 它是设计推导的真实历史，删除会失去可审计性。
- 拒绝把 419 条规则全部拆成单独 `.md` 文件 —— 会触发它自己定义的 RULE-DOC-001（文档膨胀）。
- 拒绝在本 ADR 内提升 active 比例 —— 那是阶段 2 的工作（fixtures）和阶段 3 的工作（self-audit Goal），不应在阶段 1 偷跑。

## Evidence

- 生成器: `scripts/extract_rules.py`
- 产物: `.agent/rules/registry.yaml`, `.agent/rules/iron-rules.md`, `.agent/rules/README.md`
- 验证: `python3 -c "import yaml; d=yaml.safe_load(open('.agent/rules/registry.yaml')); assert d['total_rules']==419 and d['p0_count']==119"`
