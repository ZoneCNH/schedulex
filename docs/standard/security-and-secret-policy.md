# 安全和密钥策略

基础库标准、模板和生成库必须默认安全。安全策略覆盖源码、文档、测试、CI、manifest、Issue、PR 和 Evidence。

## 密钥边界

禁止提交或粘贴：

- 真实 token、password、private key、API key、access key。
- 真实生产连接串。
- 可直接访问生产资源的 endpoint 与凭据组合。
- 从本地环境复制的 `.env` 或 kubeconfig。
- `/home/k8s/secrets/env/*` 中任何文件内容、键值、路径展开结果或脱敏前日志。

示例必须使用占位符，例如 `example-token`、`example-secret` 或 `localhost`。

## `/home/k8s/secrets/env/*` 规则

- 该路径只属于调用方组合层，例如 `x.go` 或业务服务。
- `xlib-standard`、`kernel` 和生成基础库不得读取该路径。
- 调用方可以读取该路径并把显式 `Config` 传给基础库。
- 源码、README、测试日志、release manifest、Evidence、Issue 和 PR 描述不得包含该路径下的真实内容。

## Secret Gate

`GOWORK=off make security` 必须执行：

- `govulncheck ./...`
- `scripts/check_secrets.sh`

缺少 `govulncheck` 必须失败，不能跳过。secret scan 发现疑似凭据时必须阻断。

Secret scan 会排除 `.git`、`.omc`、`.omx` 和 `vendor` 等本地或第三方目录，避免把 Agent 运行态、OMX 兼容状态或 vendored 依赖误判为源码凭据。该排除只用于降低误报，不代表这些目录可以提交真实凭据；任何进入 git 历史、manifest、Issue、PR 或日志的 secret 都必须视为违规。

## 日志和 Evidence

- 日志不得输出敏感字段原值。
- release manifest 不得记录 secret。
- PR 描述和 Issue 模板不得要求粘贴真实凭据。
- `DONE with evidence:` 可以说明 secret scan 通过，但不得复制敏感样本。

## 依赖安全

- 新依赖必须有明确用途。
- 依赖变更后运行 `GOWORK=off make security` 和 `GOWORK=off make boundary`。
- 发现漏洞时记录影响面、修复版本和验证命令。

## 例外

安全例外必须有 ADR 或 Issue 记录，且只能放宽非 secret、非生产、非凭据相关约束。生产凭据、真实密钥和 `/home/k8s/secrets/env/*` 内容没有例外路径。
