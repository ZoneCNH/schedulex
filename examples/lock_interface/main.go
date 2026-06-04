package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

type noopLocker struct{}
type noopLease struct{}

func (noopLocker) Acquire(context.Context, string, time.Duration) (schedulex.Lease, error) { return noopLease{}, nil }
func (noopLease) Release(context.Context) error { return nil }

func main() {
	lease, _ := noopLocker{}.Acquire(context.Background(), "job", time.Second)
	fmt.Println(lease.Release(context.Background()) == nil)
}
