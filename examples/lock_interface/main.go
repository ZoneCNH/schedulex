package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

type memLocker struct{ mu sync.Mutex; held bool }
type lease struct{ l *memLocker }
func (m *memLocker) Lock(context.Context, string) (schedulex.Lease, error) { m.mu.Lock(); defer m.mu.Unlock(); if m.held { return nil, schedulex.ErrLockUnavailable }; m.held = true; return lease{m}, nil }
func (l lease) Release(context.Context) error { l.l.mu.Lock(); defer l.l.mu.Unlock(); l.l.held = false; return nil }

func main() {
	_, err := (&memLocker{}).Lock(context.Background(), "job")
	fmt.Println(err == nil)
}
