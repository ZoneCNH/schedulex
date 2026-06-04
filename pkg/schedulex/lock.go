package schedulex

import (
	"context"
	"time"
)

// Locker is the distributed-lock extension point. v0.1 ships no Redis/Postgres implementation.
type Locker interface {
	TryLock(ctx context.Context, key string, ttl time.Duration) (Lease, error)
}

// Lease represents an acquired lock that must be released by the adapter.
type Lease interface {
	Release(ctx context.Context) error
}
