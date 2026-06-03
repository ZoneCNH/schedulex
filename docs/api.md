# API 模板

## 占位符

- `{{MODULE_NAME}}`：生成的仓库名称。
- `{{MODULE_PATH}}`：生成的 Go module 路径。
- `{{PACKAGE_NAME}}`：生成的包名。

## 公共 API

- `Config`：由用户显式提供的配置。
- `Validate`：拒绝无效配置，并返回 `ErrorKindValidation`。
- `Sanitize`：在日志或 Evidence 采集前屏蔽敏感值。
- `New`：基于显式配置创建客户端；拒绝 `nil`、canceled 和 expired context；成功时记录 `client_created_total`。
- `Close`：释放资源，并且必须幂等；成功首次关闭时记录 `client_closed_total`。
- `HealthCheck`：报告客户端健康状态，JSON 字段必须匹配 `contracts/health.schema.json`；当本次检查的 context deadline 预算短于 `Config.Timeout` 时返回 `degraded`。
- `Error`：稳定 error contract，支持 `errors.Is` / `errors.As` 和 `IsKind`。
- `NewError` / `WrapError`：创建或包装稳定错误，包装时必须保留 cause。
- `Metrics`：注入式指标钩子；指标名必须匹配 `contracts/metrics.md`。
- `Version`：发布版本。

生成的基础库不得依赖 `x.go`。

## 生成对齐

使用 `scripts/render_template.sh` 生成具体基础库时，公共包目录会从 `pkg/templatex` 移动到 `pkg/{{PACKAGE_NAME}}`，代码 imports、文档占位符和 module path 会同步替换。
