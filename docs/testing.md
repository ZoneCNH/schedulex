# 测试策略

`xlib-standard` 的测试策略同时约束标准仓库、Go 参考模板、generator、Harness、Evidence Runtime 和所有生成基础库。

## 必需覆盖范围

- `go test ./...` 必须覆盖公共包、`internal/`、`contracts/`、`testkit/` 和 `examples/`。
- 配置校验、脱敏、typed error kind、wrapped cause。
- 客户端创建、取消 context、过期 context、幂等关闭和 zero-value client。
- 健康检查 JSON 字段 contract。
- 生命周期 metrics 和健康 metrics。
- `contracts/` 与公共常量同步。
- `contracts/config.schema.json` 与 `Config` 字段映射同步。
- `scripts/render_template.sh` 生成的临时 `kernel` 可以通过 `GOWORK=off go test ./...`。
- `Config.Sanitize` 的 secret 不变量必须由 property test 覆盖。
- `Config` 边界输入必须由 fuzz-smoke 覆盖。
- `HealthStatus` JSON 公共输出必须由 golden test 锁定。
- `internal/releasequality` 必须覆盖 `Compute`、`Verify` 和 `Marshal` 的正常与失败路径。
- Release manifest 测试必须在临时 fixture 仓库构造所需 `.omc` state，不得依赖当前工作区的 Agent 运行态文件。

## 示例与 testkit Smoke

- `examples/basic` 必须输出当前 module name。
- `examples/config` 必须输出脱敏后的 secret 值。
- `examples/health` 必须输出 `healthy`。
- `testkit` 必须验证 `Config("fixture")` 生成可通过 `Validate` 的测试配置。

生成的基础库必须保持测试独立于 `x.go`，且不得读取 `/home/k8s/secrets/env/*`。
