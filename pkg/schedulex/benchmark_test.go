package schedulex

import (
	"context"
	"sync"
	"testing"
	"time"
)

// ────────────────────────────────────────────────────────────────
// 辅助：用于 benchmark 的最小 Job 实现
// ────────────────────────────────────────────────────────────────

type noopJob struct{ name string }

func (n noopJob) Name() string              { return n.name }
func (noopJob) Run(_ context.Context) error { return nil }

// 辅助：用于 benchmark 的内存锁实现
type benchLocker struct {
	mu   sync.Mutex
	held map[string]bool
}

type benchLease struct {
	l   *benchLocker
	key string
}

func newBenchLocker() *benchLocker { return &benchLocker{held: map[string]bool{}} }

func (b *benchLocker) TryLock(_ context.Context, key string, _ time.Duration) (Lease, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.held[key] {
		return nil, ErrLockUnavailable
	}
	b.held[key] = true
	return &benchLease{b, key}, nil
}

func (b *benchLease) Release(_ context.Context) error {
	b.l.mu.Lock()
	defer b.l.mu.Unlock()
	delete(b.l.held, b.key)
	return nil
}

// ────────────────────────────────────────────────────────────────
// 1. Scheduler 基准
// ────────────────────────────────────────────────────────────────

// BenchmarkNewScheduler 测量创建调度器的开销。
func BenchmarkNewScheduler(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewScheduler()
	}
}

// BenchmarkAddJob 测量添加任务的开销。
func BenchmarkAddJob(b *testing.B) {
	if testing.Short() {
		b.Skip("short 模式跳过")
	}
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	trigger := Every(time.Hour)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewScheduler(WithClock(clock))
		j := JobFunc{NameValue: "bench", RunFunc: func(context.Context) error { return nil }}
		_ = s.AddJob(j, trigger)
	}
}

// BenchmarkSchedulerStartStop 测量启动→停机的完整生命周期。
func BenchmarkSchedulerStartStop(b *testing.B) {
	if testing.Short() {
		b.Skip("short 模式跳过")
	}
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	job := noopJob{name: "life"}
	trigger := Every(time.Hour)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewScheduler(WithClock(clock))
		_ = s.AddJob(job, trigger)
		_ = s.Start(context.Background())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_ = s.Shutdown(ctx)
		cancel()
	}
}

// ────────────────────────────────────────────────────────────────
// 2. Trigger 基准
// ────────────────────────────────────────────────────────────────

// BenchmarkTriggerOnce 测量 Once trigger 的 Next 调用。
func BenchmarkTriggerOnce(b *testing.B) {
	now := time.Now()
	t := Once(now.Add(time.Hour))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.Next(now)
	}
}

// BenchmarkTriggerEvery 测量 Every trigger 的 Next 调用。
func BenchmarkTriggerEvery(b *testing.B) {
	now := time.Now()
	t := Every(time.Minute)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.Next(now)
	}
}

// BenchmarkTriggerDailyAt 测量 DailyAt trigger 的 Next 调用。
func BenchmarkTriggerDailyAt(b *testing.B) {
	now := time.Now()
	t := DailyAt(9, 0, time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.Next(now)
	}
}

// BenchmarkTriggerCron 测量 Cron trigger 的 Next 调用。
func BenchmarkTriggerCron(b *testing.B) {
	now := time.Now()
	t, _ := Cron("*/5 * * * *", time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = t.Next(now)
	}
}

// ────────────────────────────────────────────────────────────────
// 3. 调度核心基准
// ────────────────────────────────────────────────────────────────

// BenchmarkDispatchRun 测量 dispatchRun 执行链路。
func BenchmarkDispatchRun(b *testing.B) {
	if testing.Short() {
		b.Skip("short 模式跳过")
	}
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	s, _ := NewScheduler(WithClock(clock), WithMaxConcurrent(64))
	job := noopJob{name: "dispatch"}
	_ = s.AddJob(job, Every(time.Hour))
	_ = s.Start(context.Background())
	state := s.jobs["dispatch"]
	now := clock.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.mu.Lock()
		state.running = 0
		state.queued = false
		state.queuedDispatching = false
		s.mu.Unlock()
		s.dispatchRun(state, now, i+1, false)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_ = s.Shutdown(ctx)
	cancel()
}

// BenchmarkRunJob 测量单任务执行（runJob 函数）。
func BenchmarkRunJob(b *testing.B) {
	ctx := context.Background()
	job := noopJob{name: "run"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runJob(ctx, job)
	}
}

// BenchmarkEventEmit 测量事件发射（含 nil sink 路径）。
func BenchmarkEventEmit(b *testing.B) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	s, _ := NewScheduler(WithClock(clock))
	job := noopJob{name: "emit"}
	_ = s.AddJob(job, Every(time.Hour))
	_ = s.Start(context.Background())
	state := s.jobs["emit"]
	now := clock.Now()
	ctx := context.Background()
	evt := s.event(state, EventScheduled, now)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.emit(ctx, evt)
	}
	ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
	_ = s.Shutdown(ctx2)
	cancel()
}

// BenchmarkSnapshot 测量快照生成。
func BenchmarkSnapshot(b *testing.B) {
	if testing.Short() {
		b.Skip("short 模式跳过")
	}
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	s, _ := NewScheduler(WithClock(clock))
	for i := 0; i < 10; i++ {
		j := JobFunc{
			NameValue: "snap-" + string(rune('a'+i)),
			RunFunc:   func(context.Context) error { return nil },
		}
		_ = s.AddJob(j, Every(time.Duration(i+1)*time.Hour))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Snapshot()
	}
}

// ────────────────────────────────────────────────────────────────
// 4. Jitter / Misfire 基准
// ────────────────────────────────────────────────────────────────

// BenchmarkApplyDeterministicJitter 测量抖动计算。
func BenchmarkApplyDeterministicJitter(b *testing.B) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := JitterPolicy{Max: time.Second, Seed: 42}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyDeterministicJitter(base, p, "bench-job", int64(i))
	}
}

// BenchmarkPlanMisfire 测量 misfire 决策。
func BenchmarkPlanMisfire(b *testing.B) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := make([]time.Time, 10)
	for i := range missed {
		missed[i] = base.Add(time.Duration(i) * time.Minute)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PlanMisfire(MisfireCatchUp, missed, time.Time{}, false)
	}
}

// BenchmarkReconcileMisfire 测量 misfire 协调。
func BenchmarkReconcileMisfire(b *testing.B) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := make([]time.Time, 10)
	for i := range missed {
		missed[i] = base.Add(time.Duration(i) * time.Minute)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ReconcileMisfire(MisfireCatchUp, missed)
	}
}

// ────────────────────────────────────────────────────────────────
// 5. Lock 基准
// ────────────────────────────────────────────────────────────────

// BenchmarkLockAcquireRelease 测量锁获取释放的开销。
func BenchmarkLockAcquireRelease(b *testing.B) {
	locker := newBenchLocker()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lease, err := locker.TryLock(ctx, "bench-key", time.Minute)
		if err != nil {
			b.Fatal(err)
		}
		_ = lease.Release(ctx)
	}
}
