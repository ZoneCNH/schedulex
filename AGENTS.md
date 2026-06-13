# 仓库贡献指南

## 项目概述

本仓库是 `github.com/ZoneCNH/schedulex`，Go 1.23 的 L1 deterministic scheduler。生产代码只依赖 Go 标准库，提供触发器、Clock 注入、misfire/overlap 策略、确定性 jitter、事件输出、快照和 `Locker` 扩展接口。

## 项目结构

- `pkg/schedulex/`: 公共调度器 API 与实现，版本常量位于 `scheduler.go`。
- `examples/`: 可运行示例与 smoke 测试，示例不得引入业务域词或外部基础设施依赖。
- `contracts/`: 公共 API snapshot、release manifest schema 和契约测试。
- `scripts/`: 本地 gate、release preflight、manifest/evidence 生成脚本。
- `release/downstream-adoption/`: 下游采纳 fixture/evidence；`latest.json` 可提交，`release/manifest/latest.json` 是生成产物不提交。
- `docs/`: API、spec、release、standard、goal 与测试策略文档。

## 常用命令

```bash
GOWORK=off make fmt
GOWORK=off make vet
GOWORK=off make test
GOWORK=off make race
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off make docs-check
GOWORK=off make release-preflight VERSION=v1.0.0
GOWORK=off make release-check VERSION=v1.0.0
```

## 发布规则

- 发布版本统一从 `Makefile` 的 `VERSION`、`pkg/schedulex.Version`、contracts、release docs 和 downstream fixture 读取；v1.0.0 不允许出现活跃 v0.1.x 锚点。
- `scripts/generate_schedulex_manifest.sh` 生成 `release/manifest/latest.json` 和 `latest.json.sha256`；这两个文件是本地发布产物，不提交。
- `release/downstream-adoption/fixture/go.mod` 不允许 `replace`。tag 发布前只跑非网络 smoke；tag 推送后再跑 `SCHEDULEX_DOWNSTREAM_NETWORK=1 VERSION=v1.0.0 ./scripts/check_downstream_smoke.sh`。
- 示例、文档和状态文件必须保持通用调度器语义，不写入订单、交易、账户等业务域名。

## 代码约束

- `pkg/schedulex` 不依赖标准库以外的生产依赖。
- 时间相关测试必须通过可注入 Clock 或有界超时保持确定性。
- 新公共 API 必须更新 `contracts/public_api.snapshot`、API 文档和契约测试。
- 发布前保持 `git diff --check`、fmt、vet、unit、race、boundary、contracts、docs-check 通过。
