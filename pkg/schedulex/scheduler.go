package schedulex

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ModuleName = "github.com/ZoneCNH/schedulex"
	Version    = "v0.1.0"

	maxMisfireCatchUp = 128
	misfireGrace      = 100 * time.Millisecond
)

var (
	// ErrSchedulerClosed is returned when mutating a shut-down scheduler.
	ErrSchedulerClosed = errors.New("schedulex: scheduler closed")
	// ErrJobExists is returned when a duplicate job id is registered.
	ErrJobExists = errors.New("schedulex: job already exists")
	// ErrInvalidJob is returned for incomplete job definitions.
	ErrInvalidJob = errors.New("schedulex: job name and trigger are required")
	// ErrInvalidOption is returned when an option violates scheduler contracts.
	ErrInvalidOption = errors.New("schedulex: invalid option")
	// ErrLockUnavailable signals that a lock adapter could not acquire a lease.
	ErrLockUnavailable = errors.New("schedulex: lock unavailable")
)

// Options captures scheduler construction settings.
type Options struct {
	Clock         Clock
	EventSink     EventSink
	MaxConcurrent int
}

// Option configures a Scheduler.
type Option func(*Options) error

// WithClock injects a deterministic clock.
func WithClock(clock Clock) Option {
	return func(opts *Options) error {
		if clock == nil {
			return ErrInvalidOption
		}
		opts.Clock = clock
		return nil
	}
}

// WithEventSink sets the scheduler-level event sink.
func WithEventSink(sink EventSink) Option {
	return func(opts *Options) error {
		opts.EventSink = sink
		return nil
	}
}

// WithMaxConcurrent sets the scheduler-wide execution limit.
func WithMaxConcurrent(limit int) Option {
	return func(opts *Options) error {
		if limit <= 0 {
			return ErrInvalidOption
		}
		opts.MaxConcurrent = limit
		return nil
	}
}

type Scheduler struct {
	clock   Clock
	sink    EventSink
	sem     chan struct{}
	mu      sync.Mutex
	jobs    map[string]*jobState
	started bool
	closed  bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type jobState struct {
	cfg               jobConfig
	running           int
	queued            bool
	queuedDispatching bool
	queuedScheduled   time.Time
	queuedAttempt     int
	attempt           atomic.Int64
}

// NewScheduler constructs a scheduler with standard-library defaults.
func NewScheduler(opts ...Option) (*Scheduler, error) {
	cfg := Options{Clock: NewRealClock(), MaxConcurrent: 1}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	clock := cfg.Clock
	if clock == nil {
		clock = NewRealClock()
	}
	limit := cfg.MaxConcurrent
	if limit <= 0 {
		limit = 1
	}
	return &Scheduler{clock: clock, sink: cfg.EventSink, sem: make(chan struct{}, limit), jobs: map[string]*jobState{}}, nil
}

// AddJob registers a job and deterministic trigger.
func (s *Scheduler) AddJob(job Job, trigger Trigger, opts ...JobOption) error {
	if job == nil || job.Name() == "" || trigger == nil {
		return ErrInvalidJob
	}
	cfg := jobConfig{
		id:            job.Name(),
		job:           job,
		trigger:       trigger,
		misfirePolicy: MisfireSkip,
		overlapPolicy: OverlapSkip,
		lockKey:       job.Name(),
		lockTTL:       time.Minute,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return err
		}
	}
	if cfg.lockKey == "" {
		cfg.lockKey = cfg.id
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSchedulerClosed
	}
	if _, ok := s.jobs[cfg.id]; ok {
		return ErrJobExists
	}
	state := &jobState{cfg: cfg}
	s.jobs[cfg.id] = state
	if s.started {
		s.startLocked(state)
	}
	return nil
}

// Start launches registered job loops. It is idempotent until Shutdown.
func (s *Scheduler) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSchedulerClosed
	}
	if s.started {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.started = true
	for _, state := range s.jobs {
		s.startLocked(state)
	}
	return nil
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.mu.Lock()
	already := s.closed
	if !already {
		s.closed = true
		if s.cancel != nil {
			s.cancel()
		}
	}
	s.mu.Unlock()
	if already {
		return nil
	}
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		s.emit(ctx, Event{Type: EventShutdown, At: s.clock.Now()})
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	snap := Snapshot{
		Version:  Version,
		Now:      now,
		Started:  s.started,
		Running:  s.started && !s.closed,
		Closed:   s.closed,
		Shutdown: s.closed,
		JobCount: len(s.jobs),
		Jobs:     make([]JobSnapshot, 0, len(s.jobs)),
	}
	for id, state := range s.jobs {
		next, ok := state.cfg.trigger.Next(now)
		snap.Jobs = append(snap.Jobs, JobSnapshot{
			ID:            id,
			Name:          state.cfg.job.Name(),
			Next:          next,
			HasNext:       ok,
			MisfirePolicy: state.cfg.misfirePolicy,
			OverlapPolicy: state.cfg.overlapPolicy,
			Running:       state.running > 0,
			Queued:        state.queued || state.queuedDispatching,
		})
	}
	snap.sort()
	return snap
}

func (s *Scheduler) startLocked(state *jobState) {
	s.wg.Add(1)
	go s.loop(state)
}

type eventOption func(*Event)

func withAttempt(attempt int) eventOption {
	return func(e *Event) { e.Attempt = attempt }
}

func withReason(reason string) eventOption {
	return func(e *Event) { e.Reason = reason }
}

func withAttributes(attributes map[string]string) eventOption {
	return func(e *Event) {
		if len(attributes) == 0 {
			return
		}
		e.Attributes = make(map[string]string, len(attributes))
		for key, value := range attributes {
			e.Attributes[key] = value
		}
	}
}

func withErr(err error) eventOption {
	return func(e *Event) {
		if err != nil {
			e.Err = err.Error()
		}
	}
}

func withStarted(at time.Time) eventOption {
	return func(e *Event) {
		e.StartedAt = at
		if !e.ScheduledAt.IsZero() {
			e.Lag = at.Sub(e.ScheduledAt)
		}
	}
}

func withFinished(at time.Time) eventOption {
	return func(e *Event) {
		e.FinishedAt = at
		if !e.StartedAt.IsZero() {
			e.Duration = at.Sub(e.StartedAt)
		}
	}
}

func (s *Scheduler) event(state *jobState, eventType EventType, scheduled time.Time, opts ...eventOption) Event {
	event := Event{
		Type:        eventType,
		JobID:       state.cfg.id,
		JobName:     state.cfg.job.Name(),
		At:          s.clock.Now(),
		ScheduledAt: scheduled,
	}
	for _, opt := range opts {
		opt(&event)
	}
	return event
}

func (s *Scheduler) emit(ctx context.Context, e Event) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.sink != nil {
		s.sink.OnEvent(ctx, e)
	}
	if state := s.jobForEvent(e.JobID); state != nil && state.cfg.eventSink != nil {
		state.cfg.eventSink.OnEvent(ctx, e)
	}
}

func (s *Scheduler) jobForEvent(id string) *jobState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobs[id]
}
