package schedulex

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMisfirePolicyContracts(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	missed := []time.Time{base, base.Add(time.Minute), base.Add(2 * time.Minute)}
	got := map[string]int{
		"skip_runs":     len(PlanMisfire(MisfireSkip, missed, time.Time{}, false).Runs),
		"skip_skipped":  len(PlanMisfire(MisfireSkip, missed, time.Time{}, false).Skipped),
		"run_once_runs": len(PlanMisfire(MisfireRunOnce, missed, time.Time{}, false).Runs),
		"catch_up_runs": len(PlanMisfire(MisfireCatchUp, missed, time.Time{}, false).Runs),
	}
	assertGoldenJSON(t, "../../contracts/misfire_cases/l1_golden.json", got)
}

func TestJitterDeterministic(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := JitterPolicy{Max: time.Second, Seed: 42}
	a := ApplyDeterministicJitter(base, p, "job", 7)
	b := ApplyDeterministicJitter(base, p, "job", 7)
	if !a.Equal(b) {
		t.Fatalf("jitter not deterministic: %s vs %s", a, b)
	}
	if a.Sub(base) < 0 || a.Sub(base) > time.Second {
		t.Fatalf("jitter out of range: %s", a.Sub(base))
	}
}

func TestSchedulerShutdownIdempotentAndEvents(t *testing.T) {
	var eventsMu sync.Mutex
	var events []EventType
	s, err := NewScheduler(WithMaxConcurrent(1), WithEventSink(EventSinkFunc(func(_ context.Context, e Event) {
		eventsMu.Lock()
		events = append(events, e.Type)
		eventsMu.Unlock()
	})))
	if err != nil {
		t.Fatal(err)
	}
	var ran atomic.Int32
	job := JobFunc{NameValue: "once", RunFunc: func(context.Context) error {
		ran.Add(1)
		return nil
	}}
	if err := s.AddJob(job, Once(time.Now().Add(20*time.Millisecond))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	eventually(t, time.Second, func() bool { return ran.Load() == 1 })
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if snap := s.Snapshot(); !snap.Closed || snap.JobCount != 1 {
		t.Fatalf("bad snapshot: %+v", snap)
	}
}

func TestMaxConcurrencyAndOverlapSkip(t *testing.T) {
	start := make(chan struct{})
	finish := make(chan struct{})
	var active, maxActive atomic.Int32
	s, err := NewScheduler(WithMaxConcurrent(1))
	if err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "j", RunFunc: func(context.Context) error {
		a := active.Add(1)
		if a > maxActive.Load() {
			maxActive.Store(a)
		}
		close(start)
		<-finish
		active.Add(-1)
		return nil
	}}
	if err := s.AddJob(job, Every(time.Millisecond), WithOverlapPolicy(OverlapSkip)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	eventually(t, time.Second, func() bool {
		select {
		case <-start:
			return true
		default:
			return false
		}
	})
	time.Sleep(5 * time.Millisecond)
	close(finish)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if maxActive.Load() > 1 {
		t.Fatalf("max concurrency exceeded: %d", maxActive.Load())
	}
}

func TestLockerInterfaceContract(t *testing.T) {
	locker := &memoryLocker{held: map[string]bool{}}
	lease, err := locker.TryLock(context.Background(), "k", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := locker.TryLock(context.Background(), "k", time.Minute); !errors.Is(err, ErrLockUnavailable) {
		t.Fatalf("want ErrLockUnavailable, got %v", err)
	}
	if err := lease.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := locker.TryLock(context.Background(), "k", time.Minute); err != nil {
		t.Fatal(err)
	}
}

func TestSchedulerLeakBudget(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 5; i++ {
		s, err := NewScheduler()
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := s.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
	}
	eventually(t, time.Second, func() bool { return runtime.NumGoroutine() <= before+4 })
}

func eventually(t *testing.T, d time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met")
}

type memoryLocker struct {
	mu   sync.Mutex
	held map[string]bool
}
type memoryLease struct {
	l   *memoryLocker
	key string
}

func (m *memoryLocker) TryLock(_ context.Context, key string, _ time.Duration) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.held[key] {
		return nil, ErrLockUnavailable
	}
	m.held[key] = true
	return &memoryLease{m, key}, nil
}
func (m *memoryLease) Release(context.Context) error {
	m.l.mu.Lock()
	defer m.l.mu.Unlock()
	delete(m.l.held, m.key)
	return nil
}
