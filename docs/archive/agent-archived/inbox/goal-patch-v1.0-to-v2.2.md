下面是 **Goal 完整规则 v1.0**，可直接作为：

```text
.agent/rules/goal-rules.md
.agent/rules/harness-rules.md
.agent/rules/evidence-rules.md
.agent/rules/worktree-rules.md
CONSTITUTION.md
```

的规则源。

---

# Goal 完整规则 v1.0

## 0. 总定义

**Goal 不是任务清单，而是 Goal Runtime Execution System。**

标准链路必须是：

```text
Goal
→ Context Recovery
→ Spec
→ Design
→ Plan
→ Tasks
→ Execution
→ Verification
→ Evidence
→ Review
→ Release
→ Retrospective
→ Self-improving
```

任何 Goal 都必须满足：

```text
可理解
可拆解
可执行
可验证
可追踪
可发布
可回滚
可复利改进
```

---

# 1. 第一性原理铁律

## RULE-CORE-001：没有证据，不允许 DONE

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

---

## RULE-CORE-002：Goal 必须从真实上下文开始

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

---

## RULE-CORE-003：需求必须可验证

所有 Requirement 必须绑定 Acceptance Criteria。

```text
Requirement → Acceptance Criteria → Test → Evidence
```

没有 AC 的需求，不允许进入实现。

---

## RULE-CORE-004：所有变更必须可追踪

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

---

## RULE-CORE-005：Harness 是机器裁判

Harness 负责判断是否允许进入下一阶段。

任何人工判断不能绕过 P0 Harness Gate。

---

## RULE-CORE-006：Self-improving 是强制环节

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

# 2. Goal Runtime 模式规则

## RULE-MODE-001：必须声明执行模式

每个 Goal 必须声明模式：

```text
Lite
Standard
Full
```

---

## Lite Mode

适用于：

```text
小文档
小修复
小脚本
低风险调整
```

最低要求：

```text
Goal
Task
Acceptance Criteria
Evidence
Review
```

---

## Standard Mode

适用于：

```text
普通功能
模块实现
Issue 修复
中等复杂度改造
```

最低要求：

```text
Goal
Context
Spec
Plan
Tasks
Tests
Evidence
PR
Review
Retrospective
```

---

## Full Mode

适用于：

```text
架构升级
基础库标准
跨仓库改造
运行时控制平面
自动化系统
goalcli / harness / xlib-standard / kernel / x.go
```

最低要求：

```text
Goal
Context Recovery
Spec
Design
ADR
Plan
Tasks
Issues
Worktrees
Commits
PRs
Harness Gates
Evidence
Release Manifest
Retrospective
Self-improving Patches
```

---

# 3. Goal 对象模型规则

## RULE-OBJECT-001：Goal 必须包含完整字段

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

---

## RULE-OBJECT-002：统一对象关系

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

# 4. ID 规则

## RULE-ID-001：所有核心对象必须有稳定 ID

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

---

## RULE-ID-002：禁止无 ID 的需求、任务、证据

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

# 5. 状态机规则

## RULE-STATE-001：Goal 必须经过状态机

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

---

## RULE-STATE-002：异常状态必须显式记录

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

---

## RULE-STATE-003：禁止跳状态

例如：

```text
禁止 INIT → EXECUTING
禁止 SPEC_READY → RELEASING
禁止 EXECUTING → DONE
禁止 VERIFYING 失败后继续 RELEASE
```

---

# 6. Context Recovery 规则

## RULE-CONTEXT-001：必须恢复真实项目状态

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

---

## RULE-CONTEXT-002：禁止引用不存在的能力

禁止出现：

```text
文档说 make docs-check 存在，但 Makefile 没有
文档说 evidence-check 存在，但脚本不存在
文档说 release manifest 存在，但目录不存在
文档说 harness 已接入，但 CI 未执行
```

---

## RULE-CONTEXT-003：上下文冲突必须进入 Decision Log

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

# 7. Spec 规则

## RULE-SPEC-001：Spec 必须包含 Requirement

每个 Requirement 必须包含：

```text
req_id
description
priority
rationale
source
acceptance_criteria
verification_method
risk
```

---

## RULE-SPEC-002：Requirement 必须分优先级

```text
P0: 阻断级，必须完成
P1: 核心能力
P2: 增强能力
P3: 可延后
```

---

## RULE-SPEC-003：Acceptance Criteria 必须可验证

每个 AC 必须包含：

```text
ac_id
statement
verification_type
pass_condition
fail_condition
required_evidence
```

验证类型：

```text
semantic
executable
hybrid
```

---

## RULE-SPEC-004：禁止抽象不可验收需求

禁止：

```text
提升质量
完善体系
优化结构
增强能力
更加健壮
```

必须改为：

```text
增加 make evidence-check
所有 Task 必须绑定 Evidence
PR 模板必须包含 Traceability Matrix
main 分支禁止直接 commit
```

---

# 8. Design 规则

## RULE-DESIGN-001：设计必须降低执行歧义

Design 必须包含：

```text
architecture
module boundaries
interfaces
data flow
control flow
config model
error handling
observability
security
compatibility
migration
rollback
```

---

## RULE-DESIGN-002：关键设计必须写 ADR

必须写 ADR 的情况：

```text
目录结构变化
公共 API 变化
存储模型变化
配置模型变化
CI/Harness 变化
安全策略变化
跨仓库规则变化
发布流程变化
```

---

## RULE-DESIGN-003：Design 不能替代 Task

设计只回答“怎么做”，不能当成执行结果。

---

# 9. Plan / Task 规则

## RULE-TASK-001：Task 是最小可执行单元

每个 Task 必须满足：

```text
可独立执行
可独立验证
可独立回滚
可独立收集 Evidence
可映射到 Requirement
```

---

## RULE-TASK-002：Task 必须包含完整字段

```text
task_id
title
type
priority
related_requirement
related_ac
files_to_change
steps
commands_to_run
tests_to_pass
evidence_to_collect
done_definition
rollback_plan
risks
```

---

## RULE-TASK-003：禁止超大 Task

一个 Task 不应该同时做：

```text
改架构
改 CI
改文档
改测试
改发布
改规则
```

必须拆分。

---

## RULE-TASK-004：Task 必须可生成 Issue

每个 Task 都应该能映射为 GitHub Issue。

---

# 10. Worktree-Only 规则

## RULE-WORKTREE-001：禁止 main 开发

```text
main / master 只能作为同步基线与发布基线。
禁止在 main / master 上直接开发、commit、push。
```

---

## RULE-WORKTREE-002：所有开发必须使用 git worktree

每个 Goal / Issue / Task 必须创建独立 worktree。

推荐结构：

```text
~/code/<repo>                      # main worktree，只同步
~/code/.worktrees/<repo>/<goal>/<task>
```

---

## RULE-WORKTREE-003：worktree 分支必须绑定 Goal / Task

分支命名：

```text
goal/<GOAL-ID>/<TASK-ID>
issue/<ISSUE-ID>
task/<TASK-ID>
```

---

## RULE-WORKTREE-004：必须提供 worktree gate

必须有：

```text
make worktree-check
scripts/harness/no-main-dev.sh
.githooks/pre-commit
.githooks/pre-push
```

---

## RULE-WORKTREE-005：PR 合并后必须清理 worktree

```text
git worktree remove <path>
git worktree prune
```

---

# 11. Harness 规则

## RULE-HARNESS-001：Harness Gate 是阶段裁判

每个阶段必须有 Gate：

```text
Context Gate
Goal Gate
Spec Gate
Design Gate
Plan Gate
Task Gate
Issue Gate
Worktree Gate
Commit Gate
PR Gate
CI Gate
Evidence Gate
Review Gate
Release Gate
Retrospective Gate
Self-improving Gate
```

---

## RULE-HARNESS-002：Gate 必须有标准结构

```yaml
gate_id:
name:
type: semantic | executable | hybrid
severity: P0 | P1 | P2 | P3
target:
rules:
commands:
pass_condition:
fail_condition:
evidence_required:
```

---

## RULE-HARNESS-003：P0 Gate 失败必须阻断

P0 Gate 失败时：

```text
禁止 commit
禁止 PR ready
禁止 merge
禁止 release
禁止 DONE
```

---

## RULE-HARNESS-004：Harness 结果必须归档

必须生成：

```text
reports/context-check.json
reports/spec-check.json
reports/design-check.json
reports/task-check.json
reports/worktree-check.txt
reports/evidence-check.json
reports/release-check.json
```

---

# 12. Evidence 规则

## RULE-EVIDENCE-001：Evidence 是完成证明

Evidence 必须包含：

```text
evidence_id
related_goal
related_task
related_requirement
related_ac
command
output
artifact_path
timestamp
status
```

---

## RULE-EVIDENCE-002：Evidence 必须可复查

禁止：

```text
测试通过了
已完成
应该没问题
已修复
```

必须有：

```text
命令
输出
日志
报告
文件路径
PR 链接
CI 链接
```

---

## RULE-EVIDENCE-003：Evidence 必须进入 Traceability Matrix

```text
Requirement → AC → Task → Test → Evidence → Status
```

---

# 13. Issue 自动化规则

## RULE-ISSUE-001：Issue 必须从 Task 生成

Issue 必须包含：

```text
Goal ID
Task ID
Requirement ID
Acceptance Criteria
Implementation Scope
Files to Change
Commands to Run
Evidence Required
DoD
Risk
Rollback Plan
```

---

## RULE-ISSUE-002：Issue 必须有标准 Label

推荐 Label：

```text
goal
spec
design
task
harness
evidence
self-improving
release
risk
blocked
needs-research
needs-decision
```

---

## RULE-ISSUE-003：Issue 关闭必须有 Evidence

没有 Evidence 的 Issue 不允许关闭。

---

# 14. Commit 规则

## RULE-COMMIT-001：Commit 必须绑定 Task / Issue / Evidence

推荐格式：

```text
<type>(<scope>): <summary>

Refs: TASK-xxx
Closes: #123
Evidence: EVID-xxx
```

---

## RULE-COMMIT-002：禁止无语义提交

禁止：

```text
fix
update
wip
stuff
misc
temp
```

---

## RULE-COMMIT-003：禁止大杂烩提交

一个 commit 应该对应一个明确变更意图。

---

## RULE-COMMIT-004：Commit 前必须通过本地 Gate

至少：

```text
make worktree-check
make lint
make test
make evidence-check
```

---

# 15. PR 规则

## RULE-PR-001：PR 必须是可审查交付单元

PR 必须包含：

```text
Goal
Related Issues
Requirements Covered
Changes
Tests
Evidence
Risk
Rollback
Checklist
```

---

## RULE-PR-002：PR 必须包含 Traceability

```text
Requirement | AC | Task | Test | Evidence | Status
```

---

## RULE-PR-003：PR 合并条件

必须满足：

```text
CI passed
Harness passed
Evidence complete
Traceability complete
No P0 risk open
Review approved
main up to date
```

---

## RULE-PR-004：禁止直接合并未验证 PR

没有 Evidence / Harness / Review 的 PR 不允许 merge。

---

# 16. Release 规则

## RULE-RELEASE-001：Release 必须有 Release Manifest

Release Manifest 必须包含：

```text
version
date
goal
commit
tag
included issues
included PRs
changes
evidence summary
test summary
compatibility
migration notes
risks
rollback plan
known issues
retrospective
```

---

## RULE-RELEASE-002：Release 前必须全部验证

必须通过：

```text
make release-check
make evidence-check
make ci
```

---

## RULE-RELEASE-003：Release 必须可回滚

必须有：

```text
rollback command
rollback condition
last known good version
affected components
risk note
```

---

# 17. Retrospective / Self-improving 规则

## RULE-RETRO-001：每个 Goal 必须有 Retrospective

必须回答：

```text
什么有效
什么失败
根因是什么
哪个 Gate 缺失
哪个 Rule 缺失
哪个 Prompt 需要修复
哪个 CI 需要新增
下轮如何自动避免
```

---

## RULE-RETRO-002：必须生成 Patch

至少生成：

```text
Prompt Patch
Harness Patch
Rule Patch
CI Gate Suggestion
New Issue Candidates
```

---

## RULE-RETRO-003：重复问题必须升级为规则

如果同类问题出现两次，必须：

```text
加入 rule
加入 harness gate
加入 CI check
加入 template
```

---

# 18. AutoResearch 规则

## RULE-RESEARCH-001：未知项必须进入 AutoResearch

以下情况必须研究：

```text
API 行为不确定
依赖版本不确定
Issue 描述不完整
架构冲突
文档与代码不一致
测试失败原因不明确
外部系统可能变化
安全规则不明确
```

---

## RULE-RESEARCH-002：Research 必须产出 Decision

不能只输出资料摘要，必须形成：

```text
事实
假设
风险
选项
推荐决策
证据
DEC-xxx
```

---

# 19. Change Propagation 规则

任何变更必须同步下游对象。

| 变更对象        | 必须同步                               |
| ----------- | ---------------------------------- |
| Goal        | Spec / Plan / Tasks / Issues       |
| Spec        | Design / Plan / Tasks / Tests      |
| Requirement | AC / Tasks / Tests / Evidence      |
| Design      | ADR / Plan / Risk / Docs           |
| Task        | Issue / Branch / Commit / Evidence |
| Public API  | Docs / Examples / Tests            |
| Config      | Schema / Docs / Migration          |
| CI Gate     | Makefile / Workflow / Reports      |
| Release     | Changelog / Manifest / Tag         |
| Rule        | Harness / Templates / Docs         |

---

# 20. Risk / Decision / Rollback 规则

## RULE-RISK-001：P0/P1 风险必须登记

Risk Register 字段：

```text
risk_id
description
probability
impact
severity
affected_objects
mitigation
fallback
owner
status
```

---

## RULE-DECISION-001：关键选择必须记录

Decision Log 字段：

```text
decision_id
context
options
selected_option
reason
tradeoff
affected_objects
rollback_condition
```

---

## RULE-ROLLBACK-001：高风险变更必须可回滚

涉及以下内容必须写 rollback：

```text
CI
release
storage
config
public API
security
automation
harness
rules
```

---

# 21. Security 规则

## RULE-SECURITY-001：禁止提交密钥

禁止进入：

```text
source code
README
tests
logs
release manifest
PR description
evidence report
```

---

## RULE-SECURITY-002：密钥必须走外部注入

对于你的基础库体系，默认使用：

```text
/home/k8s/secrets/env/*
```

---

## RULE-SECURITY-003：Evidence 不能泄漏敏感信息

Evidence 中必须过滤：

```text
token
password
secret
private key
access key
cookie
authorization header
```

---

# 22. xlib-standard / kernel / x.go 专用规则

## RULE-XSTACK-001：xlib-standard 是标准源

```text
xlib-standard = 基础库标准源
.agent = 运行时控制平面
goalcli = 机器裁判与执行器
Evidence = 完成证明
downstream adoption = 扩张方式
self-improving = 复利机制
```

---

## RULE-XSTACK-002：kernel 是 L0 内核库

```text
kernel 只能沉淀跨库通用、稳定、低依赖的 L0 能力。
```

---

## RULE-XSTACK-003：L1/L2 必须遵守分层

```text
L0: kernel
L1: configx / observex / testkitx
L2: redisx / kafkax / postgresx / taosx / ossx / clickhousex
```

禁止：

```text
L0 依赖 L1/L2
L1 依赖 L2
L2 横向强耦合
```

---

## RULE-XGO-001：x.go 专用架构约束

```text
Market Data 不直接决定 Regime
Macro Data 不直接依赖 Market Data 内部实现
Regime Engine 只消费标准化状态输入
Storage 通过 interface 隔离
Config 不使用隐式全局状态
CI Gate 优先 Go 化
```

---

# 23. 必须提供的 Makefile Gate

最小集合：

```makefile
make worktree-check
make goal-check
make context-check
make spec-check
make design-check
make task-check
make issue-check
make pr-check
make evidence-check
make release-check
make retro-check
make ci
```

---

# 24. 必须提供的 goalcli 命令

从 `goalcli v0.1.0` 开始，建议最小命令集：

```bash
goalcli goal init
goalcli context scan
goalcli spec check
goalcli design check
goalcli tasks check

goalcli worktree create
goalcli worktree check
goalcli worktree clean

goalcli issues create
goalcli issues sync
goalcli issues status

goalcli pr create
goalcli pr update
goalcli pr ready

goalcli evidence collect
goalcli evidence check

goalcli release prepare
goalcli release publish

goalcli retro generate
goalcli patch propose
```

---

# 25. 最小验收清单

## Goal 验收

```text
[ ] Goal 有 ID
[ ] Goal 有 mode
[ ] Goal 有 scope / non-goals
[ ] Goal 有 success criteria
[ ] Goal 有 constraints
[ ] Goal 有 risk level
[ ] Goal 有 state
```

---

## Spec 验收

```text
[ ] 每个 Requirement 有 AC
[ ] 每个 AC 有验证方法
[ ] 每个 AC 有 Evidence 要求
[ ] 没有不可验证需求
```

---

## Task 验收

```text
[ ] 每个 Task 绑定 Requirement
[ ] 每个 Task 绑定 AC
[ ] 每个 Task 有测试命令
[ ] 每个 Task 有 Evidence
[ ] 每个 Task 有 rollback
```

---

## Worktree 验收

```text
[ ] main 禁止开发
[ ] main 禁止直接 push
[ ] 每个 Task 使用独立 worktree
[ ] make worktree-check 通过
```

---

## PR 验收

```text
[ ] PR 绑定 Issue
[ ] PR 绑定 Goal
[ ] PR 包含 Evidence
[ ] PR 包含 Traceability
[ ] CI 通过
[ ] Harness 通过
```

---

## Release 验收

```text
[ ] Issues 已关闭
[ ] PRs 已合并
[ ] Evidence 已归档
[ ] CHANGELOG 已更新
[ ] VERSION 已更新
[ ] Release Manifest 已生成
[ ] Rollback Plan 已记录
```

---

## Self-improving 验收

```text
[ ] Retrospective 已生成
[ ] Prompt Patch 已生成
[ ] Harness Patch 已生成
[ ] Rule Patch 已生成
[ ] New Issue Candidates 已生成
```

---

# 26. 交付清单

```text
.agent/rules/goal-rules.md
.agent/rules/worktree-rules.md
.agent/rules/evidence-rules.md
.agent/rules/harness-rules.md
.agent/rules/self-improving-rules.md

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

# 27. 最终铁律摘要

最重要的是这 12 条：

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

最终目标：

> **让任何 Goal 都能自动拆解、自动执行、自动验证、自动交付、自动复盘，并把每一次执行经验固化为下一轮更强的工程系统。**



继续补齐 **Goal 完整规则 v1.1：可执行化规则、自动化规则、治理规则、评分规则**。

上一版解决的是 **Goal 应该有哪些规则**。
这一版继续解决：

```text
这些规则如何落到文件？
如何被 goalcli / Harness / CI 执行？
如何阻断违规？
如何形成长期复利？
如何判断一个 Goal 是否合格？
```

---

# 28. 规则分级体系

## RULE-CLASS-001：所有规则必须分级

规则必须分为：

```text
P0: 铁律，不允许绕过，失败即阻断
P1: 核心规则，默认阻断，特殊情况需 Decision Log
P2: 推荐规则，失败不阻断但必须记录
P3: 建议规则，用于优化和评分
```

---

## P0 规则

P0 失败时必须阻断：

```text
禁止继续执行
禁止 commit
禁止 PR ready
禁止 merge
禁止 release
禁止 DONE
```

典型 P0：

```text
没有 Evidence
main 直接开发
Requirement 没有 AC
PR 没有 Traceability
Release 没有 Manifest
提交包含密钥
Harness Gate 失败
```

---

## P1 规则

P1 失败时默认阻断，但可以通过 Decision Log 临时放行：

```text
Design 缺 ADR
Task 过大
文档未同步
Risk 未登记
Rollback 不完整
```

放行必须生成：

```text
DEC-YYYYMMDD-NNN
```

---

## P2 / P3 规则

P2 / P3 不一定阻断，但必须进入评分和 Retrospective。

例如：

```text
模板不够清晰
命名不够统一
文档可读性不足
报告格式不一致
```

---

# 29. 规则裁决优先级

## RULE-PRIORITY-001：规则冲突时按优先级裁决

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

---

## RULE-PRIORITY-002：本地规则不得覆盖 P0 全局规则

仓库局部规则可以增强，但不能削弱：

```text
不能允许 main 开发
不能允许无 Evidence DONE
不能允许跳过 PR
不能允许无 Traceability merge
```

---

# 30. 规则文件结构

建议 `.agent/rules/` 拆成：

```text
.agent/rules/
├── 00-index.md
├── 01-core-rules.md
├── 02-goal-runtime-rules.md
├── 03-context-rules.md
├── 04-spec-rules.md
├── 05-design-rules.md
├── 06-task-rules.md
├── 07-worktree-rules.md
├── 08-issue-rules.md
├── 09-commit-rules.md
├── 10-pr-rules.md
├── 11-evidence-rules.md
├── 12-release-rules.md
├── 13-retrospective-rules.md
├── 14-self-improving-rules.md
├── 15-security-rules.md
├── 16-xstack-rules.md
├── 17-downstream-adoption-rules.md
├── 18-deprecation-rules.md
├── 19-scoring-rules.md
└── 20-rule-change-protocol.md
```

---

# 31. 规则索引文件

`.agent/rules/00-index.md` 必须作为规则入口。

```md
# Goal Runtime Rules Index

## Rule Sources

| File | Domain | Severity |
|---|---|---|
| 01-core-rules.md | Core Runtime | P0 |
| 07-worktree-rules.md | Worktree-only Development | P0 |
| 11-evidence-rules.md | Evidence Protocol | P0 |
| 12-release-rules.md | Release Protocol | P0/P1 |
| 14-self-improving-rules.md | Retrospective & Patch | P1 |
| 15-security-rules.md | Secret / Safety | P0 |

## Mandatory P0 Gates

- worktree-check
- evidence-check
- traceability-check
- release-check
- secret-check
- no-main-dev-check
```

---

# 32. 规则机器化结构

每条规则除了 Markdown 描述，还应该有 YAML 机器规则。

目录：

```text
.agent/policies/
├── core.yaml
├── worktree.yaml
├── evidence.yaml
├── pr.yaml
├── release.yaml
├── security.yaml
└── self_improving.yaml
```

示例：

```yaml
rules:
  - id: RULE-WORKTREE-001
    title: No development on main
    severity: P0
    scope:
      - commit
      - pr
      - release
    check:
      type: executable
      command: make worktree-check
    fail_behavior:
      block_commit: true
      block_pr: true
      block_release: true
    evidence_required:
      - reports/worktree-check.txt
```

---

# 33. Goal Registry 规则

## RULE-REGISTRY-001：所有 Goal 必须登记

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

---

## RULE-REGISTRY-002：Goal 状态必须同步

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

# 34. Task Registry 规则

必须存在：

```text
.agent/registries/tasks.yaml
```

示例：

```yaml
tasks:
  - task_id: TASK-GOAL-20260603-001-001
    goal_id: GOAL-20260603-001
    req_id: REQ-SPEC-goal-runtime-v1.0-001
    ac_ids:
      - AC-REQ-SPEC-goal-runtime-v1.0-001-001
    issue: 123
    branch: goal/GOAL-20260603-001/TASK-001
    worktree: ~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001
    status: READY
    evidence:
      - EVID-TASK-GOAL-20260603-001-001-20260603-001
```

---

# 35. Evidence Registry 规则

必须存在：

```text
.agent/registries/evidence.yaml
```

示例：

```yaml
evidence:
  - evidence_id: EVID-TASK-GOAL-20260603-001-001-20260603-001
    goal_id: GOAL-20260603-001
    task_id: TASK-GOAL-20260603-001-001
    command: make evidence-check
    status: passed
    artifact:
      - reports/evidence-check.json
      - .agent/goals/GOAL-20260603-001/evidence/TASK-001.md
    timestamp: 2026-06-03T16:00:00+09:00
```

---

# 36. Traceability 强制规则

## RULE-TRACE-001：Traceability Matrix 是 Goal 的事实主链

必须存在：

```text
.agent/goals/<GOAL-ID>/traceability.md
```

结构：

```md
| Req | AC | Design | Task | Issue | Commit | PR | Test | Evidence | Status |
|---|---|---|---|---|---|---|---|---|---|
| REQ-001 | AC-001 | DESIGN-1.1 | TASK-001 | #123 | abc123 | #130 | make test | EVID-001 | Done |
```

---

## RULE-TRACE-002：Traceability 缺链必须阻断 Release

以下缺任意一个，Release Gate 失败：

```text
Requirement
Acceptance Criteria
Task
Test
Evidence
Status
```

---

# 37. Issue 自动化详细规则

## RULE-ISSUE-AUTO-001：Issue 必须由 Task 生成，不允许手写漂移

推荐命令：

```bash
goalcli issues create --goal GOAL-20260603-001
```

生成逻辑：

```text
tasks.yaml
→ issue body
→ labels
→ milestone
→ assignee
→ acceptance checklist
→ evidence checklist
```

---

## RULE-ISSUE-AUTO-002：Issue 内容变更必须回写 Task Registry

如果 GitHub Issue 中发生：

```text
scope 变化
AC 变化
优先级变化
状态变化
```

必须同步：

```text
tasks.yaml
traceability.md
goals.yaml
```

---

# 38. Branch / Worktree 自动化规则

## RULE-WORKTREE-AUTO-001：不允许手工随意创建开发分支

必须通过：

```bash
goalcli worktree create \
  --goal GOAL-20260603-001 \
  --task TASK-GOAL-20260603-001-001 \
  --repo xlib-standard
```

---

## RULE-WORKTREE-AUTO-002：worktree 名称必须可追踪

格式：

```text
~/code/.worktrees/<repo>/<goal-id>/<task-id>
```

示例：

```text
~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-GOAL-20260603-001-001
```

---

## RULE-WORKTREE-AUTO-003：worktree 创建后必须运行 preflight

```bash
make worktree-check
make context-check
```

不通过则自动删除 worktree 或标记为 BLOCKED。

---

# 39. Commit 自动化规则

## RULE-COMMIT-AUTO-001：commit 必须由 goalcli 包装

推荐命令：

```bash
goalcli commit create \
  --task TASK-GOAL-20260603-001-001 \
  --evidence EVID-TASK-GOAL-20260603-001-001-20260603-001 \
  --type feat \
  --scope harness \
  --message "add worktree gate"
```

---

## RULE-COMMIT-AUTO-002：commit message 自动生成

格式：

```text
feat(harness): add worktree gate

Goal: GOAL-20260603-001
Task: TASK-GOAL-20260603-001-001
Issue: #123
Evidence: EVID-TASK-GOAL-20260603-001-001-20260603-001
```

---

## RULE-COMMIT-AUTO-003：commit 前必须执行本地 Gate

至少：

```bash
make worktree-check
make lint
make test
make evidence-check
```

如果是文档类任务：

```bash
make docs-check
make evidence-check
make traceability-check
```

---

# 40. PR 自动化规则

## RULE-PR-AUTO-001：PR 必须由 goalcli 生成或校验

推荐命令：

```bash
goalcli pr create --task TASK-GOAL-20260603-001-001
```

PR 必须自动填充：

```text
Goal ID
Task ID
Issue ID
Requirements Covered
AC
Evidence
Risk
Rollback
Checklist
```

---

## RULE-PR-AUTO-002：PR 更新必须同步 Evidence

当新 commit push 后，必须更新：

```text
PR Evidence Summary
Traceability Matrix
Evidence Registry
```

推荐命令：

```bash
goalcli pr update --pr 130
```

---

# 41. Merge 规则

## RULE-MERGE-001：禁止直接 merge 未通过 Gate 的 PR

必须满足：

```text
CI passed
Harness passed
review approved
evidence complete
traceability complete
release impact checked
no P0 risk open
```

---

## RULE-MERGE-002：Merge 策略必须统一

建议：

```text
Squash merge: 小任务 / 文档 / 简单修复
Merge commit: 多 commit 保留必要历史
Rebase merge: 禁止用于复杂 Goal，避免证据链丢失
```

对于 Goal Runtime，推荐：

```text
Squash merge + PR body 保留完整 Evidence Summary
```

---

# 42. Release 自动化规则

## RULE-RELEASE-AUTO-001：Release 必须从 Manifest 生成

推荐命令：

```bash
goalcli release prepare --goal GOAL-20260603-001
```

生成：

```text
release/REL-20260603-goal-runtime/manifest.md
release/REL-20260603-goal-runtime/changelog.md
release/REL-20260603-goal-runtime/evidence-summary.md
release/REL-20260603-goal-runtime/rollback.md
```

---

## RULE-RELEASE-AUTO-002：发布命令必须二次确认 Gate

```bash
goalcli release publish --release REL-20260603-goal-runtime
```

内部必须执行：

```bash
make ci
make evidence-check
make release-check
make secret-check
```

---

# 43. Review 规则

## RULE-REVIEW-001：Review 不是读代码，而是审事实链

Review 必须检查：

```text
Goal 是否满足
Requirement 是否覆盖
AC 是否验证
Task 是否完成
Evidence 是否可信
Risk 是否关闭
Release 是否可回滚
Retro 是否形成补丁
```

---

## RULE-REVIEW-002：Review 结论必须结构化

```yaml
review_id: REV-GOAL-20260603-001-20260603-001
target: GOAL-20260603-001
decision: APPROVED | NEEDS_FIX | NEEDS_REPLAN | REJECTED
blockers:
  - id:
    severity:
    description:
required_fixes:
  - task:
evidence_checked:
  - EVID-001
```

---

# 44. Self-improving 运行规则

## RULE-SI-001：Retro 必须生成可执行 Patch，不只是总结

每个 Retro 至少产出：

```text
1 个 Prompt Patch
1 个 Harness Patch 候选
1 个 Rule Patch 候选
1 个 CI Gate Suggestion
1 个 New Issue Candidate
```

---

## RULE-SI-002：Patch 必须分状态

```text
PROPOSED
ACCEPTED
REJECTED
SUPERSEDED
IMPLEMENTED
```

---

## RULE-SI-003：Patch 必须进入 Patch Registry

```text
.agent/registries/patches.yaml
```

示例：

```yaml
patches:
  - patch_id: PATCH-HARNESS-20260603-001
    source_goal: GOAL-20260603-001
    type: harness
    status: PROPOSED
    problem: PR allowed without traceability evidence
    proposed_gate: traceability-check
    target_files:
      - .agent/harness/gates/pr-gate.yaml
      - Makefile
```

---

# 45. Downstream Adoption 规则

这是 xlib-standard 作为标准工厂必须有的规则。

## RULE-DOWNSTREAM-001：标准必须能被下游库采用

下游库包括：

```text
kernel
configx
observex
testkitx
redisx
kafkax
postgresx
taosx
ossx
clickhousex
x.go
```

---

## RULE-DOWNSTREAM-002：每次 xlib-standard 规则更新必须说明下游影响

必须输出：

```text
affected downstream repos
required migration
breaking or non-breaking
minimum adoption steps
verification commands
```

---

## RULE-DOWNSTREAM-003：下游采用必须有 Adoption Manifest

```md
# Adoption Manifest

## Source Standard
- xlib-standard version:
- rule version:
- template version:

## Target Repo
- repo:
- current state:
- adoption scope:

## Adopted Components
- .agent/rules
- .agent/templates
- harness gates
- Makefile gates
- GitHub workflows

## Evidence
-

## Remaining Gaps
-
```

---

# 46. Drift Detection 规则

## RULE-DRIFT-001：必须检测文档、规则、代码、CI 漂移

典型漂移：

```text
README 写了 make docs-check，但 Makefile 没有
rules 写了 worktree-only，但 hooks 没启用
PR 模板要求 Evidence，但 CI 不检查
Release Manifest 要求 rollback，但 release 目录没有
```

---

## RULE-DRIFT-002：漂移必须阻断 Release

如果漂移影响 P0/P1，必须：

```text
标记 INCONSISTENT_STATE
生成 Drift Report
阻断 Release
创建修复 Issue
```

---

# 47. Deprecation 规则

## RULE-DEPRECATION-001：旧规则不能直接删除

必须经过：

```text
PROPOSED_DEPRECATION
MIGRATION_AVAILABLE
DOWNSTREAM_NOTIFIED
DEPRECATED
REMOVED
```

---

## RULE-DEPRECATION-002：规则废弃必须说明替代方案

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

# 48. Human Approval 规则

## RULE-HUMAN-001：高风险变更必须人工批准

以下必须进入：

```text
NEEDS_HUMAN_APPROVAL
```

场景：

```text
删除规则
弱化 P0 Gate
修改 release 流程
修改 secret 策略
修改 L0/L1/L2 分层规则
跨仓库批量变更
破坏兼容性
```

---

# 49. Failure Budget 规则

## RULE-FAILURE-001：Goal 必须有失败预算

建议字段：

```yaml
failure_budget:
  max_retries_per_task: 3
  max_failed_gates: 2
  max_days_blocked: 2
  require_replan_after_failed_attempts: 3
```

---

## RULE-FAILURE-002：超过失败预算必须 replan

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

# 50. 评分规则

## RULE-SCORE-001：每个 Goal 必须评分

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

## 评分等级

```text
90-100: 工厂级，可复用
80-89: 可执行，少量改进
70-79: 勉强可用，存在结构缺口
60-69: 高风险，不建议执行
<60: 不合格，必须重构
```

---

## 10 分标准

一个 10 分 Goal 必须满足：

```text
所有 P0 Gate 通过
所有 Requirement 有 AC
所有 Task 有 Evidence
所有 PR 有 Traceability
所有 Release 有 Manifest
所有风险有处理
所有重复问题转 Patch
worktree-only 全链路生效
```

---

# 51. 最小可执行规则集 MVA

第一阶段只落这 12 条即可形成闭环：

```text
1. RULE-CORE-001：没有 Evidence 不允许 DONE
2. RULE-CONTEXT-001：必须恢复真实项目状态
3. RULE-SPEC-003：AC 必须可验证
4. RULE-TASK-001：Task 是最小可执行单元
5. RULE-WORKTREE-001：禁止 main 开发
6. RULE-WORKTREE-002：所有开发必须用 git worktree
7. RULE-HARNESS-003：P0 Gate 失败必须阻断
8. RULE-EVIDENCE-003：Evidence 必须进入 Traceability
9. RULE-PR-003：PR 合并条件
10. RULE-RELEASE-001：Release 必须有 Manifest
11. RULE-RETRO-002：必须生成 Patch
12. RULE-SECURITY-001：禁止提交密钥
```

---

# 52. 最终推荐落地顺序

## 第 1 层：铁律

```text
CONSTITUTION.md
.agent/rules/01-core-rules.md
.agent/rules/07-worktree-rules.md
.agent/rules/11-evidence-rules.md
```

---

## 第 2 层：机器裁判

```text
.agent/harness/gates/
scripts/harness/
Makefile gates
GitHub Actions
```

---

## 第 3 层：自动化

```text
goalcli goal
goalcli worktree
goalcli issues
goalcli pr
goalcli evidence
goalcli release
goalcli retro
```

---

## 第 4 层：复利系统

```text
retrospective
patch registry
rule patch
harness patch
prompt patch
downstream adoption
drift detection
```

---

# 53. 最终规则总纲

最终 Goal 规则可以压缩成一条：

> **任何 Goal 都必须在真实上下文中生成可验证 Spec，经 Design 和 Task 拆解后，在独立 worktree 中执行，通过 Harness Gate 验证，用 Evidence 证明完成，经 PR 和 Release Manifest 交付，最后通过 Retrospective 生成 Prompt / Harness / Rule Patch，使下一轮执行自动变强。**

这就是完整 Goal 规则体系的核心闭环。
继续补齐 **Goal 完整规则 v1.2：规则即代码、Agent 执行协议、自动化门禁、并发隔离、审计与度量**。

上一版已经定义了规则体系。
这一版重点解决：

```text
规则如何真正变成机器可执行？
Agent 如何不乱跑？
多个 worktree / 多个 Issue 如何并发不冲突？
如何防止文档膨胀、规则漂移、Evidence 作假？
如何衡量 Goal Runtime 是否越来越强？
```

---

# 54. Rule as Code 规则

## RULE-CODE-001：所有 P0 / P1 规则必须机器化

Markdown 规则只能解释，不能作为唯一裁判。

每条 P0 / P1 规则必须至少落到以下一种机器形式：

```text
YAML Policy
Harness Gate
Makefile Target
GitHub Action
goalcli Command
Shell Script
JSON Schema
CI Check
```

例如：

```text
“禁止 main 开发”
不能只写在文档里
必须落到：
- scripts/harness/no-main-dev.sh
- make worktree-check
- .githooks/pre-commit
- .githooks/pre-push
- GitHub branch protection
- GitHub Actions guard
```

---

## RULE-CODE-002：规则必须有机器 ID

每条规则必须具备：

```yaml
id: RULE-WORKTREE-001
title: No main development
severity: P0
domain: worktree
enforced_by:
  - make worktree-check
  - scripts/harness/no-main-dev.sh
  - .githooks/pre-commit
  - .github/workflows/worktree-guard.yml
evidence:
  - reports/worktree-check.txt
```

---

# 55. Schema-first 规则

## RULE-SCHEMA-001：Goal 核心对象必须有 Schema

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

---

## RULE-SCHEMA-002：没有通过 Schema 校验的对象不得进入下一阶段

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

# 56. Goal Pack 规则

## RULE-GOALPACK-001：每个 Goal 必须形成 Goal Pack

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

---

## RULE-GOALPACK-002：Goal Pack 必须可离线审计

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

# 57. Agent 执行协议

## RULE-AGENT-001：Agent 不允许自由发挥执行

Agent 必须遵循固定执行协议：

```text
Read Goal
→ Recover Context
→ Validate Spec
→ Validate Design
→ Validate Plan
→ Create Worktree
→ Execute Task
→ Run Gates
→ Collect Evidence
→ Update Traceability
→ Create PR
→ Wait for Review/Gates
→ Release
→ Retrospective
```

---

## RULE-AGENT-002：Agent 每一步必须写 Execution Log

必须记录：

```text
执行时间
执行目录
当前分支
当前 worktree
执行命令
命令结果
变更文件
失败原因
修复动作
Evidence ID
```

推荐文件：

```text
.agent/goals/<GOAL-ID>/execution-log.md
```

---

## RULE-AGENT-003：Agent 不能跳过失败

如果命令失败，Agent 必须：

```text
记录失败
分析根因
判断是否可自动修复
若可修复，最多重试 N 次
若不可修复，进入 BLOCKED / NEEDS_RESEARCH / NEEDS_DECISION
```

禁止：

```text
忽略失败继续执行
只说“应该没问题”
伪造 Evidence
跳过测试
```

---

# 58. Agent 权限边界规则

## RULE-AGENT-AUTH-001：Agent 只能操作当前 Goal 授权范围

每个 Goal 必须定义：

```yaml
allowed_repos:
  - xlib-standard
allowed_paths:
  - .agent/
  - scripts/harness/
  - Makefile
forbidden_paths:
  - .git/
  - secrets/
  - production/
allowed_commands:
  - make ci
  - make evidence-check
  - goalcli evidence collect
forbidden_commands:
  - rm -rf /
  - git push origin main
  - export SECRET=...
```

---

## RULE-AGENT-AUTH-002：越权必须阻断

如果 Agent 修改非授权路径，必须：

```text
标记 POLICY_VIOLATION
撤销变更
生成 Risk
进入 NEEDS_HUMAN_APPROVAL
```

---

# 59. 并发执行规则

## RULE-CONCURRENCY-001：每个 Task 独立 worktree

并发执行时：

```text
一个 Task = 一个 worktree = 一个 branch = 一个 PR
```

禁止多个 Agent 在同一 worktree 并发开发。

---

## RULE-CONCURRENCY-002：必须有 Lock 文件

推荐：

```text
.agent/locks/
├── GOAL-20260603-001.lock
├── TASK-GOAL-20260603-001-001.lock
└── release.lock
```

Lock 内容：

```yaml
lock_id: LOCK-TASK-GOAL-20260603-001-001
owner: agent-01
goal_id: GOAL-20260603-001
task_id: TASK-GOAL-20260603-001-001
worktree: /home/zone/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001
created_at: 2026-06-03T16:30:00+09:00
expires_at: 2026-06-03T20:30:00+09:00
```

---

## RULE-CONCURRENCY-003：Release 必须串行

允许并发：

```text
Task
Issue
PR
Evidence collection
```

必须串行：

```text
Version bump
Release Manifest
Tag
Publish
Downstream adoption baseline update
```

---

# 60. Branch 命名规则

## RULE-BRANCH-001：分支必须可追踪

允许格式：

```text
goal/<GOAL-ID>/<TASK-ID>
issue/<ISSUE-ID>
fix/<ISSUE-ID>-<summary>
feat/<GOAL-ID>-<summary>
chore/<TASK-ID>-<summary>
```

推荐：

```text
goal/GOAL-20260603-001/TASK-001
```

---

## RULE-BRANCH-002：禁止无来源分支

禁止：

```text
test
tmp
dev
new
fix
wip
local
```

---

# 61. 文件变更规则

## RULE-FILE-001：Task 必须声明预计变更文件

Task 中必须包含：

```yaml
files_to_change:
  - .agent/rules/07-worktree-rules.md
  - scripts/harness/no-main-dev.sh
  - Makefile
```

---

## RULE-FILE-002：实际变更超出范围必须解释

如果实际变更文件不在 `files_to_change` 中，必须：

```text
更新 Task
更新 Traceability
更新 Risk
必要时生成 Decision Log
```

否则 PR Gate 失败。

---

# 62. Evidence 反作弊规则

## RULE-EVIDENCE-ANTI-FAKE-001：Evidence 不能只写结果，必须包含原始命令

有效 Evidence 必须包含：

```text
command
cwd
branch
commit
timestamp
exit_code
stdout/stderr 摘要
artifact path
```

---

## RULE-EVIDENCE-ANTI-FAKE-002：Evidence 必须绑定 commit

没有 commit hash 的 Evidence 只能作为临时证据，不能用于 Release。

```yaml
evidence_id: EVID-TASK-001-20260603-001
commit: abc123def
command: make ci
exit_code: 0
status: passed
```

---

## RULE-EVIDENCE-ANTI-FAKE-003：Release Evidence 必须可复跑

Release 前的核心 Evidence 必须能通过命令复跑：

```bash
make ci
make evidence-check
make release-check
```

---

# 63. 文档膨胀控制规则

## RULE-DOC-001：禁止无边界文档膨胀

任何新增文档必须回答：

```text
它服务哪个 Goal？
它验证哪个 Requirement？
它替代还是补充已有文档？
它是否需要进入导航？
它是否需要 Harness 检查？
```

---

## RULE-DOC-002：文档必须分类

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

---

## RULE-DOC-003：长期规则文档必须有 SSOT

如果多个文档描述同一规则，必须指定唯一事实源：

```text
SSOT: .agent/rules/07-worktree-rules.md
```

其他文档只能引用，不允许复制漂移。

---

# 64. 上下文压缩规则

## RULE-CONTEXT-COMPRESSION-001：Goal 必须支持上下文压缩

大型 Goal 必须提供：

```text
context-summary.md
decision-summary.md
evidence-summary.md
current-state.md
next-actions.md
```

用于 Agent 在上下文不足时恢复。

---

## RULE-CONTEXT-COMPRESSION-002：上下文摘要不能替代原始 Evidence

摘要只能帮助阅读，不能作为完成证明。

---

# 65. CI Gate 编排规则

## RULE-CI-001：CI 必须分层

推荐 CI 分层：

```text
L0 Fast Gate:
- format
- lint
- worktree-check
- schema-check

L1 Verification Gate:
- test
- evidence-check
- traceability-check

L2 Release Gate:
- release-check
- changelog-check
- secret-check
- manifest-check

L3 Downstream Gate:
- adoption-check
- compatibility-check
```

---

## RULE-CI-002：Fast Gate 必须足够快

L0 Gate 应该控制在短时间内完成，避免 Agent 频繁等待。

```text
L0 失败：立即修
L1 失败：分析任务实现
L2 失败：禁止发布
L3 失败：禁止下游推广
```

---

# 66. 本地与远端门禁一致性规则

## RULE-GATE-CONSISTENCY-001：本地 Gate 与 CI Gate 不能冲突

例如：

```text
本地 make evidence-check 通过
CI evidence-check 失败
```

必须进入 Drift Detection。

---

## RULE-GATE-CONSISTENCY-002：Makefile 是统一入口

Agent、开发者、CI 都必须使用 Makefile / goalcli 入口，不允许各跑各的命令。

---

# 67. GitHub Actions 必备规则

## RULE-GHA-001：必须有 Goal Gates Workflow

推荐：

```text
.github/workflows/goal-gates.yml
.github/workflows/worktree-guard.yml
.github/workflows/release-gate.yml
.github/workflows/secret-scan.yml
```

---

## RULE-GHA-002：PR 必须跑 Goal Gates

PR 上至少执行：

```bash
make ci
make evidence-check
make traceability-check
make pr-check
```

---

# 68. Issue 生命周期规则

## RULE-ISSUE-LIFECYCLE-001：Issue 状态必须有限状态机

```text
OPEN
→ READY
→ IN_PROGRESS
→ IN_REVIEW
→ BLOCKED
→ DONE
→ CLOSED
```

异常：

```text
NEEDS_RESEARCH
NEEDS_DECISION
NEEDS_REPLAN
DUPLICATE
WONT_DO
```

---

## RULE-ISSUE-LIFECYCLE-002：Issue 进入 DONE 前必须满足 DoD

```text
Task 完成
测试通过
Evidence 归档
Traceability 更新
PR 合并
```

---

# 69. PR 生命周期规则

## RULE-PR-LIFECYCLE-001：PR 状态必须结构化

```text
DRAFT
READY_FOR_REVIEW
CHANGES_REQUESTED
APPROVED
MERGED
CLOSED
```

---

## RULE-PR-LIFECYCLE-002：Draft PR 可以没有完整 Evidence，Ready PR 不可以

Draft 阶段允许：

```text
Evidence pending
Tests pending
Risk pending
```

Ready 阶段必须：

```text
Evidence complete
Tests passed
Traceability complete
Risk updated
Rollback documented
```

---

# 70. Release Channel 规则

## RULE-RELEASE-CHANNEL-001：发布必须区分通道

```text
dev
alpha
beta
rc
stable
```

---

## RULE-RELEASE-CHANNEL-002：stable 必须最高门禁

stable 发布必须满足：

```text
所有 P0/P1 Gate 通过
Release Manifest 完整
Rollback 完整
Known Issues 明确
Downstream impact 明确
Retrospective 已创建
```

---

# 71. 版本规则

## RULE-VERSION-001：版本更新必须与变更类型一致

```text
PATCH: bugfix / docs / non-breaking gate
MINOR: new feature / new gate / new template
MAJOR: breaking rule / breaking API / breaking workflow
```

---

## RULE-VERSION-002：规则版本必须独立记录

建议：

```text
project version: v0.3.0
rule version: rules-v1.2
harness version: harness-v0.1
template version: templates-v0.2
```

---

# 72. Rollback 深化规则

## RULE-ROLLBACK-002：Rollback 必须可执行

不能只写“回滚到上一版”。

必须写：

```bash
git revert <commit>
git tag -d <tag>
git push origin :refs/tags/<tag>
goalcli release rollback --release REL-xxx
```

---

## RULE-ROLLBACK-003：Rollback 也必须有 Evidence

回滚完成后必须生成：

```text
EVID-ROLLBACK-YYYYMMDD-NNN
```

---

# 73. 规则变更协议

## RULE-CHANGE-001：修改规则必须走 Rule Change Protocol

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

---

## RULE-CHANGE-002：P0 规则变更必须人工批准

任何弱化 P0 的行为都必须进入：

```text
NEEDS_HUMAN_APPROVAL
```

---

# 74. 模板同步规则

## RULE-TEMPLATE-001：规则变更必须同步模板

例如新增 Evidence 字段后必须同步：

```text
issue-template.md
pr-template.md
evidence-template.md
release-manifest-template.md
retrospective-template.md
```

---

## RULE-TEMPLATE-002：模板必须有版本号

```yaml
template_id: TEMPLATE-PR-v1.2
version: v1.2
compatible_rules:
  - RULE-EVIDENCE-001
  - RULE-TRACE-001
```

---

# 75. Harness 自测试规则

## RULE-HARNESS-TEST-001：Harness 本身必须有测试

不能只测试业务代码，也要测试规则系统。

必须验证：

```text
main 分支开发会失败
缺 Evidence 会失败
缺 Traceability 会失败
缺 Release Manifest 会失败
合法 Goal 会通过
```

---

## RULE-HARNESS-TEST-002：必须有违规样例

推荐：

```text
tests/fixtures/violations/
├── missing-evidence/
├── main-branch-dev/
├── missing-traceability/
├── missing-release-manifest/
└── secret-leak/
```

---

# 76. Golden Path 规则

## RULE-GOLDEN-001：必须维护一个最小成功样例

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

---

## RULE-GOLDEN-002：每次规则升级都必须跑 Golden Test

防止规则升级把正常流程误杀。

---

# 77. 度量指标规则

## RULE-METRIC-001：Goal Runtime 必须度量

核心指标：

```text
Goal completion rate
Gate pass rate
Evidence coverage
Traceability coverage
PR cycle time
Rework rate
Rollback rate
Rule drift count
Repeated issue count
Self-improving patch adoption rate
Downstream adoption rate
```

---

## RULE-METRIC-002：低于阈值必须触发治理

例如：

```text
Evidence coverage < 95% → 阻断 Release
Traceability coverage < 100% → 阻断 Release
Repeated issue count > 2 → 必须新增 Rule/Harness
Gate false positive > 10% → 需要调参
```

---

# 78. 审计规则

## RULE-AUDIT-001：Goal 必须可审计

审计者应能回答：

```text
这个 Goal 为什么做？
谁批准？
改了什么？
如何验证？
证据在哪里？
是否影响下游？
怎么回滚？
复盘产生了什么改进？
```

---

## RULE-AUDIT-002：审计报告必须可生成

推荐命令：

```bash
goalcli audit goal --goal GOAL-20260603-001
```

生成：

```text
reports/audit/GOAL-20260603-001.md
```

---

# 79. Downstream Sync 规则

## RULE-DOWNSTREAM-SYNC-001：xlib-standard 更新后必须生成同步任务

当以下内容变化：

```text
rules
templates
harness
schemas
Makefile gates
GitHub workflows
```

必须为下游生成 adoption issue：

```text
kernel
configx
observex
testkitx
redisx
kafkax
postgresx
taosx
ossx
clickhousex
x.go
```

---

## RULE-DOWNSTREAM-SYNC-002：下游不同步必须记录原因

```text
not applicable
blocked
requires migration
breaking change
waiting for release
```

---

# 80. 最终补强后的 Goal Runtime 闭环

完整闭环应该变成：

```text
Goal Pack
→ Schema Validate
→ Context Recovery
→ Spec Gate
→ Design Gate
→ Task Gate
→ Issue Auto Create
→ Worktree Auto Create
→ Agent Execute
→ Local Gates
→ Evidence Collect
→ Commit with Evidence
→ PR Auto Create
→ CI Gates
→ Review
→ Merge
→ Release Manifest
→ Release Gate
→ Publish
→ Retrospective
→ Prompt/Harness/Rule Patch
→ Metrics
→ Audit
→ Downstream Adoption
→ Drift Detection
→ Next Goal
```

---

# 81. 最小落地优先级

你现在最应该先实现的不是完整系统，而是 **最小不可绕过闭环**：

```text
P0-1: worktree-only
P0-2: evidence-check
P0-3: traceability-check
P0-4: pr-check
P0-5: release-manifest-check
P0-6: secret-check
P0-7: retrospective-check
```

对应 Makefile：

```makefile
.PHONY: ci
ci:
	$(MAKE) worktree-check
	$(MAKE) schema-check
	$(MAKE) traceability-check
	$(MAKE) evidence-check
	$(MAKE) secret-check
	$(MAKE) test

.PHONY: release-check
release-check:
	$(MAKE) ci
	$(MAKE) manifest-check
	$(MAKE) retrospective-check
```

---

# 82. 最终铁律 v1.2

最终补强后的铁律是：

```text
1. 规则必须机器化，否则只是建议。
2. P0/P1 规则必须有 Harness Gate。
3. Goal 必须形成可离线审计的 Goal Pack。
4. Agent 只能按执行协议运行，不能自由发挥。
5. 每个 Task 必须独立 worktree、独立 Evidence、独立 PR。
6. Evidence 必须绑定命令、commit、输出和 artifact。
7. Traceability 是事实主链，缺链禁止 Release。
8. Release 必须串行、可回滚、可审计。
9. Retrospective 必须产出可执行 Patch。
10. 规则升级必须同步 Harness、模板、CI、下游采用说明。
11. Harness 本身必须测试，不能只测试业务代码。
12. Metrics 决定系统是否真的在复利，而不是文档变多。
```

最终目标：

> **把 Goal 从“写计划”升级为一个可机器裁判、可证据证明、可并发执行、可审计发布、可持续自我增强的工程操作系统。**

继续补齐 **Goal 完整规则 v1.3：goalcli 命令契约、自动化执行协议、异常豁免、Agent Team、审计交付与 MVA Issue 拆解**。

这一版重点把规则进一步压成 **可实现的工程接口**。

---

# 83. goalcli 命令契约规则

## RULE-GOALCLI-001：goalcli 必须是 Goal Runtime 的唯一执行入口之一

所有自动化动作必须优先通过：

```bash
goalcli <domain> <action>
```

而不是每个 Agent 自己拼命令。

核心原因：

```text
统一上下文
统一日志
统一 Evidence
统一 Traceability
统一错误处理
统一 Gate 行为
```

---

## RULE-GOALCLI-002：goalcli 命令必须幂等

重复执行不应破坏状态。

例如：

```bash
goalcli issues create --goal GOAL-20260603-001
```

第二次执行时必须：

```text
已存在 Issue → 更新或跳过
不存在 Issue → 创建
状态冲突 → 报告 INCONSISTENT_STATE
```

禁止：

```text
重复创建一批相同 Issue
覆盖人工更新内容
丢失 Issue 与 Task 映射
```

---

## RULE-GOALCLI-003：goalcli 命令必须统一输出机器结果

所有命令输出至少包含：

```json
{
  "command": "goalcli evidence check",
  "status": "passed",
  "goal_id": "GOAL-20260603-001",
  "reports": ["reports/evidence-check.json"],
  "errors": [],
  "warnings": []
}
```

并写入：

```text
reports/<command>.json
.agent/goals/<GOAL-ID>/execution-log.md
```

---

# 84. goalcli Exit Code 规则

## RULE-GOALCLI-EXIT-001：退出码必须标准化

```text
0  = PASS
1  = GENERAL_FAILURE
2  = POLICY_VIOLATION
3  = SCHEMA_INVALID
4  = EVIDENCE_MISSING
5  = TRACEABILITY_BROKEN
6  = WORKTREE_INVALID
7  = SECRET_DETECTED
8  = RELEASE_BLOCKED
9  = NEEDS_HUMAN_APPROVAL
10 = INCONSISTENT_STATE
```

CI、Agent、脚本必须基于退出码做决策。

例如：

```text
exit 4 → 自动进入 Evidence 修复流程
exit 6 → 阻断 commit / PR
exit 9 → 进入人工审批
exit 10 → 进入状态修复，不允许继续执行
```

---

# 85. 报告产物规则

## RULE-REPORT-001：每个 Gate 必须生成报告

推荐结构：

```text
reports/
├── context-check.json
├── schema-check.json
├── worktree-check.txt
├── spec-check.json
├── design-check.json
├── task-check.json
├── traceability-check.json
├── evidence-check.json
├── pr-check.json
├── release-check.json
├── secret-check.json
├── retro-check.json
└── ci-summary.md
```

---

## RULE-REPORT-002：报告必须进入 Evidence

凡是用于证明完成的报告，都必须被 Evidence 引用。

```yaml
evidence_id: EVID-TASK-001-20260603-001
artifacts:
  - reports/traceability-check.json
  - reports/evidence-check.json
  - reports/ci-summary.md
```

---

# 86. GitHub Issue 自动化协议

## RULE-GITHUB-ISSUE-001：Issue 必须从 Task Registry 生成

输入：

```text
.agent/goals/<GOAL-ID>/tasks.yaml
```

输出：

```text
GitHub Issues
.agent/goals/<GOAL-ID>/issues.yaml
.agent/registries/tasks.yaml
.agent/goals/<GOAL-ID>/traceability.md
```

---

## RULE-GITHUB-ISSUE-002：Issue Body 必须包含固定区块

```md
## Goal
## Task
## Requirement
## Acceptance Criteria
## Scope
## Files to Change
## Commands to Run
## Evidence Required
## Worktree Requirement
## Risk
## Rollback
## Definition of Done
```

缺任一区块，Issue Gate 失败。

---

## RULE-GITHUB-ISSUE-003：Issue 不能成为新事实源

Issue 是 Task 的执行镜像，不是需求 SSOT。

事实源优先级：

```text
Spec / Task Registry
→ Issue
→ PR
→ Commit
→ Release Notes
```

如果 Issue 被人工修改，必须回写 Task Registry。

---

# 87. GitHub Label 规则

## RULE-LABEL-001：Label 必须结构化

推荐 label 体系：

```text
type:goal
type:spec
type:design
type:task
type:harness
type:evidence
type:release
type:self-improving

priority:P0
priority:P1
priority:P2
priority:P3

state:ready
state:blocked
state:needs-research
state:needs-decision
state:needs-replan
state:needs-review

risk:security
risk:breaking
risk:compatibility
risk:ci
risk:release
```

---

## RULE-LABEL-002：P0 Issue 必须被显式标记

P0 Issue 必须包含：

```text
priority:P0
```

且不得被自动关闭，除非 Evidence 完整。

---

# 88. Milestone 规则

## RULE-MILESTONE-001：Goal 必须映射到 Milestone

推荐：

```text
MILESTONE-GOAL-20260603-001
```

每个 Goal 下的 Issue 必须绑定同一 Milestone。

---

## RULE-MILESTONE-002：Milestone 关闭条件

Milestone 关闭前必须满足：

```text
所有 P0/P1 Issue closed
所有 PR merged
Release Manifest generated
Retrospective generated
Patch candidates recorded
```

---

# 89. Commit Evidence Binding 规则

## RULE-COMMIT-EVID-001：最终 commit 必须绑定 Evidence

最终可合并 commit message 必须包含：

```text
Goal:
Task:
Issue:
Evidence:
```

示例：

```text
feat(harness): add worktree-only gate

Goal: GOAL-20260603-001
Task: TASK-GOAL-20260603-001-003
Issue: #123
Evidence: EVID-TASK-GOAL-20260603-001-003-20260603-001
```

---

## RULE-COMMIT-EVID-002：临时 commit 必须在 PR 前整理

允许开发中存在：

```text
wip
fixup
debug
```

但 PR Ready 前必须 squash / rewrite 成合规 commit。

---

# 90. PR 同步协议

## RULE-PR-SYNC-001：PR Body 必须由 goalcli 可重复生成

PR Body 不应完全手写。

推荐命令：

```bash
goalcli pr render --task TASK-GOAL-20260603-001-003
goalcli pr update --pr 123
```

---

## RULE-PR-SYNC-002：PR 更新必须同步四个对象

```text
PR Body
Traceability Matrix
Evidence Registry
Task Registry
```

任何一个不同步，PR Gate 失败。

---

# 91. PR Comment Bot 规则

## RULE-PR-BOT-001：Harness 必须在 PR 中评论结果摘要

PR 每次 Gate 执行后应评论：

```md
## Harness Summary

| Gate | Status | Report |
|---|---|---|
| worktree-check | PASS | reports/worktree-check.txt |
| evidence-check | PASS | reports/evidence-check.json |
| traceability-check | PASS | reports/traceability-check.json |
| release-impact-check | PASS | reports/release-impact-check.json |

Decision: READY_FOR_REVIEW
```

---

## RULE-PR-BOT-002：失败必须给出可执行修复建议

不能只说失败，必须指出：

```text
失败规则
失败文件
失败原因
修复命令
所需 Evidence
```

---

# 92. Review Bot 与人工 Review 边界

## RULE-REVIEW-BOT-001：机器 Review 负责事实链

机器检查：

```text
Schema
Traceability
Evidence
CI
Secrets
Branch
Worktree
Release impact
```

人工 Review 检查：

```text
架构合理性
需求理解
取舍是否正确
长期维护性
是否值得进入标准
```

---

## RULE-REVIEW-BOT-002：机器通过不等于人工通过

对于 P0/P1、架构、规则、发布相关变更，仍需人工 Review。

---

# 93. Worktree 清理协议

## RULE-WORKTREE-CLEAN-001：PR 合并后必须清理 worktree

清理命令：

```bash
goalcli worktree clean --task TASK-GOAL-20260603-001-003
```

内部执行：

```bash
git worktree remove <path>
git worktree prune
git fetch --prune
```

---

## RULE-WORKTREE-CLEAN-002：未合并 worktree 不允许直接删除

如果 worktree 仍有未提交变更：

```text
必须生成 abandoned report
必须记录原因
必须确认没有 Evidence 丢失
必须更新 Task 状态
```

---

# 94. 主线同步协议

## RULE-MAIN-SYNC-001：main 只能做同步与发布基线

允许：

```bash
git fetch origin
git pull --ff-only
git tag
git status
goalcli release prepare
```

禁止：

```bash
git commit
git push origin main
编辑业务文件
运行实现型 Agent
```

---

## RULE-MAIN-SYNC-002：每个 worktree 创建前必须基于最新 main

```bash
git fetch origin
git worktree add <path> -b <branch> origin/main
```

如果 main 落后，必须先同步。

---

# 95. Release Artifact 规则

## RULE-REL-ARTIFACT-001：Release 必须生成 Artifact 包

推荐结构：

```text
release/REL-20260603-goal-runtime/
├── manifest.md
├── changelog.md
├── evidence-summary.md
├── test-summary.md
├── risk-summary.md
├── rollback.md
├── known-issues.md
├── adoption-impact.md
└── checksums.txt
```

---

## RULE-REL-ARTIFACT-002：Release Artifact 必须不可变

发布后不得直接修改 release artifact。

若发现错误：

```text
生成新 patch release
或生成 correction note
不得静默覆盖
```

---

# 96. Changelog 规则

## RULE-CHANGELOG-001：Changelog 必须从 PR / Release Manifest 生成

禁止手写遗漏。

结构：

```md
## Added
## Changed
## Fixed
## Deprecated
## Removed
## Security
## Migration
## Evidence
```

---

## RULE-CHANGELOG-002：规则变更必须显式标记 Breaking / Non-breaking

例如：

```text
Breaking: P0 worktree-only enforcement is now mandatory.
Non-breaking: Added optional retrospective template.
```

---

# 97. Exception / Waiver 规则

## RULE-WAIVER-001：任何绕过规则都必须写 Waiver

Waiver 不是随便放行，而是受控异常。

结构：

```yaml
waiver_id: WAIVER-20260603-001
rule_id: RULE-DESIGN-002
reason: "ADR temporarily deferred for urgent P1 fix"
scope: "TASK-GOAL-20260603-001-004"
expires_at: "2026-06-10"
approved_by: "human"
risk: "architecture decision may be undocumented"
follow_up_issue: "#155"
```

---

## RULE-WAIVER-002：P0 规则默认不可豁免

以下规则不可豁免：

```text
禁止提交密钥
禁止 main 直接开发
禁止无 Evidence DONE
禁止无 Release Manifest 发布
禁止无 Traceability 合并
```

如确需例外，必须进入：

```text
NEEDS_HUMAN_APPROVAL
```

并生成安全审计记录。

---

# 98. Policy Violation 处理规则

## RULE-VIOLATION-001：违规必须进入统一处理流程

流程：

```text
Detect Violation
→ Stop Execution
→ Record Violation
→ Generate Evidence
→ Open Fix Issue
→ Apply Fix
→ Re-run Gate
→ Update Retrospective
```

---

## RULE-VIOLATION-002：违规记录必须保留

推荐目录：

```text
.agent/violations/
├── VIOLATION-20260603-001.md
└── VIOLATION-20260603-002.md
```

---

# 99. AutoResearch 深化规则

## RULE-AR-001：AutoResearch 必须区分事实与假设

研究输出必须包含：

```md
## Facts
## Assumptions
## Unknowns
## Options
## Risks
## Recommendation
## Decision Needed
## Evidence
```

---

## RULE-AR-002：研究结论必须进入 Decision Log

任何会影响设计、依赖、API、CI、发布的研究结论，必须落成：

```text
DEC-YYYYMMDD-NNN
```

---

# 100. Agent Team 规则

## RULE-AGENT-TEAM-001：复杂 Goal 必须拆 Agent 角色

推荐角色：

```text
Context Agent      # 恢复仓库事实
Spec Agent         # 生成需求与 AC
Design Agent       # 设计与 ADR
Task Agent         # 拆 Task / Issue
Implement Agent    # 实现
Test Agent         # 测试
Evidence Agent     # 收集证据
Review Agent       # 复查事实链
Release Agent      # 发布
Retro Agent        # 复盘与补丁
```

---

## RULE-AGENT-TEAM-002：角色之间必须通过文件交接

禁止靠聊天记忆交接。

交接文件：

```text
context.md
spec.yaml
design.md
tasks.yaml
traceability.md
evidence/
review.md
release-manifest.md
retrospective.md
```

---

# 101. Agent Handoff 规则

## RULE-HANDOFF-001：每次 Agent 交接必须写 Handoff Note

```md
# Handoff Note

## From
Context Agent

## To
Design Agent

## Current State
CONTEXT_READY

## Completed
-

## Open Risks
-

## Required Next Actions
-

## Evidence
-
```

---

## RULE-HANDOFF-002：没有 Handoff 不允许切换 Agent

否则容易出现：

```text
上下文丢失
重复实现
Evidence 断链
PR 内容漂移
```

---

# 102. 依赖升级规则

## RULE-DEPENDENCY-001：依赖升级必须有独立 Task

禁止在业务功能 PR 中顺手升级大量依赖。

---

## RULE-DEPENDENCY-002：依赖升级必须包含

```text
current version
target version
breaking changes
security notes
test result
rollback plan
```

---

# 103. Supply Chain 规则

## RULE-SUPPLY-001：Release 必须支持供应链审计

对于基础库标准工厂，Release 建议包含：

```text
dependency list
checksums
generated artifacts
source commit
tag
CI run
```

---

## RULE-SUPPLY-002：禁止未知来源二进制进入 Release

任何二进制 artifact 必须能追踪到源码和构建命令。

---

# 104. Secret Scan 规则

## RULE-SECRET-001：secret-check 是 P0 Gate

必须检查：

```text
.env
token
password
secret
private key
access key
authorization
cookie
```

---

## RULE-SECRET-002：发现 Secret 必须立即阻断

处理流程：

```text
stop
remove secret
rotate credential if leaked
rewrite history if needed
record violation
add regression gate
```

---

# 105. 结构债规则

## RULE-DEBT-001：结构债必须可检测

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

---

## RULE-DEBT-002：结构债必须进入 Debt Registry

```yaml
debt_id: DEBT-20260603-001
type: layering_violation
severity: P1
location: internal/foo/bar.go
detected_by: make structure-check
fix_issue: "#166"
```

---

# 106. Standard Factory 规则

## RULE-FACTORY-001：xlib-standard 的最终输出不是仓库，而是工厂能力

xlib-standard 必须能够生产或约束：

```text
kernel
configx
observex
testkitx
redisx
kafkax
postgresx
taosx
ossx
clickhousex
x.go
```

---

## RULE-FACTORY-002：标准输出必须具备可迁移性

任何新增规则、模板、Harness，都要回答：

```text
是否能被下游复用？
下游最小接入成本是什么？
是否破坏现有库？
如何验证下游采用成功？
```

---

# 107. Adoption Gate 规则

## RULE-ADOPTION-GATE-001：下游采用必须有 Gate

推荐命令：

```bash
make adoption-check
```

检查：

```text
.agent/rules 存在
.agent/templates 存在
Makefile gates 存在
GitHub workflows 存在
Evidence Protocol 生效
worktree-only 生效
```

---

## RULE-ADOPTION-GATE-002：下游采用必须有证据

```text
reports/adoption-check.json
.agent/adoption/adoption-manifest.md
```

---

# 108. Goal Runtime 版本兼容规则

## RULE-RUNTIME-COMPAT-001：Goal Runtime 升级必须声明兼容范围

```yaml
runtime_version: goal-runtime-v1.3
compatible_with:
  - rules-v1.2
  - harness-v0.1
  - templates-v0.2
breaking_changes:
  - worktree-only enforcement changed from warning to blocking
```

---

## RULE-RUNTIME-COMPAT-002：不兼容升级必须提供迁移脚本

例如：

```bash
goalcli migrate rules --from rules-v1.2 --to rules-v1.3
```

---

# 109. goalcli v0.1.0 最小功能边界

## 必须实现

```text
goal init
schema validate
worktree check
evidence check
traceability check
pr check
release check
retro check
audit goal
```

---

## 暂缓实现

```text
自动写代码
自动发布 stable
跨仓库批量改造
复杂 Agent 编排
自动依赖升级
```

原因：

```text
v0.1.0 先做裁判系统，不先做全自动执行器。
先确保不会做错，再扩大自动化范围。
```

---

# 110. goalcli v0.1.0 推荐 Issue 拆解

## ISSUE-001：建立 Schema 基础

```text
目标：定义 Goal / Task / Evidence / Release / Patch schema
验收：make schema-check 通过
证据：reports/schema-check.json
```

---

## ISSUE-002：实现 worktree-check

```text
目标：禁止 main 开发
验收：main 分支失败，合法 worktree 通过
证据：reports/worktree-check.txt
```

---

## ISSUE-003：实现 evidence-check

```text
目标：检查 Evidence 是否绑定 Task / Command / Commit / Artifact
验收：缺 Evidence 失败，完整 Evidence 通过
证据：reports/evidence-check.json
```

---

## ISSUE-004：实现 traceability-check

```text
目标：检查 Req → AC → Task → Evidence 链路
验收：缺链失败，完整链通过
证据：reports/traceability-check.json
```

---

## ISSUE-005：实现 pr-check

```text
目标：检查 PR 模板、Evidence、Traceability、Risk、Rollback
验收：PR Ready 前必须通过
证据：reports/pr-check.json
```

---

## ISSUE-006：实现 release-check

```text
目标：检查 Release Manifest、Evidence Summary、Rollback、Known Issues
验收：缺任意 P0 项失败
证据：reports/release-check.json
```

---

## ISSUE-007：实现 retro-check

```text
目标：检查 Retrospective 是否产生 Patch 候选
验收：无 Prompt/Harness/Rule Patch 候选则失败
证据：reports/retro-check.json
```

---

## ISSUE-008：实现 audit goal

```text
目标：生成 Goal 审计报告
验收：可以从 Goal Pack 重建事实链
证据：reports/audit/<GOAL-ID>.md
```

---

# 111. 最终验收矩阵

| 模块           | 必须产物                      | Gate               | Evidence            |
| ------------ | ------------------------- | ------------------ | ------------------- |
| Goal Pack    | `.agent/goals/<GOAL-ID>/` | schema-check       | schema report       |
| Worktree     | 独立 worktree               | worktree-check     | worktree report     |
| Evidence     | evidence registry         | evidence-check     | evidence report     |
| Traceability | traceability.md           | traceability-check | traceability report |
| PR           | PR body                   | pr-check           | PR report           |
| Release      | manifest                  | release-check      | release report      |
| Retro        | retrospective.md          | retro-check        | retro report        |
| Audit        | audit report              | audit goal         | audit report        |

---

# 112. 最终执行原则 v1.3

压缩成 10 条：

```text
1. goalcli 是执行入口，Harness 是裁判入口。
2. 所有 P0/P1 规则必须机器化。
3. 所有命令必须幂等、可审计、可产出报告。
4. Issue 从 Task 生成，PR 从 Issue/Task 生成，Release 从 Manifest 生成。
5. main 只做基线，Task 必须在独立 worktree 中执行。
6. Commit 必须绑定 Goal / Task / Issue / Evidence。
7. PR Ready 必须 Evidence + Traceability + Risk + Rollback 完整。
8. Release 必须生成不可变 Artifact 包。
9. 任何绕过必须有 Waiver，P0 默认不可豁免。
10. v0.1.0 先做裁判系统，再做全自动执行系统。
```

最终结论：

> **Goal Runtime 的第一阶段不是追求“全自动写代码”，而是建立一个不会污染 main、不会无证据 DONE、不会断链发布、不会规则漂移的机器裁判系统。goalcli v0.1.0 应先成为 Harness + Evidence + Traceability + Worktree 的执行入口。**
继续补齐 **Goal 完整规则 v1.4：工程落地目录、检查器契约、goalcli 内核设计、CI 工作流、迁移与防过度工程化规则**。

这一版解决的问题是：

```text
规则已经完整了，怎么真正实现第一版？
哪些文件先建？
哪些命令先做？
哪些 Gate 先接入？
goalcli 内部怎么设计？
如何防止 Goal Runtime 本身变成新的复杂度债？
```

---

# 113. 仓库落地总目录规则

## RULE-REPO-LAYOUT-001：Goal Runtime 必须有固定目录结构

推荐在 `xlib-standard` 中落地：

```text
.agent/
├── constitution/
│   └── CONSTITUTION.md
│
├── rules/
│   ├── 00-index.md
│   ├── 01-core-rules.md
│   ├── 02-goal-runtime-rules.md
│   ├── 03-context-rules.md
│   ├── 04-spec-rules.md
│   ├── 05-design-rules.md
│   ├── 06-task-rules.md
│   ├── 07-worktree-rules.md
│   ├── 08-issue-rules.md
│   ├── 09-commit-rules.md
│   ├── 10-pr-rules.md
│   ├── 11-evidence-rules.md
│   ├── 12-release-rules.md
│   ├── 13-retrospective-rules.md
│   ├── 14-self-improving-rules.md
│   ├── 15-security-rules.md
│   ├── 16-xstack-rules.md
│   ├── 17-downstream-adoption-rules.md
│   ├── 18-deprecation-rules.md
│   ├── 19-scoring-rules.md
│   └── 20-rule-change-protocol.md
│
├── schemas/
│   ├── goal.schema.json
│   ├── spec.schema.json
│   ├── task.schema.json
│   ├── evidence.schema.json
│   ├── traceability.schema.json
│   ├── release.schema.json
│   ├── retrospective.schema.json
│   └── patch.schema.json
│
├── harness/
│   ├── gates/
│   ├── policies/
│   ├── fixtures/
│   │   ├── golden/
│   │   └── violations/
│   └── reports/
│
├── templates/
│   ├── goal-template.md
│   ├── task-template.md
│   ├── issue-template.md
│   ├── pr-template.md
│   ├── evidence-template.md
│   ├── release-manifest-template.md
│   └── retrospective-template.md
│
├── registries/
│   ├── goals.yaml
│   ├── tasks.yaml
│   ├── evidence.yaml
│   ├── patches.yaml
│   ├── debt.yaml
│   └── adoption.yaml
│
├── goals/
│   └── GOAL-YYYYMMDD-NNN/
│
├── violations/
├── patches/
├── adoption/
└── locks/
```

---

# 114. 根目录工具规则

## RULE-ROOT-001：根目录必须提供统一入口

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

# 115. `goalcli.yaml` 配置规则

## RULE-GOALCLI-CONFIG-001：goalcli 必须有统一配置文件

```yaml
version: goalcli-v0.1.0

runtime:
  timezone: Asia/Tokyo
  default_mode: Standard
  default_branch: main
  worktree_root: ~/code/.worktrees

repositories:
  default: xlib-standard
  root: .
  remote: origin

rules:
  severity_policy:
    P0: block
    P1: block_with_waiver
    P2: warn
    P3: score_only

gates:
  required:
    - schema-check
    - worktree-check
    - traceability-check
    - evidence-check
    - secret-check
    - pr-check
    - release-check
    - retro-check

evidence:
  require_commit: true
  require_command: true
  require_artifact: true
  output_dir: .agent/goals/{goal_id}/evidence

release:
  output_dir: release
  require_manifest: true
  require_rollback: true

security:
  secret_patterns:
    - token
    - password
    - secret
    - private_key
    - access_key
    - authorization
```

---

# 116. Makefile 契约规则

## RULE-MAKE-001：Makefile 是本地、CI、Agent 的统一命令入口

最小版本必须包含：

```makefile
.PHONY: schema-check
schema-check:
	goalcli schema validate --all

.PHONY: worktree-check
worktree-check:
	goalcli worktree check

.PHONY: traceability-check
traceability-check:
	goalcli traceability check

.PHONY: evidence-check
evidence-check:
	goalcli evidence check

.PHONY: secret-check
secret-check:
	goalcli secret check

.PHONY: pr-check
pr-check:
	goalcli pr check

.PHONY: release-check
release-check:
	goalcli release check

.PHONY: retro-check
retro-check:
	goalcli retro check

.PHONY: goal-check
goal-check:
	goalcli goal check

.PHONY: ci
ci: schema-check worktree-check traceability-check evidence-check secret-check
```

---

## RULE-MAKE-002：禁止 CI 直接绕过 Makefile

GitHub Actions 不应直接拼复杂脚本。
CI 应该调用：

```bash
make ci
make pr-check
make release-check
```

这样本地、Agent、CI 的行为一致。

---

# 117. Git Hooks 安装规则

## RULE-HOOKS-001：必须提供 hooks 安装脚本

`scripts/git/install-hooks.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail

git config core.hooksPath .githooks

chmod +x .githooks/pre-commit
chmod +x .githooks/pre-push

echo "Git hooks installed."
```

---

## RULE-HOOKS-002：hooks 只做快速阻断

`pre-commit` 不应跑完整 CI，只跑 P0 快速检查：

```bash
make worktree-check
make secret-check
```

`pre-push` 跑：

```bash
make worktree-check
make traceability-check
make evidence-check
```

完整验证交给 PR CI。

---

# 118. GitHub Actions 工作流规则

## RULE-GHA-WORKFLOW-001：必须拆分工作流

推荐四个 workflow：

```text
.github/workflows/worktree-guard.yml
.github/workflows/goal-gates.yml
.github/workflows/release-gate.yml
.github/workflows/secret-scan.yml
```

---

## RULE-GHA-WORKFLOW-002：PR 必须运行 Goal Gates

`goal-gates.yml`：

```yaml
name: Goal Gates

on:
  pull_request:

jobs:
  goal-gates:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Install goalcli
        run: |
          echo "install goalcli here"

      - name: Run CI gates
        run: |
          make ci

      - name: Run PR gate
        run: |
          make pr-check
```

---

## RULE-GHA-WORKFLOW-003：Release 必须单独 Gate

`release-gate.yml`：

```yaml
name: Release Gate

on:
  workflow_dispatch:
  push:
    tags:
      - "v*"

jobs:
  release-gate:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Run release gates
        run: |
          make ci
          make release-check
          make retro-check
```

---

# 119. Branch Protection 规则

## RULE-BRANCH-PROTECTION-001：main 必须启用保护

必须启用：

```text
Require pull request before merging
Require status checks to pass
Require branches to be up to date
Require conversation resolution
Restrict direct pushes
Disallow force push
Disallow deletion
Do not allow bypassing
```

---

## RULE-BRANCH-PROTECTION-002：P0 工作流必须成为 required checks

Required checks 至少包括：

```text
Goal Gates
Worktree Guard
Secret Scan
Release Gate
```

---

# 120. goalcli 内核架构规则

## RULE-GOALCLI-ARCH-001：goalcli v0.1.0 先做裁判，不做复杂 Agent

内核模块建议：

```text
cmd/goalcli/
internal/
├── config/
├── schema/
├── worktree/
├── evidence/
├── traceability/
├── pr/
├── release/
├── retro/
├── audit/
├── report/
├── policy/
└── gitutil/
```

---

## RULE-GOALCLI-ARCH-002：每个 Checker 必须实现统一接口

伪接口：

```go
type CheckResult struct {
    Name      string
    Status    string
    Severity  string
    Errors    []CheckError
    Warnings  []CheckWarning
    Reports   []string
    Evidence  []string
}

type Checker interface {
    Name() string
    Check(ctx context.Context, input CheckInput) CheckResult
}
```

---

# 121. Checker 输出规则

## RULE-CHECKER-001：Checker 输出必须同时支持人读和机器读

每个检查器必须输出：

```text
reports/<checker>.json
reports/<checker>.md 或 .txt
```

例如：

```text
reports/worktree-check.json
reports/worktree-check.txt
```

---

## RULE-CHECKER-002：JSON 报告必须统一结构

```json
{
  "checker": "worktree-check",
  "status": "failed",
  "severity": "P0",
  "goal_id": "GOAL-20260603-001",
  "errors": [
    {
      "rule_id": "RULE-WORKTREE-001",
      "message": "development on main is forbidden",
      "file": "",
      "line": 0
    }
  ],
  "warnings": [],
  "artifacts": [
    "reports/worktree-check.txt"
  ],
  "timestamp": "2026-06-03T17:30:00+09:00"
}
```

---

# 122. Schema 最小字段规则

## RULE-SCHEMA-MIN-001：Goal Schema 最小字段

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

---

## RULE-SCHEMA-MIN-002：Task Schema 最小字段

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

---

## RULE-SCHEMA-MIN-003：Evidence Schema 最小字段

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

# 123. Traceability 检查算法规则

## RULE-TRACE-ALG-001：Traceability Check 必须检查完整链路

检查逻辑：

```text
读取 spec.yaml
读取 tasks.yaml
读取 evidence.yaml
读取 traceability.md 或 traceability.yaml

验证：
1. 每个 Requirement 至少有一个 AC
2. 每个 AC 至少有一个 Task
3. 每个 Task 至少有一个 Evidence
4. 每个 Evidence 绑定 command / commit / artifact
5. Status 与 Evidence 状态一致
```

---

## RULE-TRACE-ALG-002：Traceability 缺链退出码必须固定

```text
缺 Requirement → exit 5
缺 AC → exit 5
缺 Task → exit 5
缺 Evidence → exit 4
Evidence 无 commit → exit 4
```

---

# 124. Evidence Check 算法规则

## RULE-EVID-ALG-001：Evidence Check 必须验证文件真实存在

不能只检查 registry 中写了路径。

必须检查：

```text
artifact path exists
artifact is non-empty
command exists
exit_code recorded
commit exists
timestamp exists
```

---

## RULE-EVID-ALG-002：Release Evidence 必须强制 commit hash

开发阶段 Evidence 可以没有 commit。
Release 阶段 Evidence 必须有 commit hash。

---

# 125. Secret Check 规则

## RULE-SECRET-CHECK-001：secret-check 必须至少支持关键词扫描

最小扫描范围：

```text
.env
*.yaml
*.yml
*.json
*.md
*.go
*.rs
*.sh
.github/
.agent/
release/
reports/
```

---

## RULE-SECRET-CHECK-002：secret-check 必须允许白名单

例如文档中出现 `SECRET=***` 示例，不应误杀。

推荐：

```text
.agent/security/secret-allowlist.yaml
```

结构：

```yaml
allowlist:
  - file: docs/examples/env.md
    pattern: "SECRET=***"
    reason: "masked example"
```

---

# 126. PR Check 规则

## RULE-PR-CHECK-001：PR Check 必须检查 PR Body 固定区块

必需区块：

```text
Goal
Related Issues
Requirements Covered
Changes
Tests
Evidence
Risk
Rollback
Checklist
```

---

## RULE-PR-CHECK-002：PR Check 必须检查 Ready 状态

Draft PR 可以缺 Evidence。
Ready PR 不允许缺：

```text
Evidence
Traceability
Risk
Rollback
CI result
```

---

# 127. Release Check 规则

## RULE-RELEASE-CHECK-001：Release Check 必须检查不可缺产物

```text
release/<REL-ID>/manifest.md
release/<REL-ID>/changelog.md
release/<REL-ID>/evidence-summary.md
release/<REL-ID>/rollback.md
release/<REL-ID>/known-issues.md
```

---

## RULE-RELEASE-CHECK-002：Release Check 必须检查 Evidence 覆盖率

```text
Traceability coverage = 100%
Evidence coverage = 100%
P0/P1 Risk open = 0
Rollback exists = true
```

---

# 128. Retro Check 规则

## RULE-RETRO-CHECK-001：Retrospective 不能只是总结

必须包含：

```text
What Worked
What Failed
Root Cause
Missing Gates
Missing Rules
Prompt Patch
Harness Patch
Rule Patch
New Issue Candidates
```

---

## RULE-RETRO-CHECK-002：缺 Patch 候选则 Retro Gate 失败

除非 Goal 是 Lite Mode 且无失败、无漂移、无新规则需求。

---

# 129. Audit Check 规则

## RULE-AUDIT-CHECK-001：Audit 必须能从 Goal Pack 重建事实链

`goalcli audit goal` 必须输出：

```text
Goal summary
Requirement coverage
Task coverage
Evidence coverage
PR summary
Release summary
Risk summary
Decision summary
Retrospective summary
Open gaps
Final score
```

---

## RULE-AUDIT-CHECK-002：Audit Score 低于 90 不允许 stable Release

```text
>= 90 stable allowed
80-89 rc allowed
70-79 beta allowed
<70 release blocked
```

---

# 130. 防过度工程化规则

## RULE-SIMPLICITY-001：v0.1.0 不实现全自动写代码

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

---

## RULE-SIMPLICITY-002：规则不能无限增长而无机器约束

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

# 131. Lite / Standard / Full Gate 矩阵

## RULE-MODE-GATE-001：不同模式 Gate 不同

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

# 132. Goal 质量评分规则 v1.4

## RULE-SCORE-V14-001：评分必须根据证据，不根据文字漂亮程度

评分项：

```text
Context Recovery: 10
Spec & AC: 10
Design & ADR: 10
Task Decomposition: 10
Worktree Isolation: 10
Harness Coverage: 10
Evidence Quality: 15
Traceability: 10
Release Readiness: 10
Self-improving: 5
```

扣分规则：

```text
无 worktree evidence: -10
缺任意 P0 Evidence: -15
Traceability 不完整: -10
Release 无 rollback: -10
Retro 无 Patch: -5
引用不存在命令: -10
文档与 Makefile 漂移: -10
```

---

# 133. v0.1.0 最小交付清单

## 第一批必须交付

```text
[ ] goalcli.yaml
[ ] Makefile gate targets
[ ] scripts/harness/no-main-dev.sh
[ ] .githooks/pre-commit
[ ] .githooks/pre-push
[ ] .github/workflows/worktree-guard.yml
[ ] .github/workflows/goal-gates.yml
[ ] .agent/schemas/goal.schema.json
[ ] .agent/schemas/task.schema.json
[ ] .agent/schemas/evidence.schema.json
[ ] .agent/templates/issue-template.md
[ ] .agent/templates/pr-template.md
[ ] .agent/templates/evidence-template.md
[ ] .agent/harness/fixtures/golden/
[ ] .agent/harness/fixtures/violations/
```

---

# 134. v0.1.0 最小命令清单

```bash
goalcli schema validate --all
goalcli worktree check
goalcli evidence check
goalcli traceability check
goalcli pr check
goalcli release check
goalcli retro check
goalcli audit goal --goal <GOAL-ID>
```

---

# 135. v0.1.0 最小测试清单

```text
[ ] main 分支执行 worktree-check 必须失败
[ ] 合法 worktree 执行 worktree-check 必须通过
[ ] 缺 Evidence 的 Task 必须失败
[ ] 缺 commit 的 Release Evidence 必须失败
[ ] 缺 Traceability 的 PR 必须失败
[ ] 缺 Release Manifest 的 Release 必须失败
[ ] 含 secret 的文件必须失败
[ ] Golden Goal Pack 必须通过所有 Gate
```

---

# 136. v0.1.0 最小 Issue 拆解最终版

```text
ISSUE-001: 初始化 .agent 目录与规则索引
ISSUE-002: 增加 goalcli.yaml 与 Makefile gates
ISSUE-003: 实现 worktree-only Gate
ISSUE-004: 实现 schema-check
ISSUE-005: 实现 evidence-check
ISSUE-006: 实现 traceability-check
ISSUE-007: 实现 pr-check
ISSUE-008: 实现 release-check
ISSUE-009: 实现 retro-check
ISSUE-010: 实现 secret-check
ISSUE-011: 增加 Git hooks 与安装脚本
ISSUE-012: 增加 GitHub Actions workflows
ISSUE-013: 增加 Golden / Violation fixtures
ISSUE-014: 实现 audit goal
ISSUE-015: 生成 Release Manifest 与 v0.1.0 自审计报告
```

---

# 137. 执行顺序规则

## RULE-ORDER-001：先做 Gate，再做自动化

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

# 138. 当前阶段最终目标

当前阶段不是把所有 Goal Runtime 都做完。
当前阶段目标是：

> **建立一个最小、不可绕过、可审计的 Goal 裁判系统。**

完成标准：

```text
main 开发会被阻断
缺 Evidence 会被阻断
缺 Traceability 会被阻断
缺 Release Manifest 会被阻断
含 Secret 会被阻断
合法 Golden Goal 可以通过
所有结果都有 reports
所有结果可以被 audit goal 汇总
```

---

# 139. 最终推荐路径 v1.4

按优先级：

```text
P0：worktree-only + secret-check
P0：schema-check + evidence-check
P0：traceability-check
P1：pr-check + release-check
P1：retro-check
P1：audit goal
P2：GitHub bot comment
P2：Issue / PR 自动生成
P3：Downstream adoption automation
```

---

# 140. v1.4 总结铁律

```text
1. 先裁判，后自动化。
2. 先阻断错误，再追求效率。
3. 先离线可审计，再接 GitHub 自动化。
4. 先 worktree-only，再并发 Agent。
5. 先 Evidence + Traceability，再 PR / Release。
6. 先 Golden / Violation fixtures，再扩展规则。
7. 先 goalcli v0.1.0 最小命令，再做复杂 Agent Team。
8. 规则增长必须被机器化，否则就是新文档债。
```

最终结论：

> **Goal Runtime v0.1.0 的正确形态不是“自动化平台”，而是“最小机器裁判内核”：防 main 污染、防无证据完成、防断链发布、防密钥泄漏、防规则漂移。先把不会错做到，再把速度做快。**
继续补齐 **Goal 完整规则 v1.5：执行 Runbook、DoR/DoD、状态迁移、Agent 操作边界、最小发布闭环、治理闭环**。

这一版把前面的规则进一步收敛成：

```text
从“规则体系”
→ 到“每天怎么跑”
→ 到“每个 Issue 怎么进入 Ready”
→ 到“每个 PR 怎么进入 Merge”
→ 到“每个 Release 怎么进入 Done”
→ 到“每次 Retro 怎么反哺规则”
```

---

# 141. Definition of Ready 规则

## RULE-DOR-001：没有 Ready，不允许进入执行

任何 Task / Issue 在进入执行前必须满足 **DoR**。

```text
Definition of Ready = 可以开工的最低条件
```

Task Ready 条件：

```text
[ ] Goal ID 已存在
[ ] Task ID 已存在
[ ] Requirement 已绑定
[ ] Acceptance Criteria 已绑定
[ ] files_to_change 已声明
[ ] commands_to_run 已声明
[ ] evidence_required 已声明
[ ] rollback_plan 已声明
[ ] worktree 创建命令已明确
[ ] 无 P0 blocker
```

---

## RULE-DOR-002：Issue Ready 必须由 Harness 判定

Issue 进入 `state:ready` 前必须通过：

```bash
make task-check
make issue-check
make traceability-check
```

如果失败：

```text
Issue state = BLOCKED
Goal state = NEEDS_REPLAN 或 INCONSISTENT_STATE
```

---

# 142. Definition of Done 规则

## RULE-DOD-001：Task Done 必须有证据

Task Done 条件：

```text
[ ] 实现完成
[ ] 本地 Gate 通过
[ ] 测试通过
[ ] Evidence 已生成
[ ] Evidence 已绑定 commit
[ ] Traceability 已更新
[ ] PR 已创建或更新
[ ] Risk / Rollback 已更新
```

---

## RULE-DOD-002：Issue Done 必须经过 PR

Issue Done 条件：

```text
[ ] 对应 PR 已 merge
[ ] 所有关联 Task 完成
[ ] Evidence 完整
[ ] Traceability 完整
[ ] Issue comment 中包含 Evidence Summary
[ ] 无未关闭 P0/P1 风险
```

禁止：

```text
手动关闭 Issue
无 PR 关闭 Issue
无 Evidence 关闭 Issue
```

---

## RULE-DOD-003：Goal Done 必须经过 Release / Review / Retro

Goal Done 条件：

```text
[ ] 所有 P0/P1 Issue closed
[ ] 所有 PR merged
[ ] Release Manifest 完整
[ ] Audit Score 达标
[ ] Retrospective 完整
[ ] Prompt / Harness / Rule Patch 候选已登记
[ ] Downstream impact 已判断
[ ] Goal Registry 状态更新为 DONE
```

---

# 143. 状态迁移门禁规则

## RULE-STATE-GATE-001：每个状态迁移必须有 Gate

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

---

## RULE-STATE-GATE-002：状态迁移必须写入 Registry

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

# 144. Agent Daily Runbook 规则

## RULE-RUNBOOK-001：Agent 每次执行必须先跑 Preflight

标准启动流程：

```bash
git status
git branch --show-current
make worktree-check
make schema-check
goalcli goal status --goal <GOAL-ID>
goalcli task status --task <TASK-ID>
```

禁止直接开始改文件。

---

## RULE-RUNBOOK-002：Agent 执行顺序必须固定

```text
1. 读取 Goal Pack
2. 检查当前状态
3. 检查 worktree
4. 检查 Task DoR
5. 执行最小变更
6. 运行本地 Gate
7. 收集 Evidence
8. 更新 Traceability
9. 创建 commit
10. 更新 PR
11. 写 Execution Log
```

---

## RULE-RUNBOOK-003：Agent 每次结束必须写收尾记录

收尾记录必须包含：

```text
当前状态
已完成内容
运行命令
失败命令
Evidence ID
变更文件
下一步
是否存在 blocker
```

文件：

```text
.agent/goals/<GOAL-ID>/execution-log.md
```

---

# 145. Agent Stop Conditions 规则

## RULE-STOP-001：以下情况必须停止执行

```text
P0 Gate 失败
main / master 分支开发
检测到 secret
Evidence 缺失
Traceability 断链
Schema 无效
工作目录不干净且无法解释
修改超出授权范围
出现架构冲突
需要人工审批
```

---

## RULE-STOP-002：停止后必须生成 Blocker 记录

```md
# BLOCKER-YYYYMMDD-NNN

## Goal
GOAL-xxx

## Task
TASK-xxx

## Reason
-

## Failed Gate
-

## Evidence
-

## Required Decision
-

## Suggested Fix
-
```

---

# 146. Agent Repair Loop 规则

## RULE-REPAIR-001：失败后最多自动修复 N 次

建议：

```yaml
repair_policy:
  max_auto_retries: 3
  retry_requires_new_evidence: true
  repeated_failure_state: NEEDS_REPLAN
```

---

## RULE-REPAIR-002：每次修复必须保留失败证据

禁止覆盖失败报告。

应该保留：

```text
reports/failures/<timestamp>-<gate>.json
reports/failures/<timestamp>-<gate>.txt
```

失败证据同样有价值，因为它会进入 Retrospective。

---

# 147. 文件所有权规则

## RULE-OWNERSHIP-001：关键文件必须有 Owner

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

---

## RULE-OWNERSHIP-002：修改关键文件必须触发 Review

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

# 148. 变更类型规则

## RULE-CHANGE-TYPE-001：每个 Task 必须声明变更类型

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

---

## RULE-CHANGE-TYPE-002：不同类型触发不同 Gate

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

# 149. Impact Analysis 规则

## RULE-IMPACT-001：P0/P1 变更必须有影响分析

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

---

## RULE-IMPACT-002：影响分析必须进入 PR

PR 中必须有：

```md
## Impact Analysis

| Area | Impact | Required Action |
|---|---|---|
| Rules | yes/no | |
| Harness | yes/no | |
| CI | yes/no | |
| Templates | yes/no | |
| Downstream | yes/no | |
| Release | yes/no | |
```

---

# 150. Backward Compatibility 规则

## RULE-COMPAT-001：规则升级必须兼容旧 Goal Pack

除非明确标记 breaking，否则新版本 goalcli 必须能读取旧版本 Goal Pack。

---

## RULE-COMPAT-002：Breaking 规则必须有迁移计划

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

# 151. Migration Script 规则

## RULE-MIGRATION-001：结构性变更必须有迁移脚本

目录：

```text
.agent/migrations/
├── MIGRATION-20260603-rules-v1.4-to-v1.5.md
└── scripts/
    └── migrate-rules-v1.4-to-v1.5.sh
```

---

## RULE-MIGRATION-002：迁移必须可 dry-run

```bash
goalcli migrate --from rules-v1.4 --to rules-v1.5 --dry-run
goalcli migrate --from rules-v1.4 --to rules-v1.5 --apply
```

---

# 152. Registry Consistency 规则

## RULE-REGISTRY-CONSISTENCY-001：Registry 之间必须一致

必须检查：

```text
goals.yaml 中的 Goal 存在对应 .agent/goals/<GOAL-ID>/
tasks.yaml 中的 Task 存在于对应 Goal Pack
evidence.yaml 中的 Evidence 文件真实存在
patches.yaml 中的 Patch 文件真实存在
adoption.yaml 中的下游状态有证据
```

---

## RULE-REGISTRY-CONSISTENCY-002：Registry 漂移必须阻断 Release

```bash
make registry-check
```

如果失败：

```text
Goal state = INCONSISTENT_STATE
Release blocked
```

---

# 153. Evidence Coverage 规则

## RULE-EVIDENCE-COVERAGE-001：Evidence 覆盖率必须量化

计算：

```text
Evidence Coverage = 有有效 Evidence 的 Task 数 / 总 Task 数
Traceability Coverage = 完整链路数量 / 总 Requirement 数
```

Release 要求：

```text
Evidence Coverage = 100%
Traceability Coverage = 100%
P0/P1 Risk Closure = 100%
```

---

## RULE-EVIDENCE-COVERAGE-002：Coverage 结果必须进入 Audit

```text
reports/audit/<GOAL-ID>.md
```

必须包含：

```text
task_count
task_with_evidence_count
requirement_count
requirement_with_full_trace_count
coverage_percent
open_gaps
```

---

# 154. Test Pyramid for Goal Runtime 规则

## RULE-GOAL-TEST-001：Goal Runtime 自身也需要测试金字塔

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

---

## RULE-GOAL-TEST-002：每个新增 Gate 必须有正反样例

新增 Gate 时必须同时增加：

```text
golden fixture
violation fixture
expected report
```

---

# 155. Golden Goal Pack 标准

## RULE-GOLDEN-PACK-001：Golden Pack 是系统最小正确样例

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

---

## RULE-GOLDEN-PACK-002：Golden Pack 必须进入 CI

CI 必须跑：

```bash
goalcli audit goal --fixture .agent/harness/fixtures/golden/minimal-goal-pack
```

---

# 156. Violation Fixture 标准

## RULE-VIOLATION-FIXTURE-001：每个 P0 规则必须有违规样例

至少：

```text
missing-evidence/
missing-traceability/
main-branch-dev/
secret-leak/
missing-release-manifest/
invalid-schema/
```

---

## RULE-VIOLATION-FIXTURE-002：违规样例必须断言失败原因

每个 violation fixture 必须定义：

```yaml
expected:
  exit_code: 4
  rule_id: RULE-EVIDENCE-001
  status: failed
```

---

# 157. PR Merge Queue 规则

## RULE-MERGE-QUEUE-001：多个 PR 必须进入 Merge Queue

原因：

```text
防止 PR 单独通过，合并后互相破坏
```

要求：

```text
每个 PR 合并前基于最新 main 重跑 required checks
```

---

## RULE-MERGE-QUEUE-002：合并后必须更新 Goal 状态

PR merge 后自动更新：

```text
Task status
Issue status
Traceability status
Evidence commit
Goal progress
```

---

# 158. Release Train 规则

## RULE-RELEASE-TRAIN-001：多个 Goal 不应随意混发

Release 应绑定：

```text
一个主 Goal
或一组明确兼容的 Goal
```

禁止：

```text
把未完成 Goal 混进 stable release
```

---

## RULE-RELEASE-TRAIN-002：Release 必须声明包含范围

```yaml
release:
  rel_id: REL-20260603-goal-runtime
  included_goals:
    - GOAL-20260603-001
  excluded_goals:
    - GOAL-20260603-002
```

---

# 159. Partial Release 规则

## RULE-PARTIAL-RELEASE-001：允许部分发布，但必须显式声明

如果某些 Task 未完成，Release Manifest 必须写：

```text
included
excluded
known gaps
follow-up issues
risk
```

---

## RULE-PARTIAL-RELEASE-002：P0 缺口不允许 stable partial release

P0 未完成时最多：

```text
alpha / beta / rc
```

不得 stable。

---

# 160. Goal Freeze 规则

## RULE-GOAL-FREEZE-001：进入 Release 前必须冻结 Goal Scope

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

---

## RULE-GOAL-FREEZE-002：冻结后只允许修复 release blocker

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

# 161. Goal Cancellation 规则

## RULE-CANCEL-001：Goal 可以取消，但不能静默消失

取消时必须生成：

```text
cancellation report
reason
completed work
remaining work
evidence retained
follow-up decision
```

状态：

```text
CANCELLED
```

---

## RULE-CANCEL-002：取消 Goal 的 worktree / branch / issue 必须清理

必须处理：

```text
open PR
open Issue
worktree
branch
partial evidence
lock file
```

---

# 162. Archival 规则

## RULE-ARCHIVE-001：完成 Goal 必须归档

完成后归档：

```text
.agent/goals/<GOAL-ID>/
release/<REL-ID>/
reports/audit/<GOAL-ID>.md
```

---

## RULE-ARCHIVE-002：归档后不能改历史证据

如果需要修正，创建：

```text
correction note
new evidence
new patch release
```

---

# 163. Prompt Patch 生效规则

## RULE-PROMPT-PATCH-001：Prompt Patch 不能只生成，必须裁决

状态流：

```text
PROPOSED
→ REVIEWED
→ ACCEPTED / REJECTED
→ APPLIED
→ VERIFIED
```

---

## RULE-PROMPT-PATCH-002：Prompt Patch 生效必须有验证

验证方式：

```text
新 Prompt 在 golden goal 上通过
原失败场景不再失败
没有引入新的 P0 误杀
```

---

# 164. Harness Patch 生效规则

## RULE-HARNESS-PATCH-001：Harness Patch 必须带测试

任何新增 Gate 必须附带：

```text
checker implementation
golden fixture
violation fixture
report sample
Makefile target
CI workflow update
```

---

## RULE-HARNESS-PATCH-002：Gate 从 warn 升级 block 必须公告

因为会影响下游采用。

必须更新：

```text
CHANGELOG
adoption-impact.md
migration notes
```

---

# 165. Rule Patch 生效规则

## RULE-RULE-PATCH-001：Rule Patch 必须同步四处

```text
.agent/rules/
.agent/policies/
.agent/templates/
.agent/harness/gates/
```

如果只改 Markdown，不算完成。

---

## RULE-RULE-PATCH-002：Rule Patch 必须进入 Rule Index

```text
.agent/rules/00-index.md
```

---

# 166. New Issue Candidate 规则

## RULE-ISSUE-CANDIDATE-001：Retro 产生的新 Issue 不能丢

必须登记：

```text
.agent/registries/issue-candidates.yaml
```

字段：

```yaml
candidate_id:
source_goal:
title:
reason:
priority:
suggested_owner:
suggested_acceptance_criteria:
```

---

## RULE-ISSUE-CANDIDATE-002：Issue Candidate 必须被裁决

状态：

```text
PROPOSED
ACCEPTED
REJECTED
MERGED_WITH_EXISTING
```

---

# 167. Downstream Adoption Scoring 规则

## RULE-ADOPTION-SCORE-001：下游采用必须评分

每个下游仓库评分：

```text
Rules adopted: 20
Templates adopted: 15
Harness gates adopted: 25
Makefile gates adopted: 15
CI workflows adopted: 15
Evidence protocol adopted: 10
```

满分 100。

---

## RULE-ADOPTION-SCORE-002：低于 80 不算完成采用

```text
>= 90: full adoption
80-89: usable adoption
60-79: partial adoption
<60: failed adoption
```

---

# 168. xstack 分层准入规则

## RULE-XSTACK-ADMISSION-001：基础库进入 xstack 必须过准入

每个库必须有：

```text
README
CONSTITUTION
.agent/rules
Makefile gates
CI
Evidence Protocol
Release Manifest
API stability policy
```

---

## RULE-XSTACK-ADMISSION-002：未通过准入不能作为标准库推广

尤其是：

```text
kernel
configx
observex
testkitx
redisx
kafkax
postgresx
taosx
ossx
clickhousex
```

---

# 169. Anti-Cargo-Cult 规则

## RULE-ANTI-CARGO-001：不得复制规则而不接 Gate

下游仓库不能只复制 `.agent/rules`。

必须同步：

```text
Makefile gates
scripts/harness
GitHub workflows
templates
schemas
evidence examples
```

---

## RULE-ANTI-CARGO-002：规则采用必须有运行证据

```text
make ci
make evidence-check
make release-check
```

必须在下游通过。

---

# 170. 最终 MVA 执行战役

把当前阶段拆成 5 个战役：

```text
Campaign 1: Worktree & Secret Safety
Campaign 2: Schema & Goal Pack
Campaign 3: Evidence & Traceability
Campaign 4: PR & Release Gates
Campaign 5: Retro & Self-improving
```

---

## Campaign 1：Worktree & Secret Safety

目标：

```text
防止 main 污染
防止 secret 泄漏
```

交付：

```text
worktree-check
secret-check
hooks
branch protection
GitHub workflow
```

验收：

```text
main commit 被阻断
direct push main 被阻断
secret fixture 被阻断
合法 worktree 通过
```

---

## Campaign 2：Schema & Goal Pack

目标：

```text
让 Goal 对象机器可读
```

交付：

```text
goal.schema.json
task.schema.json
evidence.schema.json
Goal Pack 模板
schema-check
golden fixture
```

验收：

```text
合法 Goal Pack 通过
缺字段 Goal Pack 失败
```

---

## Campaign 3：Evidence & Traceability

目标：

```text
防止无证据完成
防止需求断链
```

交付：

```text
evidence-check
traceability-check
evidence registry
traceability matrix
violation fixtures
```

验收：

```text
缺 Evidence 失败
缺 Traceability 失败
完整链路通过
```

---

## Campaign 4：PR & Release Gates

目标：

```text
防止无审查合并
防止断链发布
```

交付：

```text
pr-check
release-check
release artifact structure
release manifest template
rollback template
```

验收：

```text
缺 PR Evidence 失败
缺 Release Manifest 失败
缺 rollback 失败
```

---

## Campaign 5：Retro & Self-improving

目标：

```text
每次执行都产生复利资产
```

交付：

```text
retro-check
patch registry
issue-candidates registry
audit goal
metrics report
```

验收：

```text
无 Patch 候选的 Retro 失败
audit 能重建事实链
```

---

# 171. 30 天实现节奏

## Day 1-3

```text
完成 Campaign 1
```

## Day 4-7

```text
完成 Campaign 2
```

## Day 8-14

```text
完成 Campaign 3
```

## Day 15-21

```text
完成 Campaign 4
```

## Day 22-30

```text
完成 Campaign 5
生成 v0.1.0 Release
对 kernel 做第一个 downstream adoption 试点
```

---

# 172. v1.5 总结铁律

```text
1. 没有 DoR，不开工。
2. 没有 DoD，不关闭。
3. 状态迁移必须有 Gate。
4. Agent 每次执行必须有 Runbook 和 Execution Log。
5. 失败不能覆盖，要保留为复利资产。
6. 关键文件必须有 Owner。
7. P0/P1 变更必须做 Impact Analysis。
8. 规则升级必须考虑兼容与迁移。
9. Golden / Violation fixtures 是 Harness 的生命线。
10. Release 前必须冻结 Scope。
11. Retro Patch 必须被裁决，不能生成后丢弃。
12. 下游采用不能只复制文档，必须跑 Gate。
```

最终结论：

> **Goal Runtime 的成熟形态，是一个带 DoR、DoD、状态迁移、Evidence、Traceability、Release Freeze、Retro Patch 和 Downstream Adoption 的工程治理闭环。它不是为了写更多文档，而是为了让每一次执行都能被机器裁判、被人审计、被下游复用、被下一轮复利增强。**
继续补齐 **Goal 完整规则 v1.6：控制平面、权限模型、命令事务、审计留痕、发布不可变性、长期治理与降复杂度机制**。

这一版重点解决：

```text
Goal Runtime 如何长期运行？
goalcli / .agent / Harness 谁是事实源？
自动化命令失败如何恢复？
权限如何最小化？
Evidence 如何长期保存？
规则如何避免无限膨胀？
```

---

# 173. 控制平面规则

## RULE-CONTROL-001：`.agent` 是 Goal Runtime 控制平面

`.agent` 不是普通文档目录，而是运行时控制平面。

它必须承载：

```text
规则
模板
Schema
Harness Gate
Goal Pack
Registry
Evidence Index
Patch Registry
Adoption State
Violation Log
Audit Report
```

---

## RULE-CONTROL-002：控制平面优先于散落文档

优先级：

```text
.agent/schemas      # 机器对象定义
.agent/rules        # 规则事实源
.agent/policies     # 机器策略
.agent/harness      # 裁判实现
.agent/registries   # 状态事实源
.agent/goals        # Goal 事实包
docs/               # 人类阅读文档
README              # 导航入口
```

如果 README 与 `.agent/rules` 冲突，以 `.agent/rules` 为准。

---

# 174. SSOT 事实源规则

## RULE-SSOT-001：每类事实必须只有一个主源

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

---

## RULE-SSOT-002：非 SSOT 文档只能引用，不得复制事实

例如：

```text
README 可以写：
详见 .agent/rules/07-worktree-rules.md

但不应复制完整 worktree 规则，避免漂移。
```

---

# 175. 命令事务规则

## RULE-CMD-TXN-001：goalcli 命令必须事务化

任何会修改状态的命令必须遵循：

```text
precheck
→ plan
→ apply
→ verify
→ write report
→ update registry
```

例如：

```bash
goalcli worktree create --task TASK-001
```

内部必须执行：

```text
检查 Goal / Task 存在
检查 main 最新
检查目标路径不存在或可复用
创建 worktree
运行 worktree-check
更新 tasks.yaml
写 reports/worktree-create.json
```

---

## RULE-CMD-TXN-002：命令失败必须可恢复

失败时必须输出：

```text
已完成哪些步骤
失败在哪一步
是否有部分状态
如何回滚
如何重试
```

禁止：

```text
命令失败但留下未知状态
Registry 更新了但 worktree 没创建
Issue 创建了但 tasks.yaml 没回写
```

---

# 176. Dry-run 规则

## RULE-DRYRUN-001：所有破坏性命令必须支持 dry-run

必须支持：

```bash
goalcli release publish --dry-run
goalcli migrate --dry-run
goalcli worktree clean --dry-run
goalcli issues sync --dry-run
```

---

## RULE-DRYRUN-002：dry-run 必须输出计划

输出必须包含：

```text
将修改哪些文件
将调用哪些外部 API
将创建哪些 Issue / PR / Tag
将删除哪些 worktree
风险是什么
阻断条件是什么
```

---

# 177. 外部系统权限规则

## RULE-PERMISSION-001：GitHub 权限必须最小化

goalcli / Agent 不应默认拥有全部权限。

权限分级：

```text
read-only:
  读取 repo / issue / PR / release

issue-writer:
  创建和更新 Issue

pr-writer:
  创建和更新 PR

release-writer:
  创建 tag / release

admin:
  修改 branch protection / secrets / workflow 权限
```

---

## RULE-PERMISSION-002：发布权限必须单独隔离

自动化执行可以创建 draft release，但 stable 发布必须使用独立权限。

```text
普通 Agent：不得 stable publish
Release Agent：可 prepare
Human / trusted release token：可 publish stable
```

---

# 178. 自动化安全边界规则

## RULE-AUTO-SAFETY-001：自动化只能扩大确定性，不能扩大不确定性

允许自动化：

```text
生成 Issue
生成 PR 模板
收集 Evidence
跑 Gate
更新 Traceability
生成 Release Manifest
生成 Audit Report
```

不允许默认自动化：

```text
绕过 Review
自动合并 P0/P1 PR
自动发布 stable
自动删除历史 Evidence
自动修改 P0 规则
```

---

## RULE-AUTO-SAFETY-002：高风险自动化必须先进入 Simulation

高风险自动化包括：

```text
跨仓库批量同步
批量创建 PR
批量修改规则
批量 release
批量迁移 schema
```

必须先执行：

```bash
goalcli simulate adoption --target kernel
goalcli simulate migrate --from rules-v1.5 --to rules-v1.6
```

---

# 179. Evidence 长期保存规则

## RULE-EVIDENCE-RETENTION-001：Evidence 必须有保存策略

推荐：

```yaml
retention:
  failed_gate_reports: 180d
  successful_task_evidence: 365d
  release_evidence: permanent
  audit_reports: permanent
```

---

## RULE-EVIDENCE-RETENTION-002：Release Evidence 永久保存

Release 证据包括：

```text
manifest
evidence-summary
test-summary
rollback
audit report
tag
commit hash
CI run id
```

不得自动清理。

---

# 180. Evidence Hash 规则

## RULE-EVIDENCE-HASH-001：Release Artifact 必须生成 hash

发布产物必须包含：

```text
checksums.txt
```

示例：

```text
sha256  manifest.md
sha256  evidence-summary.md
sha256  rollback.md
sha256  audit-report.md
```

---

## RULE-EVIDENCE-HASH-002：审计时必须校验 hash

```bash
goalcli audit release --release REL-20260603-goal-runtime
```

必须验证：

```text
artifact 存在
hash 匹配
manifest 未被修改
Evidence 未缺失
```

---

# 181. 审计等级规则

## RULE-AUDIT-LEVEL-001：审计分三级

```text
L1: Task Audit
L2: Goal Audit
L3: Release Audit
```

---

## RULE-AUDIT-LEVEL-002：不同对象不同审计要求

| 审计级别          | 必须检查                                                   |
| ------------- | ------------------------------------------------------ |
| Task Audit    | Task / AC / Evidence / Commit                          |
| Goal Audit    | Requirement / Task / PR / Risk / Retro                 |
| Release Audit | Manifest / Tag / Rollback / Evidence / Adoption Impact |

---

# 182. 违规等级规则

## RULE-VIOLATION-SEVERITY-001：违规必须分级

```text
V0: 安全事故，例如 secret 泄漏
V1: 发布阻断，例如无 Evidence / 无 Traceability
V2: 规则漂移，例如 README 与 Makefile 不一致
V3: 风格或模板问题
```

---

## RULE-VIOLATION-SEVERITY-002：V0 必须进入安全流程

V0 必须执行：

```text
停止执行
移除泄漏内容
轮换凭据
生成安全事件记录
阻断发布
新增回归测试
Retrospective 必须生成 Rule / Harness Patch
```

---

# 183. 规则膨胀控制规则

## RULE-RULE-BLOAT-001：新增规则必须给出机器化路径

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

---

## RULE-RULE-BLOAT-002：规则必须定期清理

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

# 184. 文档债规则

## RULE-DOC-DEBT-001：文档重复即债务

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

---

## RULE-DOC-DEBT-002：文档必须有生命周期

每个长期文档建议包含：

```yaml
status: active | draft | deprecated | archived
owner:
last_reviewed_at:
ssot:
related_rules:
```

---

# 185. Registry Lock 规则

## RULE-REGISTRY-LOCK-001：Registry 更新必须加锁

修改以下文件时必须持有 lock：

```text
.agent/registries/goals.yaml
.agent/registries/tasks.yaml
.agent/registries/evidence.yaml
.agent/registries/patches.yaml
.agent/registries/adoption.yaml
```

---

## RULE-REGISTRY-LOCK-002：锁超时必须可恢复

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

# 186. Multi-Repo Goal 规则

## RULE-MULTIREPO-001：跨仓库 Goal 必须有协调 Goal Pack

跨仓库 Goal 不能只在一个仓库记录。

必须有：

```text
orchestration goal pack
per-repo goal pack
cross-repo traceability
cross-repo release plan
rollback plan
```

---

## RULE-MULTIREPO-002：跨仓库发布必须声明顺序

例如：

```text
1. xlib-standard 发布规则版本
2. kernel 采用规则
3. configx / observex / testkitx 采用
4. L2 基础库采用
5. x.go 采用
```

---

# 187. Downstream Contract 规则

## RULE-DOWNSTREAM-CONTRACT-001：xlib-standard 必须发布下游契约

契约包括：

```text
必须复制哪些目录
必须实现哪些 Make target
必须启用哪些 GitHub workflow
必须遵守哪些 P0 规则
允许下游覆盖哪些规则
不允许覆盖哪些规则
```

---

## RULE-DOWNSTREAM-CONTRACT-002：下游不得弱化 P0

下游可增强：

```text
增加更多 Gate
增加更严格测试
增加 repo-specific rule
```

但不得弱化：

```text
worktree-only
evidence required
traceability required
secret blocking
release manifest required
```

---

# 188. Rule Compatibility Matrix 规则

## RULE-COMPAT-MATRIX-001：规则版本必须有兼容矩阵

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

---

## RULE-COMPAT-MATRIX-002：不兼容必须阻断 adoption

如果下游：

```text
rules-v1.6 + harness-v0.0
```

不兼容，则：

```text
adoption-check failed
```

---

# 189. Goal Runtime Dashboard 规则

## RULE-DASHBOARD-001：必须能生成静态 Dashboard

不需要第一阶段做 Web 服务，先生成静态报告：

```bash
goalcli dashboard generate
```

输出：

```text
reports/dashboard/index.md
```

---

## RULE-DASHBOARD-002：Dashboard 至少展示

```text
Active Goals
Blocked Goals
Open P0/P1 Issues
Gate Pass Rate
Evidence Coverage
Traceability Coverage
Open Violations
Pending Patches
Downstream Adoption Score
```

---

# 190. Metrics Governance 规则

## RULE-METRIC-GOV-001：指标必须驱动治理，不只是展示

触发规则：

```text
Evidence Coverage < 100%：阻断 release
Traceability Coverage < 100%：阻断 release
Open V0 > 0：阻断所有 release
Open P0 violation > 0：阻断 stable
Patch adoption rate < 50%：触发 governance review
Rule drift count > 3：触发 drift cleanup goal
```

---

## RULE-METRIC-GOV-002：指标必须进入 Retrospective

Retro 必须包含：

```text
本轮 Gate 失败次数
重复失败类型
Evidence 覆盖率
Traceability 覆盖率
修复耗时
新增 Patch 数
被采纳 Patch 数
```

---

# 191. Release Promotion 规则

## RULE-PROMOTION-001：发布必须支持晋级

发布通道：

```text
alpha → beta → rc → stable
```

---

## RULE-PROMOTION-002：晋级必须有新增 Evidence

不能同一套未更新证据直接晋级。

晋级 stable 前必须重新跑：

```bash
make ci
make release-check
goalcli audit release --release <REL-ID>
```

---

# 192. Roll-forward 规则

## RULE-ROLLFORWARD-001：小问题优先 roll-forward

如果 stable 已发布且问题低风险：

```text
优先新 patch release
不直接修改旧 release artifact
```

---

## RULE-ROLLFORWARD-002：高风险问题必须 rollback

高风险包括：

```text
secret 泄漏
破坏 P0 gate
错误发布 breaking rule
下游大面积失败
```

---

# 193. Agent 记忆与文件事实规则

## RULE-AGENT-MEMORY-001：Agent 记忆不能作为事实源

Agent 可以参考历史记忆，但不能替代：

```text
当前仓库文件
Goal Pack
Registry
Evidence
CI 报告
Release Manifest
```

---

## RULE-AGENT-MEMORY-002：重要上下文必须写入文件

不能只存在聊天中。

必须落地：

```text
decision-log.md
context-summary.md
current-state.md
next-actions.md
```

---

# 194. Context Window 防爆规则

## RULE-CONTEXT-WINDOW-001：大型 Goal 必须分层摘要

必须提供：

```text
00-current-state.md
01-decision-summary.md
02-open-blockers.md
03-next-actions.md
04-evidence-summary.md
```

---

## RULE-CONTEXT-WINDOW-002：Agent 恢复时先读摘要，再读原文

读取顺序：

```text
current-state
decision-summary
open-blockers
traceability
tasks
evidence
full design
```

---

# 195. Goal Split 规则

## RULE-GOAL-SPLIT-001：Goal 过大必须拆分

拆分信号：

```text
超过 15 个 P0/P1 Task
超过 5 个子系统
超过 3 个仓库
预计 Release 需要多个阶段
Traceability Matrix 超过 100 行
```

---

## RULE-GOAL-SPLIT-002：拆分后必须保留父子关系

```yaml
parent_goal: GOAL-20260603-001
child_goals:
  - GOAL-20260603-002
  - GOAL-20260603-003
```

---

# 196. Goal Merge 规则

## RULE-GOAL-MERGE-001：相似 Goal 可以合并，但必须保留来源

合并必须记录：

```text
原 Goal ID
合并原因
保留 Scope
舍弃 Scope
迁移后的 Task
迁移后的 Evidence
```

---

# 197. No-Orphan 规则

## RULE-ORPHAN-001：禁止孤儿对象

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

---

## RULE-ORPHAN-002：orphan-check 必须进入 CI

```bash
make orphan-check
```

---

# 198. Naming Normalization 规则

## RULE-NAMING-001：命名必须统一

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

# 199. Glossary 规则

## RULE-GLOSSARY-001：核心术语必须进入 Glossary

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

---

## RULE-GLOSSARY-002：文档新增术语必须同步 Glossary

否则 docs-check 警告。

---

# 200. 最终闭环成熟度等级

## RULE-MATURITY-001：Goal Runtime 成熟度分五级

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

---

## RULE-MATURITY-002：当前目标应先达到 L2

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

# 201. v1.6 最终新增交付清单

```text
[ ] .agent/policies/*.yaml
[ ] .agent/glossary.md
[ ] .agent/ownership.yaml
[ ] .agent/locks/
[ ] .agent/violations/
[ ] .agent/migrations/
[ ] .agent/registries/adoption.yaml
[ ] .agent/registries/issue-candidates.yaml
[ ] reports/dashboard/index.md
[ ] reports/audit/
[ ] release/<REL-ID>/checksums.txt
[ ] rule-prune-report.md
[ ] downstream-contract.md
[ ] compatibility-matrix.md
```

---

# 202. v1.6 最终新增 Make Target

```makefile
.PHONY: registry-check
registry-check:
	goalcli registry check

.PHONY: orphan-check
orphan-check:
	goalcli orphan check

.PHONY: drift-check
drift-check:
	goalcli drift check

.PHONY: audit-check
audit-check:
	goalcli audit goal --all

.PHONY: dashboard
dashboard:
	goalcli dashboard generate

.PHONY: adoption-check
adoption-check:
	goalcli adoption check
```

---

# 203. v1.6 最终铁律

```text
1. .agent 是控制平面，不是文档垃圾桶。
2. 每类事实必须只有一个 SSOT。
3. goalcli 命令必须事务化、幂等、可 dry-run、可恢复。
4. 权限必须最小化，发布权限必须隔离。
5. Evidence 必须长期保存，Release Evidence 永久保存。
6. Release Artifact 必须 hash 化、不可变、可审计。
7. 新增规则必须有机器化路径，否则不能成为 P0/P1。
8. Registry 更新必须加锁，防并发污染。
9. Agent 记忆不是事实源，文件和 Evidence 才是事实源。
10. Goal 过大必须拆，孤儿对象必须清理。
11. 当前阶段目标是 L2，不是一步到 L5。
```

---

# 204. 当前阶段最终可执行目标

现在最合理的目标不是继续无限扩展规则，而是收敛成：

> **Goal Runtime v0.1.0 = L2 证据型机器裁判内核。**

它的完成定义：

```text
main 开发会被阻断
secret 会被阻断
无 Evidence 会被阻断
无 Traceability 会被阻断
无 Release Manifest 会被阻断
所有 Gate 有 reports
所有 reports 可进入 Evidence
Goal Pack 可被 audit
Release Artifact 不可变
Retro Patch 可登记
```

---

# 205. v1.6 结论

Goal Runtime 的长期治理不是靠“写更多规则”，而是靠：

```text
SSOT
Schema
Harness
Evidence
Traceability
Registry
Audit
Release Immutability
Retrospective Patch
Downstream Adoption
```

最终判断：

> **规则只是输入，机器裁判才是系统；Evidence 只是证明，Audit 才是信任；Retrospective 只是复盘，Patch 被采纳才是复利。**
继续补齐 **Goal 完整规则 v1.7：运行时启动、规则编译、状态修复、事故响应、长期防腐化与最终收敛边界**。

这一版重点解决最后一类问题：

```text
系统已经有规则、Harness、Evidence、Traceability、Registry、Release、Retro。
但如何保证它长期不会腐化？
如何在状态不一致时恢复？
如何让规则可安装、可升级、可下游采用？
如何防止 Goal Runtime 自己变成新的复杂系统？
```

---

# 206. Runtime Bootstrap 规则

## RULE-BOOTSTRAP-001：任何仓库接入 Goal Runtime 必须先 bootstrap

接入命令：

```bash
goalcli bootstrap repo --profile xlib-standard
```

必须生成或检查：

```text
.agent/
goalcli.yaml
Makefile gates
scripts/harness/
.githooks/
.github/workflows/
.agent/schemas/
.agent/templates/
.agent/rules/
.agent/harness/fixtures/
```

---

## RULE-BOOTSTRAP-002：bootstrap 必须幂等

重复执行：

```bash
goalcli bootstrap repo
```

不得重复覆盖用户修改。必须输出：

```text
created
updated
skipped
conflict
requires-decision
```

冲突进入：

```text
NEEDS_DECISION
```

---

# 207. Runtime Doctor 规则

## RULE-DOCTOR-001：必须提供一键诊断

```bash
goalcli doctor
```

检查：

```text
目录结构是否完整
Makefile gate 是否存在
Git hooks 是否启用
GitHub workflow 是否存在
Schema 是否有效
Golden fixture 是否通过
Violation fixture 是否失败
Registry 是否一致
是否存在孤儿对象
是否存在 drift
```

---

## RULE-DOCTOR-002：doctor 不能修改状态

`doctor` 只诊断，不修复。

修复必须走：

```bash
goalcli repair
```

---

# 208. Repair 规则

## RULE-REPAIR-001：repair 必须基于 doctor report

禁止盲修。

```bash
goalcli doctor --output reports/doctor.json
goalcli repair --from reports/doctor.json --dry-run
goalcli repair --from reports/doctor.json --apply
```

---

## RULE-REPAIR-002：repair 必须生成修复证据

```text
reports/repair-plan.json
reports/repair-result.json
.agent/violations/VIOLATION-xxx.md
```

---

# 209. Rule Compiler 规则

## RULE-COMPILER-001：Markdown Rule 必须可编译成 Policy Index

输入：

```text
.agent/rules/*.md
.agent/policies/*.yaml
.agent/harness/gates/*.yaml
```

输出：

```text
.agent/compiled/rules-index.json
.agent/compiled/gates-index.json
.agent/compiled/required-checks.json
```

---

## RULE-COMPILER-002：编译失败必须阻断 Release

```bash
make rule-compile
```

失败场景：

```text
规则 ID 重复
规则缺 severity
P0/P1 规则无 enforced_by
Gate 引用不存在命令
Policy 引用不存在 rule_id
```

---

# 210. Rule Coverage 规则

## RULE-COVERAGE-001：P0/P1 规则必须有覆盖率

计算：

```text
Rule Coverage = 有机器 Gate 的 P0/P1 规则数 / P0/P1 总规则数
```

Release 要求：

```text
P0 Rule Coverage = 100%
P1 Rule Coverage >= 90%
```

---

## RULE-COVERAGE-002：无覆盖规则必须降级

如果规则无法机器化：

```text
不能是 P0
不能是 P1
最多 P2/P3
```

---

# 211. Gate Dependency Graph 规则

## RULE-GATE-DAG-001：Gate 必须有依赖图

示例：

```text
schema-check
→ registry-check
→ traceability-check
→ evidence-check
→ pr-check
→ release-check
→ audit-check
```

---

## RULE-GATE-DAG-002：禁止循环依赖

例如：

```text
evidence-check 依赖 release-check
release-check 又依赖 evidence-check
```

必须阻断。

---

# 212. State Reconciliation 规则

## RULE-RECONCILE-001：必须能修复状态不一致

命令：

```bash
goalcli reconcile --goal GOAL-20260603-001
```

检查并修复：

```text
goals.yaml
tasks.yaml
evidence.yaml
traceability.md
issues.yaml
release-manifest.md
execution-log.md
```

---

## RULE-RECONCILE-002：reconcile 不能静默改状态

必须输出：

```text
before
after
reason
confidence
manual_review_required
```

低置信修复必须进入：

```text
NEEDS_HUMAN_APPROVAL
```

---

# 213. Source of Truth Conflict 规则

## RULE-CONFLICT-001：SSOT 冲突必须自动检测

典型冲突：

```text
tasks.yaml 标记 DONE，但 Issue 仍 OPEN
evidence.yaml 有 Evidence，但文件不存在
release manifest 引用不存在 PR
traceability.md 引用不存在 Requirement
README 复制了过期规则
```

---

## RULE-CONFLICT-002：冲突必须有裁决顺序

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

# 214. Runtime Install Profile 规则

## RULE-PROFILE-001：必须区分安装 Profile

```text
minimal
standard
full
downstream
xgo
```

---

## RULE-PROFILE-002：不同 Profile 安装不同能力

| Profile    | 适用            | 必装                                        |
| ---------- | ------------- | ----------------------------------------- |
| minimal    | 小库            | worktree/evidence/schema                  |
| standard   | 普通基础库         | minimal + traceability/pr/release         |
| full       | xlib-standard | standard + retro/adoption/audit/dashboard |
| downstream | 下游库           | adoption-check + contract                 |
| xgo        | x.go          | x.go 架构专用 gates                           |

---

# 215. Local Parity 规则

## RULE-PARITY-001：本地与 CI 必须同源

本地：

```bash
make ci
```

CI：

```yaml
run: make ci
```

禁止 CI 另写一套逻辑。

---

## RULE-PARITY-002：CI-only Gate 必须说明原因

例如：

```text
GitHub token 权限
release tag 权限
branch protection API
```

否则必须本地可复现。

---

# 216. Agent Lease 规则

## RULE-LEASE-001：Agent 执行必须持有 Lease

执行 Task 前：

```bash
goalcli lease acquire --task TASK-001 --agent agent-01
```

Lease 包含：

```yaml
task_id:
agent_id:
worktree:
branch:
expires_at:
heartbeat_at:
```

---

## RULE-LEASE-002：Lease 超时必须释放或接管

```bash
goalcli lease recover
```

防止僵尸 Agent 占用任务。

---

# 217. Agent Heartbeat 规则

## RULE-HEARTBEAT-001：长任务必须写心跳

```bash
goalcli heartbeat --task TASK-001
```

心跳记录：

```text
当前命令
当前状态
最新 Evidence
最近失败
下一步
```

---

## RULE-HEARTBEAT-002：无心跳任务进入 STALE

超过 TTL：

```text
Task state = STALE
```

必须人工或自动恢复。

---

# 218. Worktree Garbage Collection 规则

## RULE-WT-GC-001：必须定期清理 worktree

```bash
goalcli worktree gc --dry-run
goalcli worktree gc --apply
```

可清理：

```text
merged PR 对应 worktree
cancelled task worktree
stale abandoned worktree
```

---

## RULE-WT-GC-002：清理前必须保护未归档 Evidence

如果 worktree 中存在未归档 Evidence，禁止删除。

---

# 219. PR Size Limit 规则

## RULE-PR-SIZE-001：PR 必须控制大小

建议阈值：

```text
文件数 <= 20
核心变更行 <= 800
P0/P1 Requirement <= 3
Task <= 3
```

超过必须拆 PR 或写 Waiver。

---

## RULE-PR-SIZE-002：大 PR 必须加强 Review

大 PR 必须增加：

```text
architecture-review
harness-review
release-impact-review
```

---

# 220. Change Batch 规则

## RULE-BATCH-001：不同风险类型不能混批

禁止一个 PR 同时包含：

```text
规则变更 + CI 变更 + release 变更 + 大量文档重构
```

必须拆成：

```text
rule PR
harness PR
ci PR
docs PR
release PR
```

---

# 221. Release Freeze Deepening 规则

## RULE-FREEZE-003：Freeze 后禁止新增非 blocker 变更

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

---

## RULE-FREEZE-004：Freeze 解除必须有 Decision

如果要扩大 Scope：

```text
Goal state = NEEDS_REPLAN
生成 DEC-xxx
重新跑 Plan Gate
```

---

# 222. Incident Response 规则

## RULE-INCIDENT-001：V0/V1 事故必须开 Incident

```text
INCIDENT-YYYYMMDD-NNN
```

事故包括：

```text
secret 泄漏
错误 stable release
P0 Gate 被绕过
main 被污染
Evidence 丢失
Release artifact 被篡改
```

---

## RULE-INCIDENT-002：Incident 必须进入 Retro Patch

Incident 结束后必须生成：

```text
root cause
timeline
blast radius
fix
new gate
new rule
new fixture
new issue
```

---

# 223. Main Pollution Recovery 规则

## RULE-MAIN-RECOVERY-001：main 被污染必须立即冻结

检测到 main 直接 commit：

```text
冻结 release
冻结 merge
生成 INCIDENT
生成 revert plan
```

---

## RULE-MAIN-RECOVERY-002：恢复必须有 Evidence

```bash
git revert <bad-commit>
make ci
make audit-check
```

生成：

```text
EVID-MAIN-RECOVERY-YYYYMMDD-NNN
```

---

# 224. Evidence Loss Recovery 规则

## RULE-EVID-LOSS-001：Evidence 丢失必须阻断 Release

如果 registry 引用的 Evidence 文件不存在：

```text
release blocked
goal state = INCONSISTENT_STATE
```

---

## RULE-EVID-LOSS-002：Evidence 可重建但必须标记

重建 Evidence 必须标记：

```text
reconstructed: true
source:
  - command rerun
  - CI artifact
  - git log
confidence:
```

Release Evidence 不得低置信重建。

---

# 225. Trust Score 规则

## RULE-TRUST-001：Goal 必须有 Trust Score

Trust Score 由以下组成：

```text
Evidence 完整性
Traceability 完整性
Gate 通过率
Release 可回滚性
Secret 风险
Drift 数量
Open violation 数量
```

---

## RULE-TRUST-002：低 Trust Score 限制发布通道

```text
>= 90 stable
80-89 rc
70-79 beta
<70 blocked
```

---

# 226. Drift Budget 规则

## RULE-DRIFT-BUDGET-001：允许少量低风险漂移，但必须量化

```yaml
drift_budget:
  P0: 0
  P1: 0
  P2: 3
  P3: 10
```

---

## RULE-DRIFT-BUDGET-002：超过预算必须开 Drift Cleanup Goal

```text
GOAL-YYYYMMDD-drift-cleanup
```

---

# 227. Rule Sunset 规则

## RULE-SUNSET-001：规则必须允许退役

退役流程：

```text
mark deprecated
add replacement
notify downstream
wait one release cycle
remove enforcement
archive rule
```

---

## RULE-SUNSET-002：无执行价值规则必须退役

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

# 228. Governance Cadence 规则

## RULE-GOV-CADENCE-001：必须有固定治理节奏

建议：

```text
每周：open violations / blocked goals / pending patches
每双周：drift / rule bloat / fixture health
每月：downstream adoption score / release audit
每季度：rule sunset / maturity review
```

---

## RULE-GOV-CADENCE-002：治理会议必须产出对象

不能只讨论，必须产出：

```text
Decision
Patch
Issue
Rule Change
Deprecation
Migration
```

---

# 229. Dashboard Health 规则

## RULE-DASHBOARD-HEALTH-001：Dashboard 必须显示红线指标

红线：

```text
Open V0 > 0
Open P0 violation > 0
Evidence Coverage < 100%
Traceability Coverage < 100%
P0 Rule Coverage < 100%
main pollution detected
release artifact hash mismatch
```

---

## RULE-DASHBOARD-HEALTH-002：红线指标必须阻断 stable

Dashboard 不是展示板，而是治理入口。

---

# 230. Downstream Promotion 规则

## RULE-PROMOTE-001：xlib-standard 新规则不能直接推全量下游

先按顺序推广：

```text
1. xlib-standard self-check
2. kernel pilot
3. L1 pilot: configx / observex / testkitx
4. L2 pilot: redisx / kafkax / postgresx
5. x.go adoption
```

---

## RULE-PROMOTE-002：每一层通过后才能推广下一层

必须满足：

```text
adoption score >= 80
P0 violation = 0
make ci passed
release-check passed
```

---

# 231. Compatibility Guard 规则

## RULE-COMPAT-GUARD-001：下游兼容性失败不得阻断 xlib-standard 内部发布

但必须限制发布通道：

```text
内部 stable 可以发布
downstream promotion blocked
```

---

## RULE-COMPAT-GUARD-002：Breaking Rule 必须分阶段推进

```text
warn-only
dual-run
blocking
mandatory
```

不能直接从不存在变成强阻断。

---

# 232. Anti-Fragile Retro 规则

## RULE-ANTI-FRAGILE-001：每次失败必须至少增强一个系统部件

失败不能只修当前 bug。

必须至少增强一个：

```text
rule
gate
fixture
schema
template
runbook
dashboard
audit
```

---

## RULE-ANTI-FRAGILE-002：重复失败必须升级 severity

同类失败两次：

```text
P3 → P2
P2 → P1
P1 → P0 candidate
```

---

# 233. 最小长期运行命令集

Goal Runtime 长期运行至少需要：

```bash
goalcli doctor
goalcli repair --dry-run
goalcli reconcile --goal <GOAL-ID>
goalcli rule compile
goalcli rule coverage
goalcli drift check
goalcli registry check
goalcli orphan check
goalcli audit goal --all
goalcli dashboard generate
goalcli worktree gc --dry-run
```

对应 Makefile：

```makefile
.PHONY: doctor
doctor:
	goalcli doctor

.PHONY: rule-compile
rule-compile:
	goalcli rule compile

.PHONY: rule-coverage
rule-coverage:
	goalcli rule coverage

.PHONY: reconcile
reconcile:
	goalcli reconcile --all

.PHONY: maintenance
maintenance:
	$(MAKE) doctor
	$(MAKE) rule-compile
	$(MAKE) registry-check
	$(MAKE) orphan-check
	$(MAKE) drift-check
	$(MAKE) audit-check
	$(MAKE) dashboard
```

---

# 234. v1.7 新增交付清单

```text
[ ] goalcli doctor
[ ] goalcli repair
[ ] goalcli reconcile
[ ] goalcli rule compile
[ ] goalcli rule coverage
[ ] goalcli lease acquire/release/recover
[ ] goalcli heartbeat
[ ] goalcli worktree gc
[ ] .agent/compiled/
[ ] .agent/leases/
[ ] .agent/incidents/
[ ] reports/doctor.json
[ ] reports/rule-coverage.json
[ ] reports/dashboard/index.md
[ ] reports/maintenance-summary.md
```

---

# 235. v1.7 最终铁律

```text
1. Goal Runtime 必须可安装、可诊断、可修复、可升级。
2. 规则必须可编译，P0/P1 必须有覆盖率。
3. 状态不一致必须 reconcile，不能靠人工猜。
4. Agent 执行必须有 lease 和 heartbeat，防止并发污染。
5. main 污染、Evidence 丢失、Release 篡改必须按 Incident 处理。
6. Dashboard 的红线指标必须阻断 stable。
7. 规则必须有 sunset 机制，否则会变成规则债。
8. 下游推广必须分阶段，不能一次性全量推。
9. 每次失败都必须增强系统，不能只修表面问题。
10. 长期目标不是规则越来越多，而是错误越来越难发生。
```

---

# 236. 当前体系到这里的收敛判断

到 v1.7，Goal Runtime 的规则已经覆盖：

```text
目标建模
上下文恢复
Spec / AC
Design / ADR
Task / Issue
Worktree-only
Commit / PR
Evidence
Traceability
Release
Retrospective
Self-improving
Harness
Schema
Registry
Audit
Dashboard
Downstream Adoption
Incident
Governance
Sunset
```

现在应该进入下一步：

```text
从“继续扩规则”
→ 收敛为“可执行落地文档 + goalcli v0.1.0 Issue 拆解 + 文件清单 + 下载版”
```

最终判断：

> **Goal Runtime v1.7 的核心已经足够完整。下一步不应继续堆规则，而应把 v1.0-v1.7 合并成一个可执行规范，并拆成 goalcli v0.1.0、xlib-standard `.agent`、Harness Gates、CI Workflows、下游 adoption 五条落地线。**
继续，但从 **v1.8** 开始不再继续无限扩规则，而是进入 **Rule Freeze + Build Mode**：把 v1.0-v1.7 收敛成可执行规范、文件树、Issue、Gate、CI、Release 与 Adoption。

---

# Goal 完整规则 v1.8：规则冻结、落地打包、执行战役

## 237. Rule Freeze 规则

### RULE-FREEZE-RULE-001：规则体系必须进入冻结态

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

---

### RULE-FREEZE-RULE-002：冻结后的新增规则必须经过 Rule Change

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

---

## 238. Goal Runtime 最终分层

完整体系收敛为 9 层：

```text
L0 Constitution        # 宪法与不可违背铁律
L1 Rules               # 人类可读规则
L2 Policies            # 机器可读策略
L3 Schemas             # 对象结构定义
L4 Harness Gates       # 机器裁判
L5 Runtime Registries  # 状态事实源
L6 Goal Packs          # 每个 Goal 的事实包
L7 Automation          # goalcli / GitHub / CI
L8 Evidence & Audit    # 证据与审计
L9 Self-improving      # 复盘补丁与下游采用
```

核心判断：

> **L0-L4 决定系统不会乱，L5-L8 决定系统可信，L9 决定系统会不会复利。**

---

## 239. xlib-standard 落地包结构

最终应该在 `xlib-standard` 中形成这个运行时包：

```text
.agent/
├── constitution/
│   └── CONSTITUTION.md
│
├── rules/
│   ├── 00-index.md
│   ├── 01-core-rules.md
│   ├── 02-goal-runtime-rules.md
│   ├── 03-worktree-rules.md
│   ├── 04-evidence-rules.md
│   ├── 05-traceability-rules.md
│   ├── 06-pr-rules.md
│   ├── 07-release-rules.md
│   ├── 08-retrospective-rules.md
│   ├── 09-self-improving-rules.md
│   ├── 10-security-rules.md
│   ├── 11-adoption-rules.md
│   ├── 12-governance-rules.md
│   └── 13-sunset-rules.md
│
├── policies/
│   ├── core.yaml
│   ├── worktree.yaml
│   ├── evidence.yaml
│   ├── traceability.yaml
│   ├── pr.yaml
│   ├── release.yaml
│   ├── security.yaml
│   └── adoption.yaml
│
├── schemas/
│   ├── goal.schema.json
│   ├── spec.schema.json
│   ├── task.schema.json
│   ├── evidence.schema.json
│   ├── traceability.schema.json
│   ├── release.schema.json
│   ├── retrospective.schema.json
│   └── patch.schema.json
│
├── harness/
│   ├── gates/
│   ├── fixtures/
│   │   ├── golden/
│   │   └── violations/
│   └── reports/
│
├── templates/
│   ├── goal-template.md
│   ├── task-template.md
│   ├── issue-template.md
│   ├── pr-template.md
│   ├── evidence-template.md
│   ├── release-manifest-template.md
│   └── retrospective-template.md
│
├── registries/
│   ├── goals.yaml
│   ├── tasks.yaml
│   ├── evidence.yaml
│   ├── patches.yaml
│   ├── adoption.yaml
│   ├── debt.yaml
│   └── issue-candidates.yaml
│
├── goals/
├── patches/
├── violations/
├── incidents/
├── migrations/
├── adoption/
├── locks/
├── leases/
├── compiled/
├── glossary.md
└── ownership.yaml
```

根目录必须配套：

```text
Makefile
goalcli.yaml
scripts/harness/
scripts/git/
.github/workflows/
.githooks/
release/
reports/
```

---

## 240. goalcli v0.1.0 最终定位

### 不是：

```text
自动写代码平台
全自动 Agent 编排系统
自动发布 stable 的机器人
复杂项目管理平台
```

### 而是：

```text
Goal Runtime 的最小机器裁判内核
```

它先解决 7 个硬问题：

```text
1. main 不能被污染
2. secret 不能泄漏
3. Goal Pack 必须结构化
4. Evidence 必须存在且可验证
5. Traceability 不能断链
6. Release 必须有 Manifest 和 Rollback
7. Audit 必须能重建事实链
```

---

## 241. goalcli v0.1.0 命令边界

第一版只做这些：

```bash
goalcli bootstrap repo
goalcli doctor
goalcli schema validate --all
goalcli worktree check
goalcli evidence check
goalcli traceability check
goalcli secret check
goalcli pr check
goalcli release check
goalcli retro check
goalcli registry check
goalcli orphan check
goalcli audit goal --goal <GOAL-ID>
goalcli dashboard generate
```

暂缓：

```bash
goalcli code generate
goalcli pr merge
goalcli release publish stable
goalcli multi-repo rewrite
goalcli auto-fix-all
```

原则：

> **先做不会错的裁判，再做提效的自动化。**

---

## 242. P0 Harness Gate 最小集合

第一阶段必须先实现 8 个 Gate：

| Gate                 | 目标             | 阻断级别 |
| -------------------- | -------------- | ---- |
| `worktree-check`     | 防 main 开发      | P0   |
| `secret-check`       | 防密钥泄漏          | P0   |
| `schema-check`       | 防对象结构无效        | P0   |
| `evidence-check`     | 防无证据完成         | P0   |
| `traceability-check` | 防需求断链          | P0   |
| `pr-check`           | 防无审查合并         | P1   |
| `release-check`      | 防无 Manifest 发布 | P0   |
| `retro-check`        | 防无复利闭环         | P1   |

Makefile 最小入口：

```makefile
.PHONY: ci
ci:
	$(MAKE) worktree-check
	$(MAKE) secret-check
	$(MAKE) schema-check
	$(MAKE) evidence-check
	$(MAKE) traceability-check

.PHONY: release-check-all
release-check-all:
	$(MAKE) ci
	$(MAKE) pr-check
	$(MAKE) release-check
	$(MAKE) retro-check
```

---

## 243. v0.1.0 Issue 拆解最终版

### EPIC-001：Goal Runtime 裁判内核

```text
目标：建立 Goal Runtime v0.1.0 的最小机器裁判系统。
```

#### ISSUE-001：初始化 `.agent` 控制平面

```text
交付：
- .agent/constitution/
- .agent/rules/
- .agent/policies/
- .agent/schemas/
- .agent/templates/
- .agent/harness/
- .agent/registries/

验收：
- 目录完整
- README / index 指向 SSOT
- 无重复事实源
```

---

#### ISSUE-002：建立 `goalcli.yaml`

```text
交付：
- runtime 配置
- gate 配置
- worktree 配置
- evidence 配置
- release 配置
- security 配置

验收：
- goalcli doctor 可读取配置
- 配置缺字段会失败
```

---

#### ISSUE-003：实现 `worktree-check`

```text
目标：
- 禁止 main / master 开发
- 强制 worktree-only

验收：
- main 分支执行失败
- 合法 worktree 执行通过
- reports/worktree-check.json 生成
```

---

#### ISSUE-004：实现 `secret-check`

```text
目标：
- 防止 token/password/secret/private key 进入代码、文档、Evidence、Release

验收：
- secret violation fixture 失败
- masked example 通过
- reports/secret-check.json 生成
```

---

#### ISSUE-005：实现 `schema-check`

```text
目标：
- Goal / Task / Evidence / Release 对象机器校验

验收：
- golden goal pack 通过
- 缺字段 fixture 失败
- reports/schema-check.json 生成
```

---

#### ISSUE-006：实现 `evidence-check`

```text
目标：
- Evidence 必须绑定 task、command、cwd、branch、commit、exit_code、artifact

验收：
- 缺 Evidence 失败
- artifact 不存在失败
- Release Evidence 无 commit 失败
- reports/evidence-check.json 生成
```

---

#### ISSUE-007：实现 `traceability-check`

```text
目标：
- Requirement → AC → Task → Evidence 完整链路

验收：
- 缺 Requirement 失败
- 缺 AC 失败
- 缺 Task 失败
- 缺 Evidence 失败
- golden 链路通过
```

---

#### ISSUE-008：实现 `pr-check`

```text
目标：
- PR 必须包含 Goal / Issue / Requirement / AC / Evidence / Risk / Rollback

验收：
- 缺 Evidence 的 Ready PR 失败
- Draft PR 可警告
- reports/pr-check.json 生成
```

---

#### ISSUE-009：实现 `release-check`

```text
目标：
- Release 必须有 Manifest / Evidence Summary / Rollback / Known Issues

验收：
- 缺 manifest 失败
- 缺 rollback 失败
- Evidence coverage <100% 失败
```

---

#### ISSUE-010：实现 `retro-check`

```text
目标：
- Retrospective 必须产生 Prompt/Harness/Rule Patch 候选

验收：
- 无 Patch 候选失败
- Lite Mode 可降级为 warning
```

---

#### ISSUE-011：实现 `audit goal`

```text
目标：
- 从 Goal Pack 重建完整事实链

输出：
- Goal Summary
- Requirement Coverage
- Task Coverage
- Evidence Coverage
- Release Readiness
- Open Risks
- Final Score

验收：
- reports/audit/<GOAL-ID>.md 生成
```

---

#### ISSUE-012：接入 Git Hooks

```text
交付：
- .githooks/pre-commit
- .githooks/pre-push
- scripts/git/install-hooks.sh

验收：
- main commit 被阻断
- secret 被阻断
- evidence 缺失时 pre-push 失败
```

---

#### ISSUE-013：接入 GitHub Actions

```text
交付：
- worktree-guard.yml
- goal-gates.yml
- release-gate.yml
- secret-scan.yml

验收：
- PR 自动运行 make ci
- Release 自动运行 release-check
- direct push main 被阻断
```

---

#### ISSUE-014：建立 Golden / Violation Fixtures

```text
交付：
- golden/minimal-goal-pack
- violations/missing-evidence
- violations/missing-traceability
- violations/main-branch-dev
- violations/secret-leak
- violations/missing-release-manifest

验收：
- golden 全部通过
- violation 全部失败且失败原因正确
```

---

#### ISSUE-015：发布 v0.1.0

```text
交付：
- release/REL-<date>-goalcli-v0.1.0/manifest.md
- changelog.md
- evidence-summary.md
- rollback.md
- audit report
- checksums.txt

验收：
- release-check 通过
- audit score >= 90
- no P0 violation
```

---

## 244. 1 天 / 7 天 / 30 天行动计划

### 1 天：建立不可绕过基线

```text
[ ] 创建 .agent 基础目录
[ ] 写 CONSTITUTION.md
[ ] 写 worktree-rules / evidence-rules / security-rules
[ ] 增加 goalcli.yaml
[ ] 增加 Makefile gate 空壳
[ ] 实现 no-main-dev.sh
[ ] 增加 pre-commit / pre-push
[ ] 增加 worktree-guard.yml
```

完成后必须达到：

```text
main 不能直接开发
worktree-only 规则开始生效
```

---

### 7 天：建立 L2 裁判内核

```text
[ ] 实现 schema-check
[ ] 实现 secret-check
[ ] 实现 evidence-check
[ ] 实现 traceability-check
[ ] 建立 golden fixture
[ ] 建立 violation fixtures
[ ] 接入 goal-gates.yml
[ ] 生成第一份 audit report
```

完成后必须达到：

```text
无 Evidence 不可 DONE
Traceability 断链不可 Release
Golden Goal Pack 可通过
Violation Fixtures 可失败
```

---

### 30 天：形成 v0.1.0 Release

```text
[ ] 实现 pr-check
[ ] 实现 release-check
[ ] 实现 retro-check
[ ] 实现 audit goal
[ ] 生成 dashboard
[ ] 生成 release artifact
[ ] 发布 goalcli v0.1.0
[ ] 选择 kernel 作为第一个 downstream adoption pilot
```

完成后必须达到：

```text
xlib-standard 自身达到 Goal Runtime L2
kernel 可开始试点采用
```

---

## 245. 最终评分标准

### 90 分以上

```text
worktree-only 生效
secret-check 生效
schema-check 生效
evidence-check 生效
traceability-check 生效
release-check 生效
audit goal 可运行
Golden/Violation fixtures 完整
Release Manifest 完整
```

### 80-89 分

```text
核心 Gate 生效
Evidence/Traceability 基本完整
但 Retro / Dashboard / Adoption 不完整
```

### 70-79 分

```text
文档完整，但机器 Gate 不足
可作为设计稿，不可作为执行系统
```

### 60-69 分

```text
规则多，但 SSOT 混乱
Evidence 不稳定
容易形成文档债
```

### <60 分

```text
普通文档，不是 Goal Runtime
```

---

## 246. 最终推荐路径

现在最优路径不是继续扩写规则，而是立即进入：

```text
Phase 1：冻结规则
Phase 2：整理 SSOT 文档
Phase 3：实现 P0 Gates
Phase 4：建立 Golden / Violation fixtures
Phase 5：发布 goalcli v0.1.0
Phase 6：kernel downstream adoption pilot
Phase 7：Retro 后再决定 v0.2.0 自动化范围
```

---

# 247. v1.8 最终铁律

```text
1. 规则到 v1.8 必须冻结，不能继续无边界膨胀。
2. 下一步必须从规则写作转为 Gate 实现。
3. goalcli v0.1.0 只做裁判内核，不做全自动平台。
4. P0 Gate 必须先于 Issue/PR/Release 自动化。
5. Golden/Violation fixtures 是系统可信度的根。
6. xlib-standard 先自证，再推广 kernel。
7. 没有 audit report，不算真正完成。
8. 没有 downstream adoption pilot，不算标准工厂。
```

最终结论：

> **Goal Runtime v1.8 的正确动作是冻结规则、实现最小裁判内核、发布 v0.1.0，并用 kernel 做第一次下游采用验证。继续增加规则的边际收益已经下降，开始实现 Gate 的收益最高。**
继续，但这里开始进入 **v1.9：执行打包版**。

从 v1.9 起，不再新增大批规则，而是把前面 v1.0-v1.8 压缩成 **真正可执行的落地方案**：哪些文件先建、哪些 Gate 先实现、哪些 Issue 先开、哪些命令必须跑、如何验收。

---

# Goal Runtime v1.9：执行打包版

## 248. 当前阶段定位

现在的目标不是继续“补规则”，而是进入：

```text
Rule Freeze
→ Implementation Pack
→ P0 Gate First
→ Golden / Violation Fixtures
→ v0.1.0 Release
→ kernel Adoption Pilot
```

最终要完成的不是一份更大的文档，而是一个能运行的系统：

```text
.agent 控制平面
+ goalcli 裁判内核
+ Harness Gates
+ Evidence Protocol
+ Traceability Matrix
+ Worktree-only Guard
+ CI Required Checks
+ Release Manifest
+ Retrospective Patch
```

---

# 249. 最终目标重新收敛

## 当前阶段目标

> 把 xlib-standard 从“规则文档集合”升级为“可机器裁判的 Goal Runtime 标准源”。

更具体地说：

```text
1. main 不能直接开发
2. 缺 Evidence 不能 DONE
3. 缺 Traceability 不能 Release
4. 缺 Release Manifest 不能发布
5. Secret 不能进入代码 / 文档 / Evidence / Release
6. Goal Pack 必须可审计
7. Golden Goal Pack 必须通过
8. Violation Fixtures 必须失败
9. 所有 Gate 必须输出 reports
10. goalcli v0.1.0 必须能执行最小裁判闭环
```

---

# 250. v0.1.0 必须先完成的 5 条主线

## Track A：控制平面

```text
目标：建立 .agent 作为 Goal Runtime 控制平面。
```

交付：

```text
.agent/constitution/CONSTITUTION.md
.agent/rules/
.agent/policies/
.agent/schemas/
.agent/templates/
.agent/harness/
.agent/registries/
.agent/goals/
.agent/glossary.md
.agent/ownership.yaml
```

验收：

```bash
make schema-check
make registry-check
make docs-check
```

---

## Track B：P0 安全门禁

```text
目标：先防止系统犯不可接受的错误。
```

交付：

```text
worktree-check
secret-check
schema-check
```

验收：

```bash
make worktree-check
make secret-check
make schema-check
```

必须证明：

```text
main 分支开发失败
合法 worktree 通过
secret fixture 失败
golden schema 通过
invalid schema 失败
```

---

## Track C：证据与追踪门禁

```text
目标：防止“做了但无法证明”。
```

交付：

```text
evidence-check
traceability-check
evidence registry
traceability matrix
```

验收：

```bash
make evidence-check
make traceability-check
```

必须证明：

```text
缺 Evidence 失败
Evidence artifact 不存在失败
Release Evidence 无 commit 失败
Req → AC → Task → Evidence 断链失败
Golden Goal Pack 完整链路通过
```

---

## Track D：PR / Release 门禁

```text
目标：防止断链合并和断链发布。
```

交付：

```text
pr-check
release-check
release manifest template
rollback template
known issues template
```

验收：

```bash
make pr-check
make release-check
```

必须证明：

```text
Ready PR 缺 Evidence 失败
PR 缺 Rollback 失败
Release 缺 Manifest 失败
Release 缺 Rollback 失败
Evidence coverage < 100% 失败
```

---

## Track E：复利与审计

```text
目标：让每次 Goal 都能被审计，并生成下一轮改进资产。
```

交付：

```text
retro-check
audit goal
patch registry
issue-candidates registry
dashboard report
```

验收：

```bash
make retro-check
make audit-check
make dashboard
```

必须证明：

```text
Retro 无 Patch 候选失败
audit 能重建 Goal 事实链
dashboard 能显示 Evidence / Traceability / Gate 状态
```

---

# 251. 第一批文件创建顺序

不要一次性铺满所有目录。按这个顺序落地。

## 第 1 批：不可绕过基线

```text
01. CONSTITUTION.md
02. goalcli.yaml
03. Makefile
04. scripts/harness/no-main-dev.sh
05. .githooks/pre-commit
06. .githooks/pre-push
07. scripts/git/install-hooks.sh
08. .github/workflows/worktree-guard.yml
```

目标：

```text
先让 main 不能被污染。
```

---

## 第 2 批：对象结构

```text
09. .agent/schemas/goal.schema.json
10. .agent/schemas/task.schema.json
11. .agent/schemas/evidence.schema.json
12. .agent/schemas/release.schema.json
13. .agent/templates/goal-template.md
14. .agent/templates/task-template.md
15. .agent/templates/evidence-template.md
16. .agent/templates/release-manifest-template.md
```

目标：

```text
让 Goal / Task / Evidence / Release 可以被机器读取。
```

---

## 第 3 批：核心规则 SSOT

```text
17. .agent/rules/00-index.md
18. .agent/rules/01-core-rules.md
19. .agent/rules/02-worktree-rules.md
20. .agent/rules/03-evidence-rules.md
21. .agent/rules/04-traceability-rules.md
22. .agent/rules/05-release-rules.md
23. .agent/rules/06-security-rules.md
24. .agent/rules/07-retrospective-rules.md
```

目标：

```text
先只写核心规则，不继续无限拆文档。
```

---

## 第 4 批：Golden / Violation Fixtures

```text
25. .agent/harness/fixtures/golden/minimal-goal-pack/
26. .agent/harness/fixtures/violations/missing-evidence/
27. .agent/harness/fixtures/violations/missing-traceability/
28. .agent/harness/fixtures/violations/main-branch-dev/
29. .agent/harness/fixtures/violations/secret-leak/
30. .agent/harness/fixtures/violations/missing-release-manifest/
```

目标：

```text
先证明系统能识别正确与错误。
```

---

## 第 5 批：CI 与 Release

```text
31. .github/workflows/goal-gates.yml
32. .github/workflows/secret-scan.yml
33. .github/workflows/release-gate.yml
34. release/REL-YYYYMMDD-goalcli-v0.1.0/manifest.md
35. release/REL-YYYYMMDD-goalcli-v0.1.0/evidence-summary.md
36. release/REL-YYYYMMDD-goalcli-v0.1.0/rollback.md
37. release/REL-YYYYMMDD-goalcli-v0.1.0/checksums.txt
```

目标：

```text
形成第一个可发布、可审计的 v0.1.0。
```

---

# 252. Makefile 最小最终版

第一版不要复杂，必须稳定。

```makefile
.PHONY: worktree-check
worktree-check:
	goalcli worktree check

.PHONY: secret-check
secret-check:
	goalcli secret check

.PHONY: schema-check
schema-check:
	goalcli schema validate --all

.PHONY: evidence-check
evidence-check:
	goalcli evidence check

.PHONY: traceability-check
traceability-check:
	goalcli traceability check

.PHONY: pr-check
pr-check:
	goalcli pr check

.PHONY: release-check
release-check:
	goalcli release check

.PHONY: retro-check
retro-check:
	goalcli retro check

.PHONY: registry-check
registry-check:
	goalcli registry check

.PHONY: orphan-check
orphan-check:
	goalcli orphan check

.PHONY: audit-check
audit-check:
	goalcli audit goal --all

.PHONY: dashboard
dashboard:
	goalcli dashboard generate

.PHONY: ci
ci:
	$(MAKE) worktree-check
	$(MAKE) secret-check
	$(MAKE) schema-check
	$(MAKE) evidence-check
	$(MAKE) traceability-check

.PHONY: release-check-all
release-check-all:
	$(MAKE) ci
	$(MAKE) pr-check
	$(MAKE) release-check
	$(MAKE) retro-check
	$(MAKE) audit-check
```

---

# 253. goalcli v0.1.0 内核模块

建议 Go 实现结构：

```text
cmd/goalcli/
└── main.go

internal/
├── config/
│   └── config.go
├── report/
│   └── report.go
├── gitutil/
│   └── git.go
├── schema/
│   └── checker.go
├── worktree/
│   └── checker.go
├── secret/
│   └── checker.go
├── evidence/
│   └── checker.go
├── traceability/
│   └── checker.go
├── pr/
│   └── checker.go
├── release/
│   └── checker.go
├── retro/
│   └── checker.go
├── registry/
│   └── checker.go
├── orphan/
│   └── checker.go
├── audit/
│   └── audit.go
└── dashboard/
    └── dashboard.go
```

v0.1.0 不做复杂插件系统，先用清晰模块。

---

# 254. Checker 统一接口

所有 Gate 用同一种结果结构。

```go
type Status string

const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	StatusWarn   Status = "warn"
)

type Severity string

const (
	SeverityP0 Severity = "P0"
	SeverityP1 Severity = "P1"
	SeverityP2 Severity = "P2"
	SeverityP3 Severity = "P3"
)

type CheckError struct {
	RuleID  string `json:"rule_id"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type CheckResult struct {
	Checker   string       `json:"checker"`
	Status    Status       `json:"status"`
	Severity  Severity     `json:"severity"`
	GoalID    string       `json:"goal_id,omitempty"`
	Errors    []CheckError `json:"errors"`
	Warnings  []CheckError `json:"warnings"`
	Artifacts []string     `json:"artifacts"`
	Timestamp string       `json:"timestamp"`
}

type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}
```

---

# 255. Gate 输出标准

每个 Gate 必须输出：

```text
reports/<gate>.json
reports/<gate>.md 或 reports/<gate>.txt
```

例如：

```text
reports/worktree-check.json
reports/worktree-check.txt
reports/evidence-check.json
reports/traceability-check.json
reports/release-check.json
```

所有报告必须能进入 Evidence。

---

# 256. worktree-check 合约

## 输入

```text
当前 git repo
当前分支
当前 repo root
goalcli.yaml worktree_root
```

## 检查

```text
当前分支不是 main/master
当前路径在 worktree_root 下
不是 primary main worktree
不是 detached HEAD
```

## 失败

```text
exit code = 6
rule_id = RULE-WORKTREE-001
severity = P0
```

## 输出

```text
reports/worktree-check.json
reports/worktree-check.txt
```

---

# 257. secret-check 合约

## 输入

```text
git tracked files
.agent/security/secret-allowlist.yaml
```

## 检查范围

```text
*.go
*.rs
*.sh
*.md
*.yaml
*.yml
*.json
.env*
.github/
.agent/
release/
reports/
```

## 检查关键词

```text
token
password
secret
private_key
access_key
authorization
cookie
```

## 失败

```text
exit code = 7
rule_id = RULE-SECRET-001
severity = P0
```

---

# 258. schema-check 合约

## 输入

```text
.agent/schemas/*.schema.json
.agent/goals/*/*.yaml
.agent/registries/*.yaml
```

## 检查

```text
Goal Pack 是否符合 schema
Task 是否符合 schema
Evidence 是否符合 schema
Release 是否符合 schema
Registry 是否符合 schema
```

## 失败

```text
exit code = 3
rule_id = RULE-SCHEMA-001
severity = P0
```

---

# 259. evidence-check 合约

## 输入

```text
.agent/registries/evidence.yaml
.agent/goals/<GOAL-ID>/evidence/
reports/
```

## 检查

```text
Evidence ID 存在
Goal ID 存在
Task ID 存在
command 存在
cwd 存在
branch 存在
exit_code 存在
artifact path 存在且非空
Release Evidence 必须有 commit hash
```

## 失败

```text
exit code = 4
rule_id = RULE-EVIDENCE-001
severity = P0
```

---

# 260. traceability-check 合约

## 输入

```text
spec.yaml
tasks.yaml
evidence.yaml
traceability.md 或 traceability.yaml
```

## 检查链路

```text
Requirement → AC
AC → Task
Task → Evidence
Evidence → Artifact
Status 与 Evidence 一致
```

## 失败

```text
exit code = 5
rule_id = RULE-TRACE-001
severity = P0
```

---

# 261. release-check 合约

## 输入

```text
release/<REL-ID>/manifest.md
release/<REL-ID>/evidence-summary.md
release/<REL-ID>/rollback.md
release/<REL-ID>/known-issues.md
.agent/registries/evidence.yaml
traceability.md
```

## 检查

```text
manifest 存在
evidence summary 存在
rollback 存在
known issues 存在
Evidence coverage = 100%
Traceability coverage = 100%
P0/P1 open risk = 0
checksums.txt 存在
```

## 失败

```text
exit code = 8
rule_id = RULE-RELEASE-001
severity = P0
```

---

# 262. retro-check 合约

## 输入

```text
.agent/goals/<GOAL-ID>/retrospective.md
.agent/registries/patches.yaml
.agent/registries/issue-candidates.yaml
```

## 检查

```text
What Worked 存在
What Failed 存在
Root Cause 存在
Prompt Patch 候选存在
Harness Patch 候选存在
Rule Patch 候选存在
New Issue Candidate 存在
```

## Lite Mode 例外

Lite Mode 可降级为 warning，但必须说明原因。

---

# 263. audit goal 合约

## 输入

```text
.agent/goals/<GOAL-ID>/
.agent/registries/
release/
reports/
```

## 输出

```text
reports/audit/<GOAL-ID>.md
reports/audit/<GOAL-ID>.json
```

必须包含：

```text
Goal Summary
State Timeline
Requirement Coverage
Task Coverage
Evidence Coverage
Traceability Coverage
PR Summary
Release Summary
Risk Summary
Retrospective Summary
Open Gaps
Score
Decision
```

评分小于 90：

```text
不允许 stable release
```

---

# 264. GitHub Actions 最小版

## worktree-guard.yml

```yaml
name: Worktree Guard

on:
  push:
  pull_request:

jobs:
  guard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Block direct push to main
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        run: |
          echo "Direct push to main is forbidden."
          exit 1
```

---

## goal-gates.yml

```yaml
name: Goal Gates

on:
  pull_request:

jobs:
  gates:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install goalcli
        run: |
          echo "Install goalcli v0.1.0"

      - name: Run CI Gates
        run: |
          make ci

      - name: Run PR Gate
        run: |
          make pr-check
```

---

## release-gate.yml

```yaml
name: Release Gate

on:
  workflow_dispatch:
  push:
    tags:
      - "v*"

jobs:
  release-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install goalcli
        run: |
          echo "Install goalcli v0.1.0"

      - name: Run Release Gates
        run: |
          make release-check-all
```

---

# 265. Git worktree 标准 Runbook

## 创建任务 worktree

```bash
cd ~/code/xlib-standard

git fetch origin

git worktree add \
  ~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001 \
  -b goal/GOAL-20260603-001/TASK-001 \
  origin/main

cd ~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001

make worktree-check
```

---

## 提交前

```bash
make worktree-check
make secret-check
make schema-check
make evidence-check
make traceability-check
```

---

## PR 前

```bash
make ci
make pr-check
```

---

## Release 前

```bash
make release-check-all
```

---

## 合并后清理

```bash
git worktree remove ~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001
git worktree prune
```

---

# 266. Commit 标准

最终 commit 必须包含：

```text
<type>(<scope>): <summary>

Goal: GOAL-YYYYMMDD-NNN
Task: TASK-GOAL-YYYYMMDD-NNN-XXX
Issue: #123
Evidence: EVID-TASK-GOAL-YYYYMMDD-NNN-XXX-YYYYMMDD-001
```

示例：

```text
feat(harness): add worktree-only gate

Goal: GOAL-20260603-001
Task: TASK-GOAL-20260603-001-003
Issue: #123
Evidence: EVID-TASK-GOAL-20260603-001-003-20260603-001
```

---

# 267. PR 模板最终字段

````md
# PR: <title>

## Goal
- GOAL:

## Related Issues
- Closes #

## Requirements Covered
| Requirement | AC | Task | Evidence |
|---|---|---|---|

## Changes
-

## Tests
```bash
make ci
make pr-check
````

## Evidence

*

## Risk

*

## Rollback

*

## Worktree Evidence

* Branch:
* Worktree path:
* `make worktree-check`: PASS/FAIL

## Checklist

* [ ] Goal linked
* [ ] Issue linked
* [ ] Requirement mapped
* [ ] AC verified
* [ ] Evidence attached
* [ ] Traceability updated
* [ ] Risk documented
* [ ] Rollback documented
* [ ] CI passed

````

---

# 268. Release Manifest 最终字段

```md
# Release Manifest

## Release
- Release ID:
- Version:
- Date:
- Channel:
- Commit:
- Tag:

## Included Goals
-

## Included Issues
-

## Included PRs
-

## Changes
-

## Evidence Summary
| Task | Evidence | Status |
|---|---|---|

## Test Summary
-

## Traceability Coverage
-

## Risk Summary
-

## Compatibility
-

## Migration Notes
-

## Rollback Plan
-

## Known Issues
-

## Checksums
-

## Retrospective
-
````

---

# 269. v0.1.0 发布前最终验收清单

```text
[ ] `.agent` 控制平面存在
[ ] `goalcli.yaml` 存在
[ ] `Makefile` 有所有 Gate target
[ ] `worktree-check` 可阻断 main
[ ] `secret-check` 可阻断 secret
[ ] `schema-check` 可校验 Goal Pack
[ ] `evidence-check` 可阻断缺证据
[ ] `traceability-check` 可阻断断链
[ ] `pr-check` 可校验 Ready PR
[ ] `release-check` 可校验 Release Manifest
[ ] `retro-check` 可校验 Patch 候选
[ ] `audit goal` 可生成报告
[ ] Golden fixture 全通过
[ ] Violation fixtures 全失败
[ ] Git hooks 已可安装
[ ] GitHub Actions 已接入
[ ] Branch protection 已启用
[ ] Release artifact 已生成
[ ] Audit Score >= 90
[ ] P0 violation = 0
```

---

# 270. kernel Adoption Pilot

xlib-standard 自证通过后，下一步不是全量推广，而是先拿 `kernel` 做试点。

## 目标

```text
验证 xlib-standard Goal Runtime 是否能被 L0 内核库采用。
```

## kernel 需要采用

```text
.agent/rules/
.agent/templates/
.agent/schemas/
Makefile gates
scripts/harness/
.github/workflows/
Evidence Protocol
Release Manifest
```

## kernel 不一定马上采用

```text
dashboard
multi-repo orchestration
full downstream scoring
complex Agent Team
```

## kernel Adoption 验收

```bash
make ci
make evidence-check
make traceability-check
make release-check
```

必须生成：

```text
.agent/adoption/adoption-manifest.md
reports/adoption-check.json
```

评分：

```text
>= 80：可用采用
>= 90：完整采用
< 80：不允许推广到 L1/L2
```

---

# 271. 最终执行顺序

```text
1. Freeze rules
2. Create .agent control plane
3. Add goalcli.yaml
4. Add Makefile gates
5. Implement worktree-check
6. Implement secret-check
7. Implement schema-check
8. Implement evidence-check
9. Implement traceability-check
10. Add golden / violation fixtures
11. Add PR / release / retro checks
12. Add audit goal
13. Add Git hooks
14. Add GitHub Actions
15. Generate v0.1.0 release artifact
16. Run self-audit
17. Pilot adoption in kernel
```

---

# 272. 当前最小可行行动 MVA

今天就做这 8 件事：

```text
1. 建 .agent 最小目录
2. 写 CONSTITUTION.md
3. 写 goalcli.yaml
4. 写 Makefile gate target
5. 写 no-main-dev.sh
6. 写 pre-commit / pre-push
7. 写 worktree-guard.yml
8. 创建第一个合法 worktree，并证明 main 开发被阻断
```

这 8 件完成后，系统从“文档规则”进入“执行规则”。

---

# 273. v1.9 最终铁律

```text
1. 不再扩规则，开始实现 Gate。
2. 不先做自动化 Issue / PR / Release，先做裁判。
3. 不先推广下游，先让 xlib-standard 自证。
4. 不把 goalcli v0.1.0 做成大平台，只做最小裁判内核。
5. 不允许 main 开发，这是所有 Agent 自动化的地基。
6. 不允许无 Evidence DONE，这是所有交付可信度的地基。
7. 不允许 Traceability 断链，这是所有 Goal 可审计性的地基。
8. 不允许 Release 无 Manifest，这是所有标准工厂扩张的地基。
```

最终结论：

> **现在真正该执行的不是继续写第 100 条规则，而是把 worktree-check、secret-check、schema-check、evidence-check、traceability-check 五个 Gate 做出来。它们一旦生效，Goal Runtime 就从“治理文档”变成“工程操作系统内核”。**
继续，进入 **Goal Runtime v2.0：Build Mode / 可实现版本**。

从这里开始，重点不再是“规则更多”，而是把前面所有规则变成 **可编码、可测试、可验收、可发布** 的工程任务。

---

# Goal Runtime v2.0：最小裁判内核实现方案

## 274. v2.0 定位

v2.0 的目标：

> **把 Goal Runtime 从规则体系推进到可运行的 goalcli v0.1.0 裁判内核。**

v2.0 不追求全自动 Agent，不追求复杂平台，不追求一次覆盖所有下游库。

只做一件事：

```text
建立 xlib-standard 的最小机器裁判系统。
```

也就是：

```text
main 污染 → 阻断
secret 泄漏 → 阻断
Goal Pack 无结构 → 阻断
Evidence 缺失 → 阻断
Traceability 断链 → 阻断
Release 无 Manifest → 阻断
Retro 无 Patch → 警告或阻断
```

---

# 275. goalcli v0.1.0 最小产品定义

## 不是平台

```text
goalcli v0.1.0 不做：
- 自动写代码
- 自动修复所有问题
- 自动合并 PR
- 自动 stable 发布
- 自动跨仓库批量改造
- 复杂多 Agent 调度
```

## 是裁判内核

```text
goalcli v0.1.0 必须做：
- 读取 goalcli.yaml
- 执行 Gate
- 输出标准 reports
- 识别违规
- 生成 audit
- 给 CI / Agent / 人类统一判断结果
```

一句话：

> **goalcli v0.1.0 = Goal Runtime 的 `make test` + `make audit` + `policy gate`。**

---

# 276. goalcli CLI 最小命令树

```text
goalcli
├── bootstrap
│   └── repo
├── doctor
├── worktree
│   └── check
├── secret
│   └── check
├── schema
│   └── validate
├── evidence
│   └── check
├── traceability
│   └── check
├── pr
│   └── check
├── release
│   └── check
├── retro
│   └── check
├── registry
│   └── check
├── orphan
│   └── check
├── audit
│   └── goal
└── dashboard
    └── generate
```

暂时不要做更多命令。

---

# 277. goalcli 项目结构

建议在 `xlib-standard` 内先作为本地工具实现：

```text
tools/goalcli/
├── go.mod
├── cmd/
│   └── goalcli/
│       └── main.go
├── internal/
│   ├── app/
│   ├── config/
│   ├── report/
│   ├── gitutil/
│   ├── fsutil/
│   ├── checker/
│   ├── worktree/
│   ├── secret/
│   ├── schema/
│   ├── evidence/
│   ├── traceability/
│   ├── pr/
│   ├── release/
│   ├── retro/
│   ├── registry/
│   ├── orphan/
│   ├── audit/
│   └── dashboard/
└── testdata/
    ├── golden/
    └── violations/
```

后续稳定后，再考虑独立仓库。

---

# 278. 第一版 Go 模块边界

## `config`

职责：

```text
读取 goalcli.yaml
解析 worktree_root
解析 reports_dir
解析 severity_policy
解析 required_gates
```

## `report`

职责：

```text
统一 CheckResult
统一 JSON report 输出
统一 Markdown/Text report 输出
统一 exit code
```

## `gitutil`

职责：

```text
获取当前分支
获取 repo root
获取 commit hash
判断是否 detached HEAD
判断是否 main/master
判断 tracked files
```

## `checker`

职责：

```text
定义 Checker interface
聚合执行多个 checker
生成 ci-summary
```

## 各 Gate 模块

```text
worktree
secret
schema
evidence
traceability
pr
release
retro
registry
orphan
```

每个模块只做一件事。

---

# 279. 统一 CheckResult 契约

所有 Gate 返回同一种结果。

```go
type Status string

const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	StatusWarn   Status = "warn"
)

type Severity string

const (
	SeverityP0 Severity = "P0"
	SeverityP1 Severity = "P1"
	SeverityP2 Severity = "P2"
	SeverityP3 Severity = "P3"
)

type Finding struct {
	RuleID  string `json:"rule_id"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type CheckResult struct {
	Checker   string    `json:"checker"`
	Status    Status    `json:"status"`
	Severity  Severity  `json:"severity"`
	GoalID    string    `json:"goal_id,omitempty"`
	Errors    []Finding `json:"errors"`
	Warnings  []Finding `json:"warnings"`
	Artifacts []string  `json:"artifacts"`
	Timestamp string    `json:"timestamp"`
}
```

---

# 280. 统一退出码

```text
0  PASS
1  GENERAL_FAILURE
2  POLICY_VIOLATION
3  SCHEMA_INVALID
4  EVIDENCE_MISSING
5  TRACEABILITY_BROKEN
6  WORKTREE_INVALID
7  SECRET_DETECTED
8  RELEASE_BLOCKED
9  NEEDS_HUMAN_APPROVAL
10 INCONSISTENT_STATE
```

CI、Agent、Makefile 都按退出码处理。

---

# 281. 第一优先级 Gate：worktree-check

## 目标

```text
防止 main / master 直接开发。
```

## 检查项

```text
当前分支不是 main/master
当前不是 detached HEAD
当前 repo root 位于 worktree_root 下
当前不是 primary main worktree
```

## 最小算法

```text
1. git rev-parse --show-toplevel
2. git symbolic-ref --short HEAD
3. git rev-parse --git-common-dir
4. 读取 goalcli.yaml worktree_root
5. 判断 branch 是否 main/master
6. 判断 repo root 是否在 worktree_root 下
7. 输出 reports/worktree-check.json
```

## 失败条件

```text
branch = main
branch = master
detached HEAD
repo root 不在 worktree_root
```

## 验收

```text
[ ] 在 main worktree 下执行失败
[ ] 在合法 task worktree 下执行通过
[ ] 输出 reports/worktree-check.json
[ ] 输出 reports/worktree-check.txt
[ ] exit code = 6
```

---

# 282. 第二优先级 Gate：secret-check

## 目标

```text
防止 secret 进入代码、文档、Evidence、Release。
```

## 检查范围

```text
*.go
*.rs
*.sh
*.md
*.yaml
*.yml
*.json
.env*
.github/
.agent/
release/
reports/
```

## 最小关键词

```text
token
password
secret
private_key
access_key
authorization
cookie
```

## 注意

第一版不要做复杂误报消除，但必须支持 allowlist。

```text
.agent/security/secret-allowlist.yaml
```

## 验收

```text
[ ] secret-leak fixture 失败
[ ] masked example 通过
[ ] reports/secret-check.json 生成
[ ] exit code = 7
```

---

# 283. 第三优先级 Gate：schema-check

## 目标

```text
让 Goal / Task / Evidence / Release 成为机器可读对象。
```

## 第一版 Schema

只实现四个：

```text
goal.schema.json
task.schema.json
evidence.schema.json
release.schema.json
```

不要一开始上太多 schema。

## 最小字段

### Goal

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

### Task

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

### Evidence

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

### Release

```yaml
release_id:
version:
channel:
commit:
tag:
included_goals:
included_issues:
included_prs:
evidence_summary:
rollback_plan:
known_issues:
```

## 验收

```text
[ ] golden goal pack 通过
[ ] invalid schema fixture 失败
[ ] reports/schema-check.json 生成
[ ] exit code = 3
```

---

# 284. 第四优先级 Gate：evidence-check

## 目标

```text
防止“已完成但无法证明”。
```

## 检查项

```text
Evidence ID 存在
Goal ID 存在
Task ID 存在
command 非空
cwd 非空
branch 非空
exit_code 存在
artifact path 存在
artifact 非空
Release Evidence 必须有 commit
```

## 第一版输入

```text
.agent/registries/evidence.yaml
.agent/goals/*/evidence/
reports/
```

## 验收

```text
[ ] 缺 Evidence 失败
[ ] artifact 不存在失败
[ ] artifact 空文件失败
[ ] Release Evidence 无 commit 失败
[ ] golden evidence 通过
[ ] reports/evidence-check.json 生成
[ ] exit code = 4
```

---

# 285. 第五优先级 Gate：traceability-check

## 目标

```text
防止 Requirement → AC → Task → Evidence 断链。
```

## 第一版可以简化

不要一开始做复杂 Markdown 解析。

优先使用 YAML：

```text
.agent/goals/<GOAL-ID>/traceability.yaml
```

结构：

```yaml
links:
  - requirement_id: REQ-001
    ac_id: AC-001
    task_id: TASK-001
    evidence_id: EVID-001
    status: done
```

## 检查项

```text
每个 Requirement 有 AC
每个 AC 有 Task
每个 Task 有 Evidence
Evidence 在 evidence registry 中存在
status 与 Evidence status 一致
```

## 验收

```text
[ ] 缺 AC 失败
[ ] 缺 Task 失败
[ ] 缺 Evidence 失败
[ ] Evidence registry 不存在该 ID 失败
[ ] golden traceability 通过
[ ] reports/traceability-check.json 生成
[ ] exit code = 5
```

---

# 286. PR-check 第一版边界

第一版不要调用 GitHub API，先做本地 PR 模板检查。

## 输入

```text
.agent/goals/<GOAL-ID>/pr.md
```

或：

```text
.github/pull_request_template.md
```

## 必须区块

```text
Goal
Related Issues
Requirements Covered
Changes
Tests
Evidence
Risk
Rollback
Worktree Evidence
Checklist
```

## 验收

```text
[ ] 缺 Goal 区块失败
[ ] 缺 Evidence 区块失败
[ ] 缺 Rollback 区块失败
[ ] reports/pr-check.json 生成
```

后续 v0.2.0 再接 GitHub API。

---

# 287. release-check 第一版边界

## 输入

```text
release/<REL-ID>/
```

必须包含：

```text
manifest.md
evidence-summary.md
rollback.md
known-issues.md
checksums.txt
```

## 检查项

```text
文件存在
文件非空
Evidence coverage = 100%
Traceability coverage = 100%
Rollback 存在
checksums.txt 存在
```

## 验收

```text
[ ] 缺 manifest 失败
[ ] 缺 rollback 失败
[ ] 缺 checksums 失败
[ ] Evidence coverage <100% 失败
[ ] reports/release-check.json 生成
```

---

# 288. retro-check 第一版边界

## 输入

```text
.agent/goals/<GOAL-ID>/retrospective.md
```

必须区块：

```text
What Worked
What Failed
Root Cause
Prompt Patch
Harness Patch
Rule Patch
New Issue Candidates
```

## 验收

```text
[ ] 缺 Root Cause 失败
[ ] 缺 Patch 候选失败
[ ] Lite Mode 可 warning
[ ] Standard/Full Mode 必须阻断
```

---

# 289. audit goal 第一版

## 目标

```text
从 Goal Pack 重建事实链。
```

## 输出

```text
reports/audit/<GOAL-ID>.md
reports/audit/<GOAL-ID>.json
```

## 内容

```text
Goal Summary
State
Task Count
Evidence Count
Evidence Coverage
Traceability Coverage
Release Readiness
Open Risks
Open Gaps
Score
Decision
```

## 评分

```text
>= 90：可以 stable
80-89：只允许 rc
70-79：只允许 beta
<70：阻断 release
```

---

# 290. Golden Fixture 最小结构

```text
.agent/harness/fixtures/golden/minimal-goal-pack/
├── goal.yaml
├── spec.yaml
├── tasks.yaml
├── evidence.yaml
├── traceability.yaml
├── release/
│   └── REL-TEST/
│       ├── manifest.md
│       ├── evidence-summary.md
│       ├── rollback.md
│       ├── known-issues.md
│       └── checksums.txt
└── retrospective.md
```

Golden 必须能通过：

```bash
goalcli schema validate --fixture .agent/harness/fixtures/golden/minimal-goal-pack
goalcli evidence check --fixture .agent/harness/fixtures/golden/minimal-goal-pack
goalcli traceability check --fixture .agent/harness/fixtures/golden/minimal-goal-pack
goalcli release check --fixture .agent/harness/fixtures/golden/minimal-goal-pack
goalcli retro check --fixture .agent/harness/fixtures/golden/minimal-goal-pack
goalcli audit goal --fixture .agent/harness/fixtures/golden/minimal-goal-pack
```

---

# 291. Violation Fixtures 最小集合

```text
.agent/harness/fixtures/violations/
├── invalid-schema/
├── missing-evidence/
├── missing-traceability/
├── missing-release-manifest/
├── missing-rollback/
└── secret-leak/
```

每个 violation 必须有：

```yaml
expected.yaml
```

示例：

```yaml
expected:
  checker: evidence-check
  status: failed
  exit_code: 4
  rule_id: RULE-EVIDENCE-001
```

---

# 292. 第一阶段 Issue 拆解

## ISSUE-001：goalcli 项目初始化

```text
交付：
- tools/goalcli/go.mod
- cmd/goalcli/main.go
- internal/report
- internal/config

验收：
- goalcli --version 可运行
- goalcli doctor 可运行
```

## ISSUE-002：实现 report 标准输出

```text
交付：
- CheckResult
- JSON report writer
- text report writer
- exit code mapper

验收：
- reports/*.json 结构统一
```

## ISSUE-003：实现 worktree-check

```text
验收：
- main 失败
- legal worktree 通过
- exit code = 6
```

## ISSUE-004：实现 secret-check

```text
验收：
- secret-leak fixture 失败
- allowlist 生效
- exit code = 7
```

## ISSUE-005：实现 schema-check

```text
验收：
- golden 通过
- invalid-schema 失败
- exit code = 3
```

## ISSUE-006：实现 evidence-check

```text
验收：
- missing-evidence 失败
- golden 通过
- exit code = 4
```

## ISSUE-007：实现 traceability-check

```text
验收：
- missing-traceability 失败
- golden 通过
- exit code = 5
```

## ISSUE-008：实现 release-check

```text
验收：
- missing-release-manifest 失败
- missing-rollback 失败
- golden 通过
```

## ISSUE-009：实现 retro-check

```text
验收：
- retro 缺 Patch 候选失败
- golden 通过
```

## ISSUE-010：实现 audit goal

```text
验收：
- reports/audit/<GOAL-ID>.md 生成
- score 正确计算
```

## ISSUE-011：接入 Makefile / Git hooks / GitHub Actions

```text
验收：
- make ci 调通
- pre-commit 阻断 main
- PR workflow 调用 make ci
```

## ISSUE-012：生成 v0.1.0 Release Artifact

```text
验收：
- release-check-all 通过
- Audit Score >= 90
```

---

# 293. 第一阶段验收命令

```bash
make worktree-check
make secret-check
make schema-check
make evidence-check
make traceability-check
make release-check
make retro-check
make audit-check
make ci
make release-check-all
```

全部通过，才算 v0.1.0 完成。

---

# 294. v2.0 最终验收标准

```text
[ ] goalcli --version 可运行
[ ] goalcli doctor 可运行
[ ] worktree-check 可阻断 main
[ ] secret-check 可阻断 secret
[ ] schema-check 可识别 invalid schema
[ ] evidence-check 可识别 missing evidence
[ ] traceability-check 可识别 broken traceability
[ ] release-check 可识别 missing manifest / rollback
[ ] retro-check 可识别 missing patch candidates
[ ] audit goal 可生成 score
[ ] golden fixture 全通过
[ ] violation fixtures 全失败
[ ] Makefile 统一入口可用
[ ] Git hooks 可安装
[ ] GitHub Actions 调用 Makefile
[ ] v0.1.0 release artifact 完整
```

---

# 295. 下一步最小行动

现在真正应该执行的是：

```text
1. 在 xlib-standard 新建 tools/goalcli/
2. 实现 CheckResult 与 report writer
3. 实现 worktree-check
4. 接入 Makefile
5. 用 main 分支失败 + worktree 通过证明规则生效
```

这一步完成后，Goal Runtime 才从“规则文档”进入“机器裁判”。

---

# 296. v2.0 最终铁律

```text
1. 只做最小裁判内核。
2. 只实现可测试 Gate。
3. 只接受有 Golden / Violation fixture 的规则。
4. 只用 reports 作为 Evidence 输入。
5. 所有 Gate 都必须有统一 CheckResult。
6. 所有失败都必须有 rule_id、exit_code、修复方向。
7. 先 worktree-check，再做其它自动化。
8. v0.1.0 完成前，不做自动 merge、自动 stable publish、复杂 Agent 编排。
```

最终结论：

> **Goal Runtime v2.0 的核心不是再写规则，而是把 `worktree-check`、`secret-check`、`schema-check`、`evidence-check`、`traceability-check` 做成真正可运行的 goalcli v0.1.0。只要这五个 Gate 生效，整个 Goal 系统的可信度会发生质变。**
继续进入 **Goal Runtime v2.1：第一批代码级落地骨架**。

这一版不再增加抽象规则，而是把 `goalcli v0.1.0` 的第一批可提交文件、最小 Go 代码结构、Makefile、worktree-check、report writer、fixture 测试方式整理出来。

---

# Goal Runtime v2.1：goalcli v0.1.0 第一批可执行骨架

## 297. v2.1 的目标

v2.1 只完成一个目标：

> **让 `goalcli worktree check` 成为第一个真实可运行的 P0 Gate。**

这一步完成后，Goal Runtime 从“规则文档”进入“机器裁判”。

第一批必须做到：

```text
goalcli --version 可运行
goalcli doctor 可运行
goalcli worktree check 可运行
make worktree-check 可运行
main 分支执行失败
合法 worktree 执行通过
reports/worktree-check.json 生成
reports/worktree-check.txt 生成
```

---

# 298. 第一批 Commit 拆解

不要一个 PR 做完所有 Gate。第一批按 5 个 Commit / Task 落地。

## Commit 1：初始化 goalcli 工具目录

```text
feat(goalcli): initialize goalcli cli skeleton
```

交付：

```text
tools/goalcli/go.mod
tools/goalcli/cmd/goalcli/main.go
tools/goalcli/internal/report/result.go
tools/goalcli/internal/report/writer.go
tools/goalcli/internal/config/config.go
```

验收：

```bash
cd tools/goalcli
go run ./cmd/goalcli --version
go run ./cmd/goalcli doctor
```

---

## Commit 2：实现 report 输出协议

```text
feat(goalcli): add standard check result reports
```

交付：

```text
internal/report/result.go
internal/report/writer.go
reports/.gitkeep
```

验收：

```bash
go run ./cmd/goalcli doctor
ls reports/
```

---

## Commit 3：实现 gitutil

```text
feat(goalcli): add git repository inspection helpers
```

交付：

```text
internal/gitutil/git.go
```

验收：

```bash
go run ./cmd/goalcli doctor
```

doctor 输出：

```text
repo_root
branch
commit
is_detached
git_common_dir
```

---

## Commit 4：实现 worktree-check

```text
feat(harness): add worktree-only gate
```

交付：

```text
internal/worktree/checker.go
reports/worktree-check.json
reports/worktree-check.txt
```

验收：

```bash
go run ./cmd/goalcli worktree check
```

---

## Commit 5：接入 Makefile / hooks

```text
chore(harness): wire worktree gate into make and git hooks
```

交付：

```text
Makefile
.githooks/pre-commit
.githooks/pre-push
scripts/git/install-hooks.sh
.github/workflows/worktree-guard.yml
```

验收：

```bash
make worktree-check
scripts/git/install-hooks.sh
```

---

# 299. `tools/goalcli` 初始目录

推荐第一版结构：

```text
tools/goalcli/
├── go.mod
├── cmd/
│   └── goalcli/
│       └── main.go
└── internal/
    ├── config/
    │   └── config.go
    ├── gitutil/
    │   └── git.go
    ├── report/
    │   ├── result.go
    │   └── writer.go
    └── worktree/
        └── checker.go
```

根目录配套：

```text
Makefile
goalcli.yaml
reports/.gitkeep
.githooks/pre-commit
.githooks/pre-push
scripts/git/install-hooks.sh
```

---

# 300. `goalcli.yaml` 最小版本

先不要复杂化。

```yaml
version: goalcli-v0.1.0

runtime:
  timezone: Asia/Tokyo
  reports_dir: reports

git:
  default_branch: main
  protected_branches:
    - main
    - master

worktree:
  required: true
  root: ~/code/.worktrees

rules:
  severity_policy:
    P0: block
    P1: block
    P2: warn
    P3: score_only
```

第一版只需要 `reports_dir`、`protected_branches`、`worktree.root`。

---

# 301. `CheckResult` 标准结构

`tools/goalcli/internal/report/result.go`

```go
package report

type Status string

const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	StatusWarn   Status = "warn"
)

type Severity string

const (
	SeverityP0 Severity = "P0"
	SeverityP1 Severity = "P1"
	SeverityP2 Severity = "P2"
	SeverityP3 Severity = "P3"
)

type Finding struct {
	RuleID  string `json:"rule_id"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type CheckResult struct {
	Checker   string    `json:"checker"`
	Status    Status    `json:"status"`
	Severity  Severity  `json:"severity"`
	GoalID    string    `json:"goal_id,omitempty"`
	Errors    []Finding `json:"errors"`
	Warnings  []Finding `json:"warnings"`
	Artifacts []string  `json:"artifacts"`
	Timestamp string    `json:"timestamp"`
}
```

---

# 302. Report Writer

`tools/goalcli/internal/report/writer.go`

```go
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteJSON(path string, result CheckResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func WriteText(path string, result CheckResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("checker: %s\n", result.Checker))
	b.WriteString(fmt.Sprintf("status: %s\n", result.Status))
	b.WriteString(fmt.Sprintf("severity: %s\n", result.Severity))
	b.WriteString(fmt.Sprintf("timestamp: %s\n", result.Timestamp))

	if len(result.Errors) > 0 {
		b.WriteString("\nerrors:\n")
		for _, e := range result.Errors {
			b.WriteString(fmt.Sprintf("- [%s] %s\n", e.RuleID, e.Message))
		}
	}

	if len(result.Warnings) > 0 {
		b.WriteString("\nwarnings:\n")
		for _, w := range result.Warnings {
			b.WriteString(fmt.Sprintf("- [%s] %s\n", w.RuleID, w.Message))
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func ExitCode(result CheckResult, fallback int) int {
	if result.Status == StatusPassed || result.Status == StatusWarn {
		return 0
	}

	switch result.Checker {
	case "worktree-check":
		return 6
	case "secret-check":
		return 7
	case "schema-check":
		return 3
	case "evidence-check":
		return 4
	case "traceability-check":
		return 5
	case "release-check":
		return 8
	default:
		return fallback
	}
}
```

---

# 303. Git 工具函数

`tools/goalcli/internal/gitutil/git.go`

```go
package gitutil

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

func RepoRoot() (string, error) {
	return runGit("rev-parse", "--show-toplevel")
}

func Branch() (string, error) {
	return runGit("symbolic-ref", "--quiet", "--short", "HEAD")
}

func Commit() (string, error) {
	return runGit("rev-parse", "HEAD")
}

func GitCommonDir() (string, error) {
	return runGit("rev-parse", "--git-common-dir")
}

func IsDetachedHead() bool {
	_, err := Branch()
	return err != nil
}
```

---

# 304. Config 最小读取

第一版可以先不引入复杂配置库。为了避免 YAML 解析复杂度，`worktree.root` 可以先允许环境变量覆盖。

优先级：

```text
GOALCLI_WORKTREE_ROOT
→ goalcli.yaml
→ 默认 ~/code/.worktrees
```

v0.1.0 可以先实现环境变量 + 默认值；v0.1.1 再完善 YAML。

`tools/goalcli/internal/config/config.go`

```go
package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ReportsDir        string
	WorktreeRoot      string
	ProtectedBranches map[string]bool
}

func Load() Config {
	worktreeRoot := os.Getenv("GOALCLI_WORKTREE_ROOT")
	if worktreeRoot == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			worktreeRoot = filepath.Join(home, "code", ".worktrees")
		} else {
			worktreeRoot = ".worktrees"
		}
	}

	worktreeRoot = expandHome(worktreeRoot)

	return Config{
		ReportsDir:   "reports",
		WorktreeRoot: worktreeRoot,
		ProtectedBranches: map[string]bool{
			"main":   true,
			"master": true,
		},
	}
}

func expandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}

	return path
}
```

---

# 305. worktree-check 实现骨架

`tools/goalcli/internal/worktree/checker.go`

```go
package worktree

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/config"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/gitutil"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/report"
)

func Check(cfg config.Config) report.CheckResult {
	result := report.CheckResult{
		Checker:  "worktree-check",
		Status:   report.StatusPassed,
		Severity: report.SeverityP0,
		Errors:   []report.Finding{},
		Warnings: []report.Finding{},
		Artifacts: []string{
			filepath.Join(cfg.ReportsDir, "worktree-check.json"),
			filepath.Join(cfg.ReportsDir, "worktree-check.txt"),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	repoRoot, err := gitutil.RepoRoot()
	if err != nil {
		return fail(result, "RULE-WORKTREE-001", "not inside a git repository: "+err.Error())
	}

	branch, err := gitutil.Branch()
	if err != nil {
		return fail(result, "RULE-WORKTREE-001", "detached HEAD is not allowed for development")
	}

	if cfg.ProtectedBranches[branch] {
		return fail(result, "RULE-WORKTREE-001", "direct development on protected branch '"+branch+"' is forbidden")
	}

	cleanRepoRoot := filepath.Clean(repoRoot)
	cleanWorktreeRoot := filepath.Clean(cfg.WorktreeRoot)

	if !isWithin(cleanRepoRoot, cleanWorktreeRoot) {
		return fail(result, "RULE-WORKTREE-002", "repo root is not inside configured worktree root: "+cleanRepoRoot)
	}

	return result
}

func fail(result report.CheckResult, ruleID string, message string) report.CheckResult {
	result.Status = report.StatusFailed
	result.Errors = append(result.Errors, report.Finding{
		RuleID:  ruleID,
		Message: message,
	})
	return result
}

func isWithin(path string, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}
```

---

# 306. CLI `main.go`

`tools/goalcli/cmd/goalcli/main.go`

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/config"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/gitutil"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/report"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/worktree"
)

const version = "v0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	switch {
	case args[0] == "--version" || args[0] == "version":
		fmt.Println("goalcli", version)

	case args[0] == "doctor":
		runDoctor()

	case len(args) >= 2 && args[0] == "worktree" && args[1] == "check":
		runWorktreeCheck()

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`goalcli v0.1.0

Usage:
  goalcli --version
  goalcli doctor
  goalcli worktree check`)
}

func runDoctor() {
	root, _ := gitutil.RepoRoot()
	branch, _ := gitutil.Branch()
	commit, _ := gitutil.Commit()
	commonDir, _ := gitutil.GitCommonDir()

	fmt.Println("goalcli doctor")
	fmt.Println("repo_root:", root)
	fmt.Println("branch:", branch)
	fmt.Println("commit:", commit)
	fmt.Println("git_common_dir:", commonDir)
}

func runWorktreeCheck() {
	cfg := config.Load()
	result := worktree.Check(cfg)

	jsonPath := filepath.Join(cfg.ReportsDir, "worktree-check.json")
	textPath := filepath.Join(cfg.ReportsDir, "worktree-check.txt")

	_ = report.WriteJSON(jsonPath, result)
	_ = report.WriteText(textPath, result)

	if result.Status == report.StatusPassed {
		fmt.Println("PASS: worktree-check")
	} else {
		fmt.Println("FAIL: worktree-check")
		for _, err := range result.Errors {
			fmt.Printf("- [%s] %s\n", err.RuleID, err.Message)
		}
	}

	os.Exit(report.ExitCode(result, 1))
}
```

---

# 307. `go.mod`

`tools/goalcli/go.mod`

```go
module github.com/ZoneCNH/xlib-standard/tools/goalcli

go 1.22
```

如果仓库主 Go 版本已经统一到更高版本，可以同步到项目标准版本。

---

# 308. 根 Makefile 接入方式

根目录 `Makefile`：

```makefile
GOALCLI := go run ./tools/goalcli/cmd/goalcli

.PHONY: goalcli-version
goalcli-version:
	$(GOALCLI) --version

.PHONY: doctor
doctor:
	$(GOALCLI) doctor

.PHONY: worktree-check
worktree-check:
	$(GOALCLI) worktree check

.PHONY: ci
ci: worktree-check
```

后续再扩：

```makefile
secret-check
schema-check
evidence-check
traceability-check
release-check
retro-check
audit-check
```

第一版只接 `worktree-check`，防止 Makefile 先声明不存在命令造成漂移。

---

# 309. Git Hooks

## `.githooks/pre-commit`

```bash
#!/usr/bin/env bash
set -euo pipefail

make worktree-check
```

## `.githooks/pre-push`

```bash
#!/usr/bin/env bash
set -euo pipefail

make worktree-check
```

## `scripts/git/install-hooks.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

git config core.hooksPath .githooks

chmod +x .githooks/pre-commit
chmod +x .githooks/pre-push

echo "Git hooks installed."
```

---

# 310. GitHub Actions：worktree-guard 最小版

`.github/workflows/worktree-guard.yml`

```yaml
name: Worktree Guard

on:
  push:
  pull_request:

jobs:
  guard:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Block direct push to main
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        run: |
          echo "Direct push to main is forbidden."
          exit 1

      - name: Check branch name on PR
        if: github.event_name == 'pull_request'
        run: |
          BRANCH="${{ github.head_ref }}"
          case "$BRANCH" in
            goal/*|task/*|issue/*|fix/*|feat/*|chore/*)
              echo "OK branch name: $BRANCH"
              ;;
            *)
              echo "Invalid branch name: $BRANCH"
              exit 1
              ;;
          esac
```

注意：GitHub Actions runner 的 checkout 路径不一定在本机 worktree_root 下，所以 CI 里不建议直接跑本地 `worktree-check` 的路径规则。CI 先负责阻断 direct push main 和分支命名，本地负责 worktree path。

---

# 311. 第一批验收脚本

本地 main 仓库中执行：

```bash
make worktree-check
```

预期失败：

```text
FAIL: worktree-check
- [RULE-WORKTREE-001] direct development on protected branch 'main' is forbidden
```

创建合法 worktree：

```bash
cd ~/code/xlib-standard

git fetch origin

git worktree add \
  ~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001 \
  -b goal/GOAL-20260603-001/TASK-001 \
  origin/main

cd ~/code/.worktrees/xlib-standard/GOAL-20260603-001/TASK-001

make worktree-check
```

预期通过：

```text
PASS: worktree-check
```

报告存在：

```bash
cat reports/worktree-check.json
cat reports/worktree-check.txt
```

---

# 312. 第一批 DoD

只有全部满足，才能进入 `secret-check`。

```text
[ ] goalcli --version 可运行
[ ] goalcli doctor 可运行
[ ] make worktree-check 可运行
[ ] main 分支执行 worktree-check 失败
[ ] 合法 worktree 执行 worktree-check 通过
[ ] reports/worktree-check.json 生成
[ ] reports/worktree-check.txt 生成
[ ] .githooks/pre-commit 调用 make worktree-check
[ ] .githooks/pre-push 调用 make worktree-check
[ ] worktree-guard.yml 已存在
[ ] README 或 CONSTITUTION 说明 main 禁止开发
```

---

# 313. 第二批：secret-check 进入条件

worktree-check 未完成前，不做 secret-check。

原因：

```text
如果 Agent 仍能在 main 开发，
后续所有 Gate 都可能污染主线。
```

进入条件：

```text
worktree-check = passed in legal worktree
worktree-check = failed in main
Git hooks installed
```

---

# 314. v2.1 最终执行命令

```bash
# 1. 初始化 goalcli
mkdir -p tools/goalcli/cmd/goalcli
mkdir -p tools/goalcli/internal/{config,gitutil,report,worktree}
mkdir -p reports
touch reports/.gitkeep

# 2. 写入 go.mod / main.go / internal packages

# 3. 验证
cd tools/goalcli
go run ./cmd/goalcli --version
go run ./cmd/goalcli doctor

# 4. 回到根目录
cd ../..
make worktree-check
```

---

# 315. v2.1 最终铁律

```text
1. 第一批只实现 worktree-check，不扩散。
2. Makefile 只声明已经存在的 Gate。
3. CI 不强行复用本地 worktree path 检查。
4. 所有 Gate 必须输出 reports。
5. main 失败 + 合法 worktree 通过，是第一阶段唯一核心证明。
```

最终结论：

> **v2.1 的唯一目标是让 `goalcli worktree check` 成为第一个真正可运行的 Harness Gate。完成它以后，Goal Runtime 才拥有第一条不可绕过的机器铁律：禁止 main 开发。**
继续进入 **Goal Runtime v2.2：实现第二个 P0 Gate —— `secret-check`**。

上一阶段 v2.1 只做一件事：让 `worktree-check` 真正跑起来，防止 main 开发。
v2.2 只做第二件事：

> **让 `goalcli secret check` 成为第二个真实可运行的 P0 Gate，防止密钥进入代码、文档、Evidence、Release。**

---

# Goal Runtime v2.2：`secret-check` 最小实现方案

## 316. v2.2 目标

`secret-check` 的目标不是替代专业密钥扫描器，而是建立第一层不可绕过的基础防线。

v2.2 必须做到：

```text
goalcli secret check 可运行
make secret-check 可运行
含 secret 的 fixture 必须失败
masked example / allowlist 示例必须通过
reports/secret-check.json 生成
reports/secret-check.txt 生成
exit code = 7
```

---

# 317. v2.2 不做什么

第一版不要做过度工程化。

暂不做：

```text
复杂熵检测
接入 GitHub Secret Scanning API
接入 trufflehog / gitleaks
历史 git commit 全量扫描
自动 rotate credential
复杂多语言语义识别
```

v2.2 只做：

```text
扫描 git tracked files
匹配常见 secret pattern
支持 allowlist
输出标准 CheckResult
阻断 P0 secret violation
```

---

# 318. v2.2 Commit 拆解

## Commit 1：扩展 CLI 命令

```text
feat(goalcli): add secret check command skeleton
```

交付：

```text
tools/goalcli/internal/secret/checker.go
tools/goalcli/internal/gitutil/git.go 更新 TrackedFiles
tools/goalcli/cmd/goalcli/main.go 增加 secret check
```

验收：

```bash
go run ./tools/goalcli/cmd/goalcli secret check
```

---

## Commit 2：实现关键词扫描

```text
feat(secret): scan tracked files for secret-like assignments
```

交付：

```text
secret checker 支持 token/password/secret/private_key/access_key/authorization/cookie
```

验收：

```bash
make secret-check
```

---

## Commit 3：增加 allowlist

```text
feat(secret): add allowlist support for masked examples
```

交付：

```text
.agent/security/secret-allowlist.yaml
```

验收：

```bash
make secret-check
```

---

## Commit 4：增加 violation fixture

```text
test(secret): add secret leak violation fixture
```

交付：

```text
.agent/harness/fixtures/violations/secret-leak/
.agent/harness/fixtures/golden/minimal-secret-example/
```

验收：

```bash
goalcli secret check --fixture .agent/harness/fixtures/violations/secret-leak
```

---

## Commit 5：接入 Makefile / hooks / CI

```text
chore(secret): wire secret gate into make and hooks
```

交付：

```text
Makefile 增加 secret-check
.githooks/pre-commit 增加 secret-check
.githooks/pre-push 增加 secret-check
```

---

# 319. 目录结构补充

新增：

```text
tools/goalcli/internal/secret/
└── checker.go

.agent/security/
└── secret-allowlist.yaml

.agent/harness/fixtures/violations/secret-leak/
├── leaked.env
└── expected.yaml

.agent/harness/fixtures/golden/minimal-secret-example/
├── example.env
└── expected.yaml
```

---

# 320. `gitutil` 增加 tracked files

更新：

```text
tools/goalcli/internal/gitutil/git.go
```

新增函数：

```go
func TrackedFiles() ([]string, error) {
	out, err := runGit("ls-files")
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(out) == "" {
		return []string{}, nil
	}

	lines := strings.Split(out, "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if file != "" {
			files = append(files, file)
		}
	}

	return files, nil
}
```

---

# 321. `secret-check` 检查范围

第一版扫描 git tracked files，但只检查这些扩展或路径：

```text
*.go
*.rs
*.sh
*.md
*.yaml
*.yml
*.json
.env
.env.*
.github/
.agent/
release/
reports/
```

跳过：

```text
.git/
vendor/
node_modules/
target/
dist/
build/
*.png
*.jpg
*.jpeg
*.gif
*.pdf
*.zip
*.tar
*.gz
*.parquet
```

---

# 322. Secret Pattern 第一版

最小规则：

```text
RULE-SECRET-001:
禁止提交明文 token / password / secret / private key / access key / authorization / cookie。
```

匹配逻辑：

```text
敏感 key + 赋值符号 + 非空真实值
```

例如应失败：

```text
OPENAI_API_KEY=sk-xxxxxxxxxxxx
password = "abc123456"
github_token: ghp_xxxxxxxxxxxx
authorization: Bearer abcdefghijklmn
private_key = "-----BEGIN PRIVATE KEY-----"
```

允许：

```text
TOKEN=***
PASSWORD=<redacted>
SECRET=REDACTED
example_token=changeme
authorization: Bearer <token>
```

---

# 323. `secret/checker.go` 最小实现

```go
package secret

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/config"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/gitutil"
	"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/report"
)

var secretAssignmentPattern = regexp.MustCompile(
	`(?i)(token|password|secret|private[_-]?key|access[_-]?key|authorization|cookie)\s*[:=]\s*["']?([^"'\s#]+)`,
)

var privateKeyPattern = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)

func Check(cfg config.Config) report.CheckResult {
	result := report.CheckResult{
		Checker:  "secret-check",
		Status:   report.StatusPassed,
		Severity: report.SeverityP0,
		Errors:   []report.Finding{},
		Warnings: []report.Finding{},
		Artifacts: []string{
			filepath.Join(cfg.ReportsDir, "secret-check.json"),
			filepath.Join(cfg.ReportsDir, "secret-check.txt"),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	files, err := gitutil.TrackedFiles()
	if err != nil {
		return fail(result, "RULE-SECRET-001", "failed to list git tracked files: "+err.Error(), "", 0)
	}

	for _, file := range files {
		if !shouldScan(file) {
			continue
		}

		findings := scanFile(file)

		for _, finding := range findings {
			result.Errors = append(result.Errors, finding)
		}
	}

	if len(result.Errors) > 0 {
		result.Status = report.StatusFailed
	}

	return result
}

func scanFile(path string) []report.Finding {
	file, err := os.Open(path)
	if err != nil {
		return []report.Finding{
			{
				RuleID:  "RULE-SECRET-001",
				Message: "failed to open file for secret scan: " + err.Error(),
				File:    path,
			},
		}
	}
	defer file.Close()

	var findings []report.Finding

	scanner := bufio.NewScanner(file)
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if isAllowedMaskedExample(line) {
			continue
		}

		if privateKeyPattern.MatchString(line) {
			findings = append(findings, report.Finding{
				RuleID:  "RULE-SECRET-001",
				Message: "private key material detected",
				File:    path,
				Line:    lineNo,
			})
			continue
		}

		matches := secretAssignmentPattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			value := strings.TrimSpace(matches[2])

			if looksLikeRealSecret(value) {
				findings = append(findings, report.Finding{
					RuleID:  "RULE-SECRET-001",
					Message: "secret-like assignment detected",
					File:    path,
					Line:    lineNo,
				})
			}
		}
	}

	return findings
}

func shouldScan(path string) bool {
	clean := filepath.ToSlash(path)

	skipPrefixes := []string{
		".git/",
		"vendor/",
		"node_modules/",
		"target/",
		"dist/",
		"build/",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(clean, prefix) {
			return false
		}
	}

	scanPrefixes := []string{
		".github/",
		".agent/",
		"release/",
		"reports/",
	}

	for _, prefix := range scanPrefixes {
		if strings.HasPrefix(clean, prefix) {
			return true
		}
	}

	ext := strings.ToLower(filepath.Ext(clean))

	switch ext {
	case ".go", ".rs", ".sh", ".md", ".yaml", ".yml", ".json":
		return true
	}

	base := filepath.Base(clean)
	return base == ".env" || strings.HasPrefix(base, ".env.")
}

func isAllowedMaskedExample(line string) bool {
	lower := strings.ToLower(line)

	allowedMarkers := []string{
		"***",
		"<redacted>",
		"redacted",
		"<token>",
		"<secret>",
		"<password>",
		"changeme",
		"example",
		"dummy",
		"placeholder",
	}

	for _, marker := range allowedMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	return false
}

func looksLikeRealSecret(value string) bool {
	value = strings.Trim(value, `"'`)

	if value == "" {
		return false
	}

	if len(value) < 8 {
		return false
	}

	placeholderValues := []string{
		"example",
		"changeme",
		"placeholder",
		"redacted",
		"<redacted>",
		"<token>",
		"<secret>",
		"<password>",
		"***",
	}

	lower := strings.ToLower(value)

	for _, placeholder := range placeholderValues {
		if lower == placeholder {
			return false
		}
	}

	return true
}

func fail(result report.CheckResult, ruleID string, message string, file string, line int) report.CheckResult {
	result.Status = report.StatusFailed
	result.Errors = append(result.Errors, report.Finding{
		RuleID:  ruleID,
		Message: message,
		File:    file,
		Line:    line,
	})
	return result
}
```

---

# 324. CLI 接入 `secret check`

更新：

```text
tools/goalcli/cmd/goalcli/main.go
```

新增 import：

```go
"github.com/ZoneCNH/xlib-standard/tools/goalcli/internal/secret"
```

新增 usage：

```text
goalcli secret check
```

新增 case：

```go
case len(args) >= 2 && args[0] == "secret" && args[1] == "check":
	runSecretCheck()
```

新增函数：

```go
func runSecretCheck() {
	cfg := config.Load()
	result := secret.Check(cfg)

	jsonPath := filepath.Join(cfg.ReportsDir, "secret-check.json")
	textPath := filepath.Join(cfg.ReportsDir, "secret-check.txt")

	_ = report.WriteJSON(jsonPath, result)
	_ = report.WriteText(textPath, result)

	if result.Status == report.StatusPassed {
		fmt.Println("PASS: secret-check")
	} else {
		fmt.Println("FAIL: secret-check")
		for _, err := range result.Errors {
			fmt.Printf("- [%s] %s", err.RuleID, err.Message)
			if err.File != "" {
				fmt.Printf(" (%s:%d)", err.File, err.Line)
			}
			fmt.Println()
		}
	}

	os.Exit(report.ExitCode(result, 1))
}
```

---

# 325. Report Exit Code 已支持

`report.ExitCode` 中前面已经定义：

```go
case "secret-check":
	return 7
```

所以 secret 检测失败时必须返回：

```text
exit code = 7
```

---

# 326. Makefile 接入

更新根目录 `Makefile`：

```makefile
GOALCLI := go run ./tools/goalcli/cmd/goalcli

.PHONY: goalcli-version
goalcli-version:
	$(GOALCLI) --version

.PHONY: doctor
doctor:
	$(GOALCLI) doctor

.PHONY: worktree-check
worktree-check:
	$(GOALCLI) worktree check

.PHONY: secret-check
secret-check:
	$(GOALCLI) secret check

.PHONY: ci
ci:
	$(MAKE) worktree-check
	$(MAKE) secret-check
```

---

# 327. Git Hooks 更新

## `.githooks/pre-commit`

```bash
#!/usr/bin/env bash
set -euo pipefail

make worktree-check
make secret-check
```

## `.githooks/pre-push`

```bash
#!/usr/bin/env bash
set -euo pipefail

make worktree-check
make secret-check
```

---

# 328. Golden Fixture

目录：

```text
.agent/harness/fixtures/golden/minimal-secret-example/
```

文件：

```text
example.env
expected.yaml
```

`example.env`：

```env
OPENAI_API_KEY=***
DATABASE_PASSWORD=<redacted>
GITHUB_TOKEN=changeme
AUTHORIZATION="Bearer <token>"
```

`expected.yaml`：

```yaml
expected:
  checker: secret-check
  status: passed
  exit_code: 0
```

---

# 329. Violation Fixture

目录：

```text
.agent/harness/fixtures/violations/secret-leak/
```

文件：

```text
leaked.env
expected.yaml
```

`leaked.env`：

```env
OPENAI_API_KEY=sk-this-is-a-fake-but-secret-like-value
DATABASE_PASSWORD=real-looking-password-123456
GITHUB_TOKEN=ghp_this_is_a_fake_secret_like_token
```

`expected.yaml`：

```yaml
expected:
  checker: secret-check
  status: failed
  exit_code: 7
  rule_id: RULE-SECRET-001
```

注意：这里用的是 fake value，但仍然要能触发扫描器。

---

# 330. 第一版 fixture 测试方式

v2.2 可以先不用 `--fixture` 参数，先用手工复制方式验证。

后续 v2.3 再实现：

```bash
goalcli secret check --fixture <path>
```

当前可先执行：

```bash
cp .agent/harness/fixtures/violations/secret-leak/leaked.env ./leaked.env
git add leaked.env
make secret-check
```

预期：

```text
FAIL: secret-check
exit code = 7
```

验证后删除：

```bash
git reset leaked.env
rm leaked.env
```

golden 示例：

```bash
cp .agent/harness/fixtures/golden/minimal-secret-example/example.env ./example.env
git add example.env
make secret-check
```

预期：

```text
PASS: secret-check
```

验证后删除：

```bash
git reset example.env
rm example.env
```

---

# 331. v2.2 DoD

只有全部满足，才进入 `schema-check`。

```text
[ ] goalcli secret check 可运行
[ ] make secret-check 可运行
[ ] secret-check 失败时 exit code = 7
[ ] secret-check 成功时 exit code = 0
[ ] reports/secret-check.json 生成
[ ] reports/secret-check.txt 生成
[ ] secret-leak fixture 能失败
[ ] masked example 能通过
[ ] pre-commit 调用 secret-check
[ ] pre-push 调用 secret-check
[ ] make ci 包含 worktree-check + secret-check
```

---

# 332. v2.2 完成后的 Gate 顺序

现在 CI 最小链路变成：

```text
worktree-check
→ secret-check
```

含义：

```text
先确保不在 main 开发
再确保没有 secret 泄漏
```

后续才能进入：

```text
schema-check
→ evidence-check
→ traceability-check
```

---

# 333. 下一阶段 v2.3

v2.3 建议做：

```text
schema-check
```

而不是继续扩 secret-check。

原因：

```text
secret-check 第一版只需要基础防线
schema-check 才能让 Goal Pack / Evidence / Release 机器可读
```

v2.3 的目标：

```text
goalcli schema validate --all
Goal / Task / Evidence / Release schema
Golden Goal Pack
invalid-schema violation fixture
reports/schema-check.json
```

---

# 334. v2.2 最终铁律

```text
1. secret-check 是 P0 Gate。
2. 第一版只扫 tracked files。
3. 第一版只做基础 pattern，不做复杂熵检测。
4. fake secret fixture 必须失败。
5. masked example 必须通过。
6. secret-check 必须进入 pre-commit / pre-push / make ci。
7. 发现 secret 后不能继续 commit / push / release。
```

最终结论：

> **v2.2 完成后，Goal Runtime 拥有第二条机器铁律：不允许 secret 进入代码、文档、Evidence、Release。worktree-check 保护主线，secret-check 保护安全边界。两者一起构成 goalcli v0.1.0 的第一层 P0 防护。**
