# schedulex

`github.com/ZoneCNH/schedulex` 是 L1 deterministic scheduler v0.1.0：一个只依赖 Go 标准库的确定性任务调度内核。

## 能力范围

- 触发器：`Once`、`Every`、五字段轻量 `Cron`、`DailyAt`。
- 时间抽象：调度决策通过 `Clock` 注入，便于黄金测试和可复现回放。
- 运行时：`Scheduler`、幂等 `Shutdown`、事件流、最大并发、重入跳过策略。
- 可靠性契约：`MisfireSkip`、`MisfireRunOnce`、`MisfireCatchUp`、确定性 jitter、`Locker` 接口。
- 边界：生产代码不依赖 x.go、L2+ 库、Redis/Postgres；锁后端由调用方实现。

## 快速开始

```go
s := schedulex.NewScheduler(schedulex.Options{MaxConcurrent: 1})
_ = s.AddJob(schedulex.Job{
    ID: "heartbeat",
    Trigger: schedulex.Every(time.Minute),
    Run: func(ctx context.Context) error { return nil },
})
_ = s.Start()
_ = s.Shutdown(context.Background())
```

## 验证

```sh
GOWORK=off go test ./...
GOWORK=off make boundary
GOWORK=off make schedulex-checks
GOWORK=off make release-preflight VERSION=v0.1.0
```
