# xlib-standard 完整 Goal 可执行方案 v2.9.3 Complete

> 版本：v2.9.3 Complete
> 目标仓库：`https://github.com/ZoneCNH/xlib-standard`  
> 首批下游验证：`github.com/ZoneCNH/kernel`、`github.com/ZoneCNH/configx`
> 执行标准：Goal Runtime v3.1 + First-Principles + Harness Engineering + Karpathy 四原则 + Self-improving + AutoResearch + Compound Engineering
> 生成日期：2026-06-02
> 替代版本：v2.9、v2.9.1、v2.9.2

---

## 0. 第三轮审计结论

本版是一个**主文档重编译版本**，不再把 v2.9.1 / v2.9.2 的补丁作为附录，而是把它们合并进主执行流。

第三轮审计发现，v2.9.2 已补齐大量落地缺口，但仍有 3 类风险需要集成进完整版：

1. **补丁附录化风险**：v2.9.2 的新增 Issue、命令、Makefile、验收标准散落在后文，Agent 实施时可能只读取前文 P0/P1/P2 Pack 而漏掉补丁。
2. **命名与自符合风险**：旧名 `baselib-template` / `foundationx` 缺少明确 guard；且下游 adoption 前应先完成 `xlib-standard` 自身 standard-source profile attestation。
3. **执行入口漂移风险**：Issue Registry、Command Registry、Makefile Target Registry、Policy Schema、Toolchain、Evidence Path、GitHub Settings、Runtime File Ownership 必须统一进入 SSOT，否则后续多 Agent 实施会漂移。

因此 v2.9.3 的裁决是：

```text
v2.9.3 Complete = v2.9 主方案
  + v2.9.1 治理对象与上下文补丁
  + v2.9.2 执行一致性补丁
  + v2.9.3 命名一致性、自符合证明、主文档重编译
```

实际执行应以 **v2.9.3 Complete** 为准。

---

## 1. 最终目标

`xlib-standard` 的目标不是做一个普通 Go 模板，而是升级为：

```text
Standard Source
  + Go Reference Template
  + Generator
  + Harness
  + Evidence Runtime
  + Constitution Runtime
  + Downstream Conformance Runtime
```

v2.9.3 完成后必须具备：

1. P0 Minimal Kernel：防止 main 开发、无 worktree、无 Evidence DONE、x.go 反向依赖、production secret default、Makefile/CI/manifest 缺失。
2. P1 Governance Hardening：Agent Team、Scope Lock、PR Contract、Acceptance Matrix、Runtime Health、GitHub Governance、Toolchain、Policy Schema、Evidence Artifact、Self-Healing Skeleton。
3. P2 Runtime & Conformance Automation：Runtime install/upgrade、release readiness、evidence replay、conformance attestation、Standard/Gate/Evidence Pack、kernel/configx adoption dry-run proof。
4. Self-improving：失败必须进入 Failure Taxonomy、Root Cause、Regression Memory、Patch Candidate。
5. Progressive adoption：P3/P4 高级治理在 P0/P1/P2 完成前冻结。

---

## 2. 第一性原理裁决

### 2.1 问题本质

`xlib-standard` 的本质问题不是“模板怎么写”，而是：

> 如何让所有基础库在多 Agent、多 worktree、多仓库、多 release 阶段中，持续遵守同一套可验证工程文明。

### 2.2 不可再拆解的基本真理

```text
TRUTH-001  规则不进入 Gate，就不是规则。
TRUTH-002  Gate 不产生 Evidence，就不能证明完成。
TRUTH-003  Evidence 不进入 Release Manifest，就不能审计。
TRUTH-004  Agent 不使用 worktree，就不能安全并发。
TRUTH-005  Agent 不受 Scope Lock 约束，就会污染无关文件。
TRUTH-006  标准不能安装到下游，就不是运行时。
TRUTH-007  下游不能生成 Attestation，就不能声称符合标准。
TRUTH-008  失败不产生 Patch，就没有 Self-improving。
TRUTH-009  方法论没有触发条件、输出物、Gate 和退出机制，就是仪式。
TRUTH-010  标准必须被采用，才有价值。
```

### 2.3 可打破限制

```text
LIMIT-001  不必一开始实现完整 v2.9.3；先 P0，再 P1，再 P2。
LIMIT-002  不必一开始做 Fleet Dashboard；P2 完成后再做。
LIMIT-003  不必一开始做 Formal Model Checking；Release Readiness Formula 先足够。
LIMIT-004  不必所有 Article 都变成 Gate；ACTIVE + BLOCKING 才必须 Gate。
LIMIT-005  不必所有任务 Full Runtime；按 C0-C5 复杂度分级。
LIMIT-006  不必所有下游同时迁移；先 xlib-standard self，再 kernel，再 configx。
```

---

## 3. 最高执行约束

### 3.1 禁止 main 主线开发

`main` 是发布主线、审计主线、Evidence 汇聚点，不是开发工作区。

阻断对象：

```text
local_write on main -> BLOCK
local_readonly on clean main -> ALLOW
ci_main_verify on clean main -> ALLOW
release_verify on clean main/tag -> ALLOW
```

### 3.2 强制 git worktree

写入型任务必须使用：

```bash
git fetch origin
git checkout main
git pull --ff-only origin main

git worktree add ../.worktrees/xlib-standard/<issue-id>   -b <branch-name> main
```

### 3.3 强制 Agent Teams

C3+ 任务必须使用 Agent Team Contract。最小角色：

```text
Lead / Implementation / Test / Harness / Evidence / Review
```

### 3.4 统一完成声明

```text
DONE with evidence:
- scope:
- issues:
- worktree:
- branch:
- changed_files:
- gates:
- evidence:
- review:
- release_impact:
- known_gaps:
- follow_up:
```

---

## 4. 冻结规则

P0/P1/P2 完成前冻结：

```text
Fleet Dashboard
Plugin Sandbox
Third-party Policy
Ecosystem Certification
Formal Model Checking
Advanced Method Effectiveness
Full Release Train Automation
Multi-repo Auto Migration
Governance Query DSL
```

允许范围：

```text
P0: Minimal Kernel
P1: Governance Hardening
P2: Runtime & Conformance Automation
```

---

## 5. Master Goal Runtime v3.1

```yaml
goal_id: GOAL-20260602-001
title: xlib-standard v2.9.3 Constitution Runtime & Conformance Automation
mode: full
owner: lead
repository: github.com/ZoneCNH/xlib-standard
state_machine:
  - INIT
  - CONTEXT_READY
  - GOAL_READY
  - SPEC_READY
  - DESIGN_READY
  - PLAN_READY
  - TASKS_READY
  - EXECUTING
  - VERIFYING
  - REVIEWING
  - RELEASING
  - RETROSPECTING
  - DONE
exception_states:
  - BLOCKED
  - FAILED
  - NEEDS_RESEARCH
  - NEEDS_DECISION
  - NEEDS_REPLAN
  - NEEDS_ROLLBACK
  - NEEDS_HUMAN_APPROVAL
  - INCONSISTENT_STATE
success_criteria:
  - P0 Issues P0-001..P0-016 completed with evidence.
  - P1 Issues P1-001..P1-021 completed with evidence.
  - P2 Issues P2-001..P2-015 completed with evidence.
  - goalcli commands in command registry are implemented or explicitly planned.
  - Makefile targets in baseline registry exist and fail non-zero on failure.
  - CI invokes Makefile and does not duplicate Gate logic.
  - Release Manifest Skeleton can be generated and checksummed.
  - xlib-standard self-conformance attestation is generated before downstream adoption.
  - kernel/configx adoption can run in patch-only/dry-run mode.
```

---

## 6. 标准目录结构目标

```text
CONSTITUTION.md
CHANGELOG.md
.tool-versions
.agent/
  minimal-kernel.yaml
  enforcement-levels.yaml
  execution-context.yaml
  issue-registry.yaml
  command-registry.yaml
  makefile-target-registry.yaml
  makefile-baseline.yaml
  boundary.yaml
  security.yaml
  done-assertion.yaml
  evidence-artifact-policy.yaml
  runtime-health.yaml
  conformance-profiles.yaml
  downstream-registry.yaml
  runtime-install.yaml
  runtime-upgrade.yaml
  runtime-file-ownership.yaml
  downstream-adoption-modes.yaml
  downstream-baseline-scan.yaml
  failure-taxonomy.yaml
  root-cause.yaml
  regression-memory.yaml
  harness-patches.yaml
  rule-patches.yaml
  prompt-patches.yaml
.github/
  CODEOWNERS
  ISSUE_TEMPLATE/
  pull_request_template.md
  workflows/ci.yml
cmd/goalcli/main.go
internal/goalcli/
  cli/
  context/
  report/
  registry/
  guards/
  evidence/
  boundary/
  security/
  manifest/
  runtime/
  conformance/
  pack/
  downstream/
  schema/
  testfixture/
contracts/
  goalcli-report.schema.json
  release-manifest.schema.json
  conformance-attestation.schema.json
  agent-policy.schema.json
  issue-registry.schema.json
  command-registry.schema.json
  execution-context.schema.json
docs/
  quickstart.md
  troubleshooting.md
  standard/
release/
  manifest/.gitkeep
  evidence/.gitkeep
testkit/governance/fixtures/
```

---

## 7. P0 Minimal Kernel Issue Pack

| Issue  | Title                                   | Type               | Complexity | Files                                                                                                                           | Gate                                                                     | Acceptance Summary                                                                                                               |
| ------ | --------------------------------------- | ------------------ | ---------- | ------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| P0-001 | Minimal Constitution                    | constitution       | C2         | CONSTITUTION.md; docs/standard/constitution.md                                                                                  | docs-check                                                               | 最小宪法、Minimal Kernel、禁止 main 开发、worktree、DONE with evidence、no x.go reverse dependency、no production secret default |
| P0-002 | Minimal Kernel Policy                   | policy             | C2         | .agent/minimal-kernel.yaml; .agent/enforcement-levels.yaml                                                                      | goalcli minimal-kernel                                                  | P0 invariants 注册；P0 不允许 local override；enforcement level 完整                                                             |
| P0-003 | goalcli CLI Skeleton                   | runtime            | C2         | cmd/goalcli/main.go; internal/goalcli/cli/\*\*                                                                                | goalcli version; goalcli doctor; go test ./...                         | version/help/doctor 可运行；exit code 稳定；预留 --json/--output/--explain                                                       |
| P0-004 | main-guard                              | guard              | C2         | internal/goalcli/guards/main_guard.go; testkit/governance/fixtures/main-guard/\*\*                                             | goalcli main-guard                                                      | local_write + main 阻断；release_verify clean main 允许；失败信息提示 worktree                                                   |
| P0-005 | worktree-guard                          | guard              | C2         | internal/goalcli/guards/worktree_guard.go; testkit/governance/fixtures/worktree-guard/\*\*                                     | goalcli worktree-guard                                                  | local_write 必须 worktree；CI / release verify 不误阻断                                                                          |
| P0-006 | evidence-check                          | evidence           | C2         | internal/goalcli/evidence/**; .agent/done-assertion.yaml; fixtures/evidence/**                                                 | goalcli evidence-check                                                  | DONE without evidence 失败；必需 command/commit/worktree/known_gaps                                                              |
| P0-007 | boundary no-xgo-import check            | boundary           | C2         | internal/goalcli/boundary/**; .agent/boundary.yaml; fixtures/boundary/**                                                       | goalcli boundary                                                        | 禁止 github.com/bytechainx/x.go 与 github.com/ZoneCNH/x.go 反向依赖                                                              |
| P0-008 | no-secret-default check                 | security           | C2         | internal/goalcli/security/**; .agent/security.yaml; fixtures/security/**                                                       | goalcli security                                                        | 禁止默认读取 /home/k8s/secrets/env/\*；允许作为调用方部署路径说明                                                                |
| P0-009 | Makefile governance-check               | harness            | C2         | Makefile                                                                                                                        | .make governance-check; make release-check                               | main/worktree/evidence/boundary/security 进入 governance-check；target 失败返回非 0                                              |
| P0-010 | CI Required Checks Skeleton             | ci                 | C2         | .github/workflows/ci.yml                                                                                                        | goalcli ci-job-matrix                                                   | CI 调用 Makefile；不 continue-on-error；最小权限；上传 evidence artifact                                                         |
| P0-011 | Release Manifest Skeleton               | release            | C2         | contracts/release-manifest.schema.json; internal/goalcli/manifest/\*\*; release/manifest/.gitkeep; .gitignore                  | goalcli manifest                                                        | manifest 字段完整；checksum 可生成；generated manifest 不提交源码历史                                                            |
| P0-012 | DONE with evidence Protocol             | evidence           | C1         | docs/standard/evidence-protocol.md; .github/pull_request_template.md; .agent/done-assertion.yaml                                | goalcli done-assertion; docs-check                                      | PR 模板和文档包含 DONE with evidence 字段                                                                                        |
| P0-013 | Execution Context Policy                | guard              | C2         | .agent/execution-context.yaml; internal/goalcli/context/**; fixtures/execution-context/**                                      | goalcli main-guard --context ...; goalcli worktree-guard --context ... | local_write/local_readonly/ci_pull_request/ci_main_verify/release_verify 语义完整                                                |
| P0-014 | goalcli CLI Contract and Report Schema | runtime            | C2         | docs/standard/goalcli-cli-contract.md; contracts/goalcli-report.schema.json; internal/goalcli/report/\*\*                    | goalcli cli-contract; goalcli contracts                                | exit code、JSON report、Finding、remediation、schema 完整                                                                        |
| P0-015 | Issue Registry and Command Registry     | governance-runtime | C2         | .agent/issue-registry.yaml; .agent/command-registry.yaml; .agent/makefile-target-registry.yaml; internal/goalcli/registry/\*\* | goalcli issue-registry; goalcli command-registry                       | Issue/Command/Makefile target 统一 SSOT，防漂移                                                                                  |
| P0-016 | Makefile Baseline Target Inventory      | harness            | C2         | .agent/makefile-baseline.yaml; Makefile; internal/goalcli/makefile/\*\*                                                        | goalcli makefile-baseline; make governance-check; make release-check    | fmt/vet/lint/test/race/boundary/security/contracts/docs-check/evidence/release targets 完整                                      |

### 7.1 P0 验收命令

```bash
XLIB_CONTEXT=local_write GOWORK=off make governance-check
XLIB_CONTEXT=release_verify GOWORK=off make release-check
GOWORK=off go test ./...
GOWORK=off go run ./cmd/goalcli doctor
GOWORK=off go run ./cmd/goalcli cli-contract
GOWORK=off go run ./cmd/goalcli issue-registry
GOWORK=off go run ./cmd/goalcli command-registry
GOWORK=off go run ./cmd/goalcli makefile-baseline
```

## 8. P1 Governance Hardening Issue Pack

| Issue  | Title                                       | Type              | Complexity | Files                                                                                                                                                                                                                                             | Gate                                                                       | Acceptance Summary                                                                             |
| ------ | ------------------------------------------- | ----------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| P1-001 | Agent Team Contract                         | agent-runtime     | C3         | .agent/team-contract.yaml; docs/standard/agent-team-contract.md; internal/goalcli/agent/\*\*                                                                                                                                                     | goalcli agent-team-contract                                               | C3+ 任务声明 team/roles/scope/worktree/gates/evidence；禁止自审                                |
| P1-002 | Scope Lock Guard                            | governance        | C3         | .agent/scope-locks.yaml; internal/goalcli/scope/**; fixtures/scope-lock/**                                                                                                                                                                       | goalcli scope-lock                                                        | owned/read_only/forbidden_paths 生效；scope leak 失败                                          |
| P1-003 | PR Template Contract                        | pr-governance     | C2         | .github/pull_request_template.md; .agent/pr-template-contract.yaml; internal/goalcli/pr/\*\*                                                                                                                                                     | goalcli pr-template                                                       | PR 包含 issue/scope/worktree/gates/evidence/known gaps/rollback/release impact                 |
| P1-004 | Acceptance Matrix                           | traceability      | C3         | .agent/acceptance-matrix.yaml; docs/standard/acceptance-matrix.md; internal/goalcli/traceability/\*\*                                                                                                                                            | goalcli acceptance-matrix                                                 | Requirement → AC → Gate/Test → Evidence → Status 无断链                                        |
| P1-005 | Runtime HealthCheck                         | runtime           | C2         | .agent/runtime-health.yaml; internal/goalcli/runtime/health.go; docs/standard/runtime-health.md                                                                                                                                                  | goalcli runtime-health; make runtime-health                               | 检查 Constitution/.agent/goalcli/Makefile/CI/contracts/manifest schema/evidence protocol      |
| P1-006 | Standard Upgrade Skeleton                   | runtime-upgrade   | C3         | .agent/standard-upgrade.yaml; docs/standard/standard-upgrade.md; internal/goalcli/upgrade/\*\*                                                                                                                                                   | goalcli upgrade-standard --dry-run                                        | current/target version、diff、rollback note、dry-run report                                    |
| P1-007 | Conformance Profiles                        | conformance       | C3         | .agent/conformance-profiles.yaml; docs/standard/conformance-profiles.md; internal/goalcli/conformance/\*\*                                                                                                                                       | goalcli conformance-profile                                               | standard-source/l0-kernel/l1-shared/l2-infra/experimental profile 完整                         |
| P1-008 | Downstream Registry                         | downstream        | C2         | .agent/downstream-registry.yaml; docs/standard/downstream-registry.md; internal/goalcli/downstream/\*\*                                                                                                                                          | goalcli downstream-registry                                               | 登记 kernel/configx/observex/testkitx/postgresx/redisx/kafkax/taosx/ossx/clickhousex           |
| P1-009 | Self-Healing Skeleton                       | self-improving    | C3         | .agent/failure-taxonomy.yaml; .agent/root-cause.yaml; .agent/regression-memory.yaml; .agent/harness-patches.yaml; .agent/rule-patches.yaml; .agent/prompt-patches.yaml                                                                            | goalcli self-healing-skeleton                                             | 失败分类、Root Cause、Regression Memory、Patch logs 最小闭环                                   |
| P1-010 | Documentation Quickstart                    | docs              | C2         | docs/quickstart.md; docs/standard/worktree-protocol.md; docs/standard/evidence-protocol.md; docs/troubleshooting.md                                                                                                                               | docs-check; docs-command-sync-check                                        | 新开发者可按文档创建 worktree、运行 governance-check、处理 failed gate                         |
| P1-011 | Goal Runtime v3.1 Governance Objects        | goal-runtime      | C3         | .agent/adr-register.yaml; .agent/decision-log.yaml; .agent/state-machine.yaml; .agent/change-propagation-matrix.yaml; .agent/human-approval-gates.yaml; .agent/failure-budget.yaml; .agent/rollback-protocol.yaml; .agent/definition-of-done.yaml | goalcli goal-runtime; goalcli state-machine; goalcli change-propagation | ADR/Decision/State/Propagation/Human Approval/Failure Budget/Rollback/DoD 完整                 |
| P1-012 | GitHub Governance Controls                  | github-governance | C2         | .github/CODEOWNERS; .github/ISSUE_TEMPLATE/\*.yml; .github/pull_request_template.md; docs/standard/branch-protection.md; .agent/github-governance.yaml                                                                                            | goalcli github-governance; goalcli codeowners                            | CODEOWNERS、Issue/PR 模板、Branch Protection 文档、Admin bypass break-glass                    |
| P1-013 | Supply Chain Security Baseline              | security          | C3         | .agent/supply-chain-security.yaml; docs/standard/supply-chain-security.md; .github/workflows/ci.yml                                                                                                                                               | goalcli supply-chain; make security; make lint                            | govulncheck/golangci-lint/action pinning/least privilege/go.mod drift                          |
| P1-014 | Release Versioning and Changelog Protocol   | release           | C2         | CHANGELOG.md; docs/standard/release-versioning.md; docs/standard/migration-note.md; .agent/release-versioning.yaml; .gitignore                                                                                                                    | goalcli changelog; goalcli versioning; goalcli generated-artifacts      | 行为变更需 changelog；breaking change 需 migration；tag protocol                               |
| P1-015 | Governance Test Strategy                    | testing           | C3         | docs/standard/governance-test-strategy.md; testkit/governance/README.md; internal/goalcli/testfixture/\*\*                                                                                                                                       | make governance-fixture-test; go test ./...                                | P0 guard valid/invalid fixture；negative tests 进入 go test                                    |
| P1-016 | AutoResearch Trigger and Record Protocol    | autoresearch      | C2         | .agent/autoresearch.yaml; docs/standard/autoresearch.md; .agent/research-record-template.md                                                                                                                                                       | goalcli autoresearch                                                      | 外部链接/工具链/依赖/不确定事实触发 research record 和 decision impact                         |
| P1-017 | Policy Schema Registry and YAML Validation  | contracts         | C3         | contracts/agent-policy.schema.json; contracts/issue-registry.schema.json; contracts/command-registry.schema.json; contracts/execution-context.schema.json; internal/goalcli/schema/\*\*                                                          | goalcli policy-schema; goalcli contracts                                 | 所有关键 .agent/\*.yaml schema validation；invalid config 不静默通过                           |
| P1-018 | GitHub Settings Apply and Verify Protocol   | github-governance | C3         | .agent/github-settings.yaml; docs/standard/github-settings.md; scripts/github/verify_settings.sh                                                                                                                                                  | goalcli github-settings --verify; goalcli codeowners                     | Required checks/branch protection/rulesets 可验证；apply 不隐式执行                            |
| P1-019 | Toolchain Pinning Baseline                  | supply-chain      | C2         | .tool-versions; .agent/toolchain.yaml; docs/standard/toolchain.md                                                                                                                                                                                 | goalcli toolchain; make lint; make security                               | 本地/CI 工具版本 SSOT；禁止 required tools 使用 latest                                         |
| P1-020 | Evidence Artifact Path and Retention Policy | evidence          | C2         | .agent/evidence-artifact-policy.yaml; docs/standard/evidence-artifacts.md                                                                                                                                                                         | goalcli evidence-artifacts                                                | release/evidence 与 release/manifest 路径、retention、DONE links 规范                          |
| P1-021 | Naming Consistency and Legacy Name Guard    | docs/governance   | C2         | .agent/naming-policy.yaml; docs/standard/naming.md; internal/goalcli/naming/\*\*                                                                                                                                                                 | goalcli naming; docs-check                                                | 默认名称统一 xlib-standard/kernel；旧名 baselib-template/foundationx 仅允许迁移/ADR/兼容上下文 |

### 8.1 P1 验收命令

```bash
GOWORK=off make p1-governance-check
GOWORK=off make governance-fixture-test
GOWORK=off make security
GOWORK=off make lint
GOWORK=off go run ./cmd/goalcli policy-schema
GOWORK=off go run ./cmd/goalcli github-settings --verify
GOWORK=off go run ./cmd/goalcli toolchain
GOWORK=off go run ./cmd/goalcli evidence-artifacts
GOWORK=off make release-check
```

## 9. P2 Runtime & Conformance Automation Issue Pack

| Issue  | Title                                      | Type                | Complexity | Files                                                                                                                                 | Gate                                                                                                           | Acceptance Summary                                                                                  |
| ------ | ------------------------------------------ | ------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| P2-001 | Runtime Install                            | runtime             | C3         | internal/goalcli/runtime/install.go; .agent/runtime-install.yaml; docs/standard/runtime-install.md                                   | goalcli install-runtime --dry-run; goalcli runtime-health                                                    | 可安装 CONSTITUTION/.agent/Makefile/CI/contracts/docs/release skeleton                              |
| P2-002 | Runtime Upgrade                            | runtime-upgrade     | C3         | internal/goalcli/runtime/upgrade.go; .agent/runtime-upgrade.yaml; docs/standard/runtime-upgrade.md                                   | goalcli upgrade-runtime --dry-run                                                                             | dry-run/apply/rollback report；失败不更新 adoption version                                          |
| P2-003 | Release Readiness Formula                  | release             | C3         | internal/goalcli/release/readiness.go; .agent/release-readiness-formula.yaml; docs/standard/release-readiness.md                     | goalcli release-ready                                                                                         | release_ready 公式；failed gate/missing evidence/dirty workspace/P0 blocker 均 not ready            |
| P2-004 | Evidence Replay                            | evidence            | C3         | internal/goalcli/evidence/replay.go; .agent/evidence-replay.yaml; docs/standard/evidence-replay.md                                   | goalcli evidence-replay                                                                                       | gate result/release manifest/runtime-health/conformance/security/boundary replay                    |
| P2-005 | Conformance Attestation                    | conformance         | C3         | internal/goalcli/conformance/attestation.go; contracts/conformance-attestation.schema.json; docs/standard/conformance-attestation.md | goalcli attest-conformance; goalcli contracts                                                                | 生成 attestation；failed gate 不得 passed attestation                                               |
| P2-006 | Standard Pack                              | pack                | C3         | internal/goalcli/pack/standard.go; .agent/standard-pack.yaml; docs/standard/standard-pack.md                                         | goalcli pack-standard                                                                                         | 打包 Constitution/policies/docs/contracts，生成 manifest/checksum                                   |
| P2-007 | Gate Pack                                  | pack                | C3         | internal/goalcli/pack/gate.go; .agent/gate-pack.yaml; docs/standard/gate-pack.md                                                     | goalcli pack-gate                                                                                             | goalcli commands/Makefile/CI snippets/fixtures/remediation/profile gates                           |
| P2-008 | Evidence Pack                              | pack                | C3         | internal/goalcli/pack/evidence.go; .agent/evidence-pack.yaml; docs/standard/evidence-pack.md                                         | goalcli pack-evidence                                                                                         | Evidence schema/manifest schema/replay/DONE protocol 打包                                           |
| P2-009 | kernel Downstream Adoption                 | downstream-adoption | C4         | patch bundle / downstream local path                                                                                                  | goalcli downstream-adoption --mode patch-only --repo kernel; goalcli attest-conformance --profile l0-kernel  | kernel 生成 adoption/runtime/profile/manifest/attestation，且不依赖 x.go                            |
| P2-010 | configx Downstream Adoption                | downstream-adoption | C4         | patch bundle / downstream local path                                                                                                  | goalcli downstream-adoption --mode patch-only --repo configx; goalcli attest-conformance --profile l1-shared | configx 生成 adoption/runtime/profile/manifest/attestation，符合显式 Config                         |
| P2-011 | SBOM and License Check Roadmap             | supply-chain        | C2         | .agent/sbom-roadmap.yaml; docs/standard/sbom-license-roadmap.md                                                                       | goalcli sbom-roadmap                                                                                          | P2 不必完整实现 SBOM，但必须有 P3/P4 roadmap 和阻断条件                                             |
| P2-012 | Downstream Adoption Modes                  | downstream-adoption | C4         | .agent/downstream-adoption-modes.yaml; docs/standard/downstream-adoption-modes.md; internal/goalcli/downstream/adoption_modes.go     | goalcli downstream-adoption --mode patch-only --repo kernel/configx                                           | 支持 local_path / patch_only / pr_plan；无授权不跨仓写                                              |
| P2-013 | Runtime File Ownership and Merge Safety    | runtime-upgrade     | C4         | .agent/runtime-file-ownership.yaml; docs/standard/runtime-file-ownership.md; internal/goalcli/runtime/file_ownership.go              | goalcli runtime-file-ownership; goalcli upgrade-runtime --dry-run                                            | XLIB_OWNED/USER_OWNED/MERGE_REQUIRED/GENERATED；不覆盖用户文件                                      |
| P2-014 | Downstream Baseline Scan Before Adoption   | downstream-adoption | C4         | .agent/downstream-baseline-scan.yaml; docs/standard/downstream-baseline-scan.md; internal/goalcli/downstream/baseline_scan.go        | goalcli downstream-baseline --repo kernel/configx --mode patch-only                                           | adoption 前扫描 module/package/Makefile/CI/contracts/docs/x.go/secret/release flow，输出 risk score |
| P2-015 | xlib-standard Self-Conformance Attestation | conformance         | C3         | release/manifest/conformance-attestation.json; docs/reports/self-conformance.md                                                       | goalcli attest-conformance --profile standard-source                                                          | 下游 adoption 前先证明 xlib-standard 自身符合 standard-source profile                               |

### 9.1 P2 验收命令

```bash
GOWORK=off make p2-runtime-check
GOWORK=off go run ./cmd/goalcli runtime-file-ownership
GOWORK=off go run ./cmd/goalcli attest-conformance --profile standard-source
GOWORK=off go run ./cmd/goalcli downstream-baseline --repo kernel --mode patch-only
GOWORK=off go run ./cmd/goalcli downstream-baseline --repo configx --mode patch-only
GOWORK=off go run ./cmd/goalcli downstream-adoption --mode patch-only --repo kernel
GOWORK=off go run ./cmd/goalcli downstream-adoption --mode patch-only --repo configx
GOWORK=off go run ./cmd/goalcli attest-conformance --profile l0-kernel
```

## 10. Execution Context Policy

```yaml
schema_version: "1.0"
contexts:
  local_write:
    branch_main_allowed: false
    worktree_required: true
    dirty_tree_allowed: true
    allowed_actions: ["edit", "generate", "test", "commit"]
  local_readonly:
    branch_main_allowed: true
    worktree_required: false
    dirty_tree_allowed: false
    allowed_actions: ["read", "inspect", "fetch", "pull"]
  ci_pull_request:
    branch_main_allowed: false
    worktree_required: false
    dirty_tree_allowed: false
    allowed_actions: ["check", "test", "artifact"]
  ci_main_verify:
    branch_main_allowed: true
    worktree_required: false
    dirty_tree_allowed: false
    allowed_actions: ["verify", "test", "artifact"]
  release_verify:
    branch_main_allowed: true
    worktree_required: false
    dirty_tree_allowed: false
    allowed_actions: ["verify", "manifest", "attest"]
```

规则：

```text
main-guard:
  BLOCK when branch == main AND context == local_write
  ALLOW when branch == main AND context in [local_readonly, ci_main_verify, release_verify] AND workspace clean

worktree-guard:
  BLOCK when context == local_write AND worktree == false
  ALLOW when context in [ci_pull_request, ci_main_verify, release_verify]
```

## 11. goalcli Command Registry

| Command                | Issue                | Domain            | Context                  | Purpose                                                |
| ---------------------- | -------------------- | ----------------- | ------------------------ | ------------------------------------------------------ |
| version                | P0-003               | runtime           | local/ci                 | 输出版本与 runtime component versions                  |
| doctor                 | P0-003               | runtime           | local/ci                 | 检查基础运行环境                                       |
| main-guard             | P0-004/P0-013        | guard             | local/ci/release         | 上下文感知 main 写入阻断                               |
| worktree-guard         | P0-005/P0-013        | guard             | local/ci/release         | 上下文感知 worktree 要求                               |
| evidence-check         | P0-006               | evidence          | local/ci                 | 验证 DONE with evidence / Evidence Bundle              |
| boundary               | P0-007               | boundary          | local/ci                 | no-xgo-import                                          |
| security               | P0-008               | security          | local/ci                 | no-secret-default + secret/supply-chain baseline hooks |
| manifest               | P0-011               | release           | local/ci/release         | 生成/校验 release manifest skeleton                    |
| cli-contract           | P0-014               | runtime           | local/ci                 | 校验 exit code / report schema                         |
| issue-registry         | P0-015               | governance        | local/ci                 | Issue SSOT 校验                                        |
| command-registry       | P0-015               | governance        | local/ci                 | Command/Makefile/CI mapping 校验                       |
| makefile-baseline      | P0-016               | harness           | local/ci                 | Makefile required targets 校验                         |
| agent-team-contract    | P1-001               | agent-runtime     | local/ci                 | Team Contract 校验                                     |
| scope-lock             | P1-002               | governance        | local/ci                 | Scope Lock 校验                                        |
| pr-template            | P1-003               | pr-governance     | ci_pr                    | PR body/template 校验                                  |
| acceptance-matrix      | P1-004               | traceability      | local/ci                 | REQ/AC/Gate/Evidence 矩阵校验                          |
| runtime-health         | P1-005               | runtime           | local/ci/downstream      | Runtime HealthCheck                                    |
| goal-runtime           | P1-011               | goal-runtime      | local/ci                 | ADR/Decision/State/DoD 对象校验                        |
| github-settings        | P1-018               | github-governance | manual/ci with token     | GitHub settings verify，不隐式 apply                   |
| policy-schema          | P1-017               | contracts         | local/ci                 | `.agent/*.yaml` schema validation                      |
| toolchain              | P1-019               | supply-chain      | local/ci                 | Toolchain pinning 校验                                 |
| evidence-artifacts     | P1-020               | evidence          | local/ci                 | Evidence path/retention 校验                           |
| naming                 | P1-021               | governance        | local/ci                 | 旧名/默认名一致性扫描                                  |
| install-runtime        | P2-001               | runtime           | local/downstream         | 标准安装 dry-run/apply                                 |
| upgrade-runtime        | P2-002               | runtime-upgrade   | local/downstream         | 标准升级 dry-run/apply/rollback                        |
| release-ready          | P2-003               | release           | local/ci/release         | Release readiness formula                              |
| evidence-replay        | P2-004               | evidence          | local/ci/release         | Evidence replay                                        |
| attest-conformance     | P2-005/P2-015        | conformance       | local/ci/downstream      | 生成符合性证明                                         |
| pack-standard          | P2-006               | pack              | release                  | Standard Pack                                          |
| pack-gate              | P2-007               | pack              | release                  | Gate Pack                                              |
| pack-evidence          | P2-008               | pack              | release                  | Evidence Pack                                          |
| downstream-baseline    | P2-014               | downstream        | local/patch-only         | 下游 adoption 前扫描                                   |
| downstream-adoption    | P2-009/P2-010/P2-012 | downstream        | local/patch-only/pr-plan | 下游 adoption patch/pr plan                            |
| runtime-file-ownership | P2-013               | runtime-upgrade   | local/downstream         | 文件所有权与覆盖安全                                   |

---

## 12. Makefile 目标设计

```makefile
XLIB_CONTEXT ?= local_write

.PHONY: fmt
fmt:
	GOWORK=off gofmt -w $$(find . -name '*.go' -not -path './.git/*')

.PHONY: vet
vet:
	GOWORK=off go vet ./...

.PHONY: lint
lint:
	GOWORK=off golangci-lint run ./...

.PHONY: test
test:
	GOWORK=off go test ./...

.PHONY: race
race:
	GOWORK=off go test -race ./...

.PHONY: main-guard
main-guard:
	GOWORK=off go run ./cmd/goalcli main-guard --context $(XLIB_CONTEXT)

.PHONY: worktree-guard
worktree-guard:
	GOWORK=off go run ./cmd/goalcli worktree-guard --context $(XLIB_CONTEXT)

.PHONY: evidence-check
evidence-check:
	GOWORK=off go run ./cmd/goalcli evidence-check

.PHONY: boundary
boundary:
	GOWORK=off go run ./cmd/goalcli boundary

.PHONY: security
security:
	GOWORK=off go run ./cmd/goalcli security

.PHONY: contracts
contracts:
	GOWORK=off go run ./cmd/goalcli contracts

.PHONY: docs-check
docs-check:
	GOWORK=off go run ./cmd/goalcli docs-check

.PHONY: governance-check
governance-check: main-guard worktree-guard evidence-check boundary security

.PHONY: p1-governance-check
p1-governance-check:
	GOWORK=off go run ./cmd/goalcli agent-team-contract
	GOWORK=off go run ./cmd/goalcli scope-lock
	GOWORK=off go run ./cmd/goalcli pr-template
	GOWORK=off go run ./cmd/goalcli acceptance-matrix
	GOWORK=off go run ./cmd/goalcli runtime-health
	GOWORK=off go run ./cmd/goalcli conformance-profile
	GOWORK=off go run ./cmd/goalcli downstream-registry
	GOWORK=off go run ./cmd/goalcli self-healing-skeleton

.PHONY: p2-runtime-check
p2-runtime-check:
	GOWORK=off go run ./cmd/goalcli install-runtime --dry-run
	GOWORK=off go run ./cmd/goalcli upgrade-runtime --dry-run
	GOWORK=off go run ./cmd/goalcli release-ready
	GOWORK=off go run ./cmd/goalcli evidence-replay
	GOWORK=off go run ./cmd/goalcli attest-conformance --profile standard-source
	GOWORK=off go run ./cmd/goalcli pack-standard
	GOWORK=off go run ./cmd/goalcli pack-gate
	GOWORK=off go run ./cmd/goalcli pack-evidence

.PHONY: release-check
release-check: governance-check test contracts docs-check

.PHONY: release-final-check
release-final-check:
	XLIB_CONTEXT=release_verify GOWORK=off make governance-check
	GOWORK=off make p1-governance-check
	GOWORK=off make p2-runtime-check
	GOWORK=off make release-check
```

---

## 13. CI 设计

CI 必须调用 Makefile，不得重复实现 Gate 逻辑。

```yaml
name: ci
permissions:
  contents: read
  pull-requests: read

on:
  pull_request:
  push:
    branches: [main]

jobs:
  governance:
    runs-on: ubuntu-latest
    env:
      XLIB_CONTEXT: ci_pull_request
    steps:
      - uses: actions/checkout@<pinned-version-or-sha>
      - uses: actions/setup-go@<pinned-version-or-sha>
      - run: GOWORK=off make governance-check

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<pinned-version-or-sha>
      - uses: actions/setup-go@<pinned-version-or-sha>
      - run: GOWORK=off make test

  contracts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<pinned-version-or-sha>
      - uses: actions/setup-go@<pinned-version-or-sha>
      - run: GOWORK=off make contracts

  release-check:
    runs-on: ubuntu-latest
    env:
      XLIB_CONTEXT: release_verify
    steps:
      - uses: actions/checkout@<pinned-version-or-sha>
      - uses: actions/setup-go@<pinned-version-or-sha>
      - run: GOWORK=off make release-check
```

原则：

```text
1. P0/P1 required checks 禁止 continue-on-error。
2. Security / Evidence / Governance 必须独立可见。
3. CI 不读取生产 secret。
4. GitHub settings apply 不隐式执行；verify 可以执行。
5. Artifact 上传只允许脱敏 Evidence。
```

---

## 14. Evidence Runtime

### 14.1 Evidence Bundle 最小 schema

```yaml
evidence:
  command:
  result:
  commit:
  branch:
  worktree:
  context:
  timestamp:
  artifacts:
  known_gaps:
  checksum:
```

### 14.2 Artifact path convention

```text
release/evidence/<date>/<issue-id>/<artifact>.json
release/manifest/<version>/manifest.json
release/manifest/<version>/manifest.json.sha256
release/manifest/latest.json              # generated, ignored unless release policy explicitly says otherwise
release/manifest/latest.json.sha256       # generated, ignored unless release policy explicitly says otherwise
```

### 14.3 Retention

```yaml
retention:
  release_evidence: permanent
  security_incident: permanent
  audit_log: 3y
  ci_artifact: 90d
  failed_rc: 30d
```

### 14.4 DONE with evidence

```text
DONE with evidence:
- scope:
- issues:
- worktree:
- branch:
- changed_files:
- gates:
- evidence:
- review:
- release_impact:
- known_gaps:
- follow_up:
```

---

## 15. Release Manifest Skeleton

```yaml
module:
version:
commit:
tree_sha:
source_digest:
tool_versions:
workspace_status:
gate_results:
evidence_artifacts:
known_gaps:
generated_at:
checksum:
```

Release Ready 公式：

```text
release_ready =
  governance_check_passed
  AND p1_governance_check_passed
  AND runtime_health_passed
  AND required_gates_passed
  AND evidence_complete
  AND manifest_valid
  AND workspace_clean
  AND no_p0_blocker
  AND no_open_security_incident
  AND review_complete
```

---

## 16. Runtime File Ownership

```yaml
file_classes:
  XLIB_OWNED:
    description: xlib-standard may generate and update
  USER_OWNED:
    description: never overwrite
  MERGE_REQUIRED:
    description: generate patch / three-way merge plan
  GENERATED:
    description: can be rebuilt; follow generated artifact policy
```

Runtime install/upgrade 必须：

```text
1. 先 classify target files。
2. 默认 dry-run。
3. USER_OWNED 不覆盖。
4. MERGE_REQUIRED 只生成 patch plan。
5. --apply 需要显式传入。
6. destructive write 需要 backup 和 rollback note。
```

---

## 17. Downstream Adoption Modes

```text
local_path:   用户本地已有 clone，goalcli 对路径执行 dry-run/apply。
patch_only:   只生成 patch bundle，不写下游仓库。
pr_plan:      生成 PR plan，不直接创建 PR，除非权限明确。
```

下游 adoption 前必须运行 baseline scan：

```text
module path
package layout
Makefile
CI
contracts
docs
x.go reverse dependency
secret default usage
existing release flow
adoption risk score
```

首批顺序：

```text
1. xlib-standard self-conformance -> standard-source
2. kernel -> l0-kernel
3. configx -> l1-shared
```

---

## 18. GitHub Governance

必须提供但不能隐式 apply：

```text
.github/CODEOWNERS
.github/ISSUE_TEMPLATE/goal.yml
.github/ISSUE_TEMPLATE/bugfix.yml
.github/pull_request_template.md
.agent/github-settings.yaml
docs/standard/branch-protection.md
scripts/github/verify_settings.sh
```

规则：

```text
1. CODEOWNERS 覆盖 CONSTITUTION.md、.agent/**、.github/**、Makefile、cmd/goalcli/**、contracts/**、release/**。
2. Branch protection / required checks 必须 verify。
3. Admin bypass 必须 break-glass + human approval。
4. Apply scripts 只可手动显式执行。
```

---

## 19. Supply Chain Security

P1 最小基线：

```text
govulncheck
golangci-lint
secret scan / gitleaks equivalent
actions pinning policy
permissions least privilege
go.mod / go.sum drift check
.tool-versions
.agent/toolchain.yaml
```

P2/P3 roadmap：

```text
SBOM
License check
Dependency risk report
Third-party policy admission
```

---

## 20. 命名策略

默认名称：

```text
standard repo: xlib-standard
L0 downstream: kernel
old template name: baselib-template only allowed in migration docs context
old downstream example: foundationx only allowed in migration docs context
```

禁止：

```text
1. README 主叙事使用 baselib-template。
2. Generator 默认输出 foundationx。
3. Release manifest 使用旧名作为当前事实。
4. 下游 adoption manifest 使用旧名。
```

---

## 21. Risk Register

| Risk ID  | Risk                                       | Level | Mitigation                               |
| -------- | ------------------------------------------ | ----: | ---------------------------------------- |
| RISK-001 | 宪法过重导致难以采用                       |    P0 | P0/P1/P2 分阶段，P3/P4 冻结              |
| RISK-002 | Agent 在 main 上开发                       |    P0 | context-aware main-guard                 |
| RISK-003 | worktree-guard 误伤 CI/release             |    P0 | execution-context policy                 |
| RISK-004 | 无 Evidence 声称完成                       |    P0 | evidence-check + DONE parser             |
| RISK-005 | 反向依赖 x.go                              |    P0 | boundary no-xgo-import check             |
| RISK-006 | 生产 secret 路径污染模板                   |    P0 | no-secret-default check                  |
| RISK-007 | Gate 无 fixture                            |    P1 | fixture-first + governance test strategy |
| RISK-008 | .agent YAML 字段错误但静默通过             |    P1 | policy schema validation                 |
| RISK-009 | GitHub settings 未真正生效                 |    P1 | verify settings protocol                 |
| RISK-010 | Runtime install 覆盖下游手写内容           |    P1 | runtime file ownership + dry-run         |
| RISK-011 | 下游 adoption 未扫描 baseline              |    P1 | downstream baseline scan                 |
| RISK-012 | 旧名污染生成库                             |    P1 | naming guard                             |
| RISK-013 | xlib-standard 自身未证明符合就要求下游符合 |    P1 | self-conformance attestation             |

---

## 22. Traceability Matrix

| Requirement           | AC                                                 | Design                       | Task                 | Gate                                                  | Evidence               |
| --------------------- | -------------------------------------------------- | ---------------------------- | -------------------- | ----------------------------------------------------- | ---------------------- |
| 禁止 main 写入开发    | local_write on main fails                          | execution-context main-guard | P0-004/P0-013        | goalcli main-guard --context local_write             | fixture output         |
| CI/release 不误阻断   | ci/release contexts pass when clean                | execution-context            | P0-013               | goalcli worktree-guard --context ci_pull_request     | fixture output         |
| 强制 worktree         | local_write outside worktree fails                 | worktree-guard               | P0-005/P0-013        | goalcli worktree-guard                               | fixture output         |
| 无 Evidence 不得 DONE | DONE no evidence fails                             | evidence-check               | P0-006/P0-012        | goalcli evidence-check                               | parser output          |
| 禁止 x.go 反向依赖    | x.go import fails                                  | boundary                     | P0-007               | goalcli boundary                                     | boundary report        |
| 禁止 secret default   | default secret path fails                          | security                     | P0-008               | goalcli security                                     | security report        |
| 命令 SSOT             | command registry has all required commands         | command registry             | P0-015               | goalcli command-registry                             | registry report        |
| Makefile baseline     | required targets exist                             | makefile baseline            | P0-016               | goalcli makefile-baseline                            | makefile report        |
| PR 合规               | PR fields complete                                 | PR contract                  | P1-003               | goalcli pr-template                                  | PR report              |
| Runtime 自检          | runtime-health passes                              | runtime health               | P1-005               | goalcli runtime-health                               | health report          |
| Policy schema         | invalid yaml fails                                 | schema validation            | P1-017               | goalcli policy-schema                                | schema report          |
| GitHub settings       | required settings verify                           | github governance            | P1-018               | goalcli github-settings --verify                     | verify report          |
| Evidence path         | artifact roots declared                            | evidence artifacts           | P1-020               | goalcli evidence-artifacts                           | artifact policy report |
| 命名一致              | stale default names blocked                        | naming policy                | P1-021               | goalcli naming                                       | naming report          |
| Release Ready         | formula returns ready only if all constraints pass | release readiness            | P2-003               | goalcli release-ready                                | readiness report       |
| 自符合证明            | standard-source attestation generated              | conformance                  | P2-015               | goalcli attest-conformance --profile standard-source | attestation            |
| 下游 baseline         | kernel/configx scanned before adoption             | downstream baseline          | P2-014               | goalcli downstream-baseline                          | scan report            |
| 下游 adoption         | patch-only adoption generated                      | adoption modes               | P2-009/P2-010/P2-012 | goalcli downstream-adoption --mode patch-only        | patch report           |

---

## 23. Definition of Done

### Task DoD

```text
1. Scope matches issue.
2. Worktree and branch recorded.
3. Required gates run.
4. Evidence bundle generated.
5. Known gaps declared.
```

### Issue DoD

```text
1. All AC passed.
2. All required gates passed.
3. Traceability updated.
4. Evidence linked.
5. Review done.
```

### P0 DoD

```text
1. P0-001..P0-016 done.
2. governance-check passes in local_write context.
3. release-check passes in release_verify context.
4. goalcli doctor/cli-contract/issue-registry/command-registry/makefile-baseline pass.
5. Release manifest skeleton generated.
```

### P1 DoD

```text
1. P1-001..P1-021 done.
2. p1-governance-check passes.
3. policy-schema, github-settings verify, toolchain, evidence-artifacts, naming pass.
4. Self-healing skeleton exists.
```

### P2 DoD

```text
1. P2-001..P2-015 done.
2. p2-runtime-check passes.
3. xlib-standard self-conformance attestation exists.
4. kernel/configx baseline scan reports exist.
5. kernel/configx patch-only adoption reports exist.
```

---

## 24. 1 天 / 7 天 / 30 天行动计划

### 1 天：冻结范围 + 生成 SSOT

```text
1. 创建 CONSTITUTION.md 最小版。
2. 创建 .agent/minimal-kernel.yaml。
3. 创建 .agent/enforcement-levels.yaml。
4. 创建 .agent/execution-context.yaml。
5. 创建 .agent/issue-registry.yaml。
6. 创建 .agent/command-registry.yaml。
7. 创建 .agent/makefile-baseline.yaml。
8. 明确 P0-001..P0-016 owner/gate/fixture/evidence。
```

验收：

```bash
GOWORK=off go run ./cmd/goalcli issue-registry
GOWORK=off go run ./cmd/goalcli command-registry
```

### 7 天：完成 P0 Minimal Kernel

```text
1. goalcli CLI skeleton。
2. execution-context-aware main/worktree guard。
3. evidence-check。
4. boundary no-xgo-import。
5. no-secret-default。
6. Makefile governance-check/release-check。
7. CI required skeleton。
8. release manifest skeleton。
9. DONE with evidence protocol。
10. CLI contract/report schema。
```

验收：

```bash
XLIB_CONTEXT=local_write GOWORK=off make governance-check
XLIB_CONTEXT=release_verify GOWORK=off make release-check
GOWORK=off go test ./...
GOWORK=off go run ./cmd/goalcli cli-contract
```

### 30 天：完成 P1 + P2 到 v2.9.3

```text
1. P1 Governance Hardening 全部完成。
2. P2 Runtime & Conformance Automation 全部完成。
3. xlib-standard self-conformance attestation 生成。
4. kernel/configx baseline scan 生成。
5. kernel/configx patch-only adoption reports 生成。
```

验收：

```bash
GOWORK=off make governance-check
GOWORK=off make p1-governance-check
GOWORK=off make p2-runtime-check
GOWORK=off make release-check
GOWORK=off go test ./...
```

---

## 25. Agent Team 可执行 Prompt

```text
You are executing GOAL-20260602-001 for github.com/ZoneCNH/xlib-standard.

Hard constraints:
- Do not develop on main.
- Use git worktree for any local write task.
- Respect .agent/execution-context.yaml.
- Do not claim DONE without evidence.
- Do not import github.com/bytechainx/x.go or github.com/ZoneCNH/x.go.
- Do not default-read /home/k8s/secrets/env/*.
- Do not implement P3/P4 features before P0/P1/P2 are done.

Execution order:
1. P0-001..P0-016
2. P1-001..P1-021
3. P2-001..P2-015

For every issue:
- Read issue-registry and command-registry.
- Create or use assigned worktree.
- Modify only declared paths.
- Add valid and invalid fixtures for any guard.
- Run required gates.
- Produce evidence bundle under declared artifact roots.
- Update traceability matrix.
- Finish only with DONE with evidence.
```

---

## 26. 最终推荐路径

```text
P0 Minimal Kernel
  = 防止灾难，保证底线。

P1 Governance Hardening
  = 支撑多人 / Agent Teams / PR / Review / GitHub / Toolchain / Evidence 管理。

P2 Runtime & Conformance Automation
  = 让标准能安装、升级、打包、证明符合，并验证 kernel/configx。

P3/P4
  = 暂时冻结，等 P2 DONE with evidence 后再启动。
```

最终裁决：

```text
xlib-standard v2.9.3 Complete 的核心目标，
是把前面所有宪法、方法论、Harness、Agent Teams、Self-improving、AutoResearch、Compound Engineering
收敛成一个可执行工程内核：
P0 先防灾难，
P1 再硬化协作治理，
P2 再完成运行时安装/升级/符合性证明；
任何高级治理功能必须等 P2 DONE with evidence 后再解锁。
```
