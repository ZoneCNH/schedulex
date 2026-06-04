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

type Trigger interface {
	Next(after time.Time) (time.Time, bool)
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

type EventType string

const (
	EventJobStarted   EventKind = "job_started"
	EventJobSucceeded EventKind = "job_succeeded"
	EventJobFailed    EventKind = "job_failed"
	EventJobSkipped   EventKind = "job_skipped"
)

type Event struct {
	Type        EventType `json:"type"`
	JobID       string    `json:"job_id"`
	At          time.Time `json:"at"`
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type EventSink interface{ Emit(Event) }
type EventSinkFunc func(Event)

func (f EventSinkFunc) Emit(e Event) { f(e) }

type Lease interface{ Release(context.Context) error }
type Locker interface {
	Lock(context.Context, string) (Lease, error)
}

type Job struct {
	ID           string
	Trigger      Trigger
	Run          JobFunc
	Misfire      MisfirePolicy
	MisfireGrace time.Duration
	Overlap      OverlapPolicy
	Jitter       JitterPolicy
	Locker       Locker
	LockKey      string
}

type Options struct {
	Clock         Clock
	EventSink     EventSink
	MaxConcurrent int
}

type Snapshot struct {
	Version  string `json:"version"`
	Started  bool   `json:"started"`
	Closed   bool   `json:"closed"`
	JobCount int    `json:"job_count"`
}
