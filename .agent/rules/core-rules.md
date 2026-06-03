# Core 规则

> 本文件由 `scripts/render_domain_rules.py` 从 [`registry.yaml`](./registry.yaml)
> 与 `.worktree/goal-patch.md` 渲染生成；冲突时以 `iron-rules.md` >
> `registry.yaml` > 本文件 > `.worktree/goal-patch.md` 为序。

本文件覆盖 Goal Runtime **核心控制层**规则：第一性铁律、对象模型、ID/状态机/模式、Context Recovery、SSOT、规则分级与评分、规则冻结等。

权威顺序参见 [`README.md`](./README.md)；P0 铁律压缩见 [`iron-rules.md`](./iron-rules.md)；机器化字段见 [`registry.yaml`](./registry.yaml)。

---

## §1 第一性原理铁律

### **[P0]** `RULE-CORE-001`：没有证据，不允许 DONE

<sub>level: P0 · status: active · enforced_by: `goalcli evidence-check` · exit: 8 · source: §1 L56</sub>

任何 Task、Issue、Goal、Release 都不能只靠描述完成。

必须使用：

```text
DONE with evidence:
- EVID-xxx
- test report
- command output
- PR link
- release manifest
```

### **[P0]** `RULE-CORE-002`：Goal 必须从真实上下文开始

<sub>level: P0 · status: active · enforced_by: `goalcli context-fast-check` · exit: 1 · source: §1 L73</sub>

禁止在没有恢复上下文的情况下直接设计方案。

必须先检查：

```text
仓库结构
已有文档
已有 Makefile target
已有 CI
已有 tests
已有 .agent
已有 issues
已有 release
已有规则
已有冲突
```

### **[P0]** `RULE-CORE-003`：需求必须可验证

<sub>level: P0 · status: active · enforced_by: `goalcli acceptance-matrix` · exit: 1 · source: §1 L94</sub>

所有 Requirement 必须绑定 Acceptance Criteria。

```text
Requirement → Acceptance Criteria → Test → Evidence
```

没有 AC 的需求，不允许进入实现。

### **[P0]** `RULE-CORE-004`：所有变更必须可追踪

<sub>level: P0 · status: active · enforced_by: `goalcli traceability-check` · exit: 9 · source: §1 L106</sub>

每个变更必须能追踪到：

```text
Goal ID
Requirement ID
Acceptance Criteria ID
Task ID
Issue ID
Commit
PR
Evidence
Release
```

### **[P0]** `RULE-CORE-005`：Harness 是机器裁判

<sub>level: P0 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §1 L124</sub>

Harness 负责判断是否允许进入下一阶段。

任何人工判断不能绕过 P0 Harness Gate。

### **[P0]** `RULE-CORE-006`：Self-improving 是强制环节

<sub>level: P0 · status: active · enforced_by: `goalcli self-improving-check` · exit: 1 · source: §1 L132</sub>

每次 Goal 完成后必须输出：

```text
Retrospective
Prompt Patch
Harness Patch
Rule Patch
CI Gate Suggestion
New Issue Candidates
```

否则 Goal 不算闭环完成。

---

## §2 Goal Runtime 模式规则

### **[P1]** `RULE-MODE-001`：必须声明执行模式

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §2 L151</sub>

每个 Goal 必须声明模式：

```text
Lite
Standard
Full
```

---

## §3 Goal 对象模型规则

### **[P1]** `RULE-OBJECT-001`：Goal 必须包含完整字段

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §3 L252</sub>

```text
goal_id
title
mode
owner
repositories
background
problem_statement
target_state
scope
non_goals
constraints
assumptions
success_criteria
risk_level
dependencies
state
```

### **[P1]** `RULE-OBJECT-002`：统一对象关系

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §3 L275</sub>

必须遵循：

```text
Goal owns Spec
Spec contains Requirements
Requirement verified_by Acceptance Criteria
Requirement implemented_by Design
Design executed_by Plan
Plan decomposes_to Tasks
Task verified_by Tests
Task proven_by Evidence
Evidence supports Review
Review unlocks Release
Release triggers Retrospective
Retrospective patches Prompt / Harness / Rules
```

---

## §4 ID 规则

### **[P1]** `RULE-ID-001`：所有核心对象必须有稳定 ID

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §4 L298</sub>

```text
GOAL-YYYYMMDD-NNN
SPEC-<domain>-vX.Y
REQ-<spec-id>-NNN
AC-<req-id>-NNN
DESIGN-<domain>-vX.Y
ADR-YYYYMMDD-NNN
PLAN-<goal-id>-vX.Y
TASK-<goal-id>-NNN
TEST-<task-id>-NNN
EVID-<task-id>-YYYYMMDD-NNN
RISK-<goal-id>-NNN
DEC-YYYYMMDD-NNN
REV-<target-id>-YYYYMMDD-NNN
REL-YYYYMMDD-<domain>
RETRO-YYYYMMDD-NNN
PATCH-PROMPT-YYYYMMDD-NNN
PATCH-HARNESS-YYYYMMDD-NNN
PATCH-RULE-YYYYMMDD-NNN
```

### **[P1]** `RULE-ID-002`：禁止无 ID 的需求、任务、证据

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §4 L323</sub>

以下对象没有 ID 不允许进入执行：

```text
Requirement
Acceptance Criteria
Task
Test
Evidence
Risk
Decision
Release
Retrospective
Patch
```

---

## §5 状态机规则

### **[P0]** `RULE-STATE-001`：Goal 必须经过状态机

<sub>level: P0 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §5 L344</sub>

标准状态机：

```text
INIT
→ CONTEXT_READY
→ GOAL_READY
→ SPEC_READY
→ DESIGN_READY
→ PLAN_READY
→ TASKS_READY
→ EXECUTING
→ VERIFYING
→ REVIEWING
→ RELEASING
→ RETROSPECTING
→ DONE
```

### **[P0]** `RULE-STATE-002`：异常状态必须显式记录

<sub>level: P0 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §5 L366</sub>

```text
BLOCKED
FAILED
NEEDS_RESEARCH
NEEDS_DECISION
NEEDS_REPLAN
NEEDS_ROLLBACK
NEEDS_HUMAN_APPROVAL
INCONSISTENT_STATE
```

### **[P0]** `RULE-STATE-003`：禁止跳状态

<sub>level: P0 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §5 L381</sub>

例如：

```text
禁止 INIT → EXECUTING
禁止 SPEC_READY → RELEASING
禁止 EXECUTING → DONE
禁止 VERIFYING 失败后继续 RELEASE
```

---

## §6 Context Recovery 规则

### **[P1]** `RULE-CONTEXT-001`：必须恢复真实项目状态

<sub>level: P1 · status: active · enforced_by: `goalcli context-fast-check` · exit: 1 · source: §6 L396</sub>

必须检查：

```text
repo root
branch
worktree
file tree
Makefile
CI workflows
docs
tests
.agent
harness
templates
rules
scripts
open issues
recent commits
release tags
```

### **[P1]** `RULE-CONTEXT-002`：禁止引用不存在的能力

<sub>level: P1 · status: active · enforced_by: `goalcli context-fast-check` · exit: 1 · source: §6 L421</sub>

禁止出现：

```text
文档说 make docs-check 存在，但 Makefile 没有
文档说 evidence-check 存在，但脚本不存在
文档说 release manifest 存在，但目录不存在
文档说 harness 已接入，但 CI 未执行
```

### **[P1]** `RULE-CONTEXT-003`：上下文冲突必须进入 Decision Log

<sub>level: P1 · status: active · enforced_by: `goalcli context-fast-check` · exit: 1 · source: §6 L434</sub>

例如：

```text
文档与代码不一致
README 与 Makefile 不一致
Spec 与 CI 不一致
Issue 与实际目录不一致
```

必须生成：

```text
DEC-YYYYMMDD-NNN
```

---

## §28 规则分级体系

### **[P1]** `RULE-CLASS-001`：所有规则必须分级

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §28 L1504</sub>

规则必须分为：

```text
P0: 铁律，不允许绕过，失败即阻断
P1: 核心规则，默认阻断，特殊情况需 Decision Log
P2: 推荐规则，失败不阻断但必须记录
P3: 建议规则，用于优化和评分
```

---

## §29 规则裁决优先级

### **[P1]** `RULE-PRIORITY-001`：规则冲突时按优先级裁决

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §29 L1581</sub>

优先级：

```text
1. Security Rules
2. Worktree Rules
3. Evidence Rules
4. Harness Gate Rules
5. Release Rules
6. Architecture Rules
7. Repository Local Rules
8. Team Convention Rules
9. Style Rules
```

例如：

```text
如果 Agent 想快速修复而直接在 main commit，
Worktree Rule 优先，必须阻断。
```

### **[P1]** `RULE-PRIORITY-002`：本地规则不得覆盖 P0 全局规则

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §29 L1606</sub>

仓库局部规则可以增强，但不能削弱：

```text
不能允许 main 开发
不能允许无 Evidence DONE
不能允许跳过 PR
不能允许无 Traceability merge
```

---

## §49 Failure Budget 规则

### **[P1]** `RULE-FAILURE-001`：Goal 必须有失败预算

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §49 L2357</sub>

建议字段：

```yaml
failure_budget:
  max_retries_per_task: 3
  max_failed_gates: 2
  max_days_blocked: 2
  require_replan_after_failed_attempts: 3
```

### **[P1]** `RULE-FAILURE-002`：超过失败预算必须 replan

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §49 L2371</sub>

进入：

```text
NEEDS_REPLAN
```

必须生成：

```text
failure report
root cause
new plan
decision log
```

---

## §50 评分规则

### **[P1]** `RULE-SCORE-001`：每个 Goal 必须评分

<sub>level: P1 · status: active · enforced_by: `goalcli score` · exit: 1 · source: §50 L2392</sub>

满分 100：

```text
Context Recovery: 10
Spec Quality: 10
Design Quality: 10
Task Decomposition: 10
Harness Coverage: 10
Evidence Quality: 15
Traceability: 10
Automation: 10
Release Readiness: 10
Self-improving: 5
```

---

## §108 Goal Runtime 版本兼容规则

### **[P1]** `RULE-RUNTIME-COMPAT-001`：Goal Runtime 升级必须声明兼容范围

<sub>level: P1 · status: active · enforced_by: `goalcli upgrade-runtime` · exit: 1 · source: §108 L4566</sub>

```yaml
runtime_version: goal-runtime-v1.3
compatible_with:
  - rules-v1.2
  - harness-v0.1
  - templates-v0.2
breaking_changes:
  - worktree-only enforcement changed from warning to blocking
```

### **[P1]** `RULE-RUNTIME-COMPAT-002`：不兼容升级必须提供迁移脚本

<sub>level: P1 · status: active · enforced_by: `goalcli upgrade-runtime` · exit: 1 · source: §108 L4580</sub>

例如：

```bash
goalcli migrate rules --from rules-v1.2 --to rules-v1.3
```

---

## §114 根目录工具规则

### **[P1]** `RULE-ROOT-001`：根目录必须提供统一入口

<sub>level: P1 · status: active · enforced_by: `goalcli boundary` · exit: 1 · source: §114 L4841</sub>

根目录至少需要：

```text
Makefile
goalcli.yaml
.github/workflows/
scripts/harness/
scripts/git/
```

推荐结构：

```text
scripts/
├── harness/
│   ├── no-main-dev.sh
│   ├── evidence-check.sh
│   ├── traceability-check.sh
│   ├── release-check.sh
│   └── secret-check.sh
│
├── git/
│   ├── install-hooks.sh
│   └── cleanup-worktrees.sh
│
└── ci/
    ├── ci-summary.sh
    └── collect-reports.sh
```

---

## §130 防过度工程化规则

### **[P1]** `RULE-SIMPLICITY-001`：v0.1.0 不实现全自动写代码

<sub>level: P1 · status: active · enforced_by: `goalcli minimal-kernel` · exit: 1 · source: §130 L5523</sub>

`goalcli v0.1.0` 的边界：

```text
做检查
做报告
做 Evidence 归档
做 Traceability 检查
做 Worktree 防护
做 Release 检查
做 Audit
```

暂不做：

```text
自动改业务代码
自动跨仓库批量提交
自动 stable 发布
自动解决复杂冲突
自动绕过人工 Review
```

### **[P1]** `RULE-SIMPLICITY-002`：规则不能无限增长而无机器约束

<sub>level: P1 · status: active · enforced_by: `goalcli minimal-kernel` · exit: 1 · source: §130 L5549</sub>

新增规则必须满足至少一个条件：

```text
能被 Harness 检查
能减少重复错误
能降低发布风险
能提升下游采用一致性
能形成自动化证据
```

否则不进入 P0/P1，只能作为 P3 建议。

---

## §131 Lite / Standard / Full Gate 矩阵

### **[P1]** `RULE-MODE-GATE-001`：不同模式 Gate 不同

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §131 L5567</sub>

| Gate               | Lite | Standard | Full |
| ------------------ | ---: | -------: | ---: |
| schema-check       |   必须 |       必须 |   必须 |
| worktree-check     |   必须 |       必须 |   必须 |
| evidence-check     |   必须 |       必须 |   必须 |
| traceability-check |   可选 |       必须 |   必须 |
| design-check       |   可选 |       推荐 |   必须 |
| risk-check         |   可选 |       必须 |   必须 |
| pr-check           |   推荐 |       必须 |   必须 |
| release-check      |   可选 |       推荐 |   必须 |
| retro-check        |   推荐 |       必须 |   必须 |
| adoption-check     |  不需要 |       可选 |   必须 |

---

## §137 执行顺序规则

### **[P1]** `RULE-ORDER-001`：先做 Gate，再做自动化

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §137 L5695</sub>

正确顺序：

```text
1. 规则文档
2. Schema
3. Worktree Gate
4. Evidence Gate
5. Traceability Gate
6. PR Gate
7. Release Gate
8. Retro Gate
9. Audit
10. Issue / PR / Release 自动化
```

错误顺序：

```text
先做自动创建 Issue / PR / Release
但没有 Gate 裁判
```

这会导致自动化扩大错误。

---

## §143 状态迁移门禁规则

### **[P0]** `RULE-STATE-GATE-001`：每个状态迁移必须有 Gate

<sub>level: P0 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §143 L5901</sub>

```text
INIT → CONTEXT_READY              需要 Context Gate
CONTEXT_READY → GOAL_READY        需要 Goal Gate
GOAL_READY → SPEC_READY           需要 Spec Gate
SPEC_READY → DESIGN_READY         需要 Design Gate
DESIGN_READY → PLAN_READY         需要 Plan Gate
PLAN_READY → TASKS_READY          需要 Task Gate
TASKS_READY → EXECUTING           需要 Worktree Gate
EXECUTING → VERIFYING             需要 Local CI Gate
VERIFYING → REVIEWING             需要 Evidence + Traceability Gate
REVIEWING → RELEASING             需要 PR + Review Gate
RELEASING → RETROSPECTING         需要 Release Gate
RETROSPECTING → DONE              需要 Retro + Audit Gate
```

### **[P0]** `RULE-STATE-GATE-002`：状态迁移必须写入 Registry

<sub>level: P0 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §143 L5920</sub>

每次迁移必须更新：

```text
.agent/registries/goals.yaml
.agent/goals/<GOAL-ID>/goal.yaml
.agent/goals/<GOAL-ID>/execution-log.md
```

示例：

```yaml
state_transitions:
  - from: TASKS_READY
    to: EXECUTING
    gate: worktree-check
    status: passed
    evidence: reports/worktree-check.txt
    timestamp: 2026-06-03T18:00:00+09:00
```

---

## §160 Goal Freeze 规则

### **[P1]** `RULE-GOAL-FREEZE-001`：进入 Release 前必须冻结 Goal Scope

<sub>level: P1 · status: active · enforced_by: `goalcli scope-lock` · exit: 1 · source: §160 L6548</sub>

状态：

```text
RELEASE_FREEZE
```

冻结后禁止：

```text
新增 P0/P1 Requirement
扩大 Scope
修改 AC
混入无关 Task
```

除非进入：

```text
NEEDS_REPLAN
```

### **[P1]** `RULE-GOAL-FREEZE-002`：冻结后只允许修复 release blocker

<sub>level: P1 · status: active · enforced_by: `goalcli scope-lock` · exit: 1 · source: §160 L6573</sub>

允许：

```text
修复 P0 Gate
补 Evidence
补 Release Manifest
补 Rollback
修复 Secret
修复 Traceability
```

---

## §174 SSOT 事实源规则

### **[P1]** `RULE-SSOT-001`：每类事实必须只有一个主源

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §174 L7143</sub>

| 事实类型        | SSOT                              |
| ----------- | --------------------------------- |
| 规则          | `.agent/rules/`                   |
| 机器策略        | `.agent/policies/`                |
| 对象结构        | `.agent/schemas/`                 |
| Goal 状态     | `.agent/registries/goals.yaml`    |
| Task 状态     | `.agent/registries/tasks.yaml`    |
| Evidence 索引 | `.agent/registries/evidence.yaml` |
| Patch 状态    | `.agent/registries/patches.yaml`  |
| Release 事实  | `release/<REL-ID>/manifest.md`    |
| 下游采用        | `.agent/registries/adoption.yaml` |

### **[P1]** `RULE-SSOT-002`：非 SSOT 文档只能引用，不得复制事实

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §174 L7159</sub>

例如：

```text
README 可以写：
详见 .agent/rules/07-worktree-rules.md

但不应复制完整 worktree 规则，避免漂移。
```

---

## §198 Naming Normalization 规则

### **[P1]** `RULE-NAMING-001`：命名必须统一

<sub>level: P1 · status: active · enforced_by: `goalcli naming` · exit: 1 · source: §198 L7945</sub>

禁止同一概念多名称：

```text
Goal Runtime / Goal System / Goal Engine 混用
Evidence / Proof / Report 混用
Harness / Gate / Checker 混用
```

必须建立 glossary：

```text
.agent/glossary.md
```

---

## §199 Glossary 规则

### **[P1]** `RULE-GLOSSARY-001`：核心术语必须进入 Glossary

<sub>level: P1 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §199 L7965</sub>

至少包括：

```text
Goal
Goal Pack
Runtime
Harness
Gate
Checker
Evidence
Traceability
Release Manifest
Retrospective
Patch
Adoption
Violation
Waiver
```

### **[P1]** `RULE-GLOSSARY-002`：文档新增术语必须同步 Glossary

<sub>level: P1 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §199 L7988</sub>

否则 docs-check 警告。

---

## §200 最终闭环成熟度等级

### **[P1]** `RULE-MATURITY-001`：Goal Runtime 成熟度分五级

<sub>level: P1 · status: active · enforced_by: `goalcli conformance-profile` · exit: 1 · source: §200 L7996</sub>

```text
L0 文档型：
只有规则文档，无机器 Gate

L1 检查型：
有 Makefile / Harness 检查

L2 证据型：
Evidence / Traceability / Audit 生效

L3 自动化型：
Issue / PR / Release 可自动生成和同步

L4 工厂型：
可向 downstream repo 规模化采用

L5 自进化型：
Retro Patch 能稳定反哺规则、Harness、模板和 CI
```

### **[P1]** `RULE-MATURITY-002`：当前目标应先达到 L2

<sub>level: P1 · status: active · enforced_by: `goalcli conformance-profile` · exit: 1 · source: §200 L8020</sub>

当前阶段不要直接追求 L5。

第一阶段目标：

```text
L2：检查型 + 证据型
```

也就是：

```text
worktree-only 生效
secret-check 生效
evidence-check 生效
traceability-check 生效
release-check 生效
audit goal 生效
```

---

## §213 Source of Truth Conflict 规则

### **[P1]** `RULE-CONFLICT-001`：SSOT 冲突必须自动检测

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §213 L8431</sub>

典型冲突：

```text
tasks.yaml 标记 DONE，但 Issue 仍 OPEN
evidence.yaml 有 Evidence，但文件不存在
release manifest 引用不存在 PR
traceability.md 引用不存在 Requirement
README 复制了过期规则
```

### **[P1]** `RULE-CONFLICT-002`：冲突必须有裁决顺序

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §213 L8445</sub>

裁决顺序：

```text
Schema
Registry
Goal Pack
Evidence Artifact
Git Commit / PR / CI
Release Manifest
Human Decision Log
README / docs
```

---

## §214 Runtime Install Profile 规则

### **[P1]** `RULE-PROFILE-001`：必须区分安装 Profile

<sub>level: P1 · status: active · enforced_by: `goalcli conformance-profile` · exit: 1 · source: §214 L8464</sub>

```text
minimal
standard
full
downstream
xgo
```

### **[P1]** `RULE-PROFILE-002`：不同 Profile 安装不同能力

<sub>level: P1 · status: active · enforced_by: `goalcli conformance-profile` · exit: 1 · source: §214 L8476</sub>

| Profile    | 适用            | 必装                                        |
| ---------- | ------------- | ----------------------------------------- |
| minimal    | 小库            | worktree/evidence/schema                  |
| standard   | 普通基础库         | minimal + traceability/pr/release         |
| full       | xlib-standard | standard + retro/adoption/audit/dashboard |
| downstream | 下游库           | adoption-check + contract                 |
| xgo        | x.go          | x.go 架构专用 gates                           |

---

## §221 Release Freeze Deepening 规则

### **[P1]** `RULE-FREEZE-003`：Freeze 后禁止新增非 blocker 变更

<sub>level: P1 · status: active · enforced_by: `goalcli scope-lock` · exit: 1 · source: §221 L8665</sub>

进入 `RELEASE_FREEZE` 后只允许：

```text
修 P0 Gate
补 Evidence
补 Traceability
补 Manifest
补 Rollback
修 Secret
修 Release blocker
```

### **[P1]** `RULE-FREEZE-004`：Freeze 解除必须有 Decision

<sub>level: P1 · status: active · enforced_by: `goalcli scope-lock` · exit: 1 · source: §221 L8681</sub>

如果要扩大 Scope：

```text
Goal state = NEEDS_REPLAN
生成 DEC-xxx
重新跑 Plan Gate
```

---

## §237 Rule Freeze 规则

### **[P1]** `RULE-FREEZE-RULE-001`：规则体系必须进入冻结态

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §237 L9153</sub>

当规则覆盖以下对象后，应进入冻结态：

```text
Goal
Context
Spec
Design
Plan
Task
Issue
Worktree
Commit
PR
Evidence
Traceability
Review
Release
Retrospective
Self-improving
Harness
Schema
Registry
Audit
Dashboard
Incident
Downstream Adoption
Governance
Sunset
```

冻结后不再随意新增规则，只允许：

```text
修正冲突
消除重复
补机器化路径
补缺失 Gate
补落地文件
补 Issue 拆解
```

### **[P1]** `RULE-FREEZE-RULE-002`：冻结后的新增规则必须经过 Rule Change

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §237 L9198</sub>

新增规则必须满足：

```text
有明确问题来源
有机器化检查方式
有违反样例
有修复路径
有下游影响说明
有是否阻断的 severity
```

否则进入 P3 建议，不进入 P0/P1。
