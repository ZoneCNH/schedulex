package schedulex

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	// ErrSchedulerShutdown is retained as a compatibility alias.
	ErrSchedulerShutdown = ErrSchedulerClosed
	// ErrJobExists is returned when a duplicate job id is registered.
	ErrJobExists = errors.New("schedulex: job already exists")
	// ErrInvalidJob is returned for incomplete job definitions.
	ErrInvalidJob = errors.New("schedulex: job name and trigger are required")
	// ErrJobInvalid is retained as a compatibility alias.
	ErrJobInvalid = ErrInvalidJob
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
		return ErrJobInvalid
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

func (s *Scheduler) loop(state *jobState) {
	defer s.wg.Done()
	after := s.clock.Now()
	for {
		next, ok := state.cfg.trigger.Next(after)
		if !ok {
			return
		}
		attempt := int(state.attempt.Add(1))
		scheduled := ApplyDeterministicJitter(next, state.cfg.jitterPolicy, state.cfg.id, int64(attempt))
		s.emit(s.ctx, s.event(state, EventScheduled, scheduled, withAttempt(attempt)))
		wait := scheduled.Sub(s.clock.Now())
		if wait < 0 {
			wait = 0
		}
		select {
		case <-s.ctx.Done():
			return
		case <-s.clock.After(wait):
		}

		now := s.clock.Now()
		if s.shouldReconcileMisfire(state, next, scheduled, now) {
			missed, capped := s.collectMissed(state, next, now)
			if len(missed) == 0 {
				after = next
				continue
			}
			decision := PlanMisfire(state.cfg.misfirePolicy, missed, time.Time{}, false)
			s.emit(s.ctx, s.event(
				state,
				EventMisfire,
				scheduled,
				withAttempt(attempt),
				withReason(string(state.cfg.misfirePolicy)),
				withAttributes(misfireAttributes(missed, decision, capped)),
			))
			for _, missedAt := range decision.Runs {
				runAttempt := attempt
				if !missedAt.Equal(next) {
					runAttempt = int(state.attempt.Add(1))
				}
				runScheduled := ApplyDeterministicJitter(missedAt, state.cfg.jitterPolicy, state.cfg.id, int64(runAttempt))
				s.dispatchReady(state, runScheduled, runAttempt)
			}
			after = missed[len(missed)-1]
			continue
		}

		s.dispatch(state, scheduled, attempt)
		after = next
	}
}

func (s *Scheduler) dispatch(state *jobState, scheduled time.Time, attempt int) {
	s.dispatchRun(state, scheduled, attempt, true)
}

func (s *Scheduler) dispatchReady(state *jobState, scheduled time.Time, attempt int) {
	s.dispatchRun(state, scheduled, attempt, false)
}

func (s *Scheduler) dispatchRun(state *jobState, scheduled time.Time, attempt int, reconcileMisfire bool) {
	runnable, reason := s.reserveOverlap(state, scheduled, attempt)
	if !runnable {
		if reason != "queued" {
			s.emit(s.ctx, s.event(state, EventSkipped, scheduled, withAttempt(attempt), withReason(reason)))
		}
		return
	}

	releaseSlot, ok := s.acquireSlot()
	if !ok {
		return
	}

	if reconcileMisfire {
		var shouldRun bool
		scheduled, shouldRun = s.reconcilePendingMisfire(state, scheduled, attempt)
		if !shouldRun {
			releaseSlot()
			return
		}
	}

	runnable, reason = s.markRunStarted(state, scheduled, attempt)
	if !runnable {
		releaseSlot()
		if reason != "queued" {
			s.emit(s.ctx, s.event(state, EventSkipped, scheduled, withAttempt(attempt), withReason(reason)))
		}
		return
	}

	s.startRun(state, scheduled, attempt, releaseSlot)
}

func (s *Scheduler) reserveOverlap(state *jobState, scheduled time.Time, attempt int) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	busy := state.running > 0 || state.queuedDispatching
	if !busy || state.cfg.overlapPolicy == OverlapAllow {
		return true, ""
	}
	if state.cfg.overlapPolicy == OverlapQueueOne {
		if state.queuedDispatching {
			return false, "overlap_queue_full"
		}
		if !state.queued {
			state.queued = true
			state.queuedScheduled = scheduled
			state.queuedAttempt = attempt
			return false, "queued"
		}
		return false, "overlap_queue_full"
	}
	return false, "overlap"
}

func (s *Scheduler) acquireSlot() (func(), bool) {
	select {
	case s.sem <- struct{}{}:
		return func() { <-s.sem }, true
	case <-s.ctx.Done():
		return nil, false
	}
}

func (s *Scheduler) reconcilePendingMisfire(state *jobState, scheduled time.Time, attempt int) (time.Time, bool) {
	now := s.clock.Now()
	if !s.shouldReconcileMisfire(state, scheduled, scheduled, now) {
		return scheduled, true
	}
	decision := PlanMisfire(state.cfg.misfirePolicy, []time.Time{scheduled}, time.Time{}, false)
	s.emit(s.ctx, s.event(
		state,
		EventMisfire,
		scheduled,
		withAttempt(attempt),
		withReason(string(state.cfg.misfirePolicy)),
		withAttributes(misfireAttributes([]time.Time{scheduled}, decision, false)),
	))
	if len(decision.Runs) == 0 {
		return time.Time{}, false
	}
	return decision.Runs[len(decision.Runs)-1], true
}

func (s *Scheduler) markRunStarted(state *jobState, scheduled time.Time, attempt int) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if state.cfg.overlapPolicy == OverlapAllow {
		state.running++
		return true, ""
	}
	if state.running == 0 && !state.queuedDispatching {
		state.running = 1
		return true, ""
	}
	if state.cfg.overlapPolicy == OverlapQueueOne {
		if state.queuedDispatching {
			return false, "overlap_queue_full"
		}
		if !state.queued {
			state.queued = true
			state.queuedScheduled = scheduled
			state.queuedAttempt = attempt
			return false, "queued"
		}
		return false, "overlap_queue_full"
	}
	return false, "overlap"
}

func (s *Scheduler) startRun(state *jobState, scheduled time.Time, attempt int, releaseSlot func()) {
	s.wg.Add(1)
	go s.run(state, scheduled, attempt, releaseSlot)
}

func (s *Scheduler) finishRun(state *jobState) {
	var queuedScheduled time.Time
	var queuedAttempt int
	var startQueued bool

	s.mu.Lock()
	if state.running > 0 {
		state.running--
	}
	canStartQueued := !s.closed && s.ctx != nil && s.ctx.Err() == nil
	if state.running == 0 && state.queued && canStartQueued {
		queuedScheduled = state.queuedScheduled
		queuedAttempt = state.queuedAttempt
		state.queued = false
		state.queuedDispatching = true
		state.queuedScheduled = time.Time{}
		state.queuedAttempt = 0
		startQueued = true
	} else if state.running == 0 && !canStartQueued {
		state.queued = false
		state.queuedDispatching = false
		state.queuedScheduled = time.Time{}
		state.queuedAttempt = 0
	}
	s.mu.Unlock()

	if startQueued {
		s.dispatchQueued(state, queuedScheduled, queuedAttempt)
	}
}

func (s *Scheduler) dispatchQueued(state *jobState, scheduled time.Time, attempt int) {
	releaseSlot, ok := s.acquireSlot()
	if !ok {
		s.clearQueuedDispatching(state)
		return
	}

	var shouldRun bool
	scheduled, shouldRun = s.reconcilePendingMisfire(state, scheduled, attempt)
	if !shouldRun {
		releaseSlot()
		s.clearQueuedDispatching(state)
		return
	}

	if !s.markQueuedRunStarted(state) {
		releaseSlot()
		return
	}
	s.startRun(state, scheduled, attempt, releaseSlot)
}

func (s *Scheduler) clearQueuedDispatching(state *jobState) {
	s.mu.Lock()
	state.queuedDispatching = false
	s.mu.Unlock()
}

func (s *Scheduler) markQueuedRunStarted(state *jobState) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !state.queuedDispatching || s.closed || s.ctx == nil || s.ctx.Err() != nil {
		state.queuedDispatching = false
		return false
	}
	state.queuedDispatching = false
	state.running++
	return true
}

func (s *Scheduler) run(state *jobState, scheduled time.Time, attempt int, releaseSlot func()) {
	defer s.wg.Done()
	defer s.finishRun(state)
	defer releaseSlot()

	var lease Lease
	if state.cfg.locker != nil {
		l, err := state.cfg.locker.TryLock(s.ctx, state.cfg.lockKey, state.cfg.lockTTL)
		if err != nil {
			eventType := EventLockFailed
			if errors.Is(err, ErrLockUnavailable) {
				eventType = EventLockSkipped
			}
			s.emit(s.ctx, s.event(state, eventType, scheduled, withAttempt(attempt), withErr(err)))
			return
		}
		lease = l
		defer func() { _ = lease.Release(context.Background()) }()
	}
	started := s.clock.Now()
	s.emit(s.ctx, s.event(state, EventStarted, scheduled, withAttempt(attempt), withStarted(started)))
	if err := runJob(s.ctx, state.cfg.job); err != nil {
		s.emit(s.ctx, s.event(state, EventFailed, scheduled, withAttempt(attempt), withStarted(started), withFinished(s.clock.Now()), withErr(err)))
		return
	}
	s.emit(s.ctx, s.event(state, EventSucceeded, scheduled, withAttempt(attempt), withStarted(started), withFinished(s.clock.Now())))
}

func runJob(ctx context.Context, job Job) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("schedulex: job panic: %v", recovered)
		}
	}()
	return job.Run(ctx)
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

func (s *Scheduler) shouldReconcileMisfire(_ *jobState, _, scheduled, now time.Time) bool {
	return now.After(scheduled.Add(misfireGrace))
}

func (s *Scheduler) collectMissed(state *jobState, first, now time.Time) ([]time.Time, bool) {
	missed := make([]time.Time, 0, 4)
	for due := first; !due.After(now); {
		missed = append(missed, due)
		if len(missed) >= maxMisfireCatchUp {
			return missed, true
		}
		next, ok := state.cfg.trigger.Next(due)
		if !ok {
			return missed, false
		}
		due = next
	}
	return missed, false
}

func misfireAttributes(missed []time.Time, decision MisfireDecision, capped bool) map[string]string {
	attributes := map[string]string{
		"missed":  strconv.Itoa(len(missed)),
		"runs":    strconv.Itoa(len(decision.Runs)),
		"skipped": strconv.Itoa(len(decision.Skipped)),
	}
	if capped {
		attributes["capped"] = "true"
	}
	if len(missed) > 0 {
		attributes["first_missed"] = missed[0].Format(time.RFC3339Nano)
		attributes["last_missed"] = missed[len(missed)-1].Format(time.RFC3339Nano)
	}
	return attributes
}

// ReconcileMisfire calculates due run instants for release golden cases.
func ReconcileMisfire(policy MisfirePolicy, missed []time.Time) []time.Time {
	if len(missed) == 0 || policy == MisfireSkip {
		return nil
	}
	if policy == MisfireRunOnce {
		return []time.Time{missed[len(missed)-1]}
	}
	return append([]time.Time(nil), missed...)
}
