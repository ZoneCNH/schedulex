package schedulex

import "time"

// Clock supplies time to the scheduler. Scheduler decisions use this interface
// so deterministic tests can avoid wall-clock time.
type Clock interface {
	Now() time.Time
	After(time.Duration) <-chan time.Time
}

type realClock struct{}

// NewRealClock returns the production wall-clock adapter. Wall-clock access is
// isolated here; scheduling decisions use Clock.
func NewRealClock() Clock { return realClock{} }

func (realClock) Now() time.Time                 { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
