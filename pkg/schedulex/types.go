package schedulex

import (
	"context"
	"errors"
	"time"
)

const Version = "v0.1.0"

var (
	ErrSchedulerClosed = errors.New("schedulex: scheduler closed")
	ErrJobExists       = errors.New("schedulex: job already exists")
	ErrInvalidJob      = errors.New("schedulex: invalid job")
	ErrLockUnavailable = errors.New("schedulex: lock unavailable")
)

type JobFunc func(context.Context) error

type Trigger interface { Next(after time.Time) (time.Time, bool) }

type MisfirePolicy int
const (
	MisfireSkip MisfirePolicy = iota
	MisfireRunOnce
	MisfireCatchUp
)

type OverlapPolicy int
const (
	OverlapAllow OverlapPolicy = iota
	OverlapSkip
	OverlapForbid
)

type JitterPolicy struct { Max time.Duration; Seed int64 }

type EventType string
const (
	EventScheduled EventType = "scheduled"
	EventStarted   EventType = "started"
	EventSucceeded EventType = "succeeded"
	EventFailed    EventType = "failed"
	EventSkipped   EventType = "skipped"
)

type Event struct {
	Type EventType `json:"type"`
	JobID string `json:"job_id"`
	At time.Time `json:"at"`
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	Error string `json:"error,omitempty"`
}

type EventSink interface { Emit(Event) }
type EventSinkFunc func(Event)
func (f EventSinkFunc) Emit(e Event) { f(e) }

type Lease interface { Release(context.Context) error }
type Locker interface { Lock(context.Context, string) (Lease, error) }

type Job struct {
	ID string
	Trigger Trigger
	Run JobFunc
	Misfire MisfirePolicy
	MisfireGrace time.Duration
	Overlap OverlapPolicy
	Jitter JitterPolicy
	Locker Locker
	LockKey string
}

type Options struct {
	Clock Clock
	EventSink EventSink
	MaxConcurrent int
}

type Snapshot struct {
	Version string `json:"version"`
	Started bool `json:"started"`
	Closed bool `json:"closed"`
	JobCount int `json:"job_count"`
}
