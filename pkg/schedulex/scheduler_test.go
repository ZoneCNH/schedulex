package schedulex

import (
	"context"
	"testing"
	"time"
)

func TestSchedulerSnapshotIsDeterministic(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC))
	s := NewScheduler(WithClock(clock))
	if err := s.AddJob(Job{ID: "hourly", Trigger: Every(time.Hour), Run: func(context.Context) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	snap := s.Snapshot()
	if len(snap.Jobs) != 1 || snap.Jobs[0].ID != "hourly" || !snap.Jobs[0].Next.Equal(clock.Now().Add(time.Hour)) {
		t.Fatalf("unexpected snapshot: %#v", snap)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	s := NewScheduler()
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.AddJob(Job{ID: "x", Trigger: Every(time.Second), Run: func(context.Context) error { return nil }}); err != ErrSchedulerShutdown {
		t.Fatalf("AddJob after shutdown = %v", err)
	}
}
