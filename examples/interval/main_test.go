package main

import (
	"context"
	"testing"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func TestExampleCompiles(t *testing.T) { main() }

func TestHeartbeatJob(t *testing.T) {
	// Exercise the closure body in heartbeatJob.RunFunc.
	if err := heartbeatJob.Run(context.Background()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if heartbeatJob.Name() != "heartbeat" {
		t.Fatalf("Name() = %q, want %q", heartbeatJob.Name(), "heartbeat")
	}
}

func TestSchedulerAddIntervalJob(t *testing.T) {
	s, err := schedulex.NewScheduler()
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	if err := s.AddJob(heartbeatJob, schedulex.Every(time.Minute)); err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	snap := s.Snapshot()
	if snap.JobCount != 1 {
		t.Fatalf("JobCount = %d, want 1", snap.JobCount)
	}
	if snap.Version == "" {
		t.Fatal("expected non-empty version after AddJob")
	}
}
