# 模块边界

本文件约束 `xlib-standard`、生成基础库、`kernel` 和 `x.go` 的模块边界。边界违规必须在 PR、review 或 release gate 中阻断。旧名 `baselib-template` 只保留为迁移兼容名，不再是主仓库角色。

## `xlib-standard` 允许内容

- P0 标准文档与版本说明。
- 仓库角色、分层、模块边界、下游兼容性规则和 module path 规则。
- Go 参考模板 `pkg/templatex`、contracts schema、metrics contract 和测试夹具。
- generator、模板元数据、渲染脚本和生成库矩阵。
- Makefile、scripts、CI、docs-check、boundary、contracts、integration、release 和 score gate。
- `.agent/` Full Goal Runtime v3.1 工件、review/release/retrospective 模板、Evidence 生成器和 release manifest 协议实现。

## `xlib-standard` 禁止内容

- 真实 `kernel`、PostgreSQL、Redis、Kafka、OSS、ClickHouse、TDengine runtime。
- `x.go` import 或以 `x.go` 为构建前提。
- 业务领域模型、业务 repository、业务消息 schema 或应用生命周期编排。
- 生产密钥、真实连接串、默认生产 endpoint，或 `/home/k8s/secrets/env/*` 内容。
- 隐藏全局 client、隐式后台进程或不可关闭资源。
- 把旧 `baselib-template` / `foundationx` 叙事作为主身份；旧名只能出现在迁移文档语境。

## 生成基础库允许内容

- 替换后的 `go.mod` module path、package name、README 和 docs。
- 该库 profile 所需的公共 API、配置、错误、健康检查、metrics 和 tests。
- 必要的 profile-specific contracts、examples、testkit 和 release Evidence。
- 对更低层基础库的显式依赖，例如 profile 库依赖 `kernel` 的稳定 primitive。

## 生成基础库禁止内容

- 反向依赖业务层或 `x.go`。
- 绕过 `Config` 隐式加载生产凭据。
- 删除 Required gate、docs-check、boundary、contracts 或 release Evidence gate。
- 把 `release/manifest/latest.json` 或 `release/manifest/latest.json.sha256` 提交进源码历史。

## `kernel` 允许内容

- L0 通用 runtime primitive：context、error、config、logging、metrics、lifecycle、health 和 test helper。
- 被 `configx`、`observex`、`testkitx` 及 profile 库显式复用的稳定 API。

## `kernel` 禁止内容

- 存储、消息、对象存储、分析数据库等 profile runtime。
- 业务语义、应用框架耦合、`x.go` 反向依赖。
- 隐式读取 `/home/k8s/secrets/env/*` 或任何生产密钥。

## `x.go` 集成边界

`x.go` 是调用方组合层，可以读取调用方授权的 `/home/k8s/secrets/env/*` 并把配置显式传给基础库。基础库、生成模板和 `kernel` 不得读取该路径、导入 `x.go` 或假设 `x.go` 存在。

## 边界验证

每次涉及包结构、依赖或模板渲染时运行：

```bash
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off make integration
```
