# Metrics Contract

标准指标用于描述 `pkg/templatex` 暴露给调用方的最小可观测面。实现可以接入任意 metrics 后端，但指标名、类型和标签语义必须保持兼容。

| 指标 | 类型 | 标签 | 说明 |
| --- | --- | --- | --- |
| `client_created_total` | counter | `name` | 成功创建 client 的次数。 |
| `client_closed_total` | counter | `name` | 成功关闭 client 的次数；重复关闭不重复计数。 |
| `client_errors_total` | counter | `op`, `kind` | client 生命周期错误次数，`kind` 必须来自 error contract。 |
| `client_health_status` | gauge | `name`, `status` | 健康状态数值，healthy 为 `1`，其他状态为 `0`。 |
| `client_health_latency_ms` | histogram | `name`, `status` | 单次健康检查耗时，单位为毫秒。 |
| `client_requests_total` | counter | `operation`, `status` | 调用方扩展请求计数。 |
| `client_request_duration_seconds` | histogram | `operation`, `status` | 调用方扩展请求耗时，单位为秒。 |
| `client_retries_total` | counter | `operation`, `kind` | 调用方扩展重试计数。 |
| `client_inflight` | gauge | `operation` | 调用方扩展并发中的请求数。 |
