# Lock 接口契约

## 概述

`Locker` 是 schedulex 的分布式锁扩展点。它允许在多实例部署中确保同一 job 同一时刻只在一个实例上执行。

## 接口定义

```go
type Locker interface {
    TryLock(ctx context.Context, key string, ttl time.Duration) (Lease, error)
}

type Lease interface {
    Release(ctx context.Context) error
}
```

## 语义要求

### 互斥（Mutual Exclusion）

- 对同一个 `key`，同一时刻最多只有一个 `Lease` 存活。
- 当已存在活跃 Lease 时，`TryLock` 必须返回 `ErrLockUnavailable` 或带有该语义的 error。
- 实现不得依赖「最后写入者胜出」语义——必须保证严格的互斥。

### 防死锁（Deadlock Prevention）

- `Lease` 必须有 TTL（time-to-live）。当持有者崩溃或网络分区时，锁在 TTL 到期后自动释放。
- 实现方必须保证 TTL 由存储侧强制执行（如 Redis 的 PX 参数），而非依赖客户端定时器。
- `TryLock` 的 `ctx` 用于控制本次调用的超时，不影响 Lease 的生命周期。

### 超时（Timeout）

- `TryLock` 必须尊重 `ctx` 的取消和超时。当 `ctx.Done()` 触发时，调用应立即返回 `ctx.Err()`。
- `Release` 同样必须尊重 `ctx`。

### 续约（Renewal）

- 当前版本不要求续约接口。TTL 设定后固定不变。
- 如果 job 执行时间可能超过 TTL，调用方应设置足够长的 `WithLockTTL`。
- 后续版本可能引入 `Renew(ctx, Lease) error` 接口。

## 实现指南

### Memory（开发/测试）

```go
type MemoryLocker struct {
    mu    sync.Mutex
    locks map[string]struct{}
}

func (l *MemoryLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (Lease, error) {
    l.mu.Lock()
    defer l.mu.Unlock()
    if _, ok := l.locks[key]; ok {
        return nil, ErrLockUnavailable
    }
    l.locks[key] = struct{}{}
    return &memoryLease{locker: l, key: key}, nil
}
```

适用于单进程集成测试和开发环境。无 TTL 自动回收——依赖显式 Release。

### Redis

```go
// 推荐使用 SET NX PX 实现：
//   SET lock:<key> <token> NX PX <ttl_ms>
// Release 使用 Lua 脚本保证原子性：
//   if redis.call("get",KEYS[1]) == ARGV[1] then
//     return redis.call("del",KEYS[1])
//   else return 0 end
```

- `token` 使用 UUID，防止误删他人锁。
- `ttl` 以毫秒为单位传给 Redis `PX` 参数。
- Release 使用 Lua 脚本保证 check-and-delete 的原子性。

### 分布式（etcd / ZooKeeper / Consul）

- 使用各系统原生的分布式锁机制（如 etcd lease + revision）。
- 保证锁的 fencing token 单调递增（用于防止锁过期后的幽灵写入）。

## 调度器集成

通过 `WithLocker` 和 `WithLockKey` 注册到 job：

```go
locker := NewRedisLocker(redisClient)
scheduler.AddJob(job, trigger,
    schedulex.WithLocker(locker),
    schedulex.WithLockKey("order-settlement"),
    schedulex.WithLockTTL(5*time.Minute),
)
```

### 行为矩阵

| 场景 | 行为 |
|------|------|
| 锁获取成功 | 执行 job，完成后 Release |
| 锁不可用 (`ErrLockUnavailable`) | 发出 `lock_skipped` 事件，跳过本次执行 |
| 锁获取失败 (其他错误) | 发出 `lock_failed` 事件，跳过本次执行 |
| 锁获取成功但 `ctx` 取消 | 立即 Release，不执行 job |

## 约束

- `key` 不能为空字符串。空 key 将由 `WithLockKey` 的默认值（job name）兜底。
- `ttl` 不能为负数。负值由 `WithLockTTL` 返回 `ErrInvalidOption`。
- 实现方不应阻塞超过 `ctx` 超时时间。
- `Release` 必须幂等——多次调用不应报错。
