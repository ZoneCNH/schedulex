# 复盘

## 改进项

- `xlib-standard` 的身份统一为 Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime。
- 默认下游从旧示例名迁移为 `kernel`，并通过下游矩阵约束 `configx`、`observex`、`testkitx` 和 profile 库。
- `.agent` 运行时从单一目标说明升级为 Full Goal Runtime v3.1 的对象、状态、traceability、gate、Evidence、release、review、rollback 和 patch 集合。

## 失败项记录规则

- 任一 required/final gate 失败不得声明完成。
- 缺少 `golangci-lint`、`govulncheck` 或 `goalcli` 必须作为 blocker 记录。
- `/home/k8s/secrets/env/*` 内容进入源码、日志、manifest、PR 或 Evidence 时必须回滚并补充规则补丁。

## 提示补丁

- 后续创建基础库必须从 `xlib-standard` 生成，旧名仅可作为迁移历史引用。
- 所有基础库必须保留 Boundary Gate、Secret Gate、Evidence Gate 和 Retrospective Patch 入口。

## Harness 补丁

- 保持 `goalcli score --min 9.8` 为最终门禁。
- 保持 kernel downstream smoke 为默认下游集成门禁。

## 规则补丁

- 禁止基础库依赖 `x.go`。
- 禁止基础库承载业务语义。
- 禁止无 Evidence 声称 `DONE`。
- 禁止读取或泄露调用方生产密钥路径 `/home/k8s/secrets/env/*`。
