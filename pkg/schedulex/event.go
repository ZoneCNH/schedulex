package schedulex

import "time"

// EventType names scheduler lifecycle and job events.
type EventType string

const (
	EventJobAdded EventType = "job_added"
	EventStarted  EventType = "started"
	EventShutdown EventType = "shutdown"
	EventMisfire  EventType = "misfire"
)

// Event is emitted to EventSink adapters.
type Event struct {
	Type  EventType `json:"type"`
	JobID string    `json:"job_id,omitempty"`
	At    time.Time `json:"at"`
}

// EventSink receives scheduler events.
type EventSink interface {
	Emit(Event)
}
