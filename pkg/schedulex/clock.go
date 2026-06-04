package schedulex

import (
	"sync"
	"time"
)

// Clock supplies time to deterministic scheduling decisions.
type Clock interface {
	Now() time.Time
	After(time.Duration) <-chan time.Time
}

type realClock struct{}

// NewRealClock returns a Clock backed by the Go standard library timer.
func NewRealClock() Clock { return realClock{} }

// SystemClock returns a Clock backed by the Go standard library timer.
func SystemClock() Clock { return NewRealClock() }

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// StaticClock is a deterministic clock for snapshot and contract tests.
type StaticClock struct {
	mu      sync.Mutex
	now     time.Time
	waiters []staticClockWaiter
}

type staticClockWaiter struct {
	at time.Time
	ch chan time.Time
}

// NewStaticClock creates a deterministic clock fixed at the supplied instant.
func NewStaticClock(now time.Time) *StaticClock {
	return &StaticClock{now: now}
}

// Now returns the current deterministic instant.
func (c *StaticClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// After returns a channel that receives the target instant once the clock reaches it.
func (c *StaticClock) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	c.mu.Lock()
	target := c.now.Add(d)
	if !target.After(c.now) {
		c.mu.Unlock()
		ch <- target
		return ch
	}
	c.waiters = append(c.waiters, staticClockWaiter{at: target, ch: ch})
	c.mu.Unlock()
	return ch
}

// Set moves the deterministic clock to an exact instant.
func (c *StaticClock) Set(now time.Time) {
	c.mu.Lock()
	c.now = now
	ready := c.readyWaitersLocked()
	c.mu.Unlock()
	deliverStaticClockWaiters(ready)
}

// Advance moves the deterministic clock forward by d.
func (c *StaticClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	ready := c.readyWaitersLocked()
	c.mu.Unlock()
	deliverStaticClockWaiters(ready)
}

func (c *StaticClock) readyWaitersLocked() []staticClockWaiter {
	var ready []staticClockWaiter
	var pending []staticClockWaiter
	for _, waiter := range c.waiters {
		if waiter.at.After(c.now) {
			pending = append(pending, waiter)
			continue
		}
		ready = append(ready, waiter)
	}
	c.waiters = pending
	return ready
}

func deliverStaticClockWaiters(waiters []staticClockWaiter) {
	for _, waiter := range waiters {
		waiter.ch <- waiter.at
	}
}
