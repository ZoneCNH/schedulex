package schedulex

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// waitForScheduled 等待 EventScheduled 并给 goroutine 时间注册 After() waiter。
// 解决 EventScheduled 在 After() 之前发出导致的竞争条件。
func waitForScheduled(t *testing.T, events *eventRecorder, jobID string) {
	t.Helper()
	events.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == jobID })
	time.Sleep(10 * time.Millisecond)
}

// ────────────────────────────────────────────────────────────────
// ReconcileMisfire — 0% → 100%
// ────────────────────────────────────────────────────────────────

// 空错过列表 → 返回 nil
func TestReconcileMisfire_EmptyMissed(t *testing.T) {
	got := ReconcileMisfire(MisfireCatchUp, nil)
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	got = ReconcileMisfire(MisfireCatchUp, []time.Time{})
	if got != nil {
		t.Fatalf("expected nil for empty slice, got %v", got)
	}
}

// MisfireSkip 策略 → 返回 nil
func TestReconcileMisfire_SkipPolicy(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base, base.Add(time.Minute)}
	got := ReconcileMisfire(MisfireSkip, missed)
	if got != nil {
		t.Fatalf("expected nil for MisfireSkip, got %v", got)
	}
}

// MisfireRunOnce → 返回最后一个
func TestReconcileMisfire_RunOnce(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base, base.Add(time.Minute), base.Add(2 * time.Minute)}
	got := ReconcileMisfire(MisfireRunOnce, missed)
	if len(got) != 1 {
		t.Fatalf("expected 1 run, got %d", len(got))
	}
	if !got[0].Equal(missed[2]) {
		t.Fatalf("expected last missed time, got %v", got[0])
	}
}

// MisfireCatchUp → 返回所有错过的
func TestReconcileMisfire_CatchUp(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base, base.Add(time.Minute)}
	got := ReconcileMisfire(MisfireCatchUp, missed)
	if len(got) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(got))
	}
	for i, m := range missed {
		if !got[i].Equal(m) {
			t.Fatalf("run[%d] = %v; want %v", i, got[i], m)
		}
	}
}

// 验证 CatchUp 返回的切片不修改原切片
func TestReconcileMisfire_CatchUpDoesNotAliasInput(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base}
	got := ReconcileMisfire(MisfireCatchUp, missed)
	got[0] = base.Add(time.Hour)
	if missed[0].Equal(got[0]) {
		t.Fatal("ReconcileMisfire should not alias the input slice")
	}
}

// RunOnce 单个元素
func TestReconcileMisfire_RunOnceSingle(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := ReconcileMisfire(MisfireRunOnce, []time.Time{base})
	if len(got) != 1 || !got[0].Equal(base) {
		t.Fatalf("expected [%v], got %v", base, got)
	}
}

// ────────────────────────────────────────────────────────────────
// PlanMisfire — 补充边界
// ────────────────────────────────────────────────────────────────

// 空错过列表
func TestPlanMisfire_EmptyMissed(t *testing.T) {
	d := PlanMisfire(MisfireCatchUp, nil, time.Time{}, false)
	if len(d.Runs) != 0 || len(d.Skipped) != 0 {
		t.Fatalf("expected empty decision for nil missed: %+v", d)
	}
	d = PlanMisfire(MisfireCatchUp, []time.Time{}, time.Time{}, true)
	if len(d.Runs) != 0 || !d.HasNext {
		t.Fatalf("expected empty runs with HasNext=true: %+v", d)
	}
}

// Skip 策略所有都跳过
func TestPlanMisfire_SkipAll(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base, base.Add(time.Minute)}
	next := base.Add(2 * time.Minute)
	d := PlanMisfire(MisfireSkip, missed, next, true)
	if len(d.Runs) != 0 {
		t.Fatalf("expected 0 runs, got %d", len(d.Runs))
	}
	if len(d.Skipped) != 2 {
		t.Fatalf("expected 2 skipped, got %d", len(d.Skipped))
	}
	if !d.Next.Equal(next) || !d.HasNext {
		t.Fatalf("unexpected Next/HasNext: %+v", d)
	}
}

// ────────────────────────────────────────────────────────────────
// emit — 补充 sink nil、job-level sink、nil ctx
// ────────────────────────────────────────────────────────────────

// sink 为 nil 不应 panic
func TestEmit_NilSink(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	s, err := NewScheduler(WithClock(clock))
	if err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "nil-sink", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := s.jobs["nil-sink"]
	s.emit(context.Background(), s.event(state, EventScheduled, clock.Now()))
	shutdownScheduler(t, s)
}

// scheduler-level sink 正常接收事件
func TestEmit_SchedulerSink(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	var received int32
	s, err := NewScheduler(WithClock(clock), WithEventSink(EventSinkFunc(func(_ context.Context, e Event) {
		atomic.AddInt32(&received, 1)
	})))
	if err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "sink-test", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := s.jobs["sink-test"]
	s.emit(context.Background(), s.event(state, EventScheduled, clock.Now()))
	if atomic.LoadInt32(&received) != 1 {
		t.Fatalf("expected 1 event, got %d", received)
	}
	shutdownScheduler(t, s)
}

// job-level sink 也被调用
func TestEmit_JobLevelSink(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	var schedulerEvents, jobEvents int32
	s, err := NewScheduler(WithClock(clock), WithEventSink(EventSinkFunc(func(_ context.Context, _ Event) {
		atomic.AddInt32(&schedulerEvents, 1)
	})))
	if err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "job-sink", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Hour), WithJobEventSink(EventSinkFunc(func(_ context.Context, _ Event) {
		atomic.AddInt32(&jobEvents, 1)
	}))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := s.jobs["job-sink"]
	s.emit(context.Background(), s.event(state, EventScheduled, clock.Now()))
	if atomic.LoadInt32(&schedulerEvents) != 1 {
		t.Fatalf("scheduler sink expected 1, got %d", schedulerEvents)
	}
	if atomic.LoadInt32(&jobEvents) != 1 {
		t.Fatalf("job sink expected 1, got %d", jobEvents)
	}
	shutdownScheduler(t, s)
}

// nil ctx → 不应 panic
func TestEmit_NilContext(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	s, err := NewScheduler(WithClock(clock))
	if err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "nil-ctx", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	state := s.jobs["nil-ctx"]
	s.emit(nil, s.event(state, EventScheduled, clock.Now()))
	shutdownScheduler(t, s)
}

// ────────────────────────────────────────────────────────────────
// run() — 补充 locker 路径
// ────────────────────────────────────────────────────────────────

// locker 返回 ErrLockUnavailable → 发出 EventLockSkipped
func TestRun_LockSkipped(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	events := newEventRecorder()
	locker := &alwaysLockFail{err: ErrLockUnavailable}

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "lock-skip", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Once(clock.Now().Add(time.Second)), WithLocker(locker)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "lock-skip")
	clock.Advance(time.Second)

	skipped := events.waitFor(t, EventLockSkipped, func(e Event) bool { return e.JobID == "lock-skip" })
	if skipped.Err == "" {
		t.Fatal("expected error in lock_skipped event")
	}
}

// locker 返回其他 error → 发出 EventLockFailed
func TestRun_LockFailed(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	events := newEventRecorder()
	lockErr := errors.New("connection refused")
	locker := &alwaysLockFail{err: lockErr}

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "lock-fail", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Once(clock.Now().Add(time.Second)), WithLocker(locker)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "lock-fail")
	clock.Advance(time.Second)

	failed := events.waitFor(t, EventLockFailed, func(e Event) bool { return e.JobID == "lock-fail" })
	if failed.Err == "" {
		t.Fatal("expected error in lock_failed event")
	}
}

// locker 成功获取锁，任务执行成功
func TestRun_LockSuccess(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	events := newEventRecorder()
	locker := &memoryLocker{held: map[string]bool{}}
	var ran int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "lock-ok", RunFunc: func(context.Context) error {
		atomic.AddInt32(&ran, 1)
		return nil
	}}
	if err := s.AddJob(job, Once(clock.Now().Add(time.Second)), WithLocker(locker)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "lock-ok")
	clock.Advance(time.Second)

	events.waitFor(t, EventSucceeded, func(e Event) bool { return e.JobID == "lock-ok" })
	if atomic.LoadInt32(&ran) != 1 {
		t.Fatalf("expected job to run once, got %d", ran)
	}
}

// 任务执行失败 → EventFailed
func TestRun_JobFailed(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	events := newEventRecorder()

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "fail-job", RunFunc: func(context.Context) error {
		return errors.New("intentional failure")
	}}
	if err := s.AddJob(job, Once(clock.Now().Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "fail-job")
	clock.Advance(time.Second)

	failed := events.waitFor(t, EventFailed, func(e Event) bool { return e.JobID == "fail-job" })
	if failed.Err == "" {
		t.Fatal("expected error in failed event")
	}
	// Duration 在 StaticClock 下为 0（StartedAt == FinishedAt），不检查
}

// Release 返回 error → 不影响调度
func TestRun_LeaseReleaseError(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	var ran int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	locker := &releaseErrLocker{}
	job := JobFunc{NameValue: "release-err", RunFunc: func(context.Context) error {
		atomic.AddInt32(&ran, 1)
		return nil
	}}
	if err := s.AddJob(job, Once(start.Add(time.Second)), WithLocker(locker)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "release-err")
	clock.Advance(time.Second)
	events.waitFor(t, EventSucceeded, func(e Event) bool { return e.JobID == "release-err" })
	if atomic.LoadInt32(&ran) != 1 {
		t.Fatalf("expected 1 run, got %d", ran)
	}
}

// nil locker → 不使用锁
func TestRun_NilLocker(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	var ran int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "no-lock", RunFunc: func(context.Context) error {
		atomic.AddInt32(&ran, 1)
		return nil
	}}
	if err := s.AddJob(job, Once(start.Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "no-lock")
	clock.Advance(time.Second)
	events.waitFor(t, EventSucceeded, func(e Event) bool { return e.JobID == "no-lock" })
	if atomic.LoadInt32(&ran) != 1 {
		t.Fatalf("expected 1 run, got %d", ran)
	}
}

// ────────────────────────────────────────────────────────────────
// OverlapAllow / OverlapSkip — markRunStarted 路径
// ────────────────────────────────────────────────────────────────

// OverlapAllow 允许并发执行
func TestMarkRunStarted_OverlapAllow(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	started := make(chan struct{}, 4)
	release := make(chan struct{})

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(4))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		close(release)
		shutdownScheduler(t, s)
	})

	job := JobFunc{NameValue: "allow", RunFunc: func(context.Context) error {
		started <- struct{}{}
		<-release
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithOverlapPolicy(OverlapAllow)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 第一次触发
	waitForScheduled(t, events, "allow")
	clock.Advance(time.Second)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first run did not start")
	}

	// 第二次触发 — OverlapAllow 应该允许
	events.waitFor(t, EventScheduled, func(e Event) bool {
		return e.JobID == "allow" && e.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	time.Sleep(10 * time.Millisecond)
	clock.Advance(time.Second)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("second run did not start with OverlapAllow")
	}
}

// OverlapSkip 跳过重叠执行
func TestMarkRunStarted_OverlapSkip(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		close(release)
		shutdownScheduler(t, s)
	})

	job := JobFunc{NameValue: "skip", RunFunc: func(context.Context) error {
		started <- struct{}{}
		<-release
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithOverlapPolicy(OverlapSkip)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 第一次触发
	waitForScheduled(t, events, "skip")
	clock.Advance(time.Second)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first run did not start")
	}

	// 第二次触发 — 应该被跳过
	events.waitFor(t, EventScheduled, func(e Event) bool {
		return e.JobID == "skip" && e.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	time.Sleep(10 * time.Millisecond)
	clock.Advance(time.Second)
	time.Sleep(30 * time.Millisecond)

	_, found := events.find(EventSkipped, func(e Event) bool {
		return e.JobID == "skip" && e.Reason == "overlap"
	})
	if !found {
		t.Fatal("expected overlap skip event")
	}
}

// ────────────────────────────────────────────────────────────────
// dispatchQueued — 补充 scheduler 关闭路径
// ────────────────────────────────────────────────────────────────

func TestDispatchQueued_SchedulerClosed(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	started := make(chan struct{}, 2)
	release := make(chan struct{})

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(1))
	if err != nil {
		t.Fatal(err)
	}

	var runs int32
	job := JobFunc{NameValue: "qclosed", RunFunc: func(context.Context) error {
		run := atomic.AddInt32(&runs, 1)
		started <- struct{}{}
		if run == 1 {
			<-release
		}
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithOverlapPolicy(OverlapQueueOne)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 第一次触发
	waitForScheduled(t, events, "qclosed")
	clock.Advance(time.Second)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first run did not start")
	}

	// 排队第二次
	events.waitFor(t, EventScheduled, func(e Event) bool {
		return e.JobID == "qclosed" && e.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	time.Sleep(10 * time.Millisecond)
	clock.Advance(time.Second)

	// 关闭 scheduler
	close(release)
	shutdownScheduler(t, s)
}

// ────────────────────────────────────────────────────────────────
// QueueOne overlap_queue_full 路径
// ────────────────────────────────────────────────────────────────

func TestReserveOverlap_QueueOneFull(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	release := make(chan struct{})
	started := make(chan struct{}, 4)

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(4))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		close(release)
		shutdownScheduler(t, s)
	})

	job := JobFunc{NameValue: "q1", RunFunc: func(context.Context) error {
		started <- struct{}{}
		<-release
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithOverlapPolicy(OverlapQueueOne)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 第一次触发
	waitForScheduled(t, events, "q1")
	clock.Advance(time.Second)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first run did not start")
	}

	// 第二次触发 → 排队
	events.waitFor(t, EventScheduled, func(e Event) bool {
		return e.JobID == "q1" && e.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	time.Sleep(10 * time.Millisecond)
	clock.Advance(time.Second)
	time.Sleep(30 * time.Millisecond)

	// 第三次触发 → overlap_queue_full
	events.waitFor(t, EventScheduled, func(e Event) bool {
		return e.JobID == "q1" && e.ScheduledAt.Equal(start.Add(3*time.Second))
	})
	time.Sleep(10 * time.Millisecond)
	clock.Advance(time.Second)
	time.Sleep(30 * time.Millisecond)

	_, found := events.find(EventSkipped, func(e Event) bool {
		return e.JobID == "q1" && e.Reason == "overlap_queue_full"
	})
	if !found {
		t.Fatal("expected overlap_queue_full skip event")
	}
}

// ────────────────────────────────────────────────────────────────
// collectMissed — maxMisfireCatchUp 上限
// ────────────────────────────────────────────────────────────────

func TestCollectMissed_CappedAtMax(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	events := newEventRecorder()
	var ran int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(128))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "capped", RunFunc: func(context.Context) error {
		atomic.AddInt32(&ran, 1)
		return nil
	}}
	if err := s.AddJob(job, Every(time.Millisecond), WithMisfirePolicy(MisfireCatchUp), WithOverlapPolicy(OverlapAllow)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "capped")
	clock.Advance(200 * time.Millisecond)

	misfire := events.waitFor(t, EventMisfire, func(e Event) bool { return e.JobID == "capped" })
	if misfire.Attributes["capped"] != "true" {
		t.Fatalf("expected capped=true, got %q", misfire.Attributes["capped"])
	}
}

// ────────────────────────────────────────────────────────────────
// NewScheduler — 补充边界
// ────────────────────────────────────────────────────────────────

func TestNewScheduler_NilClock(t *testing.T) {
	_, err := NewScheduler(WithClock(nil))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption for nil clock, got %v", err)
	}
}

func TestNewScheduler_NilOption(t *testing.T) {
	s, err := NewScheduler(nil)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestNewScheduler_ZeroMaxConcurrent(t *testing.T) {
	_, err := NewScheduler(WithMaxConcurrent(0))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption for 0, got %v", err)
	}
	_, err = NewScheduler(WithMaxConcurrent(-1))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption for -1, got %v", err)
	}
}

func TestSystemClock(t *testing.T) {
	c := SystemClock()
	if c == nil {
		t.Fatal("SystemClock returned nil")
	}
}

// ────────────────────────────────────────────────────────────────
// AddJob — 补充边界
// ────────────────────────────────────────────────────────────────

func TestAddJob_NilJob(t *testing.T) {
	s, _ := NewScheduler()
	if err := s.AddJob(nil, Every(time.Second)); !errors.Is(err, ErrInvalidJob) {
		t.Fatalf("expected ErrInvalidJob, got %v", err)
	}
}

func TestAddJob_NilTrigger(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "x", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, nil); !errors.Is(err, ErrInvalidJob) {
		t.Fatalf("expected ErrInvalidJob, got %v", err)
	}
}

func TestAddJob_EmptyName(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second)); !errors.Is(err, ErrInvalidJob) {
		t.Fatalf("expected ErrInvalidJob, got %v", err)
	}
}

func TestAddJob_DuplicateID(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "dup", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := s.AddJob(job, Every(time.Second)); !errors.Is(err, ErrJobExists) {
		t.Fatalf("expected ErrJobExists, got %v", err)
	}
}

func TestAddJob_InvalidMisfirePolicy(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "bad-policy", RunFunc: func(context.Context) error { return nil }}
	err := s.AddJob(job, Every(time.Second), WithMisfirePolicy("invalid"))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption, got %v", err)
	}
}

func TestAddJob_InvalidOverlapPolicy(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "bad-overlap", RunFunc: func(context.Context) error { return nil }}
	err := s.AddJob(job, Every(time.Second), WithOverlapPolicy("invalid"))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption, got %v", err)
	}
}

func TestAddJob_EmptyMisfirePolicy(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "empty-mp", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second), WithMisfirePolicy("")); err != nil {
		t.Fatal(err)
	}
	if s.jobs["empty-mp"].cfg.misfirePolicy != MisfireSkip {
		t.Fatalf("expected MisfireSkip, got %v", s.jobs["empty-mp"].cfg.misfirePolicy)
	}
}

func TestAddJob_EmptyOverlapPolicy(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "empty-op", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second), WithOverlapPolicy("")); err != nil {
		t.Fatal(err)
	}
	if s.jobs["empty-op"].cfg.overlapPolicy != OverlapSkip {
		t.Fatalf("expected OverlapSkip, got %v", s.jobs["empty-op"].cfg.overlapPolicy)
	}
}

func TestAddJob_NegativeJitter(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "neg-jitter", RunFunc: func(context.Context) error { return nil }}
	err := s.AddJob(job, Every(time.Second), WithJitter(JitterPolicy{Max: -1}))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption, got %v", err)
	}
}

func TestAddJob_NegativeLockTTL(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "neg-ttl", RunFunc: func(context.Context) error { return nil }}
	err := s.AddJob(job, Every(time.Second), WithLockTTL(-time.Second))
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("expected ErrInvalidOption, got %v", err)
	}
}

func TestAddJob_CustomLockKey(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "lk", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second), WithLockKey("custom-key")); err != nil {
		t.Fatal(err)
	}
	if s.jobs["lk"].cfg.lockKey != "custom-key" {
		t.Fatalf("expected custom-key, got %v", s.jobs["lk"].cfg.lockKey)
	}
}

func TestAddJob_WithLocker(t *testing.T) {
	s, _ := NewScheduler()
	locker := &memoryLocker{held: map[string]bool{}}
	job := JobFunc{NameValue: "wl", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second), WithLocker(locker)); err != nil {
		t.Fatal(err)
	}
	if s.jobs["wl"].cfg.locker == nil {
		t.Fatal("expected locker to be set")
	}
}

func TestAddJob_EmptyLockKey(t *testing.T) {
	s, _ := NewScheduler()
	job := JobFunc{NameValue: "lk-job", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second), WithLockKey("")); err != nil {
		t.Fatal(err)
	}
	if s.jobs["lk-job"].cfg.lockKey != "lk-job" {
		t.Fatalf("expected lk-job, got %v", s.jobs["lk-job"].cfg.lockKey)
	}
}

func TestAddJob_AfterStart(t *testing.T) {
	// 使用真实时钟避免 StaticClock + goroutine 调度竞争
	var ran int32

	s, err := NewScheduler(WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	job := JobFunc{NameValue: "late-add", RunFunc: func(context.Context) error {
		atomic.AddInt32(&ran, 1)
		return nil
	}}
	if err := s.AddJob(job, Once(time.Now().Add(10*time.Millisecond))); err != nil {
		t.Fatal(err)
	}

	eventually(t, time.Second, func() bool { return atomic.LoadInt32(&ran) == 1 })
}

func TestAddJob_AfterShutdown(t *testing.T) {
	s, _ := NewScheduler()
	shutdownScheduler(t, s)
	job := JobFunc{NameValue: "late", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second)); !errors.Is(err, ErrSchedulerClosed) {
		t.Fatalf("expected ErrSchedulerClosed, got %v", err)
	}
}

// ────────────────────────────────────────────────────────────────
// Start — 补充边界
// ────────────────────────────────────────────────────────────────

func TestStart_Idempotent(t *testing.T) {
	s, _ := NewScheduler()
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal("second Start should be idempotent")
	}
	shutdownScheduler(t, s)
}

func TestStart_CanceledContext(t *testing.T) {
	s, _ := NewScheduler()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Start(ctx); err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestStart_NilContext(t *testing.T) {
	s, _ := NewScheduler()
	if err := s.Start(nil); err != nil {
		t.Fatal(err)
	}
	shutdownScheduler(t, s)
}

func TestStart_AfterShutdown(t *testing.T) {
	s, _ := NewScheduler()
	shutdownScheduler(t, s)
	if err := s.Start(context.Background()); !errors.Is(err, ErrSchedulerClosed) {
		t.Fatalf("expected ErrSchedulerClosed, got %v", err)
	}
}

// ────────────────────────────────────────────────────────────────
// Shutdown — 补充 timeout 路径
// ────────────────────────────────────────────────────────────────

func TestShutdown_Timeout(t *testing.T) {
	// 使用真实时钟测试 timeout 行为
	s, err := NewScheduler(WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	job := JobFunc{NameValue: "block", RunFunc: func(context.Context) error {
		<-release
		return nil
	}}
	if err := s.AddJob(job, Every(10*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond) // 等待任务开始执行

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	err = s.Shutdown(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	close(release)
	// 等待 goroutine 清理
	time.Sleep(20 * time.Millisecond)
}

func TestShutdown_NilContext(t *testing.T) {
	s, _ := NewScheduler()
	if err := s.Shutdown(nil); err != nil {
		t.Fatal(err)
	}
}

// ────────────────────────────────────────────────────────────────
// Every.Next — 补充边界
// ────────────────────────────────────────────────────────────────

func TestEveryNext_Normal(t *testing.T) {
	trig := Every(time.Minute)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next, ok := trig.Next(now)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !next.Equal(now.Add(time.Minute)) {
		t.Fatalf("expected %v, got %v", now.Add(time.Minute), next)
	}
}

func TestEveryNext_Zero(t *testing.T) {
	trig := Every(0)
	_, ok := trig.Next(time.Now())
	if ok {
		t.Fatal("expected ok=false for zero interval")
	}
}

func TestEveryNext_Negative(t *testing.T) {
	trig := Every(-time.Second)
	_, ok := trig.Next(time.Now())
	if ok {
		t.Fatal("expected ok=false for negative interval")
	}
}

func TestEveryNext_SequentialCalls(t *testing.T) {
	trig := Every(time.Hour)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	n1, ok := trig.Next(base)
	if !ok || !n1.Equal(base.Add(time.Hour)) {
		t.Fatalf("n1 = %v, ok = %v", n1, ok)
	}
	n2, ok := trig.Next(n1)
	if !ok || !n2.Equal(base.Add(2*time.Hour)) {
		t.Fatalf("n2 = %v, ok = %v", n2, ok)
	}
}

// ────────────────────────────────────────────────────────────────
// DailyAt.Next — 补充边界
// ────────────────────────────────────────────────────────────────

func TestDailyAtNext_Before(t *testing.T) {
	loc := time.UTC
	after := time.Date(2026, 6, 4, 8, 0, 0, 0, loc)
	trig := DailyAt(10, 30, loc)
	next, ok := trig.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	want := time.Date(2026, 6, 4, 10, 30, 0, 0, loc)
	if !next.Equal(want) {
		t.Fatalf("expected %v, got %v", want, next)
	}
}

func TestDailyAtNext_After(t *testing.T) {
	loc := time.UTC
	after := time.Date(2026, 6, 4, 15, 0, 0, 0, loc)
	trig := DailyAt(10, 30, loc)
	next, ok := trig.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	want := time.Date(2026, 6, 5, 10, 30, 0, 0, loc)
	if !next.Equal(want) {
		t.Fatalf("expected %v, got %v", want, next)
	}
}

func TestDailyAtNext_ExactMatch(t *testing.T) {
	loc := time.UTC
	after := time.Date(2026, 6, 4, 10, 30, 0, 0, loc)
	trig := DailyAt(10, 30, loc)
	next, ok := trig.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	want := time.Date(2026, 6, 5, 10, 30, 0, 0, loc)
	if !next.Equal(want) {
		t.Fatalf("expected %v, got %v", want, next)
	}
}

func TestDailyAtNext_NilLoc(t *testing.T) {
	after := time.Date(2026, 6, 4, 8, 0, 0, 0, time.UTC)
	trig := DailyAt(10, 0, nil)
	next, ok := trig.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if next.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", next.Location())
	}
}

func TestDailyAtNext_DifferentTimezone(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skip(err)
	}
	after := time.Date(2026, 6, 4, 8, 0, 0, 0, loc)
	trig := DailyAt(9, 0, loc)
	next, ok := trig.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if next.Location() != loc {
		t.Fatalf("expected %v, got %v", loc, next.Location())
	}
	if next.Hour() != 9 || next.Minute() != 0 {
		t.Fatalf("expected 09:00, got %02d:%02d", next.Hour(), next.Minute())
	}
}

// ────────────────────────────────────────────────────────────────
// Once trigger
// ────────────────────────────────────────────────────────────────

func TestOnceNext_Past(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trig := Once(at)
	_, ok := trig.Next(at.Add(time.Second))
	if ok {
		t.Fatal("expected ok=false for past time")
	}
}

func TestOnceNext_Future(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trig := Once(at)
	next, ok := trig.Next(at.Add(-time.Second))
	if !ok || !next.Equal(at) {
		t.Fatalf("expected %v, ok=true; got %v, %v", at, next, ok)
	}
}

// ────────────────────────────────────────────────────────────────
// JobFunc.Run — nil RunFunc
// ────────────────────────────────────────────────────────────────

func TestJobFunc_NilRunFunc(t *testing.T) {
	job := JobFunc{NameValue: "nil-func"}
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if job.Name() != "nil-func" {
		t.Fatalf("expected nil-func, got %v", job.Name())
	}
}

// ────────────────────────────────────────────────────────────────
// runJob — panic 恢复
// ────────────────────────────────────────────────────────────────

func TestRunJob_Panic(t *testing.T) {
	job := JobFunc{NameValue: "panic-job", RunFunc: func(context.Context) error {
		panic("test panic")
	}}
	err := runJob(context.Background(), job)
	if err == nil {
		t.Fatal("expected error from panic")
	}
	if got := err.Error(); got != "schedulex: job panic: test panic" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestRunJob_Success(t *testing.T) {
	job := JobFunc{NameValue: "ok", RunFunc: func(context.Context) error { return nil }}
	if err := runJob(context.Background(), job); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunJob_Error(t *testing.T) {
	jobErr := errors.New("job error")
	job := JobFunc{NameValue: "err", RunFunc: func(context.Context) error { return jobErr }}
	if err := runJob(context.Background(), job); !errors.Is(err, jobErr) {
		t.Fatalf("expected %v, got %v", jobErr, err)
	}
}

// ────────────────────────────────────────────────────────────────
// EventSinkFunc — nil func
// ────────────────────────────────────────────────────────────────

func TestEventSinkFunc_Nil(t *testing.T) {
	var f EventSinkFunc
	f.OnEvent(context.Background(), Event{Type: EventScheduled})
}

// ────────────────────────────────────────────────────────────────
// OverlapPolicy / MisfirePolicy valid()
// ────────────────────────────────────────────────────────────────

func TestOverlapPolicy_Valid(t *testing.T) {
	cases := []struct {
		p    OverlapPolicy
		want bool
	}{
		{OverlapSkip, true},
		{OverlapQueueOne, true},
		{OverlapAllow, true},
		{"invalid", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := tc.p.valid(); got != tc.want {
			t.Errorf("%q.valid() = %v; want %v", tc.p, got, tc.want)
		}
	}
}

func TestMisfirePolicy_Valid(t *testing.T) {
	cases := []struct {
		p    MisfirePolicy
		want bool
	}{
		{MisfireSkip, true},
		{MisfireRunOnce, true},
		{MisfireCatchUp, true},
		{"invalid", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := tc.p.valid(); got != tc.want {
			t.Errorf("%q.valid() = %v; want %v", tc.p, got, tc.want)
		}
	}
}

// ────────────────────────────────────────────────────────────────
// Snapshot — 状态标记
// ────────────────────────────────────────────────────────────────

func TestSnapshot_BeforeStart(t *testing.T) {
	s, _ := NewScheduler()
	snap := s.Snapshot()
	if snap.Started || snap.Running || snap.Closed || snap.Shutdown {
		t.Fatalf("expected all false before start: %+v", snap)
	}
}

func TestSnapshot_AfterStart(t *testing.T) {
	s, _ := NewScheduler()
	_ = s.Start(context.Background())
	snap := s.Snapshot()
	if !snap.Started || !snap.Running || snap.Closed {
		t.Fatalf("expected started+running after start: %+v", snap)
	}
	shutdownScheduler(t, s)
}

func TestSnapshot_AfterShutdown(t *testing.T) {
	s, _ := NewScheduler()
	shutdownScheduler(t, s)
	snap := s.Snapshot()
	if !snap.Closed || !snap.Shutdown {
		t.Fatalf("expected closed+shutdown: %+v", snap)
	}
}

func TestSnapshot_SortedByID(t *testing.T) {
	s, _ := NewScheduler()
	for _, name := range []string{"c", "a", "b"} {
		job := JobFunc{NameValue: name, RunFunc: func(context.Context) error { return nil }}
		_ = s.AddJob(job, Every(time.Hour))
	}
	snap := s.Snapshot()
	if snap.Jobs[0].ID != "a" || snap.Jobs[1].ID != "b" || snap.Jobs[2].ID != "c" {
		t.Fatalf("jobs not sorted: %v %v %v", snap.Jobs[0].ID, snap.Jobs[1].ID, snap.Jobs[2].ID)
	}
}

func TestSnapshot_RunningFlag(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	release := make(chan struct{})

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		close(release)
		shutdownScheduler(t, s)
	})

	job := JobFunc{NameValue: "snap-run", RunFunc: func(context.Context) error {
		<-release
		return nil
	}}
	if err := s.AddJob(job, Once(start.Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	waitForScheduled(t, events, "snap-run")
	clock.Advance(time.Second)
	time.Sleep(20 * time.Millisecond)

	snap := s.Snapshot()
	found := false
	for _, j := range snap.Jobs {
		if j.ID == "snap-run" {
			found = true
			if !j.Running {
				t.Fatal("expected job to be marked running")
			}
		}
	}
	if !found {
		t.Fatal("job snap-run not found in snapshot")
	}
}

func TestJobSnapshot_Fields(t *testing.T) {
	s, _ := NewScheduler()
	locker := &memoryLocker{held: map[string]bool{}}
	job := JobFunc{NameValue: "fields", RunFunc: func(context.Context) error { return nil }}
	_ = s.AddJob(job, Every(time.Hour),
		WithMisfirePolicy(MisfireCatchUp),
		WithOverlapPolicy(OverlapAllow),
		WithLocker(locker),
	)
	snap := s.Snapshot()
	if len(snap.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(snap.Jobs))
	}
	j := snap.Jobs[0]
	if j.MisfirePolicy != MisfireCatchUp {
		t.Fatalf("expected MisfireCatchUp, got %v", j.MisfirePolicy)
	}
	if j.OverlapPolicy != OverlapAllow {
		t.Fatalf("expected OverlapAllow, got %v", j.OverlapPolicy)
	}
	if !j.HasNext {
		t.Fatal("expected HasNext=true")
	}
}

// ────────────────────────────────────────────────────────────────
// event helpers — withAttributes, withErr, withStarted, withFinished
// ────────────────────────────────────────────────────────────────

func TestWithAttributes_EmptyMap(t *testing.T) {
	e := Event{}
	withAttributes(nil)(&e)
	if e.Attributes != nil {
		t.Fatal("expected nil attributes for nil map")
	}
	withAttributes(map[string]string{})(&e)
	if e.Attributes != nil {
		t.Fatal("expected nil attributes for empty map")
	}
}

func TestWithAttributes_NonEmpty(t *testing.T) {
	e := Event{}
	withAttributes(map[string]string{"key": "value"})(&e)
	if e.Attributes["key"] != "value" {
		t.Fatalf("expected value, got %v", e.Attributes["key"])
	}
}

func TestWithErr_Nil(t *testing.T) {
	e := Event{Err: "existing"}
	withErr(nil)(&e)
	if e.Err != "existing" {
		t.Fatalf("expected existing err preserved, got %v", e.Err)
	}
}

func TestWithErr_NonNil(t *testing.T) {
	e := Event{}
	withErr(errors.New("test"))(&e)
	if e.Err != "test" {
		t.Fatalf("expected test, got %v", e.Err)
	}
}

func TestWithStarted_LagCalculation(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := Event{ScheduledAt: base}
	withStarted(base.Add(100 * time.Millisecond))(&e)
	if e.Lag != 100*time.Millisecond {
		t.Fatalf("expected 100ms lag, got %v", e.Lag)
	}
}

func TestWithStarted_NoScheduledAt(t *testing.T) {
	e := Event{}
	withStarted(time.Now())(&e)
	if e.Lag != 0 {
		t.Fatalf("expected 0 lag, got %v", e.Lag)
	}
}

func TestWithFinished_DurationCalculation(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := Event{StartedAt: base}
	withFinished(base.Add(50 * time.Millisecond))(&e)
	if e.Duration != 50*time.Millisecond {
		t.Fatalf("expected 50ms duration, got %v", e.Duration)
	}
}

func TestWithFinished_NoStartedAt(t *testing.T) {
	e := Event{}
	withFinished(time.Now())(&e)
	if e.Duration != 0 {
		t.Fatalf("expected 0 duration, got %v", e.Duration)
	}
}

// ────────────────────────────────────────────────────────────────
// Cron trigger — 补充边界
// ────────────────────────────────────────────────────────────────

func TestCron_FixedMinute(t *testing.T) {
	cron, err := Cron("30 * * * *", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	next, ok := cron.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if next.Minute() != 30 {
		t.Fatalf("expected minute 30, got %d", next.Minute())
	}
}

func TestCron_FixedHour(t *testing.T) {
	cron, err := Cron("0 14 * * *", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	next, ok := cron.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if next.Hour() != 14 || next.Minute() != 0 {
		t.Fatalf("expected 14:00, got %02d:%02d", next.Hour(), next.Minute())
	}
}

func TestCron_FixedHourAndMinute(t *testing.T) {
	cron, err := Cron("30 14 * * *", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	next, ok := cron.Next(after)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if next.Hour() != 14 || next.Minute() != 30 {
		t.Fatalf("expected 14:30, got %02d:%02d", next.Hour(), next.Minute())
	}
}

func TestCron_InvalidExpressions(t *testing.T) {
	if _, err := Cron("* * *", time.UTC); err == nil {
		t.Fatal("expected error for 3 fields")
	}
	if _, err := Cron("0 0 1 * *", time.UTC); err == nil {
		t.Fatal("expected error for non-wildcard day")
	}
	if _, err := Cron("*/0 * * * *", time.UTC); err == nil {
		t.Fatal("expected error for */0")
	}
	if _, err := Cron("60 * * * *", time.UTC); err == nil {
		t.Fatal("expected error for minute 60")
	}
	if _, err := Cron("* 24 * * *", time.UTC); err == nil {
		t.Fatal("expected error for hour 24")
	}
	if _, err := Cron("abc * * * *", time.UTC); err == nil {
		t.Fatal("expected error for abc")
	}
}

// ────────────────────────────────────────────────────────────────
// misfireAttributes — 补充 capped 路径
// ────────────────────────────────────────────────────────────────

func TestMisfireAttributes_Capped(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base, base.Add(time.Minute)}
	decision := MisfireDecision{Runs: missed, Skipped: nil}
	attrs := misfireAttributes(missed, decision, true)
	if attrs["capped"] != "true" {
		t.Fatalf("expected capped=true, got %v", attrs["capped"])
	}
	if attrs["missed"] != "2" {
		t.Fatalf("expected missed=2, got %v", attrs["missed"])
	}
	if attrs["runs"] != "2" {
		t.Fatalf("expected runs=2, got %v", attrs["runs"])
	}
}

func TestMisfireAttributes_Empty(t *testing.T) {
	attrs := misfireAttributes(nil, MisfireDecision{}, false)
	if attrs["missed"] != "0" {
		t.Fatalf("expected missed=0, got %v", attrs["missed"])
	}
	if _, ok := attrs["capped"]; ok {
		t.Fatal("expected no capped key")
	}
	if _, ok := attrs["first_missed"]; ok {
		t.Fatal("expected no first_missed key")
	}
}

// ────────────────────────────────────────────────────────────────
// Jitter — 边界
// ────────────────────────────────────────────────────────────────

func TestJitter_ZeroMax(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := JitterPolicy{Max: 0, Seed: 42}
	got := ApplyDeterministicJitter(base, p, "job", 1)
	if !got.Equal(base) {
		t.Fatalf("expected base, got %v", got)
	}
}

func TestJitter_DifferentJobIDs(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := JitterPolicy{Max: time.Second, Seed: 42}
	a := ApplyDeterministicJitter(base, p, "job-a", 1)
	b := ApplyDeterministicJitter(base, p, "job-b", 1)
	if a.Equal(b) {
		t.Fatal("different job IDs should produce different jitter")
	}
}

func TestJitter_DifferentRuns(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := JitterPolicy{Max: time.Second, Seed: 42}
	a := ApplyDeterministicJitter(base, p, "job", 1)
	b := ApplyDeterministicJitter(base, p, "job", 2)
	if a.Equal(b) {
		t.Fatal("different runs should produce different jitter")
	}
}

// ────────────────────────────────────────────────────────────────
// 多任务并发执行
// ────────────────────────────────────────────────────────────────

func TestMultipleJobsConcurrentExecution(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	var completed atomic.Int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(4))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	for i := 0; i < 3; i++ {
		name := "multi-" + string(rune('a'+i))
		job := JobFunc{NameValue: name, RunFunc: func(context.Context) error {
			completed.Add(1)
			return nil
		}}
		if err := s.AddJob(job, Once(start.Add(time.Second))); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 等待所有任务的 scheduled 事件，确保 After() 已注册
	for i := 0; i < 3; i++ {
		name := "multi-" + string(rune('a'+i))
		waitForScheduled(t, events, name)
	}
	clock.Advance(time.Second)
	eventually(t, 2*time.Second, func() bool { return completed.Load() == 3 })
}

// ────────────────────────────────────────────────────────────────
// 任务 panic 后 scheduler 继续运行
// ────────────────────────────────────────────────────────────────

func TestSchedulerContinuesAfterJobPanic(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	var secondRan int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	panicJob := JobFunc{NameValue: "panicker", RunFunc: func(context.Context) error {
		panic("boom")
	}}
	safeJob := JobFunc{NameValue: "safe", RunFunc: func(context.Context) error {
		atomic.AddInt32(&secondRan, 1)
		return nil
	}}

	if err := s.AddJob(panicJob, Once(start.Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.AddJob(safeJob, Once(start.Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 等待两个任务的 scheduled 事件，确保 After() 已注册
	waitForScheduled(t, events, "panicker")
	waitForScheduled(t, events, "safe")
	clock.Advance(time.Second)
	eventually(t, 2*time.Second, func() bool { return atomic.LoadInt32(&secondRan) == 1 })
}

// ────────────────────────────────────────────────────────────────
// WithMaxConcurrent 正常值
// ────────────────────────────────────────────────────────────────

func TestWithMaxConcurrent_Valid(t *testing.T) {
	s, err := NewScheduler(WithMaxConcurrent(8))
	if err != nil {
		t.Fatal(err)
	}
	if cap(s.sem) != 8 {
		t.Fatalf("expected sem cap 8, got %d", cap(s.sem))
	}
}

// ────────────────────────────────────────────────────────────────
// ModuleConstants
// ────────────────────────────────────────────────────────────────

func TestModuleConstants(t *testing.T) {
	if ModuleName != "github.com/ZoneCNH/schedulex" {
		t.Fatalf("unexpected ModuleName: %v", ModuleName)
	}
	if Version != "v0.1.0" {
		t.Fatalf("unexpected Version: %v", Version)
	}
}

// ────────────────────────────────────────────────────────────────
// 辅助类型
// ────────────────────────────────────────────────────────────────

type alwaysLockFail struct {
	err error
}

func (a *alwaysLockFail) TryLock(_ context.Context, _ string, _ time.Duration) (Lease, error) {
	return nil, a.err
}

type releaseErrLocker struct{}

func (r *releaseErrLocker) TryLock(_ context.Context, _ string, _ time.Duration) (Lease, error) {
	return &releaseErrLease{}, nil
}

type releaseErrLease struct{}

func (r *releaseErrLease) Release(_ context.Context) error {
	return errors.New("release failed")
}
