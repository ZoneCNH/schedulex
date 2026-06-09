package main

import (
	"context"
	"testing"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func TestExampleCompiles(t *testing.T) { main() }

func TestOnceJob(t *testing.T) {
	// Exercise the closure body in onceJob.RunFunc.
	if err := onceJob.Run(context.Background()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if onceJob.Name() != "once" {
		t.Fatalf("Name() = %q, want %q", onceJob.Name(), "once")
	}
}

func TestSchedulerAddOnceJob(t *testing.T) {
	s, err := schedulex.NewScheduler()
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	if err := s.AddJob(onceJob, schedulex.Once(time.Now().Add(time.Hour))); err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	if s.Snapshot().JobCount != 1 {
		t.Fatalf("JobCount = %d, want 1", s.Snapshot().JobCount)
	}
}
