# 规格

## 需求

`xlib-standard` 同时承担五类职责：Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime。仓库必须提供可编译 Go 模板、标准文档、ADR、下游矩阵、边界规则、`.agent` 运行时协议、release Evidence 协议和复盘补丁入口。

## 非目标

- 不成为 `x.go` 或任何业务仓库的运行时依赖。
- 不包含 `x.go` 业务模型、profile-specific runtime 或真实下游实现。
- 不隐式加载生产密钥，尤其不得读取 `/home/k8s/secrets/env/*`。
- 不把旧 `baselib-template` / `foundationx` 作为主身份或默认下游；旧名仅用于迁移文档语境。
