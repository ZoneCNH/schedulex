# ADR-20260603-003: 三个域规则文件 (core / schema-registry / agent-runtime)

Status: Accepted
Date: 2026-06-03
Supersedes: 无（补充 ADR-20260603-002）

## Context

ADR-20260603-002 把 `.worktree/goal-patch.md` 中 419 条规则机器化为 `registry.yaml`，并把 119 条 P0 压缩为 `iron-rules.md` 7 条铁律。它**没有**把 v1.1–v2.2 增量的约 200 条 P1 规则正文化到人类可读文件——这些规则只在 `registry.yaml` 中以 `id + title` 形式存在。

这造成两个问题：

1. **人类无法快速理解规则的"为什么"**：reviewer 要查 enforced_by 时只能拿 ID 去 `.worktree/goal-patch.md` 翻 13856 行散文，违反"考古地位"的设定。
2. **downstream 采用难度高**：被 kernel / configx 等下游库引用时，没有可直接拷贝的规则段落。

既有 11 个域文件 `goal-rules.md` / `worktree-rules.md` / … 只覆盖 v1.0 的约 35 条 P0，覆盖率 8%。

## Decision

新增 3 个**机器渲染的**域规则文件，按"控制层 / 机器可读层 / 执行平面层"切分：

| 文件 | 覆盖规则族 | 规则数 |
|---|---|---|
| `.agent/rules/core-rules.md` | CORE / CONTEXT / STATE / SSOT / ID / MODE / FAILURE / SCORE / FREEZE / NAMING / GLOSSARY 等 | 49 (P0=11, P1=38) |
| `.agent/rules/schema-registry-rules.md` | SCHEMA / REGISTRY / GOALPACK / GOLDEN / MIGRATION / DOC / DEBT / COMPAT 等 | 60 (P0=0, P1=60) |
| `.agent/rules/agent-runtime-rules.md` | AGENT / LEASE / HEARTBEAT / CMD-TXN / DRYRUN / GOALCLI / DASHBOARD / METRIC 等 | 75 (P0=3, P1=72) |

合计 **184 条**规则被正文化（占 registry.yaml 总量的 44%）。

新增生成器：`scripts/render_domain_rules.py`，从 `registry.yaml` 取分级 + enforced_by，从 `.worktree/goal-patch.md` 取正文，按 `source_section` 分组渲染。

## Rationale

- **不动既有 11 个域文件**：它们是手写源，被 governance test 引用；新增 3 个文件以 **机器渲染** 方式叠加，互不冲突。
- **按"层"而非"对象"切分**：core / machine-readable / runtime-plane 三层映射 Goal Runtime 的概念分层，下游库可按需引用某一层。
- **保留 `### **[P0]** RULE-ID：title` 格式**：与 `goal-rules.md` 现有写法保持一致，便于复制粘贴。
- **每条规则带元数据 `<sub>` 行**：`level / status / enforced_by / exit_code / source` 一行可见，免去往返 `registry.yaml`。
- **不在本 ADR 内提升 active 比例**：active 仍是 173/419 = 41%，与 ADR-002 一致；提升到 ≥80% 是阶段 3 的工作。

## Consequences

- 三个文件总计 ~3700 行，文件级别"读得动"。
- `scripts/render_domain_rules.py` 必须与 `extract_rules.py` 同步维护：任何源章节降级都可能引起标题归属漂移（已通过"首次出现保留"策略防御 §314 步骤编号污染）。
- 未来若把某条 indexed 规则机器化（填入 `enforced_by`），重跑 `render_domain_rules.py` 会自动刷新对应 `<sub>` 行的 `status: active` 标记。
- 既有 11 个 `*-rules.md` 仍是 P0 铁律的权威叙述；新增 3 个文件主要面向 P1 / indexed 规则。

## Rejected

- 拒绝把 200 条 indexed 规则全部塞进 `goal-rules.md`：会让该文件膨胀到 4000+ 行，违反 RULE-DOC-001（文档膨胀控制）。
- 拒绝把渲染脚本与 `extract_rules.py` 合并：两者职责不同，合并会破坏 SRP。
- 拒绝在域文件中再次定义新的 enforced_by 映射：那是 `registry.yaml` 的唯一职责。

## Evidence

- 生成器：`scripts/render_domain_rules.py`
- 产物：
  - `.agent/rules/core-rules.md`（49 条，21857 bytes）
  - `.agent/rules/schema-registry-rules.md`（60 条，23223 bytes）
  - `.agent/rules/agent-runtime-rules.md`（75 条，33636 bytes）
- 验证：
  - `python3 scripts/render_domain_rules.py` 退出 0
  - `grep -c "^### .* RULE-" .agent/rules/core-rules.md` = 49
  - `grep -c "^### .* RULE-" .agent/rules/schema-registry-rules.md` = 60
  - `grep -c "^### .* RULE-" .agent/rules/agent-runtime-rules.md` = 75
  - 三个文件每条 RULE-ID 都能在 `registry.yaml` 中找到对应 entry（机器可验证：`grep -E "id: RULE-..." registry.yaml`）
