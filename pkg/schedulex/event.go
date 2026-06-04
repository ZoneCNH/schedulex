package schedulex

import (
	"context"
	"time"
)

// EventType names scheduler lifecycle and job events.
type EventType string

const (
	EventScheduled   EventType = "scheduled"
	EventStarted     EventType = "started"
	EventSucceeded   EventType = "succeeded"
	EventFailed      EventType = "failed"
	EventSkipped     EventType = "skipped"
	EventShutdown    EventType = "shutdown"
	EventMisfire     EventType = "misfire"
	EventLockSkipped EventType = "lock_skipped"
	EventLockFailed  EventType = "lock_failed"
)

// Event is emitted to EventSink adapters.
type Event struct {
	Type        EventType         `json:"type"`
	JobID       string            `json:"job_id,omitempty"`
	JobName     string            `json:"job_name,omitempty"`
	At          time.Time         `json:"at"`
	ScheduledAt time.Time         `json:"scheduled_at,omitempty"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	FinishedAt  time.Time         `json:"finished_at,omitempty"`
	Lag         time.Duration     `json:"lag,omitempty"`
	Duration    time.Duration     `json:"duration,omitempty"`
	Attempt     int               `json:"attempt,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	Err         string            `json:"err,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// EventSink receives scheduler events.
type EventSink interface {
	OnEvent(context.Context, Event)
}

// EventSinkFunc adapts a function into an EventSink.
type EventSinkFunc func(context.Context, Event)

// OnEvent calls f when f is not nil.
func (f EventSinkFunc) OnEvent(ctx context.Context, event Event) {
	if f != nil {
		f(ctx, event)
	}
}
