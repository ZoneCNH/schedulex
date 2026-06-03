# 仓库贡献指南

## 项目概述

本仓库是 Go 1.23 基础库标准与交付运行时，模块路径为 `github.com/ZoneCNH/xlib-standard`，承担五类职责：**Standard Source**、**Go Reference Template**、**Generator**、**Harness** 和 **Evidence Runtime**。旧名 `baselib-template` / `foundationx` 仅允许出现在迁移文档中。

## 项目结构与模块组织

- `pkg/templatex`：公共 API 参考实现（config、client、health、errors、metrics、options、version），渲染后移到 `pkg/<package-name>`。
- `internal/`：内部辅助——`sanitize`（配置脱敏）、`validation`（校验）、`releasequality`（分数计算）、`tools/releasemanifest`（manifest 生成器）、`goalcli`（gate 辅助）。
- `cmd/goalcli/`：统一 CLI 门禁工具，所有 Makefile gate 目标最终调用此命令。子命令分 Go 原生实现和 shell 脚本委托两类。
- `testkit/`：可复用测试夹具、断言、golden 文件工具，以及 `governance/` 治理测试夹具。
- `examples/basic`、`examples/config`、`examples/health`：最小示例，各有 smoke 测试验证输出不漂移。
- `contracts/`：JSON schema 定义和 `contracts_test.go` 验证映射。
- `scripts/`：Harness gate shell 脚本。
- `.agent/`：Full Goal Runtime v3.1 工件（goal-runtime、state-machine、traceability-matrix、harness.yaml、evidence-protocol 等）。
- `docs/`：标准、设计、API、配置、测试、发布、迁移、downstream-matrix 文档。
- `release/manifest/`：manifest 模板；`latest.json` 和 `.sha256` 是生成产物，不提交到源码历史。

## 构建、测试与开发命令

### 基础开发

- `make fmt`：执行 `go fmt ./...`。
- `make vet`：执行 `go vet ./...`。
- `make test`：运行全部单元测试（覆盖 `pkg/`、`internal/`、`contracts/`、`testkit/`、`examples/`）。
- `make race`：使用 race detector 运行测试。
- `make lint`：执行 `golangci-lint run ./...`；缺少 `golangci-lint` 时必须失败。
- `make security`：执行 `govulncheck ./...` 和 `scripts/check_secrets.sh`；缺少 `govulncheck` 时必须失败。

### 运行单个测试

```bash
go test ./pkg/templatex/ -run TestConfigValidate
go test ./internal/sanitize/ -run TestSanitize
go test ./... -run 'Test.*Property|Test.*Invariant'   # 属性测试
go test ./... -run 'Test.*Golden|Test.*Snapshot'       # golden 测试
```

### CI 与 Gate

- `make ci`：fmt + vet + lint + test + race + boundary + security + contracts + governance-check + score。
- `make ci-extended`：ci + property + golden + fuzz-smoke。
- `make boundary`：模块边界检查（x.go import ban、业务层 ban）。
- `make contracts`：JSON schema 契约检查。
- `make docs-check`：文档结构 gate（链接、占位符、命名、release manifest 协议）。
- `make dependency-check`：依赖漂移检查（读取 renovate.json、dependabot.yml、go.mod）。
- `make standard-impact-check`：生成标准影响报告。
- `make score`：`goalcli score --min 9.8`，发布门禁阈值。
- `make governance-check`：P0 治理全量检查。
- `make p1-governance-check`：P1 本地 dry-run 治理契约。
- `make p2-runtime-check`：P2 运行时/downstream/execution-context dry-run 契约。

### 发布（必须 GOWORK=off）

所有发布和验证命令必须使用 `GOWORK=off`，避免本地 `go.work` 改写 module 解析：

```bash
GOWORK=off make release-check
XLIB_CONTEXT=release_verify GOWORK=off make release-final-check
XLIB_CONTEXT=release_verify GOWORK=off make release-preflight VERSION=v0.4.6
make evidence                                    # 生成 release/manifest/latest.json
```

### 模板渲染（生成下游库）

```bash
scripts/render_template.sh \
  --module-name kernel \
  --module-path github.com/ZoneCNH/kernel \
  --package-name kernel \
  --out ../kernel
```

目标库：`kernel`、`configx`、`observex`、`testkitx`、`postgresx`、`redisx`、`kafkax`、`taosx`、`ossx`、`clickhousex`。

## 架构要点

### goalcli CLI

`cmd/goalcli` 是所有治理 gate 的统一入口（Makefile 中 `GOALCLI ?= go run ./cmd/goalcli`）。命令类别：

- **质量 gate**：score、boundary、contracts、security/secrets
- **文档 gate**：docs-check、dependency-check、standard-impact-check
- **治理 gate**：main-guard、worktree-guard、evidence-check、cli-contract、issue-registry、command-registry、makefile-baseline
- **P1 治理**：agent-team-contract、scope-lock、pr-template、acceptance-matrix、runtime-health 等（`--dry-run --verify`）
- **P2 运行时**：install-runtime、upgrade-runtime、release-ready、evidence-replay 等
- **Evidence**：evidence/manifest、release-evidence-hash、release-evidence-check

### 权威顺序（CONSTITUTION.md）

1. `docs/goal.md` 与 `docs/standard/` 描述目标、边界和标准条款。
2. `.agent/*.yaml` 与 `cmd/goalcli` 描述机器可执行的门禁契约。
3. `release/manifest/` 与 `release/evidence/` 保存发布证据。

### Goal Runtime v3.1

`.agent/` 目录包含完整的 Goal Runtime 工件：goal-runtime、object-model、state-machine、traceability-matrix、harness.yaml、evidence-protocol、release-template、retrospective、risk-register、decision-log、rollback-protocol。harness.yaml 定义了所有 required gate 及其命令。

## 编码风格与命名约定

使用标准 Go 风格：交给 `gofmt` 处理缩进，包名保持简短，导出标识符要清晰表达用途。模板占位符 `{{MODULE_NAME}}`、`{{MODULE_PATH}}`、`{{PACKAGE_NAME}}` 必须在代码和文档中保持一致。公共库能力放入 `pkg/templatex`，私有辅助逻辑放入 `internal/`。golangci-lint 启用的 linter 见 `.golangci.yml`。

## 测试规范

测试使用 Go 标准 `testing` 包，命名遵循 `TestXxx`；场景较多时优先使用表驱动测试。必须覆盖配置校验、配置脱敏、客户端创建、幂等关闭、健康和非健康状态检查、示例 smoke 输出、`testkit` 夹具和 contracts 映射。

- 小改动：至少 `go test ./...`
- 并发相关：`make race`
- 发布流程：`make integration`
- 属性/不变量：`make property`
- Golden/快照：`make golden`
- Fuzz smoke：`make fuzz-smoke`（默认 10s/fuzz target，深度 fuzz 设置 `FUZZ_SMOKE_TIME=2m`）

## 提交与 Pull Request 规范

提交信息必须遵循 Lore protocol：第一行说明变更意图，正文使用 `Constraint:`、`Rejected:`、`Confidence:`、`Scope-risk:`、`Directive:`、`Tested:` 和 `Not-tested:` 等 trailer 记录决策和验证。PR 需要说明对模板或库行为的影响，关联相关 issue，列出已运行命令，并说明生成的 Evidence artifact。

## 关键约束

- **GOWORK=off**：所有发布和验证命令必须使用，避免 `go.work` 改写 module 解析。
- **x.go 独立**：基础库不得依赖 `x.go` 或其内部包；`x.go` 只能作为调用方组合层。
- **无真实凭据**：不得提交 `.env` 或 `/home/k8s/secrets/env/*` 内容，`scripts/check_secrets.sh` 会扫描。
- **Evidence 完成**：最终完成声明必须包含 `DONE with evidence:`。
- **Release score**：`goalcli score --min 9.8` 是发布门禁阈值。
- **latest.json**：是生成产物，不提交到源码历史（已在 .gitignore）。
- **Release manifest 测试**：必须在临时 fixture 仓库中构造所需 `.omc` state，不得依赖当前工作区的 Agent 运行态文件。

## 文档语言规则

所有仓库文档必须默认使用中文叙述，包括 `README.md`、`docs/`、`.agent/`、`contracts/*.md`、变更日志、发布说明、PR 描述模板和贡献指南。专业术语、代码标识符、命令、路径、包名、外部专有名词、协议固定短语和提交标题必须保留项目惯用原文，例如 Agent、Harness、manifest、schema、CI、PR、Issue、Go module。新增或更新文档时，先检查是否存在整段英文说明；除非用户明确要求英文，否则应改写为中文，但不要翻译专业术语。

## Agent 专用说明

仓库协作、代码评审和进度更新默认使用中文，除非用户明确要求其他语言。代码、命令、路径、包名和提交标题保留项目惯用语言。治理相关任务优先参考 `.agent/` 目录中的 Goal Runtime v3.1 工件和 `CONSTITUTION.md` 中的权威顺序。
