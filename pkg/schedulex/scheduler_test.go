package schedulex

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSchedulerSnapshotIsDeterministic(t *testing.T) {
	clock := NewStaticClock(time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC))
	s, err := NewScheduler(WithClock(clock))
	if err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "hourly", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Hour)); err != nil {
		t.Fatal(err)
	}
	snap := s.Snapshot()
	if len(snap.Jobs) != 1 || snap.Jobs[0].ID != "hourly" || !snap.Jobs[0].Next.Equal(clock.Now().Add(time.Hour)) {
		t.Fatalf("unexpected snapshot: %#v", snap)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	s, err := NewScheduler()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	job := JobFunc{NameValue: "x", RunFunc: func(context.Context) error { return nil }}
	if err := s.AddJob(job, Every(time.Second)); !errors.Is(err, ErrSchedulerClosed) {
		t.Fatalf("AddJob after shutdown = %v", err)
	}
}
