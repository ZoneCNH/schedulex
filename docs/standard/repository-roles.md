# 仓库角色

本文件区分标准源、模板/generator/Harness/Evidence 运行时、生成基础库、`kernel` 和业务组合层的职责。角色边界用于防止基础库向上依赖业务层，也用于判断文档、CI 和 release Evidence 是否完整。

| 仓库 | 角色 | 必须包含 | 禁止包含 |
| --- | --- | --- | --- |
| `xlib-standard` (`https://github.com/ZoneCNH/xlib-standard`) | 标准权威源 + Go 参考模板 + 模板、generator、Harness、Evidence 实现仓库 | P0 标准文档、仓库角色、分层、模块边界、contracts 规范源、模板包、generator、CI、release manifest、`.agent/` runtime、Evidence 生成器 | 真实 L1/L2 runtime、业务模型、`x.go` 依赖、生产凭据、把旧名作为主身份 |
| `kernel` | L0 通用基础能力库 | context、error、config、logging、metrics、lifecycle、health、test helper 等通用 runtime primitives | 存储、消息、业务语义、应用框架耦合、`x.go` 反向依赖 |
| `configx` | L1 配置基础库 | 显式配置加载/校验、脱敏、配置来源 contract | 生产密钥默认值、业务配置语义、隐式读取 `/home/k8s/secrets/env/*` |
| `observex` | L1 可观测基础库 | logging、metrics、tracing contract、health 输出规范 | 业务指标模型、应用告警策略 |
| `testkitx` | L1 测试夹具库 | shared assertions、fake runtime、contract test helper | 生产 runtime、真实外部依赖默认连接 |
| `postgresx` | L2 数据库适配器库 | PostgreSQL profile、连接配置、健康检查、错误分类、测试夹具 | 业务 repository、应用 transaction 编排 |
| `redisx` | L2 cache 适配器库 | Redis profile、连接、health、metrics、testkit | 业务缓存 key 语义、应用级策略 |
| `kafkax` | L2 messaging 适配器库 | producer/consumer profile、contract、health、metrics | 业务 topic 设计、业务消息 schema |
| `taosx` | L2 时序数据适配器库 | TDengine profile、连接、health、contract | 业务指标模型 |
| `ossx` | L2 object storage 适配器库 | bucket/profile、上传下载 contract、health | 业务文件生命周期策略 |
| `clickhousex` | L2 分析数据库适配器库 | ClickHouse profile、连接、query contract、health | 产品报表语义 |
| `x.go` | 应用或框架组合层 | 组合基础库、业务 wiring、应用生命周期、调用方授权的密钥读取 | 作为基础库模板的依赖前提 |
| `baselib-template` / `foundationx` | 旧名/迁移文档语境 | 迁移文档语境中的 ADR、迁移指南、历史变更记录和兼容性说明 | 新标准中的主仓库角色或默认下游名 |

## 判定规则

- `xlib-standard` 是标准权威源，并同时承载模板、generator、Harness、Evidence 实现；不得再把 `baselib-template` 作为主实现仓库。
- 模板仓库职责已合并进 `xlib-standard`，但仍必须保持模板实现与标准正文一致。
- `kernel` 是默认 L0 下游集成目标；L1/L2 库可以显式依赖 `kernel`，但不得反向依赖 `x.go`。
- L2/profile 库可以依赖更低层基础库，但不得向业务层取配置或模型。
- 业务层可以组合所有基础库，但不得把业务规则下沉到基础库。
- 旧 `baselib-template`/`foundationx` 名称只能在迁移文档语境中出现，不能出现在新 README 主叙事、生成默认值或 release completion 声明中。
