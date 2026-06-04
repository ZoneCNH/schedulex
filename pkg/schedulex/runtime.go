package schedulex

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrSchedulerShutdown is returned when mutating a shut-down scheduler.
	ErrSchedulerShutdown = errors.New("schedulex: scheduler shut down")
	// ErrJobExists is returned when adding a duplicate job id.
	ErrJobExists = errors.New("schedulex: job already exists")
	// ErrJobInvalid is returned for incomplete job definitions.
	ErrJobInvalid = errors.New("schedulex: job id, trigger, and run function are required")
)

// Scheduler owns job registration and deterministic snapshots.
type Scheduler struct {
	mu       sync.Mutex
	clock    Clock
	jobs     map[string]Job
	running  bool
	shutdown bool
	done     chan struct{}
}

// Option configures a Scheduler.
type Option func(*Scheduler)

// WithClock injects a deterministic clock.
func WithClock(clock Clock) Option {
	return func(s *Scheduler) {
		if clock != nil {
			s.clock = clock
		}
	}
}

// NewScheduler constructs a scheduler with standard-library defaults.
func NewScheduler(opts ...Option) *Scheduler {
	s := &Scheduler{clock: SystemClock(), jobs: map[string]Job{}, done: make(chan struct{})}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// AddJob registers a job. It does not start background execution by itself.
func (s *Scheduler) AddJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shutdown {
		return ErrSchedulerShutdown
	}
	if job.ID == "" || job.Trigger == nil || job.Run == nil {
		return ErrJobInvalid
	}
	if _, exists := s.jobs[job.ID]; exists {
		return ErrJobExists
	}
	s.jobs[job.ID] = normalizeJob(job)
	return nil
}

// Start marks the scheduler as running until the context is cancelled or Shutdown is called.
// v0.1 keeps execution adapters outside the deterministic core.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return ErrSchedulerShutdown
	}
	s.running = true
	done := s.done
	s.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// Shutdown is idempotent and unblocks Start.
func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.shutdown {
		s.shutdown = true
		s.running = false
		close(s.done)
	}
	s.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Snapshot returns a stable view of registered jobs and next fire times.
func (s *Scheduler) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	snap := Snapshot{Now: now, Running: s.running, Shutdown: s.shutdown, Jobs: make([]JobSnapshot, 0, len(s.jobs))}
	for id, job := range s.jobs {
		next, ok := job.Trigger.Next(now)
		snap.Jobs = append(snap.Jobs, JobSnapshot{ID: id, Next: next, HasNext: ok, MisfirePolicy: job.MisfirePolicy, OverlapPolicy: job.OverlapPolicy})
	}
	snap.sort()
	return snap
}

// ReconcileMisfire calculates due run instants for release golden cases.
func ReconcileMisfire(policy MisfirePolicy, missed []time.Time) []time.Time {
	if len(missed) == 0 || policy == MisfireSkip {
		return nil
	}
	if policy == MisfireRunOnce {
		return []time.Time{missed[len(missed)-1]}
	}
	out := append([]time.Time(nil), missed...)
	return out
}
