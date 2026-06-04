package schedulex

import (
	"context"
	"sync"
	"time"
)

type Scheduler struct {
	clock Clock
	sink EventSink
	sem chan struct{}
	mu sync.Mutex
	jobs map[string]Job
	started bool
	closed bool
	ctx context.Context
	cancel context.CancelFunc
	wg sync.WaitGroup
	running map[string]bool
}

func NewScheduler(opts Options) *Scheduler {
	clock := opts.Clock; if clock == nil { clock = NewRealClock() }
	max := opts.MaxConcurrent; if max <= 0 { max = 1 }
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{clock: clock, sink: opts.EventSink, sem: make(chan struct{}, max), jobs: map[string]Job{}, ctx: ctx, cancel: cancel, running: map[string]bool{}}
}

func (s *Scheduler) AddJob(job Job) error {
	if job.ID == "" || job.Trigger == nil || job.Run == nil { return ErrInvalidJob }
	s.mu.Lock(); defer s.mu.Unlock()
	if s.closed { return ErrSchedulerClosed }
	if _, ok := s.jobs[job.ID]; ok { return ErrJobExists }
	if job.LockKey == "" { job.LockKey = job.ID }
	s.jobs[job.ID] = job
	if s.started { s.startLocked(job) }
	return nil
}

func (s *Scheduler) Start() error {
	s.mu.Lock(); defer s.mu.Unlock()
	if s.closed { return ErrSchedulerClosed }
	if s.started { return nil }
	s.started = true
	for _, j := range s.jobs { s.startLocked(j) }
	return nil
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock(); already := s.closed; if !already { s.closed = true; s.cancel() }; s.mu.Unlock()
	if already { return nil }
	done := make(chan struct{})
	go func(){ s.wg.Wait(); close(done) }()
	select { case <-done: return nil; case <-ctx.Done(): return ctx.Err() }
}

func (s *Scheduler) Snapshot() Snapshot {
	s.mu.Lock(); defer s.mu.Unlock()
	return Snapshot{Version: Version, Started: s.started, Closed: s.closed, JobCount: len(s.jobs)}
}

func (s *Scheduler) startLocked(job Job) { s.wg.Add(1); go s.loop(job) }

func (s *Scheduler) loop(job Job) {
	defer s.wg.Done()
	after := s.clock.Now()
	var run int64
	for {
		next, ok := job.Trigger.Next(after); if !ok { return }
		s.emit(Event{Type: EventScheduled, JobID: job.ID, At: s.clock.Now(), ScheduledAt: next})
		wait := next.Sub(s.clock.Now()); if wait < 0 { wait = 0 }
		select { case <-s.ctx.Done(): return; case <-s.clock.After(wait): }
		scheduled := ApplyDeterministicJitter(next, job.Jitter, job.ID, run)
		run++
		s.run(job, scheduled)
		after = next
	}
}

func (s *Scheduler) run(job Job, scheduled time.Time) {
	if job.Overlap != OverlapAllow {
		s.mu.Lock(); if s.running[job.ID] { s.mu.Unlock(); s.emit(Event{Type: EventSkipped, JobID: job.ID, At: s.clock.Now(), ScheduledAt: scheduled}); return }; s.running[job.ID]=true; s.mu.Unlock()
		defer func(){ s.mu.Lock(); s.running[job.ID]=false; s.mu.Unlock() }()
	}
	select { case s.sem <- struct{}{}: defer func(){<-s.sem}(); case <-s.ctx.Done(): return }
	var lease Lease
	if job.Locker != nil {
		l, err := job.Locker.Lock(s.ctx, job.LockKey); if err != nil { s.emit(Event{Type: EventSkipped, JobID: job.ID, At: s.clock.Now(), ScheduledAt: scheduled, Error: err.Error()}); return }
		lease = l; defer lease.Release(context.Background())
	}
	s.emit(Event{Type: EventStarted, JobID: job.ID, At: s.clock.Now(), ScheduledAt: scheduled})
	if err := job.Run(s.ctx); err != nil { s.emit(Event{Type: EventFailed, JobID: job.ID, At: s.clock.Now(), ScheduledAt: scheduled, Error: err.Error()}); return }
	s.emit(Event{Type: EventSucceeded, JobID: job.ID, At: s.clock.Now(), ScheduledAt: scheduled})
}

func (s *Scheduler) emit(e Event) { if s.sink != nil { s.sink.Emit(e) } }
