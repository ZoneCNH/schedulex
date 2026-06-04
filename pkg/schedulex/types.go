// Package schedulex provides a small deterministic scheduling core.
package schedulex

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ModuleName = "github.com/ZoneCNH/schedulex"
	Version    = "v0.1.0"
)

// Clock supplies time to scheduler decisions. Tests can inject ManualClock.
type Clock interface{ Now() time.Time }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// ManualClock is a deterministic test clock.
type ManualClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewManualClock(now time.Time) *ManualClock { return &ManualClock{now: now} }
func (c *ManualClock) Now() time.Time           { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *ManualClock) Advance(d time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	return c.now
}
func (c *ManualClock) Set(t time.Time) { c.mu.Lock(); c.now = t; c.mu.Unlock() }

// Trigger returns the next run strictly after the supplied time, or false when exhausted.
type Trigger interface {
	Next(after time.Time) (time.Time, bool)
}

type onceTrigger struct{ at time.Time }

func Once(at time.Time) Trigger { return onceTrigger{at: at} }
func (t onceTrigger) Next(after time.Time) (time.Time, bool) {
	if t.at.After(after) || t.at.Equal(after) {
		return t.at, true
	}
	return time.Time{}, false
}

type everyTrigger struct{ interval time.Duration }

func Every(interval time.Duration) Trigger { return everyTrigger{interval: interval} }
func (t everyTrigger) Next(after time.Time) (time.Time, bool) {
	if t.interval <= 0 {
		return time.Time{}, false
	}
	return after.Add(t.interval), true
}

type dailyAtTrigger struct {
	loc                  *time.Location
	hour, minute, second int
}

func DailyAt(loc *time.Location, hour, minute, second int) Trigger {
	if loc == nil {
		loc = time.UTC
	}
	return dailyAtTrigger{loc: loc, hour: hour, minute: minute, second: second}
}
func (t dailyAtTrigger) Next(after time.Time) (time.Time, bool) {
	local := after.In(t.loc)
	next := time.Date(local.Year(), local.Month(), local.Day(), t.hour, t.minute, t.second, 0, t.loc)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next, true
}

type cronTrigger struct {
	minutes, hours map[int]bool
	loc            *time.Location
}

func Cron(spec string, loc *time.Location) (Trigger, error) {
	if loc == nil {
		loc = time.UTC
	}
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron spec must have 5 fields")
	}
	mins, err := parseCronField(fields[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hours, err := parseCronField(fields[1], 0, 23)
	if err != nil {
		return nil, err
	}
	// v0.1 supports minute/hour matching; day/month/week fields must be '*'.
	for _, f := range fields[2:] {
		if f != "*" {
			return nil, fmt.Errorf("unsupported cron field %q", f)
		}
	}
	return cronTrigger{minutes: mins, hours: hours, loc: loc}, nil
}
func parseCronField(field string, min, max int) (map[int]bool, error) {
	out := map[int]bool{}
	if field == "*" {
		for i := min; i <= max; i++ {
			out[i] = true
		}
		return out, nil
	}
	if strings.HasPrefix(field, "*/") {
		var step int
		if _, err := fmt.Sscanf(field, "*/%d", &step); err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid cron step %q", field)
		}
		for i := min; i <= max; i += step {
			out[i] = true
		}
		return out, nil
	}
	var v int
	if _, err := fmt.Sscanf(field, "%d", &v); err != nil || v < min || v > max {
		return nil, fmt.Errorf("invalid cron value %q", field)
	}
	out[v] = true
	return out, nil
}
func (t cronTrigger) Next(after time.Time) (time.Time, bool) {
	candidate := after.In(t.loc).Truncate(time.Minute).Add(time.Minute)
	limit := candidate.AddDate(0, 0, 366)
	for candidate.Before(limit) {
		if t.hours[candidate.Hour()] && t.minutes[candidate.Minute()] {
			return candidate, true
		}
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}, false
}

type MisfirePolicy int

const (
	MisfireRunOnce MisfirePolicy = iota
	MisfireSkip
)

type OverlapPolicy int

const (
	OverlapAllow OverlapPolicy = iota
	OverlapSkip
)

type JitterPolicy struct {
	Max  time.Duration
	Seed int64
}

func (p JitterPolicy) Delay(jobID string, scheduled time.Time) time.Duration {
	if p.Max <= 0 {
		return 0
	}
	r := rand.New(rand.NewSource(p.Seed + int64(len(jobID))*7919 + scheduled.UnixNano()))
	return time.Duration(r.Int63n(int64(p.Max) + 1))
}

type EventKind string

const (
	EventJobStarted   EventKind = "job_started"
	EventJobSucceeded EventKind = "job_succeeded"
	EventJobFailed    EventKind = "job_failed"
	EventJobSkipped   EventKind = "job_skipped"
)

type Event struct {
	Kind                    EventKind
	JobID                   string
	ScheduledAt, ObservedAt time.Time
	Err                     error
	Message                 string
}
type EventSink interface{ OnEvent(Event) }
type EventSinkFunc func(Event)

func (f EventSinkFunc) OnEvent(e Event) { f(e) }

type JobFunc func(context.Context) error

type Job struct {
	ID      string
	Trigger Trigger
	Handler JobFunc
	LockKey string
}

type Locker interface {
	Lock(context.Context, string, time.Duration) (Lease, error)
}
type Lease interface{ Release(context.Context) error }

type SchedulerOption func(*Scheduler)

func WithClock(c Clock) SchedulerOption {
	return func(s *Scheduler) {
		if c != nil {
			s.clock = c
		}
	}
}
func WithEventSink(sink EventSink) SchedulerOption      { return func(s *Scheduler) { s.sink = sink } }
func WithMisfirePolicy(p MisfirePolicy) SchedulerOption { return func(s *Scheduler) { s.misfire = p } }
func WithOverlapPolicy(p OverlapPolicy) SchedulerOption { return func(s *Scheduler) { s.overlap = p } }
func WithJitter(p JitterPolicy) SchedulerOption         { return func(s *Scheduler) { s.jitter = p } }
func WithLocker(l Locker, ttl time.Duration) SchedulerOption {
	return func(s *Scheduler) { s.locker = l; s.lockTTL = ttl }
}

type Scheduler struct {
	mu      sync.Mutex
	clock   Clock
	sink    EventSink
	misfire MisfirePolicy
	overlap OverlapPolicy
	jitter  JitterPolicy
	locker  Locker
	lockTTL time.Duration
	jobs    map[string]*jobState
	running bool
	closed  bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type jobState struct {
	job     Job
	next    time.Time
	running bool
	runs    int
}

type Snapshot struct {
	Running bool
	Jobs    []JobSnapshot
}
type JobSnapshot struct {
	ID      string
	NextRun time.Time
	Running bool
	Runs    int
}

func NewScheduler(opts ...SchedulerOption) *Scheduler {
	s := &Scheduler{clock: realClock{}, misfire: MisfireRunOnce, overlap: OverlapSkip, jobs: map[string]*jobState{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Scheduler) AddJob(job Job) error {
	if job.ID == "" {
		return errors.New("job id is required")
	}
	if job.Trigger == nil {
		return errors.New("job trigger is required")
	}
	if job.Handler == nil {
		return errors.New("job handler is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("scheduler is shut down")
	}
	if _, exists := s.jobs[job.ID]; exists {
		return fmt.Errorf("job %q already exists", job.ID)
	}
	next, ok := job.Trigger.Next(s.clock.Now().Add(-time.Nanosecond))
	if !ok {
		return errors.New("job trigger is exhausted")
	}
	s.jobs[job.ID] = &jobState{job: job, next: next}
	return nil
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("scheduler is shut down")
	}
	if s.running {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true
	s.wg.Add(1)
	go func() { defer s.wg.Done(); <-ctx.Done() }()
	return nil
}

func (s *Scheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) RunDue(ctx context.Context) int {
	now := s.clock.Now()
	var due []*jobState
	s.mu.Lock()
	for _, st := range s.jobs {
		if !st.next.IsZero() && !st.next.After(now) {
			due = append(due, st)
		}
	}
	s.mu.Unlock()
	sort.Slice(due, func(i, j int) bool { return due[i].job.ID < due[j].job.ID })
	run := 0
	for _, st := range due {
		if s.runOne(ctx, st, now) {
			run++
		}
	}
	return run
}

func (s *Scheduler) runOne(ctx context.Context, st *jobState, observed time.Time) bool {
	s.mu.Lock()
	if st.running && s.overlap == OverlapSkip {
		s.emitLocked(Event{Kind: EventJobSkipped, JobID: st.job.ID, ScheduledAt: st.next, ObservedAt: observed, Message: "overlap"})
		s.mu.Unlock()
		return false
	}
	scheduled := st.next
	if s.misfire == MisfireSkip && observed.After(scheduled) {
		s.advanceLocked(st, observed)
		s.emitLocked(Event{Kind: EventJobSkipped, JobID: st.job.ID, ScheduledAt: scheduled, ObservedAt: observed, Message: "misfire"})
		s.mu.Unlock()
		return false
	}
	st.running = true
	s.mu.Unlock()

	if delay := s.jitter.Delay(st.job.ID, scheduled); delay > 0 {
		scheduled = scheduled.Add(delay)
	}
	var lease Lease
	var err error
	if s.locker != nil && st.job.LockKey != "" {
		lease, err = s.locker.Lock(ctx, st.job.LockKey, s.lockTTL)
		if err != nil {
			s.finish(st, scheduled, observed, err, false)
			return false
		}
		defer lease.Release(ctx)
	}
	s.emit(Event{Kind: EventJobStarted, JobID: st.job.ID, ScheduledAt: scheduled, ObservedAt: observed})
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		err = st.job.Handler(ctx)
	}()
	s.finish(st, scheduled, observed, err, true)
	return err == nil
}

func (s *Scheduler) finish(st *jobState, scheduled, observed time.Time, err error, counted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st.running = false
	if counted {
		st.runs++
	}
	s.advanceLocked(st, observed)
	kind := EventJobSucceeded
	if err != nil {
		kind = EventJobFailed
	}
	s.emitLocked(Event{Kind: kind, JobID: st.job.ID, ScheduledAt: scheduled, ObservedAt: observed, Err: err})
}
func (s *Scheduler) advanceLocked(st *jobState, after time.Time) {
	if next, ok := st.job.Trigger.Next(after); ok {
		st.next = next
	} else {
		st.next = time.Time{}
	}
}
func (s *Scheduler) emit(e Event) { s.mu.Lock(); defer s.mu.Unlock(); s.emitLocked(e) }
func (s *Scheduler) emitLocked(e Event) {
	if s.sink != nil {
		s.sink.OnEvent(e)
	}
}

func (s *Scheduler) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := Snapshot{Running: s.running && !s.closed, Jobs: make([]JobSnapshot, 0, len(s.jobs))}
	for _, st := range s.jobs {
		out.Jobs = append(out.Jobs, JobSnapshot{ID: st.job.ID, NextRun: st.next, Running: st.running, Runs: st.runs})
	}
	sort.Slice(out.Jobs, func(i, j int) bool { return out.Jobs[i].ID < out.Jobs[j].ID })
	return out
}
