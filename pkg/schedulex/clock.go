package schedulex

import "time"

// Clock supplies time to deterministic scheduling decisions.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

// SystemClock returns a Clock backed by time.Now for edge adapters.
func SystemClock() Clock { return systemClock{} }

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
