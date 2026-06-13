# schedulex 宪章

本仓库是 `github.com/ZoneCNH/schedulex` 的可审计发布源。目标是提供一个标准库依赖、确定性、可测试的 Go 调度器基座。

## 权威顺序

1. `pkg/schedulex/` 与 `contracts/` 定义公共 API 和兼容性边界。
2. `Makefile` 与 `scripts/` 定义本地 gate 和发布流程。
3. `docs/`、`README.md`、`STATUS.md` 定义用户可见状态和发布说明。
4. `release/` 定义 manifest、下游采纳 fixture 与发布证据。

## 发布不变量

- 生产代码只依赖 Go 标准库。
- `ModuleName` 固定为 `github.com/ZoneCNH/schedulex`。
- v1.0.0 发布面必须统一到 `schedulex.Version == "v1.0.0"`。
- 发布 manifest 必须由脚本生成并校验 checksum；不要手写生成产物。
- downstream fixture 不允许本地 `replace`。
- 活跃示例和文档不得绑定业务域词；历史归档可以保留原始上下文。

## 变更纪律

- 不在 `main` 上直接开发；使用 release/feature branch 后再合入。
- 每个提交聚焦一个发布或修复目的，并使用 Lore commit trailers。
- 公共 API 变化必须同时更新 snapshot、文档、契约测试和发布说明。
- 发布前必须运行最小证明链：`git diff --check`、fmt、vet、unit、race、boundary、contracts、docs-check、release-preflight。

## 验收

v1.0.0 可发布的最低证据：

- `GOWORK=off go test ./...`
- `GOWORK=off go vet ./...`
- `GOWORK=off make race`
- `GOWORK=off make boundary`
- `GOWORK=off make contracts`
- `GOWORK=off make docs-check`
- `GOWORK=off make release-preflight VERSION=v1.0.0`
