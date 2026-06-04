package schedulex

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerAppliesJitterBeforeDispatch(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	first := start.Add(time.Minute)
	jitter := nonZeroJitterPolicy(t, first, "jittered", 1)
	expected := ApplyDeterministicJitter(first, jitter, "jittered", 1)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	ran := make(chan struct{}, 1)

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "jittered", RunFunc: func(context.Context) error {
		select {
		case ran <- struct{}{}:
		default:
		}
		return nil
	}}
	if err := s.AddJob(job, Every(time.Minute), WithJitter(jitter)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	scheduled := events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "jittered"
	})
	if !scheduled.ScheduledAt.Equal(expected) {
		t.Fatalf("scheduled at %v; want jittered instant %v", scheduled.ScheduledAt, expected)
	}

	clock.Set(first)
	select {
	case <-ran:
		t.Fatal("job ran before jittered scheduled instant")
	case <-time.After(20 * time.Millisecond):
	}

	clock.Set(expected)
	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatal("job did not run at jittered scheduled instant")
	}
}

func TestSchedulerCatchUpMisfireDispatchesMissedRuns(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	var runs atomic.Int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(8))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "catch-up", RunFunc: func(context.Context) error {
		runs.Add(1)
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithMisfirePolicy(MisfireCatchUp), WithOverlapPolicy(OverlapAllow)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "catch-up" && event.ScheduledAt.Equal(start.Add(time.Second))
	})
	clock.Advance(3 * time.Second)

	misfire := events.waitFor(t, EventMisfire, func(event Event) bool {
		return event.JobID == "catch-up"
	})
	if got := misfire.Attributes["missed"]; got != "3" {
		t.Fatalf("misfire missed = %q; want 3", got)
	}
	if got := misfire.Attributes["runs"]; got != "3" {
		t.Fatalf("misfire runs = %q; want 3", got)
	}
	if got := misfire.Attributes["skipped"]; got != "0" {
		t.Fatalf("misfire skipped = %q; want 0", got)
	}

	eventually(t, time.Second, func() bool { return runs.Load() == 3 })
}

func TestSchedulerSkipMisfireOnSingleLateRun(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	ran := make(chan struct{}, 1)
	var runs atomic.Int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "single-late-skip", RunFunc: func(context.Context) error {
		runs.Add(1)
		ran <- struct{}{}
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithMisfirePolicy(MisfireSkip)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "single-late-skip" && event.ScheduledAt.Equal(start.Add(time.Second))
	})
	clock.Set(start.Add(1500 * time.Millisecond))

	misfire := events.waitFor(t, EventMisfire, func(event Event) bool {
		return event.JobID == "single-late-skip"
	})
	if got := misfire.Attributes["missed"]; got != "1" {
		t.Fatalf("misfire missed = %q; want 1", got)
	}
	if got := misfire.Attributes["runs"]; got != "0" {
		t.Fatalf("misfire runs = %q; want 0", got)
	}
	if got := misfire.Attributes["skipped"]; got != "1" {
		t.Fatalf("misfire skipped = %q; want 1", got)
	}
	select {
	case <-ran:
		t.Fatal("misfire skip ran the late job")
	case <-time.After(50 * time.Millisecond):
	}
	if got := runs.Load(); got != 0 {
		t.Fatalf("runs = %d; want 0", got)
	}
}

func TestSchedulerGlobalBackpressureDoesNotMarkPendingJobRunning(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	blockerStarted := make(chan struct{}, 1)
	pendingStarted := make(chan struct{}, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		shutdownScheduler(t, s)
	})

	blocker := JobFunc{NameValue: "blocker", RunFunc: func(context.Context) error {
		blockerStarted <- struct{}{}
		<-release
		return nil
	}}
	pending := JobFunc{NameValue: "pending", RunFunc: func(context.Context) error {
		pendingStarted <- struct{}{}
		return nil
	}}
	if err := s.AddJob(blocker, Once(start.Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.AddJob(pending, Once(start.Add(2*time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "blocker" && event.ScheduledAt.Equal(start.Add(time.Second))
	})
	clock.Advance(time.Second)
	select {
	case <-blockerStarted:
	case <-time.After(time.Second):
		t.Fatal("blocker did not start")
	}
	events.waitFor(t, EventStarted, func(event Event) bool {
		return event.JobID == "blocker"
	})

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "pending" && event.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	clock.Advance(time.Second)

	select {
	case <-pendingStarted:
		t.Fatal("pending job started while global concurrency slot was occupied")
	case <-time.After(50 * time.Millisecond):
	}
	assertJobSnapshot(t, s, "pending", func(job JobSnapshot) bool {
		return !job.Running
	}, "pending job should not be marked running while waiting for global slot")

	releaseOnce.Do(func() { close(release) })
	select {
	case <-pendingStarted:
	case <-time.After(time.Second):
		t.Fatal("pending job did not start after global slot was released")
	}
	if skipped, found := events.find(EventSkipped, func(event Event) bool {
		return event.JobID == "pending" && event.Reason == "overlap"
	}); found {
		t.Fatalf("pending job was skipped as overlap while only global slot was blocked: %+v", skipped)
	}
}

func TestSchedulerQueueOneRunsQueuedAfterCurrentCompletes(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	release := make(chan struct{})
	started := make(chan int, 2)
	var runs atomic.Int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "queue-one", RunFunc: func(context.Context) error {
		run := int(runs.Add(1))
		started <- run
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

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "queue-one" && event.ScheduledAt.Equal(start.Add(time.Second))
	})
	clock.Advance(time.Second)
	expectStartedRun(t, started, 1)

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "queue-one" && event.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	clock.Advance(time.Second)
	select {
	case run := <-started:
		t.Fatalf("queued run started before current run completed: %d", run)
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	expectStartedRun(t, started, 2)
	if got := runs.Load(); got != 2 {
		t.Fatalf("runs = %d; want 2", got)
	}
}

func TestSchedulerQueueOneHonorsMisfireSkipForStaleQueuedRun(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()
	release := make(chan struct{})
	var releaseOnce sync.Once
	started := make(chan int, 2)
	var runs atomic.Int32

	s, err := NewScheduler(WithClock(clock), WithEventSink(events), WithMaxConcurrent(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		shutdownScheduler(t, s)
	})

	job := JobFunc{NameValue: "queue-one-stale", RunFunc: func(context.Context) error {
		run := int(runs.Add(1))
		started <- run
		if run == 1 {
			<-release
		}
		return nil
	}}
	if err := s.AddJob(job, Every(time.Second), WithOverlapPolicy(OverlapQueueOne), WithMisfirePolicy(MisfireSkip)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "queue-one-stale" && event.ScheduledAt.Equal(start.Add(time.Second))
	})
	clock.Advance(time.Second)
	expectStartedRun(t, started, 1)

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "queue-one-stale" && event.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	clock.Advance(time.Second)
	select {
	case run := <-started:
		t.Fatalf("queued stale run started before current run completed: %d", run)
	case <-time.After(30 * time.Millisecond):
	}

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "queue-one-stale" && event.ScheduledAt.Equal(start.Add(3*time.Second))
	})
	clock.Advance(time.Second)
	releaseOnce.Do(func() { close(release) })

	misfire := events.waitFor(t, EventMisfire, func(event Event) bool {
		return event.JobID == "queue-one-stale" && event.ScheduledAt.Equal(start.Add(2*time.Second))
	})
	if got := misfire.Attributes["runs"]; got != "0" {
		t.Fatalf("queued misfire runs = %q; want 0", got)
	}
	if got := misfire.Attributes["skipped"]; got != "1" {
		t.Fatalf("queued misfire skipped = %q; want 1", got)
	}
	if event, found := events.find(EventStarted, func(event Event) bool {
		return event.JobID == "queue-one-stale" && event.ScheduledAt.Equal(start.Add(2*time.Second))
	}); found {
		t.Fatalf("stale queued run started after misfire skip: %+v", event)
	}
}

func TestSchedulerRecoversJobPanicAsFailedEvent(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	events := newEventRecorder()

	s, err := NewScheduler(WithClock(clock), WithEventSink(events))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdownScheduler(t, s) })

	job := JobFunc{NameValue: "panic", RunFunc: func(context.Context) error {
		panic("boom")
	}}
	if err := s.AddJob(job, Once(start.Add(time.Second))); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	events.waitFor(t, EventScheduled, func(event Event) bool {
		return event.JobID == "panic" && event.ScheduledAt.Equal(start.Add(time.Second))
	})
	clock.Advance(time.Second)
	failed := events.waitFor(t, EventFailed, func(event Event) bool {
		return event.JobID == "panic"
	})
	if !strings.Contains(failed.Err, "job panic: boom") {
		t.Fatalf("failed err = %q; want panic recovery", failed.Err)
	}
}

func nonZeroJitterPolicy(t *testing.T, base time.Time, jobID string, attempt int64) JitterPolicy {
	t.Helper()
	for seed := int64(1); seed < 128; seed++ {
		policy := JitterPolicy{Max: time.Second, Seed: seed}
		if !ApplyDeterministicJitter(base, policy, jobID, attempt).Equal(base) {
			return policy
		}
	}
	t.Fatal("could not find non-zero deterministic jitter policy")
	return JitterPolicy{}
}

func shutdownScheduler(t *testing.T, s *Scheduler) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func expectStartedRun(t *testing.T, started <-chan int, want int) {
	t.Helper()
	select {
	case got := <-started:
		if got != want {
			t.Fatalf("started run = %d; want %d", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("run %d did not start", want)
	}
}

func assertJobSnapshot(t *testing.T, s *Scheduler, id string, ok func(JobSnapshot) bool, message string) {
	t.Helper()
	snap := s.Snapshot()
	for _, job := range snap.Jobs {
		if job.ID != id {
			continue
		}
		if !ok(job) {
			t.Fatalf("%s: %+v", message, job)
		}
		return
	}
	t.Fatalf("job %q not found in snapshot: %+v", id, snap.Jobs)
}

type eventRecorder struct {
	mu     sync.Mutex
	events []Event
	ch     chan Event
}

func newEventRecorder() *eventRecorder {
	return &eventRecorder{ch: make(chan Event, 64)}
}

func (r *eventRecorder) OnEvent(_ context.Context, event Event) {
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()

	select {
	case r.ch <- event:
	default:
	}
}

func (r *eventRecorder) waitFor(t *testing.T, eventType EventType, ok func(Event) bool) Event {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		if event, found := r.find(eventType, ok); found {
			return event
		}
		select {
		case <-r.ch:
		case <-deadline:
			t.Fatalf("event %s not observed", eventType)
		}
	}
}

func (r *eventRecorder) find(eventType EventType, ok func(Event) bool) (Event, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, event := range r.events {
		if event.Type != eventType {
			continue
		}
		if ok == nil || ok(event) {
			return event, true
		}
	}
	return Event{}, false
}
