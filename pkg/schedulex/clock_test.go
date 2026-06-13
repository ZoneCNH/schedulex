package schedulex

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestStaticClockAfterWaitsUntilAdvanced(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	timer := clock.After(10 * time.Minute)

	select {
	case got := <-timer:
		t.Fatalf("After fired before advance: %v", got)
	default:
	}

	clock.Advance(9 * time.Minute)
	select {
	case got := <-timer:
		t.Fatalf("After fired before target: %v", got)
	default:
	}

	clock.Advance(time.Minute)
	got := <-timer
	want := start.Add(10 * time.Minute)
	if !got.Equal(want) {
		t.Fatalf("After fired at %v; want %v", got, want)
	}
}

func TestStaticClockAfterNonPositiveDurationFiresImmediately(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	timer := clock.After(0)

	select {
	case got := <-timer:
		if !got.Equal(start) {
			t.Fatalf("After(0) fired at %v; want %v", got, start)
		}
	default:
		t.Fatal("After(0) did not fire immediately")
	}
}

func TestStaticClockSchedulerWaitsForAdvance(t *testing.T) {
	start := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	scheduled := make(chan Event, 1)
	ran := make(chan struct{}, 1)
	var runCount atomic.Int32

	s, err := NewScheduler(
		WithClock(clock),
		WithEventSink(EventSinkFunc(func(_ context.Context, event Event) {
			if event.Type != EventScheduled {
				return
			}
			select {
			case scheduled <- event:
			default:
			}
		})),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			t.Fatalf("Shutdown: %v", err)
		}
	})

	job := JobFunc{NameValue: "minute", RunFunc: func(context.Context) error {
		runCount.Add(1)
		select {
		case ran <- struct{}{}:
		default:
		}
		return nil
	}}
	if err := s.AddJob(job, Every(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	select {
	case event := <-scheduled:
		want := start.Add(time.Minute)
		if !event.ScheduledAt.Equal(want) {
			t.Fatalf("scheduled at %v; want %v", event.ScheduledAt, want)
		}
	case <-time.After(time.Second):
		t.Fatal("scheduler did not schedule job")
	}

	select {
	case <-ran:
		t.Fatal("job ran before static clock advanced")
	default:
	}

	waitForStaticClockWaiter(t, clock, start.Add(time.Minute))
	clock.Advance(time.Minute)
	expectStartedSignal(t, clock, ran, "job did not run after static clock advanced")
	if got := runCount.Load(); got != 1 {
		t.Fatalf("run count = %d; want 1", got)
	}
}
