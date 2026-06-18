# Harness Gates

Harness Gate 把 `schedulex` v1.0.0 的调度器契约、release manifest、governance 文档和 CI 状态要求变成可执行检查。Makefile target 是人机和 CI 的共同入口；脚本位于 `scripts/`，只作为 target 的实现层。

## Required Gates

| Gate | 命令 | 目的 |
| --- | --- | --- |
| Format | `GOWORK=off make fmt` | 保持 Go 格式稳定 |
| Vet | `GOWORK=off make vet` | 基础静态检查 |
| Lint | `GOWORK=off make lint` | `golangci-lint` 强制检查，缺失时失败 |
| Unit | `GOWORK=off make test` | 单元、契约和示例 smoke 测试 |
| Race | `GOWORK=off make race` | 并发安全基线 |
| Build | `GOWORK=off make build` | 编译所有包 |
| Boundary | `GOWORK=off make boundary` | 模块身份、L1 边界和禁止项 |
| Contracts | `GOWORK=off make contracts` | schema、release manifest 和契约文件 |
| Docs Check | `GOWORK=off make docs-check` | 文档与标准引用一致性 |
| Security | `GOWORK=off make security` | `govulncheck` 和 secret scan |
| API Check | `GOWORK=off make api-check` | 公共 API 面与 v1.0 契约一致 |
| Downstream Smoke | `GOWORK=off make downstream-smoke` | 代表下游可编译、可运行 |
| Integration | `GOWORK=off make integration` | generator、模板渲染和下游 smoke |
| Schedulex Check | `GOWORK=off make schedulex-check` | 调度器确定性、misfire、DST、race、lock 接口 |
| Evidence | `GOWORK=off make evidence` | 生成 `release/manifest/latest.json` 与 checksum |
| Release Final | `GOWORK=off make release-final-check` | 校验 release manifest 和仓库事实 |
| Score | `GOWORK=off make score` | 要求发布评分不低于 9.8 |

## Governance Gates

| Gate | 命令 | 目的 |
| --- | --- | --- |
| P0 Governance | `GOWORK=off make governance-check` | 校验 harness 文件、release evidence、CI 文本、目标注册和 release manifest checksum |
| P1 Governance | `GOWORK=off make p1-governance-check` | 校验工具链、PR 模板、command/makefile registry 和 CI lint/security 安装 |
| P2 Runtime | `GOWORK=off make p2-runtime-check` | 校验 runtime health、安装升级文档、执行上下文和下游 fixture |

`GOWORK=off make release-check VERSION=v1.0.0` 是发布入口，内部覆盖 `ci-extended` 与 `release-preflight`。`ci-extended` 覆盖 `ci`、`schedulex-check`、`evidence`、`governance-check`、`p1-governance-check`、`p2-runtime-check`、`release-final-check` 和 `score`；`release-preflight` 绑定显式版本并再次校验 release manifest。

## CI Status Checks

发布前必须要求以下 GitHub status checks：

- `ci`
- `release-check`
- `security`
- `integration`
- `gates`
- `worktree-check`

CI、Goal Gates、Integration、Security 和 Release workflow 引用的第三方 Action 必须固定为 40 位 commit SHA，并保留来源 tag 注释。`golangci-lint` 和 `govulncheck` 必须安装固定版本；缺失时 gate 必须失败，不能降级为 skip。

## Release Evidence

`release/manifest/latest.json` 和 `release/manifest/latest.json.sha256` 是本地与 CI 的 release evidence 输入。manifest 必须记录 commit、tree SHA、命令、退出码、环境、gate、workflow artifact 信息和 checksum；GitHub Release 对象必须在 tag 推送后由 release workflow 创建或更新，并由 `gh release view` 校验为非 draft、非 prerelease。

本文件只描述当前 `schedulex` v1.0.0 可执行 gate。不存在的历史 target 不得写入发布标准；新增 gate 前必须先进入 Makefile、registry、CI 和 release manifest 检查链。
