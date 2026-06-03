# Agent Runtime / Tooling / 度量规则

> 本文件由 `scripts/render_domain_rules.py` 从 [`registry.yaml`](./registry.yaml)
> 与 `.worktree/goal-patch.md` 渲染生成；冲突时以 `iron-rules.md` >
> `registry.yaml` > 本文件 > `.worktree/goal-patch.md` 为序。

本文件覆盖 Goal Runtime **执行平面层**规则：Agent 协议与边界、并发与租约、命令事务与 dry-run、Bootstrap/Doctor/Repair、Dashboard 与度量、`goalcli` 工具链架构、控制平面与报告规范。

对应 Gate：`runtime-doctor`、`runtime-repair`、`dashboard`、`cmd-txn-check`、`cli-contract`、`gate-dag-check`。

---

## §48 Human Approval 规则

### **[P0]** `RULE-HUMAN-001`：高风险变更必须人工批准

<sub>level: P0 · status: active · enforced_by: `goalcli pr-template` · exit: 1 · source: §48 L2333</sub>

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

## §54 Rule as Code 规则

### **[P1]** `RULE-CODE-001`：所有 P0 / P1 规则必须机器化

<sub>level: P1 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §54 L2537</sub>

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

### **[P1]** `RULE-CODE-002`：规则必须有机器 ID

<sub>level: P1 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §54 L2570</sub>

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

## §57 Agent 执行协议

### **[P1]** `RULE-AGENT-001`：Agent 不允许自由发挥执行

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §57 L2699</sub>

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

### **[P1]** `RULE-AGENT-002`：Agent 每一步必须写 Execution Log

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §57 L2722</sub>

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

### **[P1]** `RULE-AGENT-003`：Agent 不能跳过失败

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §57 L2747</sub>

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

## §58 Agent 权限边界规则

### **[P1]** `RULE-AGENT-AUTH-001`：Agent 只能操作当前 Goal 授权范围

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §58 L2772</sub>

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

### **[P1]** `RULE-AGENT-AUTH-002`：越权必须阻断

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §58 L2799</sub>

如果 Agent 修改非授权路径，必须：

```text
标记 POLICY_VIOLATION
撤销变更
生成 Risk
进入 NEEDS_HUMAN_APPROVAL
```

---

## §59 并发执行规则

### **[P1]** `RULE-CONCURRENCY-001`：每个 Task 独立 worktree

<sub>level: P1 · status: active · enforced_by: `goalcli worktree-guard` · exit: 5 · source: §59 L2814</sub>

并发执行时：

```text
一个 Task = 一个 worktree = 一个 branch = 一个 PR
```

禁止多个 Agent 在同一 worktree 并发开发。

### **[P1]** `RULE-CONCURRENCY-002`：必须有 Lock 文件

<sub>level: P1 · status: active · enforced_by: `goalcli worktree-guard` · exit: 5 · source: §59 L2826</sub>

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

### **[P1]** `RULE-CONCURRENCY-003`：Release 必须串行

<sub>level: P1 · status: active · enforced_by: `goalcli worktree-guard` · exit: 5 · source: §59 L2851</sub>

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

## §64 上下文压缩规则

### **[P1]** `RULE-CONTEXT-COMPRESSION-001`：Goal 必须支持上下文压缩

<sub>level: P1 · status: active · enforced_by: `goalcli execution-context` · exit: 1 · source: §64 L3036</sub>

大型 Goal 必须提供：

```text
context-summary.md
decision-summary.md
evidence-summary.md
current-state.md
next-actions.md
```

用于 Agent 在上下文不足时恢复。

### **[P1]** `RULE-CONTEXT-COMPRESSION-002`：上下文摘要不能替代原始 Evidence

<sub>level: P1 · status: active · enforced_by: `goalcli execution-context` · exit: 1 · source: §64 L3052</sub>

摘要只能帮助阅读，不能作为完成证明。

---

## §77 度量指标规则

### **[P1]** `RULE-METRIC-001`：Goal Runtime 必须度量

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §77 L3430</sub>

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

### **[P1]** `RULE-METRIC-002`：低于阈值必须触发治理

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §77 L3450</sub>

例如：

```text
Evidence coverage < 95% → 阻断 Release
Traceability coverage < 100% → 阻断 Release
Repeated issue count > 2 → 必须新增 Rule/Harness
Gate false positive > 10% → 需要调参
```

---

## §83 goalcli 命令契约规则

### **[P1]** `RULE-GOALCLI-001`：goalcli 必须是 Goal Runtime 的唯一执行入口之一

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §83 L3644</sub>

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

### **[P1]** `RULE-GOALCLI-002`：goalcli 命令必须幂等

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §83 L3667</sub>

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

### **[P1]** `RULE-GOALCLI-003`：goalcli 命令必须统一输出机器结果

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §83 L3695</sub>

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

## §84 goalcli Exit Code 规则

### **[P1]** `RULE-GOALCLI-EXIT-001`：退出码必须标准化

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §84 L3721</sub>

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

## §85 报告产物规则

### **[P1]** `RULE-REPORT-001`：每个 Gate 必须生成报告

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §85 L3752</sub>

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

### **[P1]** `RULE-REPORT-002`：报告必须进入 Evidence

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §85 L3775</sub>

凡是用于证明完成的报告，都必须被 Evidence 引用。

```yaml
evidence_id: EVID-TASK-001-20260603-001
artifacts:
  - reports/traceability-check.json
  - reports/evidence-check.json
  - reports/ci-summary.md
```

---

## §100 Agent Team 规则

### **[P1]** `RULE-AGENT-TEAM-001`：复杂 Goal 必须拆 Agent 角色

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §100 L4305</sub>

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

### **[P1]** `RULE-AGENT-TEAM-002`：角色之间必须通过文件交接

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §100 L4324</sub>

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

## §101 Agent Handoff 规则

### **[P1]** `RULE-HANDOFF-001`：每次 Agent 交接必须写 Handoff Note

<sub>level: P1 · status: active · enforced_by: `goalcli goal-handover` · exit: 1 · source: §101 L4346</sub>

```md

### **[P1]** `RULE-HANDOFF-002`：没有 Handoff 不允许切换 Agent

<sub>level: P1 · status: active · enforced_by: `goalcli goal-handover` · exit: 1 · source: §101 L4375</sub>

否则容易出现：

```text
上下文丢失
重复实现
Evidence 断链
PR 内容漂移
```

---

## §113 仓库落地总目录规则

### **[P1]** `RULE-REPO-LAYOUT-001`：Goal Runtime 必须有固定目录结构

<sub>level: P1 · status: active · enforced_by: `goalcli boundary` · exit: 1 · source: §113 L4761</sub>

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

## §115 `goalcli.yaml` 配置规则

### **[P1]** `RULE-GOALCLI-CONFIG-001`：goalcli 必须有统一配置文件

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §115 L4877</sub>

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

## §120 goalcli 内核架构规则

### **[P1]** `RULE-GOALCLI-ARCH-001`：goalcli v0.1.0 先做裁判，不做复杂 Agent

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §120 L5149</sub>

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

### **[P1]** `RULE-GOALCLI-ARCH-002`：每个 Checker 必须实现统一接口

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §120 L5172</sub>

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

## §121 Checker 输出规则

### **[P1]** `RULE-CHECKER-001`：Checker 输出必须同时支持人读和机器读

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §121 L5197</sub>

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

### **[P1]** `RULE-CHECKER-002`：JSON 报告必须统一结构

<sub>level: P1 · status: active · enforced_by: `goalcli` · exit: 1 · source: §121 L5215</sub>

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

## §144 Agent Daily Runbook 规则

### **[P1]** `RULE-RUNBOOK-001`：Agent 每次执行必须先跑 Preflight

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §144 L5946</sub>

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

### **[P1]** `RULE-RUNBOOK-002`：Agent 执行顺序必须固定

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §144 L5963</sub>

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

### **[P1]** `RULE-RUNBOOK-003`：Agent 每次结束必须写收尾记录

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §144 L5981</sub>

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

## §145 Agent Stop Conditions 规则

### **[P0]** `RULE-STOP-001`：以下情况必须停止执行

<sub>level: P0 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §145 L6006</sub>

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

### **[P0]** `RULE-STOP-002`：停止后必须生成 Blocker 记录

<sub>level: P0 · status: active · enforced_by: `make governance-check` · exit: 1 · source: §145 L6023</sub>

```md

---

## §146 Agent Repair Loop 规则

### **[P1]** `RULE-REPAIR-001`：失败后最多自动修复 N 次

<sub>level: P1 · status: active · enforced_by: `goalcli self-healing-skeleton` · exit: 1 · source: §146 L6054</sub>

建议：

```yaml
repair_policy:
  max_auto_retries: 3
  retry_requires_new_evidence: true
  repeated_failure_state: NEEDS_REPLAN
```

### **[P1]** `RULE-REPAIR-002`：每次修复必须保留失败证据

<sub>level: P1 · status: active · enforced_by: `goalcli self-healing-skeleton` · exit: 1 · source: §146 L6067</sub>

禁止覆盖失败报告。

应该保留：

```text
reports/failures/<timestamp>-<gate>.json
reports/failures/<timestamp>-<gate>.txt
```

失败证据同样有价值，因为它会进入 Retrospective。

---

## §173 控制平面规则

### **[P1]** `RULE-CONTROL-001`：`.agent` 是 Goal Runtime 控制平面

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §173 L7100</sub>

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

### **[P1]** `RULE-CONTROL-002`：控制平面优先于散落文档

<sub>level: P1 · status: active · enforced_by: `goalcli goal-runtime` · exit: 1 · source: §173 L7122</sub>

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

## §175 命令事务规则

### **[P1]** `RULE-CMD-TXN-001`：goalcli 命令必须事务化

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §175 L7174</sub>

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

### **[P1]** `RULE-CMD-TXN-002`：命令失败必须可恢复

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §175 L7207</sub>

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

## §176 Dry-run 规则

### **[P1]** `RULE-DRYRUN-001`：所有破坏性命令必须支持 dry-run

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §176 L7231</sub>

必须支持：

```bash
goalcli release publish --dry-run
goalcli migrate --dry-run
goalcli worktree clean --dry-run
goalcli issues sync --dry-run
```

### **[P1]** `RULE-DRYRUN-002`：dry-run 必须输出计划

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §176 L7244</sub>

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

## §178 自动化安全边界规则

### **[P1]** `RULE-AUTO-SAFETY-001`：自动化只能扩大确定性，不能扩大不确定性

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §178 L7300</sub>

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

### **[P1]** `RULE-AUTO-SAFETY-002`：高风险自动化必须先进入 Simulation

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §178 L7326</sub>

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

## §189 Goal Runtime Dashboard 规则

### **[P1]** `RULE-DASHBOARD-001`：必须能生成静态 Dashboard

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §189 L7692</sub>

不需要第一阶段做 Web 服务，先生成静态报告：

```bash
goalcli dashboard generate
```

输出：

```text
reports/dashboard/index.md
```

### **[P1]** `RULE-DASHBOARD-002`：Dashboard 至少展示

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §189 L7708</sub>

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

## §190 Metrics Governance 规则

### **[P1]** `RULE-METRIC-GOV-001`：指标必须驱动治理，不只是展示

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §190 L7726</sub>

触发规则：

```text
Evidence Coverage < 100%：阻断 release
Traceability Coverage < 100%：阻断 release
Open V0 > 0：阻断所有 release
Open P0 violation > 0：阻断 stable
Patch adoption rate < 50%：触发 governance review
Rule drift count > 3：触发 drift cleanup goal
```

### **[P1]** `RULE-METRIC-GOV-002`：指标必须进入 Retrospective

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §190 L7741</sub>

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

## §193 Agent 记忆与文件事实规则

### **[P1]** `RULE-AGENT-MEMORY-001`：Agent 记忆不能作为事实源

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §193 L7811</sub>

Agent 可以参考历史记忆，但不能替代：

```text
当前仓库文件
Goal Pack
Registry
Evidence
CI 报告
Release Manifest
```

### **[P1]** `RULE-AGENT-MEMORY-002`：重要上下文必须写入文件

<sub>level: P1 · status: active · enforced_by: `goalcli agent-team-contract` · exit: 1 · source: §193 L7826</sub>

不能只存在聊天中。

必须落地：

```text
decision-log.md
context-summary.md
current-state.md
next-actions.md
```

---

## §194 Context Window 防爆规则

### **[P1]** `RULE-CONTEXT-WINDOW-001`：大型 Goal 必须分层摘要

<sub>level: P1 · status: active · enforced_by: `goalcli execution-context` · exit: 1 · source: §194 L7843</sub>

必须提供：

```text
00-current-state.md
01-decision-summary.md
02-open-blockers.md
03-next-actions.md
04-evidence-summary.md
```

### **[P1]** `RULE-CONTEXT-WINDOW-002`：Agent 恢复时先读摘要，再读原文

<sub>level: P1 · status: active · enforced_by: `goalcli execution-context` · exit: 1 · source: §194 L7857</sub>

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

## §206 Runtime Bootstrap 规则

### **[P1]** `RULE-BOOTSTRAP-001`：任何仓库接入 Goal Runtime 必须先 bootstrap

<sub>level: P1 · status: active · enforced_by: `goalcli install-runtime` · exit: 1 · source: §206 L8171</sub>

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

### **[P1]** `RULE-BOOTSTRAP-002`：bootstrap 必须幂等

<sub>level: P1 · status: active · enforced_by: `goalcli install-runtime` · exit: 1 · source: §206 L8196</sub>

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

## §207 Runtime Doctor 规则

### **[P1]** `RULE-DOCTOR-001`：必须提供一键诊断

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §207 L8224</sub>

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

### **[P1]** `RULE-DOCTOR-002`：doctor 不能修改状态

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §207 L8247</sub>

`doctor` 只诊断，不修复。

修复必须走：

```bash
goalcli repair
```

---

## §209 Rule Compiler 规则

### **[P1]** `RULE-COMPILER-001`：Markdown Rule 必须可编译成 Policy Index

<sub>level: P1 · status: active · enforced_by: `goalcli cli-contract` · exit: 1 · source: §209 L8285</sub>

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

### **[P1]** `RULE-COMPILER-002`：编译失败必须阻断 Release

<sub>level: P1 · status: active · enforced_by: `goalcli cli-contract` · exit: 1 · source: §209 L8305</sub>

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

## §211 Gate Dependency Graph 规则

### **[P1]** `RULE-GATE-DAG-001`：Gate 必须有依赖图

<sub>level: P1 · status: active · enforced_by: `goalcli makefile-baseline` · exit: 1 · source: §211 L8356</sub>

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

### **[P1]** `RULE-GATE-DAG-002`：禁止循环依赖

<sub>level: P1 · status: active · enforced_by: `goalcli makefile-baseline` · exit: 1 · source: §211 L8372</sub>

例如：

```text
evidence-check 依赖 release-check
release-check 又依赖 evidence-check
```

必须阻断。

---

## §212 State Reconciliation 规则

### **[P1]** `RULE-RECONCILE-001`：必须能修复状态不一致

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §212 L8387</sub>

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

### **[P1]** `RULE-RECONCILE-002`：reconcile 不能静默改状态

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §212 L8409</sub>

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

## §216 Agent Lease 规则

### **[P1]** `RULE-LEASE-001`：Agent 执行必须持有 Lease

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §216 L8524</sub>

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

### **[P1]** `RULE-LEASE-002`：Lease 超时必须释放或接管

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §216 L8545</sub>

```bash
goalcli lease recover
```

防止僵尸 Agent 占用任务。

---

## §217 Agent Heartbeat 规则

### **[P1]** `RULE-HEARTBEAT-001`：长任务必须写心跳

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §217 L8557</sub>

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

### **[P1]** `RULE-HEARTBEAT-002`：无心跳任务进入 STALE

<sub>level: P1 · status: active · enforced_by: `goalcli runtime-health` · exit: 1 · source: §217 L8575</sub>

超过 TTL：

```text
Task state = STALE
```

必须人工或自动恢复。

---

## §220 Change Batch 规则

### **[P1]** `RULE-BATCH-001`：不同风险类型不能混批

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §220 L8643</sub>

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

## §225 Trust Score 规则

### **[P1]** `RULE-TRUST-001`：Goal 必须有 Trust Score

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §225 L8794</sub>

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

### **[P1]** `RULE-TRUST-002`：低 Trust Score 限制发布通道

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §225 L8810</sub>

```text
>= 90 stable
80-89 rc
70-79 beta
<70 blocked
```

---

## §228 Governance Cadence 规则

### **[P1]** `RULE-GOV-CADENCE-001`：必须有固定治理节奏

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §228 L8878</sub>

建议：

```text
每周：open violations / blocked goals / pending patches
每双周：drift / rule bloat / fixture health
每月：downstream adoption score / release audit
每季度：rule sunset / maturity review
```

### **[P1]** `RULE-GOV-CADENCE-002`：治理会议必须产出对象

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §228 L8891</sub>

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

## §229 Dashboard Health 规则

### **[P1]** `RULE-DASHBOARD-HEALTH-001`：Dashboard 必须显示红线指标

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §229 L8908</sub>

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

### **[P1]** `RULE-DASHBOARD-HEALTH-002`：红线指标必须阻断 stable

<sub>level: P1 · status: indexed · enforced_by: `（待机器化）` · source: §229 L8924</sub>

Dashboard 不是展示板，而是治理入口。
