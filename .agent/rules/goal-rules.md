# Goal 规则

> 源自 Goal 完整规则 v1.0 §0-§9, §18-§19, §22-§27

## 总定义

Goal 不是任务清单，而是 Goal Runtime Execution System。

标准链路：

```text
Goal → Context Recovery → Spec → Design → Plan → Tasks → Execution
→ Verification → Evidence → Review → Release → Retrospective → Self-improving
```

任何 Goal 都必须满足：可理解、可拆解、可执行、可验证、可追踪、可发布、可回滚、可复利改进。

---

## 第一性原理铁律

### RULE-CORE-001：没有证据，不允许 DONE

任何 Task、Issue、Goal、Release 都不能只靠描述完成。必须使用：

```text
DONE with evidence:
- EVID-xxx
- test report
- command output
- PR link
- release manifest
```

### RULE-CORE-002：Goal 必须从真实上下文开始

禁止在没有恢复上下文的情况下直接设计方案。必须先检查：

```text
仓库结构、已有文档、已有 Makefile target、已有 CI、已有 tests、
已有 .agent、已有 harness、已有 issues、已有 release、已有规则、已有冲突
```

### RULE-CORE-003：需求必须可验证

```text
Requirement → Acceptance Criteria → Test → Evidence
```

没有 AC 的需求，不允许进入实现。

### RULE-CORE-004：所有变更必须可追踪

每个变更必须能追踪到：

```text
Goal ID → Requirement ID → AC ID → Task ID → Issue ID → Commit → PR → Evidence → Release
```

### RULE-CORE-005：Harness 是机器裁判

任何人工判断不能绕过 P0 Harness Gate。

### RULE-CORE-006：Self-improving 是强制环节

每次 Goal 完成后必须输出：

```text
Retrospective、Prompt Patch、Harness Patch、Rule Patch、CI Gate Suggestion、New Issue Candidates
```

否则 Goal 不算闭环完成。

---

## Goal Runtime 模式

### RULE-MODE-001：必须声明执行模式

| 模式 | 适用场景 | 最低要求 |
|------|---------|---------|
| **Lite** | 小文档、小修复、小脚本、低风险调整 | Goal、Task、AC、Evidence、Review |
| **Standard** | 普通功能、模块实现、Issue 修复 | Goal、Context、Spec、Plan、Tasks、Tests、Evidence、PR、Review、Retrospective |
| **Full** | 架构升级、基础库标准、跨仓库改造、goalcli/harness/xlib-standard | Goal、Context Recovery、Spec、Design、ADR、Plan、Tasks、Issues、Worktrees、Commits、PRs、Harness Gates、Evidence、Release Manifest、Retrospective、Self-improving Patches |

---

## Goal 对象模型

### RULE-OBJECT-001：Goal 必须包含完整字段

```text
goal_id、title、mode、owner、repositories、background、problem_statement、
target_state、scope、non_goals、constraints、assumptions、success_criteria、
risk_level、dependencies、state
```

### RULE-OBJECT-002：统一对象关系

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

## ID 规则

### RULE-ID-001：所有核心对象必须有稳定 ID

```text
GOAL-YYYYMMDD-NNN          SPEC-<domain>-vX.Y
REQ-<spec-id>-NNN          AC-<req-id>-NNN
DESIGN-<domain>-vX.Y       ADR-YYYYMMDD-NNN
PLAN-<goal-id>-vX.Y        TASK-<goal-id>-NNN
TEST-<task-id>-NNN          EVID-<task-id>-YYYYMMDD-NNN
RISK-<goal-id>-NNN          DEC-YYYYMMDD-NNN
REV-<target-id>-YYYYMMDD-NNN  REL-YYYYMMDD-<domain>
RETRO-YYYYMMDD-NNN         PATCH-PROMPT-YYYYMMDD-NNN
PATCH-HARNESS-YYYYMMDD-NNN  PATCH-RULE-YYYYMMDD-NNN
```

### RULE-ID-002：禁止无 ID 的需求、任务、证据

以下对象没有 ID 不允许进入执行：Requirement、AC、Task、Test、Evidence、Risk、Decision、Release、Retrospective、Patch。

---

## 状态机

### RULE-STATE-001：Goal 必须经过状态机

```text
INIT → CONTEXT_READY → GOAL_READY → SPEC_READY → DESIGN_READY → PLAN_READY
→ TASKS_READY → EXECUTING → VERIFYING → REVIEWING → RELEASING → RETROSPECTING → DONE
```

### RULE-STATE-002：异常状态必须显式记录

```text
BLOCKED、FAILED、NEEDS_RESEARCH、NEEDS_DECISION、NEEDS_REPLAN、
NEEDS_ROLLBACK、NEEDS_HUMAN_APPROVAL、INCONSISTENT_STATE
```

### RULE-STATE-003：禁止跳状态

```text
禁止 INIT → EXECUTING
禁止 SPEC_READY → RELEASING
禁止 EXECUTING → DONE
禁止 VERIFYING 失败后继续 RELEASE
```

---

## Context Recovery

### RULE-CONTEXT-001：必须恢复真实项目状态

必须检查：repo root、branch、worktree、file tree、Makefile、CI workflows、docs、tests、.agent、harness、templates、rules、scripts、open issues、recent commits、release tags。

### RULE-CONTEXT-002：禁止引用不存在的能力

禁止文档声称存在某个 Makefile target / 脚本 / 目录，但实际不存在。

### RULE-CONTEXT-003：上下文冲突必须进入 Decision Log

文档与代码不一致、README 与 Makefile 不一致等冲突，必须生成 `DEC-YYYYMMDD-NNN`。

---

## Spec 规则

### RULE-SPEC-001：Spec 必须包含 Requirement

每个 Requirement 必须包含：req_id、description、priority、rationale、source、acceptance_criteria、verification_method、risk。

### RULE-SPEC-002：Requirement 必须分优先级

```text
P0: 阻断级，必须完成
P1: 核心能力
P2: 增强能力
P3: 可延后
```

### RULE-SPEC-003：Acceptance Criteria 必须可验证

每个 AC 必须包含：ac_id、statement、verification_type（semantic/executable/hybrid）、pass_condition、fail_condition、required_evidence。

### RULE-SPEC-004：禁止抽象不可验收需求

禁止：提升质量、完善体系、优化结构、增强能力、更加健壮。

必须改为可执行描述，如：增加 `make evidence-check`、所有 Task 必须绑定 Evidence。

---

## Design 规则

### RULE-DESIGN-001：设计必须降低执行歧义

Design 必须包含：architecture、module boundaries、interfaces、data flow、control flow、config model、error handling、observability、security、compatibility、migration、rollback。

### RULE-DESIGN-002：关键设计必须写 ADR

必须写 ADR 的情况：目录结构变化、公共 API 变化、存储模型变化、配置模型变化、CI/Harness 变化、安全策略变化、跨仓库规则变化、发布流程变化。

### RULE-DESIGN-003：Design 不能替代 Task

设计只回答"怎么做"，不能当成执行结果。

---

## Plan / Task 规则

### RULE-TASK-001：Task 是最小可执行单元

每个 Task 必须：可独立执行、可独立验证、可独立回滚、可独立收集 Evidence、可映射到 Requirement。

### RULE-TASK-002：Task 必须包含完整字段

```text
task_id、title、type、priority、related_requirement、related_ac、
files_to_change、steps、commands_to_run、tests_to_pass、
evidence_to_collect、done_definition、rollback_plan、risks
```

### RULE-TASK-003：禁止超大 Task

一个 Task 不应该同时做：改架构、改 CI、改文档、改测试、改发布、改规则。必须拆分。

### RULE-TASK-004：Task 必须可生成 Issue

每个 Task 都应该能映射为 GitHub Issue。

---

## AutoResearch 规则

### RULE-RESEARCH-001：未知项必须进入 AutoResearch

以下情况必须研究：API 行为不确定、依赖版本不确定、Issue 描述不完整、架构冲突、文档与代码不一致、测试失败原因不明确、外部系统可能变化、安全规则不明确。

### RULE-RESEARCH-002：Research 必须产出 Decision

不能只输出资料摘要，必须形成：事实、假设、风险、选项、推荐决策、证据、DEC-xxx。

---

## Change Propagation 规则

| 变更对象 | 必须同步 |
|---------|---------|
| Goal | Spec / Plan / Tasks / Issues |
| Spec | Design / Plan / Tasks / Tests |
| Requirement | AC / Tasks / Tests / Evidence |
| Design | ADR / Plan / Risk / Docs |
| Task | Issue / Branch / Commit / Evidence |
| Public API | Docs / Examples / Tests |
| Config | Schema / Docs / Migration |
| CI Gate | Makefile / Workflow / Reports |
| Release | Changelog / Manifest / Tag |
| Rule | Harness / Templates / Docs |

---

## xlib-standard / kernel / x.go 专用规则

### RULE-XSTACK-001：xlib-standard 是标准源

xlib-standard = 基础库标准源；.agent = 运行时控制平面；goalcli = 机器裁判与执行器；Evidence = 完成证明；downstream adoption = 扩张方式；self-improving = 复利机制。

### RULE-XSTACK-002：kernel 是 L0 内核库

kernel 只能沉淀跨库通用、稳定、低依赖的 L0 能力。

### RULE-XSTACK-003：L1/L2 必须遵守分层

```text
L0: kernel
L1: configx / observex / testkitx
L2: redisx / kafkax / postgresx / taosx / ossx / clickhousex
```

禁止：L0 依赖 L1/L2、L1 依赖 L2、L2 横向强耦合。

### RULE-XGO-001：x.go 专用架构约束

Market Data 不直接决定 Regime；Macro Data 不直接依赖 Market Data 内部实现；Regime Engine 只消费标准化状态输入；Storage 通过 interface 隔离；Config 不使用隐式全局状态；CI Gate 优先 Go 化。

---

## goalcli 最小命令集

从 `goalcli v0.1.0` 开始：

```bash
goalcli goal init         goalcli context scan
goalcli spec check        goalcli design check       goalcli tasks check
goalcli worktree create   goalcli worktree check     goalcli worktree clean
goalcli issues create     goalcli issues sync        goalcli issues status
goalcli pr create         goalcli pr update          goalcli pr ready
goalcli evidence collect  goalcli evidence check
goalcli release prepare   goalcli release publish
goalcli retro generate    goalcli patch propose
```

---

## Makefile Gate 最小集合

```makefile
make worktree-check    make goal-check      make context-check
make spec-check        make design-check    make task-check
make issue-check       make pr-check        make evidence-check
make release-check     make retro-check     make ci
```

---

## 最小验收清单

### Goal 验收

```text
[ ] Goal 有 ID
[ ] Goal 有 mode
[ ] Goal 有 scope / non-goals
[ ] Goal 有 success criteria
[ ] Goal 有 constraints
[ ] Goal 有 risk level
[ ] Goal 有 state
```

### Spec 验收

```text
[ ] 每个 Requirement 有 AC
[ ] 每个 AC 有验证方法
[ ] 每个 AC 有 Evidence 要求
[ ] 没有不可验证需求
```

### Task 验收

```text
[ ] 每个 Task 绑定 Requirement
[ ] 每个 Task 绑定 AC
[ ] 每个 Task 有测试命令
[ ] 每个 Task 有 Evidence
[ ] 每个 Task 有 rollback
```

### Worktree 验收

```text
[ ] main 禁止开发
[ ] main 禁止直接 push
[ ] 每个 Task 使用独立 worktree
[ ] make worktree-check 通过
```

### PR 验收

```text
[ ] PR 绑定 Issue
[ ] PR 绑定 Goal
[ ] PR 包含 Evidence
[ ] PR 包含 Traceability
[ ] CI 通过
[ ] Harness 通过
```

---

## 交付清单

```text
.agent/rules/goal-rules.md
.agent/rules/worktree-rules.md
.agent/rules/evidence-rules.md
.agent/rules/harness-rules.md
.agent/rules/self-improving-rules.md
.agent/rules/issue-rules.md
.agent/rules/commit-rules.md
.agent/rules/pr-rules.md
.agent/rules/release-rules.md
.agent/rules/risk-decision-rules.md
.agent/rules/security-rules.md

.agent/templates/issue-template.md
.agent/templates/pr-template.md
.agent/templates/evidence-template.md
.agent/templates/release-manifest-template.md
.agent/templates/retrospective-template.md

.agent/harness/gates/context-gate.yaml
.agent/harness/gates/spec-gate.yaml
.agent/harness/gates/design-gate.yaml
.agent/harness/gates/task-gate.yaml
.agent/harness/gates/worktree-gate.yaml
.agent/harness/gates/commit-gate.yaml
.agent/harness/gates/pr-gate.yaml
.agent/harness/gates/evidence-gate.yaml
.agent/harness/gates/release-gate.yaml
.agent/harness/gates/retro-gate.yaml

scripts/harness/no-main-dev.sh
.githooks/pre-commit
.githooks/pre-push
.github/workflows/worktree-guard.yml
.github/workflows/goal-gates.yml

Makefile
goalcli CLI
reports/
release/
```

---

## 最终铁律摘要

```text
1. Goal 不是 todo，是 Runtime System。
2. 不恢复上下文，不允许设计。
3. 没有 AC 的需求，不允许实现。
4. 没有 Evidence，不允许 DONE。
5. main 只做基线，不允许开发。
6. 所有开发必须使用 git worktree。
7. 所有变更必须可追踪到 Goal / Req / AC / Task / Evidence。
8. Harness Gate 失败必须阻断。
9. PR 没有 Evidence 不允许 merge。
10. Release 没有 Manifest 不允许发布。
11. Retrospective 缺失不算闭环。
12. 重复问题必须升级为 Prompt / Harness / Rule Patch。
```

> 最终目标：让任何 Goal 都能自动拆解、自动执行、自动验证、自动交付、自动复盘，并把每一次执行经验固化为下一轮更强的工程系统。
