# 错误模板

## 占位符

- `{{MODULE_NAME}}`
- `{{PACKAGE_NAME}}`

## 错误类型

| `ErrorKind` | 字符串 | 典型场景 | Retryable |
| --- | --- | --- | --- |
| `ErrorKindConfig` | `config` | 配置来源或配置装载失败。 | 否 |
| `ErrorKindValidation` | `validation` | 配置字段缺失、格式非法、调用参数非法。 | 否 |
| `ErrorKindConnection` | `connection` | 连接建立失败。 | 通常是 |
| `ErrorKindUnavailable` | `unavailable` | context canceled、依赖暂不可用。 | 视场景 |
| `ErrorKindTimeout` | `timeout` | context deadline exceeded 或外部超时。 | 是 |
| `ErrorKindAuth` | `auth` | 认证、授权失败。 | 否 |
| `ErrorKindConflict` | `conflict` | 幂等冲突、资源状态冲突。 | 否 |
| `ErrorKindRateLimit` | `rate_limit` | 限流或配额耗尽。 | 是 |
| `ErrorKindInternal` | `internal` | 未分类内部错误。 | 否 |

## 约束

- 公共错误必须使用 `Error`、`NewError` 或 `WrapError` 表达稳定 contract。
- 包装错误必须保留 cause，使调用方可以使用 `errors.Is` / `errors.As`。
- 调用方按 `IsKind(err, ErrorKind...)` 做分支判断，不依赖错误字符串。
- 错误可以安全纳入 Evidence，但不得包含原始凭据、生产连接串或业务私密数据。
- 生成的库不得使用 `x.go` 业务模型。
