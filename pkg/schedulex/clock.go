package schedulex

import "time"

// Clock supplies time to deterministic scheduling decisions.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

// SystemClock returns a Clock backed by time.Now for edge adapters.
func SystemClock() Clock { return systemClock{} }

func (systemClock) Now() time.Time { return time.Now() }

// StaticClock is a test/replay clock with explicit advancement.
type StaticClock struct {
	now time.Time
}

// NewStaticClock constructs a replay clock pinned to t.
func NewStaticClock(t time.Time) *StaticClock { return &StaticClock{now: t} }

// Now returns the clock's current instant.
func (c *StaticClock) Now() time.Time { return c.now }

// Advance moves the clock forward by d.
func (c *StaticClock) Advance(d time.Duration) { c.now = c.now.Add(d) }
