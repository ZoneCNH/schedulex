# Harness Gates

Harness Gate 把 `xlib-standard` 的标准、模板、generator、Evidence 和 release 要求变成可执行检查。

Full Goal Runtime v3.1 以 `cmd/goalcli` 作为 Go gate runtime。Makefile target 是推荐的人机入口，内部必须委托到 `GOWORK=off go run ./cmd/goalcli ...`；`scripts/*.sh` 是兼容实现层，不再作为 CI/发布文档中的唯一权威入口。

## Required Gates

| Gate | 命令 | 目的 |
| --- | --- | --- |
| Format | `GOWORK=off make fmt` | 保持 Go 格式稳定 |
| Vet | `GOWORK=off make vet` | 基础静态检查 |
| Lint | `GOWORK=off make lint` | `golangci-lint` 强制检查，缺失时失败 |
| Unit | `GOWORK=off make test` | 单元和示例 smoke |
| Race | `GOWORK=off make race` | 并发安全基线 |
| Boundary | `GOWORK=off make boundary` | 模块边界和模板禁止项 |
| Security | `GOWORK=off make security` | `govulncheck` 和 secret scan |
| Contracts | `GOWORK=off make contracts` | schema、metrics 和 manifest contract |
| Docs Check | `GOWORK=off make docs-check` | 文档、链接、当前命名、下游同步策略、v3.1 runtime 和 release protocol |
| Integration | `GOWORK=off make integration` | generator 和 downstream smoke |
| Dependency Check | `GOWORK=off make dependency-check` | 校验 `renovate.json`、`.github/dependabot.yml` 和 Go dependency inventory |
| Standard Impact Check | `GOWORK=off make standard-impact-check` | 生成 `release/standard-impact/latest.md` 并判定 `downstream_sync_required`、`downstream_release_decision`（`required` / `not_required`）和 `repository_rules_release_decision`（`audit_required` / `not_required`） |
| Score | `GOWORK=off make score` / `GOWORK=off go run ./cmd/goalcli score --min 9.8` | 校验 v3.1 gate runtime、CI 和文档契约一致性 |
| Evidence | `CHECK_STATUS=passed GOWORK=off make evidence` | 生成 release manifest |
| Release Evidence | `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check` | 校验 manifest 与仓库事实 |


## Goal v2.9.3 Governance Gate

| Gate | 命令 | 目的 |
| --- | --- | --- |
| P0 Governance | `XLIB_CONTEXT=local_write GOWORK=off make governance-check` | 执行 main/worktree/evidence/boundary/security/CLI/registry/Makefile baseline；禁止 x.go imports 与真实 secrets。 |
| P1 Governance Dry Run | `GOWORK=off make p1-governance-check` | 验证 policy schema、GitHub settings intent、toolchain、Evidence artifacts、naming、install/upgrade runtime 与 release-ready 文档，不读取外部 secrets。 |
| P2 Runtime Dry Run | `GOWORK=off make p2-runtime-check` | 验证 standard-source/l0-kernel conformance、pack-standard/pack-gate/pack-evidence、downstream patch-only、runtime-file-ownership 和 execution-context。 |

这些 target 是 `docs/goal.md` v2.9.3 可执行方案的验收入口；CI 通过 `GOWORK=off XLIB_CONTEXT=ci_pull_request make release-check` 覆盖 `governance-check`、`p1-governance-check` 与 `p2-runtime-check`，避免 workflow 层重复执行同一治理链。

`issue-registry` 是语义 gate：它校验 issue ID 唯一且按前缀连续、状态为 `implemented`，并要求每个条目具备命令和 Evidence。`agent-team-contract`、`acceptance-matrix`、`runtime-health` 和 `execution-context` 的 dry-run 必须读取对应文件并检查 `schema_version` 等语义 marker，不能退化为文件存在检查。

`.agent/harness.yaml` 中的 `_check`、`_chain` 和 `_release_scope` 是不同 evidence node：`*_check` 对应 Makefile target，`*_chain` 对应 harness chain 集成，`*_release_scope` 对应 release 场景覆盖。三者命名相近但不得互相替代，也不得把任一节点的通过结果升级成另一个节点的 evidence。

## Context Runtime v4.0 Profile Baseline（REQ-014 当前可执行态）

本节是 `GOAL-20260602-XLIB-RUNTIME-CONSOLIDATION-V4` / `REQ-014` 的冻结守则。当前可执行事实由 `Makefile`、`cmd/goalcli` 和四个 SSOT registry（`.agent/command-registry.yaml`、`.agent/issue-registry.yaml`、`.agent/makefile-baseline.yaml`、`.agent/makefile-target-registry.yaml`）共同证明；profile wrapper 与 registry bridge 已落地为 release gate 的一部分。物理 `.agent/context/*` packs/templates 仍不得被描述为已交付，除非对应文件实际进入仓库并被 registry/evidence 覆盖。

| Profile | 当前组合 | 冻结守则 |
| --- | --- | --- |
| `context-lite` | `governance-check` | 轻量上下文入口；`context-profile-check` 必须证明 Makefile、CLI、command registry 和 Makefile registry 均包含该 wrapper。 |
| `context-standard` | `governance-check + p1-governance-check + docs-check` | `docs-check` 是显式组成项；它只能证明静态文本和链接守则，不能替代语义审查。 |
| `context-full` | `governance-check + p1-governance-check + p2-runtime-check` | 不能用 docs-only 或 score-only 结论替代 P2 runtime dry run。 |
| `context-release` | `context-full + integration + dependency-check + standard-impact-check + score-check + evidence + release-evidence-*` | `context-release` 不得包含 `release-check` 或 `release-final-check`；`release-final-check` 必须单向调用 `context-release`，不得反向递归。 |

兼容别名 `context-fast-check`、`context-standard-check`、`context-full-check` 必须保留并与 profile wrapper 指向同一 SSOT registry 语义。任何新增 profile、alias 或 context registry bridge 都必须同步更新四个 registry；registry 仍是单一事实源，不能由临时脚本、文档表格或 `.agent/context/*` 片段取代。`context-profile-check` 同时校验 unknown profile、未知 Makefile gate、重复 target、forbidden release edge、profile DAG cycle 和 `release-final-check` 自递归，intake 明确禁止的 context ID 不得进入源码、registry 或文档；若 review 发现该 ID，应视为命名污染而不是新增上下文任务。

## Extended Gate

- `GOWORK=off make property`
- `GOWORK=off make golden`
- `GOWORK=off make fuzz-smoke`
- `GOWORK=off make ci-extended`
- `GOWORK=off make release-check-extended`

## Generator Gate

Generator gate 必须证明模板能生成代表性下游，而不是只证明 `xlib-standard` 自身可用。

代表下游：

- `kernel`
- `corekit`

旧 `foundationx` 只作为迁移兼容扫描项，不再作为默认下游。

## Final Gates

- `XLIB_CONTEXT=release_verify GOWORK=off make release-final-check`
- `XLIB_CONTEXT=release_verify GOWORK=off make release-preflight VERSION=<version>`
- `GOWORK=off go run ./cmd/goalcli score --min 9.8`
- `GOWORK=off make integration DOWNSTREAM=kernel`

## Secret Gate

Secret Gate 必须确认源码、README、测试日志、release manifest、PR 描述和 Evidence 不包含 `/home/k8s/secrets/env/*` 的真实内容。该路径只能在文档中作为调用方部署路径名出现。

Secret scan 会排除 `.git`、`.omc`、`.omx` 和 `vendor` 等本地或第三方目录，避免把 Agent runtime 或 vendored 依赖误判为源码凭据；这些目录一旦内容进入 git 历史、manifest、PR、Issue 或日志，仍按 secret leak 处理。

## Workflow Supply Chain Gate

CI、Release Check、Integration 和 Security workflow 引用的第三方 Action 必须固定为 40 位 commit SHA，并保留来源 tag 注释。`govulncheck` 安装必须使用固定版本；当前发布门禁基线是 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`。

## Context Runtime v4 profile gates 发布门禁

Context Runtime v4.0 profile gate 在现有 governance harness 上追加执行：

- `context-lite` 校验 local guard、registry、CLI contract 和 profile contract coverage。
- `context-standard` 追加 P1 governance 和 documentation checks。
- `context-full` 追加 P2 runtime dry-run coverage。
- `context-release` 追加 standard impact、score、manifest generation、release evidence 和 checksum verification，但不得调用 `release-check` 或 `release-final-check`。

为保持下游兼容，必须保留 legacy aliases（`context-fast-check`、`context-standard-check`、`context-full-check`）。

## Debt gates

Debt governance gates are part of the release harness: `architecture`, `domain`, `security-debt`, `debt`, and `debt-evidence`. The manifest must include a passing `debt` block with score >= 9.8 and zero P0 findings.
