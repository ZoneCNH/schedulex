# Job Event Schema

## 概述

schedulex 的 `EventSink` 接口接收 `Event` 结构体，用于调度生命周期的可观测性。本文档定义事件字段和类型规范。

## Event 结构

```json
{
  "type": "started",
  "job_id": "retryable-maintenance",
  "job_name": "retryable-maintenance",
  "at": "2026-06-09T10:00:00.123Z",
  "scheduled_at": "2026-06-09T10:00:00.000Z",
  "started_at": "2026-06-09T10:00:00.123Z",
  "finished_at": "0001-01-01T00:00:00Z",
  "lag": 123000000,
  "duration": 0,
  "attempt": 1,
  "reason": "",
  "err": "",
  "attributes": {}
}
```

## 字段定义

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `type` | string | ✅ | 事件类型，见下方枚举 |
| `job_id` | string | ✅ | job 注册时的唯一标识（默认为 `job.Name()`） |
| `job_name` | string | ✅ | job 的人类可读名称 |
| `at` | RFC3339 | ✅ | 事件发生时间（clock 基准） |
| `scheduled_at` | RFC3339 | ⬚ | 计划触发时间 |
| `started_at` | RFC3339 | ⬚ | 实际开始执行时间 |
| `finished_at` | RFC3339 | ⬚ | 实际完成时间 |
| `lag` | int64 (ns) | ⬚ | `started_at - scheduled_at` 的延迟 |
| `duration` | int64 (ns) | ⬚ | `finished_at - started_at` 的执行耗时 |
| `attempt` | int | ⬚ | 当前第几次尝试 |
| `reason` | string | ⬚ | 附加原因（如 misfire 策略名、overlap 原因） |
| `err` | string | ⬚ | 错误信息（仅失败事件） |
| `attributes` | map | ⬚ | 附加键值对（如 misfire 统计） |

## 事件类型枚举

| 事件类型 | 触发时机 | 典型字段 |
|----------|----------|----------|
| `scheduled` | loop 计算出下次触发时间并注册 timer | `scheduled_at`, `attempt` |
| `started` | job 开始执行 | `started_at`, `lag`, `attempt` |
| `succeeded` | job 执行成功 | `started_at`, `finished_at`, `duration`, `attempt` |
| `failed` | job 执行返回 error 或 panic | `started_at`, `finished_at`, `duration`, `attempt`, `err` |
| `skipped` | overlap 策略跳过执行 | `reason`（如 "overlap"） |
| `misfire` | 检测到错过了计划触发时间 | `reason`（misfire 策略名）, `attributes`（missed/runs/skipped 计数） |
| `shutdown` | scheduler 关闭 | 无附加字段 |
| `lock_skipped` | 分布式锁不可用 (`ErrLockUnavailable`) | `err` |
| `lock_failed` | 分布式锁获取失败 (其他错误) | `err` |

## Misfire attributes

当 `type` 为 `misfire` 时，`attributes` 包含：

| Key | 类型 | 描述 |
|-----|------|------|
| `missed` | string (int) | 错过的触发次数 |
| `runs` | string (int) | 决策执行的次数 |
| `skipped` | string (int) | 决策跳过的次数 |
| `capped` | string ("true") | 仅在达到上限 128 时存在 |
| `first_missed` | string (RFC3339) | 第一个错过的时间点 |
| `last_missed` | string (RFC3339) | 最后一个错过的时间点 |

## 事件生命周期图

```
                ┌──────────┐
                │scheduled │
                └────┬─────┘
                     │
                     ▼
               ┌──────────┐
          ┌────│ started  │────┐
          │    └──────────┘    │
          ▼                    ▼
   ┌──────────┐         ┌──────────┐
   │succeeded │         │  failed  │
   └──────────┘         └──────────┘

   ┌──────────┐         ┌──────────┐
   │ skipped  │         │misfire   │  (overlap/misfire)
   └──────────┘         └──────────┘

   ┌──────────────┐     ┌──────────────┐
   │ lock_skipped │     │ lock_failed  │  (locker)
   └──────────────┘     └──────────────┘

   ┌──────────┐
   │shutdown  │  (scheduler level)
   └──────────┘
```

## 实现 EventSink

```go
type EventSink interface {
    OnEvent(ctx context.Context, event Event)
}
```

### 适配函数

```go
// EventSinkFunc 将函数适配为 EventSink
schedulex.EventSinkFunc(func(ctx context.Context, event schedulex.Event) {
    log.Printf("[%s] %s: %s", event.Type, event.JobID, event.Reason)
})
```

### 注册位置

- **Scheduler 级别**：`NewScheduler(WithEventSink(sink))` — 接收所有事件
- **Job 级别**：`AddJob(job, trigger, WithJobEventSink(sink))` — 仅接收该 job 的事件

两个 sink 同时接收同一个事件的副本。调度器级 sink 先于 job 级 sink 触发。
