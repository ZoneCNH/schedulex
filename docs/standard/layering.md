# 分层模型

`xlib-standard` 同时是 Standard 规则的独立来源和 Go 基础库模板中的实现仓库。旧 `baselib-template` 名称只用于迁移说明，不再表示独立主角色。

| 层级 | 示例 | 职责 | 禁止 |
| --- | --- | --- | --- |
| Standard/Runtime | `xlib-standard` | 标准文本、参考模板、generator、Harness、Evidence、Goal Runtime | 真实基础设施 runtime、业务语义、生产密钥 |
| L0 | `kernel` | 通用 runtime primitive：context、error、config、logging、metrics、lifecycle、health | profile runtime、业务模型、`x.go` 反向依赖 |
| L1 | `configx`、`observex`、`testkitx` | 横向基础能力库 | 应用 wiring、业务策略 |
| L2 | `postgresx`、`redisx`、`kafkax`、`taosx`、`ossx`、`clickhousex` | profile/基础设施适配器 | 业务 repository、业务消息 schema |
| App/Composition | `x.go` 和业务服务 | 组合基础库、注入配置、读取调用方授权密钥 | 作为基础库依赖前提、把业务规则下沉 |

## 依赖方向

- `xlib-standard` 产生标准、模板和 gate，不依赖生成库或 `x.go`。
- `kernel` 可依赖标准化 contracts，但不依赖 L1/L2 或 `x.go`。
- L1/L2 可以依赖更低层基础库，例如 `kernel`。
- `x.go` 可以依赖和组合所有基础库。
- 任一基础库不得读取 `/home/k8s/secrets/env/*`；该路径只属于调用方组合层。
