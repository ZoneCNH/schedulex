# 示例索引

> schedulex 使用示例，展示核心功能的典型用法。

## 运行方式

```bash
# 运行单个示例
go run ./examples/interval

# 运行所有示例（仅编译检查）
go build ./examples/...
```

---

## 示例列表

### interval — 定时任务

展示如何使用 `Every()` 触发器创建固定间隔的定时任务。

```bash
go run ./examples/interval
```

**要点：**
- `NewScheduler()` 创建默认调度器
- `JobFunc` 将函数适配为 `Job` 接口
- `Every(time.Minute)` 创建每分钟触发一次的触发器
- `Snapshot()` 获取当前调度器状态

---

### once — 一次性任务

展示如何使用 `Once()` 触发器创建仅执行一次的任务。

```bash
go run ./examples/once
```

**要点：**
- `Once(time)` 在指定时刻触发一次后不再触发
- 适用于延迟执行、定时提醒等场景

---

### cron_like — Cron 表达式

展示如何使用 `Cron()` 触发器解析五字段 Cron 表达式。

```bash
go run ./examples/cron_like
```

**要点：**
- 支持 `*`、`*/N`、固定整数三种分钟/小时形式
- 日/月/周字段必须为 `*`
- `loc` 参数控制时区（`nil` 默认 UTC）

---

### daily_at — 每日定时

展示如何使用 `DailyAt()` 触发器创建每天固定时刻执行的任务。

```bash
go run ./examples/daily_at
```

**要点：**
- `DailyAt(hour, minute, loc)` 指定每日执行时刻
- 支持任意时区（如 `Asia/Shanghai`）
- 如果当前时刻已过，自动调度到次日

---

### misfire — 失火策略

展示如何使用 `PlanMisfire()` 规划错过触发时间时的调和决策。

```bash
go run ./examples/misfire
```

**要点：**
- `MisfireSkip`：跳过所有错过时刻
- `MisfireRunOnce`：仅执行最后一个错过时刻
- `MisfireCatchUp`：依次执行所有错过时刻

---

### shutdown — 优雅关闭

展示如何使用 `Start()` 和 `Shutdown()` 管理调度器生命周期。

```bash
go run ./examples/shutdown
```

**要点：**
- `Shutdown()` 等待运行中的任务完成
- 支持 `context.WithTimeout` 控制关闭超时
- 多次调用 `Shutdown()` 是安全的（幂等）

---

### lock_interface — 分布式锁

展示如何实现 `Locker` 和 `Lease` 接口为任务添加分布式锁。

```bash
go run ./examples/lock_interface
```

**要点：**
- `Locker.TryLock()` 尝试获取锁，返回 `Lease` 或 `ErrLockUnavailable`
- `Lease.Release()` 释放锁
- 示例使用内存实现，生产环境需适配 Redis/Postgres 等
- 配合 `WithLocker()`、`WithLockKey()`、`WithLockTTL()` 使用
