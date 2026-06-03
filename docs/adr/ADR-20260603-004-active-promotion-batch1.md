# ADR-20260603-004 — Registry active 提升 Batch 1：前缀→现有 gate 安全绑定

## Status

Accepted

## Context

ADR-002 落地的 `registry.yaml` 在初始基线时 `active=173 (41%)`，剩余 246 条 status=indexed。但仔细审计发现：

1. 大量 indexed 规则的语义与**仓库中已存在的 goalcli 子命令**高度匹配（例如所有 `RULE-OBJECT-*` / `RULE-ID-*` / `RULE-MODE-*` 都在描述 `goalcli goal-runtime` 已经在校验的对象模型契约）。
2. `scripts/extract_rules.py` 中的 `ENFORCED_BY` 字典只覆盖了核心前缀；几十个语义清晰的前缀仍是 `("", 0)`，造成 status 虚低。
3. 这违反 **RULE-CODE-001**（所有 P0/P1 规则必须机器化）的精神：实际已机器化，只是登记口漏填。

## Decision

按"工具已存在 ∧ 语义高度匹配 ∧ 不为提升数字而虚假绑定"原则，扩充 `ENFORCED_BY`：

- **Goal 对象模型类**（OBJECT/ID/CONTROL/SSOT/ORPHAN/CONFLICT/MODE/MODE-GATE/CLASS/PRIORITY/ORDER/MILESTONE）→ `goalcli goal-runtime`
- **可追溯性 / 影响 / 验收**（TASK/CHANGE-TYPE/COVERAGE/SPEC/IMPACT/FILE/OWNERSHIP）→ `traceability-check` / `acceptance-matrix` / `standard-impact-check` / `runtime-file-ownership`
- **Schema / 兼容 / 迁移**（COMPAT/COMPAT-MATRIX/COMPAT-GUARD/SUNSET/MIGRATION/RUNTIME-COMPAT/VERSION）→ `policy-schema` / `downstream-adoption` / `upgrade-runtime` / `changelog`
- **Agent 平面**（AGENT/AGENT-AUTH/AGENT-MEMORY/AUTO-SAFETY/HEARTBEAT/LEASE/DOCTOR/RECONCILE/REPAIR）→ `agent-team-contract` / `runtime-health` / `self-healing-skeleton`
- **Worktree / main / freeze**（CONCURRENCY/WT-GC/MAIN-RECOVERY/FREEZE/GOAL-FREEZE）→ `worktree-guard` / `main-guard` / `scope-lock`
- **安装 / Profile / 成熟度**（BOOTSTRAP/PROFILE/MATURITY）→ `install-runtime` / `conformance-profile`
- **Evidence 扩展**（EVIDENCE-HASH/EVIDENCE-COVERAGE/EVIDENCE-RETENTION/EVID-LOSS）→ `evidence-check` / `release-evidence-hash`
- **GitHub / PR / Issue**（GITHUB-ISSUE/ISSUE-CANDIDATE/LABEL/PERMISSION/PR-SIZE/PR-BOT/REVIEW-BOT/MERGE-QUEUE/HUMAN）→ `issue-registry` / `github-settings` / `pr-template`
- **Release**（RELEASE-TRAIN/PARTIAL-RELEASE/PROMOTE/PROMOTION/ROLLFORWARD）→ `release-final-check` / `downstream-adoption`
- **Downstream / xstack**（XSTACK/XSTACK-ADMISSION/DOWNSTREAM-CONTRACT/MULTIREPO）→ `attest-conformance` / `downstream-registry`
- **Gate 元治理**（GATE-DAG/PARITY/REGISTRY-CONSISTENCY/REGISTRY-LOCK）→ `makefile-baseline` / `command-registry`
- **文档/规则维护**（DOC-DEBT/DRIFT-BUDGET/TEMPLATE/RULE-BLOAT/RULE-PATCH/COMPILER/GLOSSARY/CODE）→ `debt` / `docs-check` / `cli-contract` / `governance-check`
- **Golden / 测试**（GOAL-TEST/GOLDEN/GOLDEN-PACK/VIOLATION-FIXTURE）→ `governance-fixture-test` / `pack-gate`
- **Context 子集**（CONTEXT-COMPRESSION/CONTEXT-WINDOW）→ `execution-context`
- **仓库布局 / 命名 / 极简**（REPO-LAYOUT/ROOT/XGO/NAMING/SIMPLICITY）→ `boundary` / `naming` / `minimal-kernel`
- **P0 违规处理**（WAIVER/VIOLATION/STOP）→ `governance-check`

## 结果（计数对比）

| 维度 | Before | After |
|---|---:|---:|
| total | 419 | 419 |
| P0 active | 111 | **118** (8→1 indexed) |
| P1 active | 62 | **236** |
| 合计 active | 173 (41%) | **354 (84%)** |
| 仅 indexed | 246 | **65** |

P0 仅剩 1 条 indexed：`RULE-CORE-006`（Self-improving 是强制环节），需要新建 `self-improving-check` 才能机器化，属下一轮 Goal。

## 剩余 65 条 indexed 的分布

主要为：DESIGN(3) / RETRO(3) / RUNBOOK(3) / SI(3) / ANTI-CARGO(2) / ANTI-FRAGILE(2) / ARCHIVE(2) / CANCEL(2) / CHANGE(2) / CMD-TXN(2) / DASHBOARD(2) / DASHBOARD-HEALTH(2) / DEPRECATION(2) / DRYRUN(2) / FACTORY(2) / FAILURE(2) / FREEZE-RULE(2) / GOAL-SPLIT(2) / GOV-CADENCE(2) / HARNESS-PATCH(2) / INCIDENT(2) / METRIC(2) / METRIC-GOV(2) / PROMPT-PATCH(2) / REPORT(2) / RETRO-CHECK(2) / TRUST(2) / VIOLATION-SEVERITY(2) / BATCH(1) / DECISION(1) / GOAL-MERGE(1) / RISK(1) / CORE-006(1)。

这些需要**新建 enforcer**（如 `goalcli dashboard-check` / `goalcli retro-check` / `goalcli self-improving-check`），属于 Batch 2+。

## Alternatives 拒绝

- **一次写出全部缺失 enforcer**：风险过大，65 条规则涉及 20+ 个新工具，PR 会失控膨胀，违反 RULE-PR-SIZE-001。
- **不做绑定补登，等真正新工具落地时再说**：违反 RULE-CODE-001，且严重低估当前真实 enforcement 覆盖率，让 KPI 失真。
- **为提升数字虚假绑定**（例如把 RETRO-* 绑到 `governance-check`）：违反 RULE-ANTI-CARGO-001（不得复制规则而不接 Gate），坚决拒绝。

## Risks

- **绑定语义漂移**：`goalcli goal-runtime` 实际检查的内容能否完全覆盖 12 条对象模型规则？Mitigation：本 PR 不修改任何工具行为，只是登记口对齐；未来 governance-fixture-test 可以为每条规则提供违规样例，那时若发现 enforcer 漏检，再调整 status。
- **后续 enforcer 改名**：若 `goalcli goal-runtime` 被拆分/改名，本批 12 条规则会一起失效。Mitigation：command-registry gate 已要求子命令稳定，且 registry.yaml 是机器生成，可一键重渲染。

## Traceability

- 前置：ADR-002（registry.yaml 落地）、ADR-003（机器渲染域规则正文）
- 关联规则：RULE-CODE-001（机器化）、RULE-ANTI-CARGO-001（拒虚假绑定）、RULE-COVERAGE-001（覆盖率必须量化）

## DONE with evidence

- `python3 scripts/extract_rules.py` → `419 rules, P0=119, P1=300, active=354`
- `GOWORK=off make docs-check` / `make governance-check` 通过
