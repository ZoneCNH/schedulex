# schedulex

`github.com/ZoneCNH/schedulex` 是 `pkg/schedulex` 的 L1 deterministic scheduler v0.1.0。它只依赖 Go 标准库，提供可复现的任务调度、触发器、misfire 处理、确定性 jitter、事件输出和分布式锁扩展接口。

## 能力范围

- 触发器：`Once`、`Every`、五字段轻量 `Cron`、`DailyAt`。
- 时间抽象：所有调度决策通过 `Clock` 注入，便于 golden 测试、DST 回放和确定性验证。
- 运行时：`NewScheduler`、`AddJob`、`Start`、幂等 `Shutdown`、`Snapshot`、最大并发和重叠执行控制。
- 可靠性契约：`MisfirePolicy` 支持 `skip`、`run_once`、`catch_up`；`OverlapPolicy` 支持 `skip`、`queue_one`、`allow`。
- 扩展接口：`Locker` 只声明 `TryLock` / `Lease.Release` 合同，不内置 Redis、Postgres 或其他运行时锁后端；`EventSink` 用于生命周期事件。
- 边界：生产代码不依赖 `x.go`、L2 基础库、真实密钥路径或应用 wiring。

## 快速开始

```go
package main

import (
	"context"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s, err := schedulex.NewScheduler(schedulex.WithMaxConcurrent(1))
	if err != nil {
		panic(err)
	}

	err = s.AddJob(
		schedulex.JobFunc{
			NameValue: "heartbeat",
			RunFunc: func(context.Context) error {
				return nil
			},
		},
		schedulex.Every(time.Minute),
		schedulex.WithMisfirePolicy(schedulex.MisfireRunOnce),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		panic(err)
	}
	_ = s.Shutdown(context.Background())
}
```

## 验证

发布式验证必须显式关闭 `go.work`：

```sh
GOWORK=off make docs-check
GOWORK=off make release-final-check
GOWORK=off make identity-check fmt vet lint test race boundary contracts docs-check security evidence release-final-check
GOWORK=off make trigger-determinism-check misfire-contract-check timezone-dst-golden-check scheduler-leak-check scheduler-race-check lock-interface-check api-check downstream-smoke
GOWORK=off make release-preflight VERSION=v0.1.0
```

Evidence 输出位于 `release/downstream-adoption/latest.json`，校验和位于 `release/downstream-adoption/latest.json.sha256`。Release manifest 位于 `release/manifest/latest.json`，校验和位于 `release/manifest/latest.json.sha256`。最终完成声明必须包含 `DONE with evidence:`。
