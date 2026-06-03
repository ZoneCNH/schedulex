# 基础库总标准

本文定义 `x.go` 基础库体系的最小生产标准。`xlib-standard` 是 [`https://github.com/ZoneCNH/xlib-standard`](https://github.com/ZoneCNH/xlib-standard) 对应的统一仓库，同时承担 Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime。旧 `baselib-template` 只作为迁移文档语境中的兼容名出现，不再表示独立主实现仓库。

## 仓库定位

- `xlib-standard`：标准权威源，也是 Go 基础库模板、generator、Harness 和 Evidence 实现仓库。
- `kernel`：默认 L0 下游集成目标，用于证明模板生成和基础能力边界。
- 生成库：由 `xlib-standard` 渲染得到的具体基础库，必须遵守本标准，并通过自身 Harness 和 Evidence 证明符合性。
- 旧 `baselib-template` / `foundationx`：仅允许在迁移文档语境中出现。

## 标准目标

- 提供独立 Go module，可在没有 `x.go` 的情况下构建、测试和发布。
- 提供稳定公共 API、显式配置、可验证错误模型、健康检查、metrics contract 和 release Evidence。
- 让新基础库从创建开始就具备 Harness gate、contracts、CI、文档、评审和复盘入口。
- 用 `kernel`、`configx`、`observex`、`testkitx`、`postgresx`、`redisx`、`kafkax`、`taosx`、`ossx` 和 `clickhousex` 的下游矩阵约束复用边界。

## 分层

- Standard：`xlib-standard`，同时是 Standard 规则的独立来源和 Go 基础库模板中的实现仓库。
- L0：`kernel` 等语言级、无业务依赖的公共能力。
- L1：面向具体中间件或基础设施的库，例如 `postgresx`、`redisx`、`kafkax`、`taosx`、`ossx`、`clickhousex`。
- L2：组合多个基础能力的技术组件。
- Business：业务服务，只消费基础库，不向基础库反向注入业务模型。

## 依赖方向

- 依赖方向只能从 Business 指向 L2/L1/L0/Standard，或从具体库继承 `xlib-standard`。
- `xlib-standard` 不得依赖 `x.go`、业务仓库、profile-specific runtime 或生成库真实 runtime。
- 生成库不得依赖 `x.go`、业务模型或调用方生产密钥路径。
- `/home/k8s/secrets/env/*` 是调用方部署路径；`xlib-standard`、`kernel` 和生成库不得读取该路径作为默认配置源，也不得把其内容写入源码、README、测试日志、release manifest、PR 描述或 Evidence。

## 公共 API

公共能力放在渲染后的 `pkg/<package-name>`。导出 API 必须满足：

- 使用显式 `Config`，不得隐式读取生产环境凭据。
- `New` 返回可关闭的 client 或服务对象，并明确 validation failure。
- `Close` 必须幂等，重复调用不应 panic。
- `HealthCheck` 必须接收 `context.Context`，输出稳定 health status。
- 错误必须可分类，`ErrorKind` 由 `contracts/errors.schema.json` 约束。

## Config / Error / Logging / Metrics

- 配置必须可脱敏，覆盖 token、secret、password、key 等敏感字段。
- 默认值必须安全、可测试、可解释。
- validation、closed、unhealthy 和 internal failure 必须可区分。
- 日志不得输出 secret、token、password、private key 或连接凭据。
- metrics 名称和 label 由 `contracts/metrics.json` 约束；label 不能包含高基数字段、用户凭据或业务私有标识。

## Lifecycle / Health / Contract

- 初始化、关闭、健康检查和错误路径必须可单元测试。
- 并发或共享状态必须通过 race gate 验证。
- 不得创建隐藏全局 client、后台 goroutine 或不可关闭资源。
- health JSON 字段由 `contracts/health.schema.json` 约束。
- contracts 是公共行为的一部分；错误、health、metrics 和 manifest schema 变更必须经过 contract gate。

## Required Gates

- `GOWORK=off make ci`
- `GOWORK=off make docs-check`
- `GOWORK=off make integration`
- `CHECK_STATUS=passed GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`

## Final Gates

- `GOWORK=off make release-final-check`
- `GOWORK=off make release-preflight VERSION=<version>`
- `goalcli score --min 9.8`

最终声明必须使用 `DONE with evidence:`，并列出 manifest、checksum、score、kernel downstream smoke 和所有未运行项。
