# schedulex L1 调度器标准

`github.com/ZoneCNH/schedulex` 的当前交付面是 `pkg/schedulex`。本模块是 L1 deterministic scheduler，只提供标准库依赖的调度内核，不承载业务模型、真实外部锁后端、`x.go` wiring 或 L2 runtime 依赖。

## 模块身份

- Go module：`github.com/ZoneCNH/schedulex`
- 包路径：`github.com/ZoneCNH/schedulex/pkg/schedulex`
- Layer：`L1`
- 版本：`v1.0.0`
- 角色：deterministic scheduler

## 公共 API

调度器构造和生命周期 API 固定为：

- `NewScheduler(opts ...Option) (*Scheduler, error)`
- `AddJob(job Job, trigger Trigger, opts ...JobOption) error`
- `Start(ctx context.Context) error`
- `Shutdown(ctx context.Context) error`
- `Snapshot() Snapshot`

任务 API 固定为 `Job` 接口和 `JobFunc` 适配器：

- `Name() string`
- `Run(ctx context.Context) error`
- `JobFunc{NameValue, RunFunc}`

触发器 API 固定为 `Trigger.Next(after time.Time) (time.Time, bool)`，构造器包括 `Once`、`Every`、`Cron` 和 `DailyAt`。`Cron` 为五字段 L1 轻量表达式，只支持确定性可验证的 minute/hour 组合，其他字段必须为通配符。

## 时间与确定性

所有调度决策必须通过 `Clock` 获取时间。测试可以注入 `StaticClock`，生产默认使用真实标准库时钟。trigger、deterministic jitter、misfire reconcile、timezone 和 DST 行为必须由 golden/contract 测试覆盖。

## 策略契约

`MisfirePolicy` 的合法值为：

- `skip`
- `run_once`
- `catch_up`

`OverlapPolicy` 的合法值为：

- `skip`
- `queue_one`
- `allow`

策略必须在 `Snapshot` 中可见，并通过合同测试防止漂移。

## 锁与事件

`Locker` 只是接口扩展点：

- `TryLock(ctx context.Context, key string, ttl time.Duration) (Lease, error)`
- `Lease.Release(ctx context.Context) error`

本仓库不得实现真实 Redis/Postgres/etcd 等运行时锁后端。调用方负责注入锁实现，`schedulex` 只验证接口调用、失败事件和释放语义。

`EventSink` 负责接收调度生命周期事件，事件覆盖 scheduled、started、succeeded、failed、skipped、misfire、lock_skipped、lock_failed 和 shutdown。

## Gate 与 Evidence

发布式验证必须使用 `GOWORK=off`：

```bash
GOWORK=off make release-final-check
GOWORK=off make identity-check fmt vet lint test race boundary contracts docs-check security evidence release-final-check
GOWORK=off make trigger-determinism-check misfire-contract-check timezone-dst-golden-check scheduler-leak-check scheduler-race-check lock-interface-check api-check downstream-smoke
GOWORK=off make release-preflight VERSION=v1.0.0
```

Evidence 必须写入 `release/manifest/latest.json`，并由 `release/manifest/latest.json.sha256` 校验。最终完成声明必须包含 `DONE with evidence:`。
