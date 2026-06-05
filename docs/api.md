# API 参考

> `github.com/ZoneCNH/schedulex/pkg/schedulex`

schedulex 是一个确定性 L1 调度器，所有调度决策均由可注入的 `Clock` 和 `Trigger` 驱动，测试和发布证据可在无墙钟睡眠的情况下回放行为。

---

## 核心接口

### Clock

```go
type Clock interface {
    Now() time.Time
    After(time.Duration) <-chan time.Time
}
```

为调度决策提供时间源。生产环境使用 `NewRealClock()`，测试使用 `StaticClock` 实现确定性回放。

### Job

```go
type Job interface {
    Name() string
    Run(context.Context) error
}
```

可调度的任务单元。`Name()` 返回用于快照和锁键的稳定标识；`Run()` 执行实际业务逻辑。Job 实现应尊重 `context.Context` 的取消信号。

### Trigger

```go
type Trigger interface {
    Next(after time.Time) (time.Time, bool)
}
```

计算给定参考时间之后的下一个调度时刻。返回 `(time.Time, false)` 表示无后续触发。

### EventSink

```go
type EventSink interface {
    OnEvent(context.Context, Event)
}
```

接收调度器生命周期和任务事件的通知适配器。

### Locker

```go
type Locker interface {
    TryLock(ctx context.Context, key string, ttl time.Duration) (Lease, error)
}
```

分布式锁扩展点。v0.1.0 未内置 Redis/Postgres 实现，需自行适配。返回 `ErrLockUnavailable` 表示锁不可用。

### Lease

```go
type Lease interface {
    Release(ctx context.Context) error
}
```

已获取的锁租约，必须由适配器释放。

---

## 核心类型

### Scheduler

```go
type Scheduler struct { /* 私有字段 */ }
```

核心调度器，管理任务注册、启动、运行和关闭的完整生命周期。

**方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `AddJob` | `(job Job, trigger Trigger, opts ...JobOption) error` | 注册任务和触发器。任务名为空或重复注册返回错误。 |
| `Start` | `(ctx context.Context) error` | 启动所有已注册任务的调度循环，幂等操作。 |
| `Shutdown` | `(ctx context.Context) error` | 优雅关闭调度器，等待运行中的任务完成或 context 超时。 |
| `Snapshot` | `() Snapshot` | 获取当前调度器状态快照，用于观测和发布证据。 |

**示例：**

```go
s, err := schedulex.NewScheduler(
    schedulex.WithClock(schedulex.NewRealClock()),
    schedulex.WithMaxConcurrent(4),
)
if err != nil {
    log.Fatal(err)
}

job := schedulex.JobFunc{
    NameValue: "heartbeat",
    RunFunc: func(ctx context.Context) error {
        fmt.Println("tick")
        return nil
    },
}
if err := s.AddJob(job, schedulex.Every(time.Minute)); err != nil {
    log.Fatal(err)
}

if err := s.Start(context.Background()); err != nil {
    log.Fatal(err)
}
// ... 业务逻辑 ...
if err := s.Shutdown(context.Background()); err != nil {
    log.Fatal(err)
}
```

---

### Option

```go
type Option func(*Options) error
```

配置 `Scheduler` 的函数选项。

### JobOption

```go
type JobOption func(*jobConfig) error
```

配置单个注册任务的函数选项。

### TriggerOption

```go
type TriggerOption func(*triggerConfig)
```

预留的触发器级别配置，当前构造函数已接受但暂无实际配置项。

---

### Event

```go
type Event struct {
    Type        EventType         `json:"type"`
    JobID       string            `json:"job_id,omitempty"`
    JobName     string            `json:"job_name,omitempty"`
    At          time.Time         `json:"at"`
    ScheduledAt time.Time         `json:"scheduled_at,omitempty"`
    StartedAt   time.Time         `json:"started_at,omitempty"`
    FinishedAt  time.Time         `json:"finished_at,omitempty"`
    Lag         time.Duration     `json:"lag,omitempty"`
    Duration    time.Duration     `json:"duration,omitempty"`
    Attempt     int               `json:"attempt,omitempty"`
    Reason      string            `json:"reason,omitempty"`
    Err         string            `json:"err,omitempty"`
    Attributes  map[string]string `json:"attributes,omitempty"`
}
```

调度器事件载体，由 `EventSink` 接收。

### EventType

```go
type EventType string
```

事件类型常量：

| 常量 | 值 | 说明 |
|------|------|------|
| `EventScheduled` | `"scheduled"` | 任务已调度 |
| `EventStarted` | `"started"` | 任务开始执行 |
| `EventSucceeded` | `"succeeded"` | 任务执行成功 |
| `EventFailed` | `"failed"` | 任务执行失败 |
| `EventSkipped` | `"skipped"` | 任务因重叠策略被跳过 |
| `EventShutdown` | `"shutdown"` | 调度器关闭 |
| `EventMisfire` | `"misfire"` | 检测到失火（错过触发时间） |
| `EventLockSkipped` | `"lock_skipped"` | 锁不可用，跳过执行 |
| `EventLockFailed` | `"lock_failed"` | 锁获取失败 |

---

### Snapshot

```go
type Snapshot struct {
    Version  string
    Now      time.Time
    Started  bool
    Running  bool
    Closed   bool
    Shutdown bool
    JobCount int
    Jobs     []JobSnapshot
}
```

调度器状态快照，用于观测和发布证据。`Jobs` 按 ID 排序。

### JobSnapshot

```go
type JobSnapshot struct {
    ID            string
    Name          string
    Next          time.Time
    HasNext       bool
    MisfirePolicy MisfirePolicy
    OverlapPolicy OverlapPolicy
    Running       bool
    Queued        bool
}
```

单个任务的确定性调度状态快照。

---

### JitterPolicy

```go
type JitterPolicy struct {
    Max  time.Duration
    Seed int64
}
```

配置确定性逐任务抖动。`Max` 为最大抖动时长，`Seed` 为确定性种子。

### MisfirePolicy

```go
type MisfirePolicy string
```

描述错过触发时间时的调和策略。见 [Misfire 常量](#misfire-常量)。

### OverlapPolicy

```go
type OverlapPolicy string
```

描述任务仍在运行时再次触发的行为。见 [Overlap 常量](#overlap-常量)。

### MisfireDecision

```go
type MisfireDecision struct {
    Runs    []time.Time
    Skipped []time.Time
    Next    time.Time
    HasNext bool
}
```

失火调和的决策结果。`Runs` 为需要执行的时刻，`Skipped` 为跳过的时刻。

---

### StaticClock

```go
type StaticClock struct { /* 私有字段 */ }
```

确定性时钟，用于快照和合约测试。通过 `Set()` 和 `Advance()` 控制时间推进。

**方法：**

| 方法 | 签名 | 说明 |
|------|------|------|
| `Now` | `() time.Time` | 返回当前确定性时刻 |
| `After` | `(time.Duration) <-chan time.Time` | 返回在目标时刻接收值的 channel |
| `Set` | `(time.Time)` | 将时钟设置到指定时刻，唤醒到期的等待者 |
| `Advance` | `(time.Duration)` | 将时钟向前推进指定时长 |

**示例：**

```go
clock := schedulex.NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
s, _ := schedulex.NewScheduler(schedulex.WithClock(clock))
// ... 注册任务 ...
_ = s.Start(context.Background())
clock.Advance(time.Hour) // 推进 1 小时，触发到期任务
```

---

### JobFunc

```go
type JobFunc struct {
    NameValue string
    RunFunc   func(context.Context) error
}
```

将函数适配为 `Job` 接口的便捷结构体。

**示例：**

```go
job := schedulex.JobFunc{
    NameValue: "my-job",
    RunFunc: func(ctx context.Context) error {
        fmt.Println("running")
        return nil
    },
}
```

### EventSinkFunc

```go
type EventSinkFunc func(context.Context, Event)
```

将函数适配为 `EventSink` 接口的类型。当函数为 `nil` 时，`OnEvent` 调用为空操作。

**示例：**

```go
sink := schedulex.EventSinkFunc(func(ctx context.Context, e schedulex.Event) {
    log.Printf("[%s] %s: %s", e.Type, e.JobName, e.Reason)
})
s, _ := schedulex.NewScheduler(schedulex.WithEventSink(sink))
```

---

## 工厂函数

### NewScheduler

```go
func NewScheduler(opts ...Option) (*Scheduler, error)
```

构造调度器。默认使用 `NewRealClock()` 和 `MaxConcurrent=1`。无效选项返回 `ErrInvalidOption`。

### NewRealClock

```go
func NewRealClock() Clock
```

返回基于 Go 标准库 `time` 包的实时时钟。

### NewStaticClock

```go
func NewStaticClock(now time.Time) *StaticClock
```

创建固定在指定时刻的确定性时钟。

---

## 触发器构造函数

### Once

```go
func Once(at time.Time) Trigger
```

一次性触发器，在指定时刻触发一次后不再触发。

**示例：**

```go
trigger := schedulex.Once(time.Now().Add(time.Hour))
```

### Every

```go
func Every(d time.Duration, _ ...TriggerOption) Trigger
```

定时间隔触发器，每隔固定时长触发一次。

**示例：**

```go
trigger := schedulex.Every(30 * time.Second)
```

### DailyAt

```go
func DailyAt(hour, minute int, loc *time.Location, _ ...TriggerOption) Trigger
```

每日定时触发器，在每天指定的小时和分钟触发。`loc` 为 `nil` 时默认 UTC。

**示例：**

```go
loc, _ := time.LoadLocation("Asia/Shanghai")
trigger := schedulex.DailyAt(9, 30, loc) // 每天 09:30 CST
```

### Cron

```go
func Cron(expr string, loc *time.Location, _ ...TriggerOption) (Trigger, error)
```

Cron 表达式触发器，支持五字段格式（分 时 日 月 周）。仅支持分钟/小时字段的 `*`、`*/N` 和固定整数形式，日/月/周字段必须为 `*`。`loc` 为 `nil` 时默认 UTC。

**示例：**

```go
trigger, err := schedulex.Cron("*/15 * * * *", time.UTC) // 每 15 分钟
if err != nil {
    log.Fatal(err)
}
```

---

## 抖动与 Misfire 函数

### ApplyDeterministicJitter

```go
func ApplyDeterministicJitter(base time.Time, p JitterPolicy, jobID string, run int64) time.Time
```

对基础时刻应用确定性抖动。使用 FNV-64a 哈希基于 `jobID`、`Seed` 和 `run` 序号生成确定性偏移。`Max <= 0` 时返回原始时刻。

### PlanMisfire

```go
func PlanMisfire(policy MisfirePolicy, missed []time.Time, next time.Time, hasNext bool) MisfireDecision
```

根据失火策略规划错过时刻的调和决策。

| 策略 | 行为 |
|------|------|
| `MisfireSkip` | 跳过所有错过时刻 |
| `MisfireRunOnce` | 仅执行最后一个错过时刻 |
| `MisfireCatchUp` | 依次执行所有错过时刻（上限 128） |

### ReconcileMisfire

```go
func ReconcileMisfire(policy MisfirePolicy, missed []time.Time) []time.Time
```

为发布 golden case 计算需要执行的时刻列表。`MisfireSkip` 返回 `nil`。

---

## 错误

| 错误 | 说明 |
|------|------|
| `ErrSchedulerClosed` | 对已关闭的调度器执行变更操作 |
| `ErrJobExists` | 注册重复的任务 ID |
| `ErrInvalidJob` | 任务定义不完整（名称或触发器缺失） |
| `ErrInvalidOption` | 选项违反调度器合约 |
| `ErrLockUnavailable` | 锁适配器无法获取租约 |

---

## 常量

### Misfire 常量

| 常量 | 值 | 说明 |
|------|------|------|
| `MisfireSkip` | `"skip"` | 跳过所有错过时刻 |
| `MisfireRunOnce` | `"run_once"` | 仅执行最后一个错过时刻 |
| `MisfireCatchUp` | `"catch_up"` | 补偿执行所有错过时刻 |

### Overlap 常量

| 常量 | 值 | 说明 |
|------|------|------|
| `OverlapSkip` | `"skip"` | 跳过重叠触发 |
| `OverlapQueueOne` | `"queue_one"` | 排队一次，前次完成后执行 |
| `OverlapAllow` | `"allow"` | 允许并发执行 |

### 版本常量

| 常量 | 值 | 说明 |
|------|------|------|
| `ModuleName` | `"github.com/ZoneCNH/schedulex"` | Go module 路径 |
| `Version` | `"v0.1.0"` | 当前版本号 |

---

## With* 选项函数

### Scheduler 选项

#### WithClock

```go
func WithClock(clock Clock) Option
```

注入确定性时钟。`clock` 为 `nil` 时返回 `ErrInvalidOption`。

#### WithEventSink

```go
func WithEventSink(sink EventSink) Option
```

设置调度器级别的事件接收器。

#### WithMaxConcurrent

```go
func WithMaxConcurrent(limit int) Option
```

设置调度器全局并发执行上限。`limit <= 0` 时返回 `ErrInvalidOption`。

### Job 选项

#### WithMisfirePolicy

```go
func WithMisfirePolicy(policy MisfirePolicy) JobOption
```

设置任务的失火调和策略。空值默认为 `MisfireSkip`。

#### WithOverlapPolicy

```go
func WithOverlapPolicy(policy OverlapPolicy) JobOption
```

设置任务的重叠执行策略。空值默认为 `OverlapSkip`。

#### WithJitter

```go
func WithJitter(policy JitterPolicy) JobOption
```

为任务设置确定性抖动。`Max < 0` 时返回 `ErrInvalidOption`。

#### WithLocker

```go
func WithLocker(locker Locker) JobOption
```

为任务设置分布式锁扩展点。

#### WithLockKey

```go
func WithLockKey(key string) JobOption
```

设置分布式锁的键名。默认使用任务名。

#### WithLockTTL

```go
func WithLockTTL(ttl time.Duration) JobOption
```

设置请求的锁 TTL。未设置时使用保守的默认值（1 分钟）。`ttl < 0` 时返回 `ErrInvalidOption`。

#### WithJobEventSink

```go
func WithJobEventSink(sink EventSink) JobOption
```

为单个任务覆盖调度器级别的事件接收器。
