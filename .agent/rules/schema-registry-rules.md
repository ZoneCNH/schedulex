# Schema / Registry / Goal Pack 规则

> 本文件由 `scripts/render_domain_rules.py` 从 [`registry.yaml`](./registry.yaml)
> 与 `.worktree/goal-patch.md` 渲染生成；冲突时以 `iron-rules.md` >
> `registry.yaml` > 本文件 > `.worktree/goal-patch.md` 为序。

本文件覆盖 Goal Runtime **机器可读层**规则：Schema 校验、Registry SSOT 与一致性、Goal Pack 结构、Golden/Violation Fixtures、规则与文档生命周期管理（archive/sunset/migration）。

对应 P0 Gate：`schema-check`、`registry-check`、`goalpack-check`、`fixture-replay`（部分尚未 active，详见 [`registry.yaml`](./registry.yaml) 中 `status` 字段）。

---

## §33 Goal Registry 规则

### **[P1]** `RULE-REGISTRY-001`：所有 Goal 必须登记

<sub>level: P1 · status: active · enforced_by: `goalcli command-registry` · exit: 1 · source: §33 L1723</sub>

必须存在：

```text
.agent/registries/goals.yaml
```

示例：

```yaml
goals:
  - goal_id: GOAL-20260603-001
    title: Build Goal Runtime Harness System
    mode: Full
    state: EXECUTING
    repo: xlib-standard
    owner: zonecnh
    branch_policy: worktree-only
    evidence_required: true
    release_required: true
    created_at: 2026-06-03
```

### **[P1]** `RULE-REGISTRY-002`：Goal 状态必须同步

<sub>level: P1 · status: active · enforced_by: `goalcli command-registry` · exit: 1 · source: §33 L1749</sub>

当 Goal 进入新阶段，必须同步：

```text
goal.md
goals.yaml
traceability.md
release-manifest.md
```

禁止出现：

```text
goal.md 显示 DONE
goals.yaml 仍是 EXECUTING
release-manifest 缺 Evidence
```

---

## §47 Deprecation 规则

### **[P1]** `RULE-DEPRECATION-001`：旧规则不能直接删除

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §47 L2302</sub>

必须经过：

```text
PROPOSED_DEPRECATION
MIGRATION_AVAILABLE
DOWNSTREAM_NOTIFIED
DEPRECATED
REMOVED
```

### **[P1]** `RULE-DEPRECATION-002`：规则废弃必须说明替代方案

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §47 L2316</sub>

必须包含：

```text
deprecated rule
replacement rule
migration steps
deadline
affected repos
rollback option
```

---

## §55 Schema-first 规则

### **[P1]** `RULE-SCHEMA-001`：Goal 核心对象必须有 Schema

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §55 L2592</sub>

必须为以下对象定义 JSON Schema / YAML Schema：

```text
Goal
Spec
Requirement
Acceptance Criteria
Design
ADR
Plan
Task
Issue
Evidence
Risk
Decision
Review
Release
Retrospective
Patch
```

推荐目录：

```text
.agent/schemas/
├── goal.schema.json
├── spec.schema.json
├── task.schema.json
├── evidence.schema.json
├── release.schema.json
├── retrospective.schema.json
└── patch.schema.json
```

### **[P1]** `RULE-SCHEMA-002`：没有通过 Schema 校验的对象不得进入下一阶段

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §55 L2630</sub>

例如：

```bash
goalcli schema validate .agent/goals/GOAL-20260603-001/tasks.yaml
```

失败则：

```text
Goal state = INCONSISTENT_STATE
禁止生成 Issue
禁止创建 PR
禁止 Release
```

---

## §56 Goal Pack 规则

### **[P1]** `RULE-GOALPACK-001`：每个 Goal 必须形成 Goal Pack

<sub>level: P1 · status: active · enforced_by: `goalcli pack-gate` · exit: 1 · source: §56 L2651</sub>

Goal Pack 是一个 Goal 的完整事实包。

目录：

```text
.agent/goals/<GOAL-ID>/
├── goal.yaml
├── context.md
├── spec.yaml
├── design.md
├── adr/
├── plan.yaml
├── tasks.yaml
├── issues.yaml
├── traceability.md
├── risk-register.yaml
├── decision-log.md
├── evidence/
├── review.md
├── release-manifest.md
├── retrospective.md
└── patches/
```

### **[P1]** `RULE-GOALPACK-002`：Goal Pack 必须可离线审计

<sub>level: P1 · status: active · enforced_by: `goalcli pack-gate` · exit: 1 · source: §56 L2679</sub>

即使 GitHub、CI、外部系统不可用，仍然能从 Goal Pack 看出：

```text
目标是什么
需求是什么
任务是什么
改了什么
怎么验证
证据在哪里
是否发布
有什么风险
下一轮如何改进
```

---

## §61 文件变更规则

### **[P1]** `RULE-FILE-001`：Task 必须声明预计变更文件

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-file-ownership` · exit: 1 · source: §61 L2914</sub>

Task 中必须包含：

```yaml
files_to_change:
  - .agent/rules/07-worktree-rules.md
  - scripts/harness/no-main-dev.sh
  - Makefile
```

### **[P1]** `RULE-FILE-002`：实际变更超出范围必须解释

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-file-ownership` · exit: 1 · source: §61 L2927</sub>

如果实际变更文件不在 `files_to_change` 中，必须：

```text
更新 Task
更新 Traceability
更新 Risk
必要时生成 Decision Log
```

否则 PR Gate 失败。

---

## §63 文档膨胀控制规则

### **[P1]** `RULE-DOC-001`：禁止无边界文档膨胀

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §63 L2989</sub>

任何新增文档必须回答：

```text
它服务哪个 Goal？
它验证哪个 Requirement？
它替代还是补充已有文档？
它是否需要进入导航？
它是否需要 Harness 检查？
```

### **[P1]** `RULE-DOC-002`：文档必须分类

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §63 L3003</sub>

推荐分类：

```text
00-index / 导航
01-rules / 规则
02-specs / 需求
03-design / 设计
04-harness / 门禁
05-templates / 模板
06-evidence / 证据
07-release / 发布
08-retro / 复盘
09-adoption / 下游采用
```

### **[P1]** `RULE-DOC-003`：长期规则文档必须有 SSOT

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §63 L3022</sub>

如果多个文档描述同一规则，必须指定唯一事实源：

```text
SSOT: .agent/rules/07-worktree-rules.md
```

其他文档只能引用，不允许复制漂移。

---

## §73 规则变更协议

### **[P1]** `RULE-CHANGE-001`：修改规则必须走 Rule Change Protocol

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §73 L3309</sub>

流程：

```text
Propose Rule Change
→ Impact Analysis
→ Decision Log
→ Update Rule
→ Update Harness
→ Update Templates
→ Update CI
→ Update Downstream Adoption Notes
→ Evidence
```

### **[P1]** `RULE-CHANGE-002`：P0 规则变更必须人工批准

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §73 L3327</sub>

任何弱化 P0 的行为都必须进入：

```text
NEEDS_HUMAN_APPROVAL
```

---

## §74 模板同步规则

### **[P1]** `RULE-TEMPLATE-001`：规则变更必须同步模板

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §74 L3339</sub>

例如新增 Evidence 字段后必须同步：

```text
issue-template.md
pr-template.md
evidence-template.md
release-manifest-template.md
retrospective-template.md
```

### **[P1]** `RULE-TEMPLATE-002`：模板必须有版本号

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §74 L3353</sub>

```yaml
template_id: TEMPLATE-PR-v1.2
version: v1.2
compatible_rules:
  - RULE-EVIDENCE-001
  - RULE-TRACE-001
```

---

## §76 Golden Path 规则

### **[P1]** `RULE-GOLDEN-001`：必须维护一个最小成功样例

<sub>level: P1 · status: active · enforced_by: `goalcli governance-fixture-test` · exit: 1 · source: §76 L3400</sub>

Golden Path 是标准成功路径。

```text
tests/fixtures/golden/goal-runtime-minimal/
```

必须包含：

```text
Goal
Spec
Task
Evidence
Traceability
Release Manifest
Retrospective
```

### **[P1]** `RULE-GOLDEN-002`：每次规则升级都必须跑 Golden Test

<sub>level: P1 · status: active · enforced_by: `goalcli governance-fixture-test` · exit: 1 · source: §76 L3422</sub>

防止规则升级把正常流程误杀。

---

## §105 结构债规则

### **[P1]** `RULE-DEBT-001`：结构债必须可检测

<sub>level: P1 · status: active · enforced_by: `goalcli debt` · exit: 1 · source: §105 L4468</sub>

结构债包括：

```text
分层违规 import
L2 横向耦合
公共 API 泄漏实现细节
配置隐式全局状态
文档重复事实源
CI gate 漂移
Evidence 缺失
```

### **[P1]** `RULE-DEBT-002`：结构债必须进入 Debt Registry

<sub>level: P1 · status: active · enforced_by: `goalcli debt` · exit: 1 · source: §105 L4484</sub>

```yaml
debt_id: DEBT-20260603-001
type: layering_violation
severity: P1
location: internal/foo/bar.go
detected_by: make structure-check
fix_issue: "#166"
```

---

## §122 Schema 最小字段规则

### **[P1]** `RULE-SCHEMA-MIN-001`：Goal Schema 最小字段

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §122 L5243</sub>

```yaml
goal_id:
title:
mode:
state:
owner:
repositories:
scope:
non_goals:
constraints:
success_criteria:
created_at:
updated_at:
```

### **[P1]** `RULE-SCHEMA-MIN-002`：Task Schema 最小字段

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §122 L5262</sub>

```yaml
task_id:
goal_id:
title:
priority:
related_requirements:
related_acceptance_criteria:
files_to_change:
commands_to_run:
evidence_required:
rollback_plan:
status:
```

### **[P1]** `RULE-SCHEMA-MIN-003`：Evidence Schema 最小字段

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §122 L5280</sub>

```yaml
evidence_id:
goal_id:
task_id:
command:
cwd:
branch:
commit:
exit_code:
status:
artifacts:
timestamp:
```

---

## §147 文件所有权规则

### **[P1]** `RULE-OWNERSHIP-001`：关键文件必须有 Owner

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-file-ownership` · exit: 1 · source: §147 L6084</sub>

建议新增：

```text
.agent/ownership.yaml
```

示例：

```yaml
owners:
  - path: ".agent/rules/"
    owner: "architecture"
    review_required: true

  - path: ".github/workflows/"
    owner: "ci"
    review_required: true

  - path: "scripts/harness/"
    owner: "harness"
    review_required: true

  - path: "goalcli.yaml"
    owner: "runtime"
    review_required: true
```

### **[P1]** `RULE-OWNERSHIP-002`：修改关键文件必须触发 Review

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-file-ownership` · exit: 1 · source: §147 L6115</sub>

关键路径：

```text
.agent/rules/
.agent/schemas/
.agent/harness/
.github/workflows/
Makefile
goalcli.yaml
scripts/harness/
CONSTITUTION.md
```

修改这些路径，PR 必须进入：

```text
requires: architecture-review
requires: harness-review
requires: ci-review
```

---

## §148 变更类型规则

### **[P1]** `RULE-CHANGE-TYPE-001`：每个 Task 必须声明变更类型

<sub>level: P1 · status: active · enforced_by: `goalcli traceability-check` · exit: 9 · source: §148 L6142</sub>

允许类型：

```text
docs
rule
schema
harness
ci
template
automation
code
test
release
migration
security
```

### **[P1]** `RULE-CHANGE-TYPE-002`：不同类型触发不同 Gate

<sub>level: P1 · status: active · enforced_by: `goalcli traceability-check` · exit: 9 · source: §148 L6163</sub>

| Change Type | Required Gate                         |
| ----------- | ------------------------------------- |
| docs        | docs-check / link-check               |
| rule        | rule-check / harness-impact-check     |
| schema      | schema-check / compatibility-check    |
| harness     | harness-test / violation-fixture-test |
| ci          | ci-dry-run / workflow-check           |
| template    | template-check                        |
| automation  | goalcli-test                          |
| code        | lint / test                           |
| release     | release-check                         |
| security    | secret-check / security-review        |
| migration   | migration-check                       |

---

## §149 Impact Analysis 规则

### **[P1]** `RULE-IMPACT-001`：P0/P1 变更必须有影响分析

<sub>level: P1 · status: active · enforced_by: `goalcli standard-impact-check` · exit: 1 · source: §149 L6183</sub>

必须回答：

```text
影响哪些规则？
影响哪些 Gate？
影响哪些模板？
影响哪些下游仓库？
是否 breaking？
是否需要迁移？
是否需要文档更新？
是否需要新增测试 fixture？
```

### **[P1]** `RULE-IMPACT-002`：影响分析必须进入 PR

<sub>level: P1 · status: active · enforced_by: `goalcli standard-impact-check` · exit: 1 · source: §149 L6200</sub>

PR 中必须有：

```md

---

## §150 Backward Compatibility 规则

### **[P1]** `RULE-COMPAT-001`：规则升级必须兼容旧 Goal Pack

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §150 L6221</sub>

除非明确标记 breaking，否则新版本 goalcli 必须能读取旧版本 Goal Pack。

### **[P1]** `RULE-COMPAT-002`：Breaking 规则必须有迁移计划

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §150 L6227</sub>

Breaking 规则包括：

```text
Schema 字段必填变化
Evidence 格式变化
Traceability 格式变化
Release Manifest 格式变化
Gate 从 warn 升级为 block
```

必须提供：

```text
migration notes
migration command
rollback plan
affected repos
```

---

## §151 Migration Script 规则

### **[P1]** `RULE-MIGRATION-001`：结构性变更必须有迁移脚本

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §151 L6252</sub>

目录：

```text
.agent/migrations/
├── MIGRATION-20260603-rules-v1.4-to-v1.5.md
└── scripts/
    └── migrate-rules-v1.4-to-v1.5.sh
```

### **[P1]** `RULE-MIGRATION-002`：迁移必须可 dry-run

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §151 L6265</sub>

```bash
goalcli migrate --from rules-v1.4 --to rules-v1.5 --dry-run
goalcli migrate --from rules-v1.4 --to rules-v1.5 --apply
```

---

## §152 Registry Consistency 规则

### **[P1]** `RULE-REGISTRY-CONSISTENCY-001`：Registry 之间必须一致

<sub>level: P1 · status: active · enforced_by: `goalcli command-registry` · exit: 1 · source: §152 L6276</sub>

必须检查：

```text
goals.yaml 中的 Goal 存在对应 .agent/goals/<GOAL-ID>/
tasks.yaml 中的 Task 存在于对应 Goal Pack
evidence.yaml 中的 Evidence 文件真实存在
patches.yaml 中的 Patch 文件真实存在
adoption.yaml 中的下游状态有证据
```

### **[P1]** `RULE-REGISTRY-CONSISTENCY-002`：Registry 漂移必须阻断 Release

<sub>level: P1 · status: active · enforced_by: `goalcli command-registry` · exit: 1 · source: §152 L6290</sub>

```bash
make registry-check
```

如果失败：

```text
Goal state = INCONSISTENT_STATE
Release blocked
```

---

## §154 Test Pyramid for Goal Runtime 规则

### **[P1]** `RULE-GOAL-TEST-001`：Goal Runtime 自身也需要测试金字塔

<sub>level: P1 · status: active · enforced_by: `goalcli governance-fixture-test` · exit: 1 · source: §154 L6347</sub>

```text
Unit Tests:
- schema validator
- evidence parser
- traceability parser
- branch checker

Fixture Tests:
- golden goal pack
- missing evidence
- missing traceability
- main branch violation
- secret leak

Integration Tests:
- make ci
- PR check
- release check
- audit goal

End-to-End Tests:
- create goal pack
- create task
- create evidence
- pass release gate
```

### **[P1]** `RULE-GOAL-TEST-002`：每个新增 Gate 必须有正反样例

<sub>level: P1 · status: active · enforced_by: `goalcli governance-fixture-test` · exit: 1 · source: §154 L6378</sub>

新增 Gate 时必须同时增加：

```text
golden fixture
violation fixture
expected report
```

---

## §155 Golden Goal Pack 标准

### **[P1]** `RULE-GOLDEN-PACK-001`：Golden Pack 是系统最小正确样例

<sub>level: P1 · status: active · enforced_by: `goalcli pack-gate` · exit: 1 · source: §155 L6392</sub>

目录：

```text
.agent/harness/fixtures/golden/minimal-goal-pack/
```

必须包含：

```text
goal.yaml
spec.yaml
tasks.yaml
evidence.yaml
traceability.md
release-manifest.md
retrospective.md
```

### **[P1]** `RULE-GOLDEN-PACK-002`：Golden Pack 必须进入 CI

<sub>level: P1 · status: active · enforced_by: `goalcli pack-gate` · exit: 1 · source: §155 L6414</sub>

CI 必须跑：

```bash
goalcli audit goal --fixture .agent/harness/fixtures/golden/minimal-goal-pack
```

---

## §156 Violation Fixture 标准

### **[P1]** `RULE-VIOLATION-FIXTURE-001`：每个 P0 规则必须有违规样例

<sub>level: P1 · status: active · enforced_by: `goalcli governance-fixture-test` · exit: 1 · source: §156 L6426</sub>

至少：

```text
missing-evidence/
missing-traceability/
main-branch-dev/
secret-leak/
missing-release-manifest/
invalid-schema/
```

### **[P1]** `RULE-VIOLATION-FIXTURE-002`：违规样例必须断言失败原因

<sub>level: P1 · status: active · enforced_by: `goalcli governance-fixture-test` · exit: 1 · source: §156 L6441</sub>

每个 violation fixture 必须定义：

```yaml
expected:
  exit_code: 4
  rule_id: RULE-EVIDENCE-001
  status: failed
```

---

## §162 Archival 规则

### **[P1]** `RULE-ARCHIVE-001`：完成 Goal 必须归档

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §162 L6628</sub>

完成后归档：

```text
.agent/goals/<GOAL-ID>/
release/<REL-ID>/
reports/audit/<GOAL-ID>.md
```

### **[P1]** `RULE-ARCHIVE-002`：归档后不能改历史证据

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §162 L6640</sub>

如果需要修正，创建：

```text
correction note
new evidence
new patch release
```

---

## §183 规则膨胀控制规则

### **[P1]** `RULE-RULE-BLOAT-001`：新增规则必须给出机器化路径

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §183 L7473</sub>

新增 P0 / P1 规则必须回答：

```text
如何检查？
用哪个命令检查？
失败报告在哪里？
是否有 golden fixture？
是否有 violation fixture？
是否会影响下游？
```

无法回答的，不允许升级为 P0/P1。

### **[P1]** `RULE-RULE-BLOAT-002`：规则必须定期清理

<sub>level: P1 · status: active · enforced_by: `goalcli docs-check` · exit: 1 · source: §183 L7490</sub>

每个版本周期必须检查：

```text
重复规则
过时规则
无法机器化规则
无人使用规则
下游无法采用规则
误杀率过高规则
```

输出：

```text
rule-prune-report.md
```

---

## §184 文档债规则

### **[P1]** `RULE-DOC-DEBT-001`：文档重复即债务

<sub>level: P1 · status: active · enforced_by: `goalcli debt` · exit: 1 · source: §184 L7513</sub>

当同一事实出现在 3 个以上文档中，必须抽取 SSOT。

例如：

```text
worktree-only 规则同时出现在：
CONSTITUTION.md
README.md
goal-rules.md
pr-template.md

则必须：
.agent/rules/07-worktree-rules.md 作为 SSOT
其他文档只引用。
```

### **[P1]** `RULE-DOC-DEBT-002`：文档必须有生命周期

<sub>level: P1 · status: active · enforced_by: `goalcli debt` · exit: 1 · source: §184 L7533</sub>

每个长期文档建议包含：

```yaml
status: active | draft | deprecated | archived
owner:
last_reviewed_at:
ssot:
related_rules:
```

---

## §185 Registry Lock 规则

### **[P1]** `RULE-REGISTRY-LOCK-001`：Registry 更新必须加锁

<sub>level: P1 · status: active · enforced_by: `goalcli command-registry` · exit: 1 · source: §185 L7549</sub>

修改以下文件时必须持有 lock：

```text
.agent/registries/goals.yaml
.agent/registries/tasks.yaml
.agent/registries/evidence.yaml
.agent/registries/patches.yaml
.agent/registries/adoption.yaml
```

### **[P1]** `RULE-REGISTRY-LOCK-002`：锁超时必须可恢复

<sub>level: P1 · status: active · enforced_by: `goalcli command-registry` · exit: 1 · source: §185 L7563</sub>

锁文件必须有：

```yaml
owner:
created_at:
expires_at:
command:
pid:
```

超时后：

```bash
goalcli lock recover
```

---

## §188 Rule Compatibility Matrix 规则

### **[P1]** `RULE-COMPAT-MATRIX-001`：规则版本必须有兼容矩阵

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §188 L7656</sub>

示例：

```text
rules-v1.6
compatible:
  harness-v0.1.x
  templates-v0.2.x
  goalcli-v0.1.x
requires:
  schema-v0.1
breaking:
  none
```

### **[P1]** `RULE-COMPAT-MATRIX-002`：不兼容必须阻断 adoption

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §188 L7674</sub>

如果下游：

```text
rules-v1.6 + harness-v0.0
```

不兼容，则：

```text
adoption-check failed
```

---

## §197 No-Orphan 规则

### **[P1]** `RULE-ORPHAN-001`：禁止孤儿对象

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §197 L7919</sub>

以下对象不得孤立存在：

```text
无 Goal 的 Task
无 Task 的 Evidence
无 Requirement 的 AC
无 Evidence 的 Release
无来源的 Patch
无 Decision 的 Waiver
无 Owner 的 P0 Risk
```

### **[P1]** `RULE-ORPHAN-002`：orphan-check 必须进入 CI

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §197 L7935</sub>

```bash
make orphan-check
```

---

## §210 Rule Coverage 规则

### **[P1]** `RULE-COVERAGE-001`：P0/P1 规则必须有覆盖率

<sub>level: P1 · status: active · enforced_by: `goalcli traceability-check` · exit: 9 · source: §210 L8325</sub>

计算：

```text
Rule Coverage = 有机器 Gate 的 P0/P1 规则数 / P0/P1 总规则数
```

Release 要求：

```text
P0 Rule Coverage = 100%
P1 Rule Coverage >= 90%
```

### **[P1]** `RULE-COVERAGE-002`：无覆盖规则必须降级

<sub>level: P1 · status: active · enforced_by: `goalcli traceability-check` · exit: 9 · source: §210 L8342</sub>

如果规则无法机器化：

```text
不能是 P0
不能是 P1
最多 P2/P3
```

---

## §227 Rule Sunset 规则

### **[P1]** `RULE-SUNSET-001`：规则必须允许退役

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §227 L8845</sub>

退役流程：

```text
mark deprecated
add replacement
notify downstream
wait one release cycle
remove enforcement
archive rule
```

### **[P1]** `RULE-SUNSET-002`：无执行价值规则必须退役

<sub>level: P1 · status: active · enforced_by: `goalcli policy-schema` · exit: 6 · source: §227 L8860</sub>

满足任一条件：

```text
长期无人触发
无法机器化
误杀率高
与其他规则重复
下游采用困难
```

进入 sunset review。

---

## §231 Compatibility Guard 规则

### **[P1]** `RULE-COMPAT-GUARD-001`：下游兼容性失败不得阻断 xlib-standard 内部发布

<sub>level: P1 · status: active · enforced_by: `goalcli downstream-adoption` · exit: 1 · source: §231 L8961</sub>

但必须限制发布通道：

```text
内部 stable 可以发布
downstream promotion blocked
```

### **[P1]** `RULE-COMPAT-GUARD-002`：Breaking Rule 必须分阶段推进

<sub>level: P1 · status: active · enforced_by: `goalcli downstream-adoption` · exit: 1 · source: §231 L8972</sub>

```text
warn-only
dual-run
blocking
mandatory
```

不能直接从不存在变成强阻断。
