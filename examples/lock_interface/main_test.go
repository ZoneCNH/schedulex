package main

import (
	"context"
	"testing"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func TestExampleCompiles(t *testing.T) { main() }

func TestTryLockAndRelease(t *testing.T) {
	l := &memLocker{}
	ctx := context.Background()

	// First lock should succeed.
	lv, err := l.TryLock(ctx, "job", time.Minute)
	if err != nil {
		t.Fatalf("first TryLock: %v", err)
	}
	if lv == nil {
		t.Fatal("expected non-nil lease")
	}

	// Second lock while held should fail.
	_, err = l.TryLock(ctx, "job", time.Minute)
	if err != schedulex.ErrLockUnavailable {
		t.Fatalf("expected ErrLockUnavailable, got %v", err)
	}

	// Release the lease.
	if err := lv.Release(ctx); err != nil {
		t.Fatalf("Release: %v", err)
	}

	// After release, lock should succeed again.
	lv2, err := l.TryLock(ctx, "job", time.Minute)
	if err != nil {
		t.Fatalf("TryLock after release: %v", err)
	}
	if err := lv2.Release(ctx); err != nil {
		t.Fatalf("Release2: %v", err)
	}
}
