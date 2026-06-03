# x.go 集成边界

`x.go` 是应用或框架组合层，不是基础库的依赖前提。`xlib-standard`、`kernel` 和所有生成基础库必须保持可独立构建、测试和发布。

## 允许

- `x.go` 作为调用方依赖基础库。
- `x.go` 读取调用方授权的 `/home/k8s/secrets/env/*`，解析后把显式 `Config` 传给基础库。
- `x.go` 组合多个基础库并管理应用生命周期。

## 禁止

- 基础库导入 `x.go`。
- 基础库默认读取 `/home/k8s/secrets/env/*` 或任何生产密钥路径。
- 将 `/home/k8s/secrets/env/*` 的内容写入源码、README、测试日志、release manifest、PR 描述或 Evidence 文本。
- 将业务模型、业务 repository、业务消息 schema 下沉到基础库。

## 验证

- boundary gate 必须扫描 `x.go` import 和业务层反向依赖。
- security/release Evidence 必须声明未泄露 `/home/k8s/secrets/env/*` 内容。
- 需要生产密钥的真实集成测试必须由调用方提供显式配置，基础库只接收脱敏后的配置对象。
