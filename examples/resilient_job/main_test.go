package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestExampleCompiles(t *testing.T) { main() }

func TestResilientJobName(t *testing.T) {
	r := &ResilientJob{Name_: "test-job"}
	if r.Name() != "test-job" {
		t.Fatalf("Name() = %q, want %q", r.Name(), "test-job")
	}
}

func TestResilientJobRunSuccess(t *testing.T) {
	r := &ResilientJob{
		Name_:   "ok",
		Execute: func(ctx context.Context) error { return nil },
		Retry:   RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Second},
		Timeout: TimeoutPolicy{Timeout: time.Second},
		Breaker: CircuitBreakerPolicy{Threshold: 3, ResetAfter: time.Second},
	}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
}

func TestResilientJobRunRetryThenSuccess(t *testing.T) {
	var calls atomic.Int32
	r := &ResilientJob{
		Name_: "retry-ok",
		Execute: func(ctx context.Context) error {
			if calls.Add(1) <= 2 {
				return errors.New("transient")
			}
			return nil
		},
		Retry:   RetryPolicy{MaxAttempts: 5, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond},
		Timeout: TimeoutPolicy{Timeout: time.Second},
		Breaker: CircuitBreakerPolicy{Threshold: 10, ResetAfter: time.Second},
	}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", calls.Load())
	}
}

func TestResilientJobRunAllRetriesFail(t *testing.T) {
	r := &ResilientJob{
		Name_: "fail",
		Execute: func(ctx context.Context) error {
			return errors.New("permanent")
		},
		Retry:   RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond},
		Timeout: TimeoutPolicy{Timeout: time.Second},
		Breaker: CircuitBreakerPolicy{Threshold: 10, ResetAfter: time.Second},
	}
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from all retries failing")
	}
}

func TestResilientJobTimeout(t *testing.T) {
	r := &ResilientJob{
		Name_: "timeout",
		Execute: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
		Retry:   RetryPolicy{MaxAttempts: 1},
		Timeout: TimeoutPolicy{Timeout: 50 * time.Millisecond},
		Breaker: CircuitBreakerPolicy{Threshold: 10, ResetAfter: time.Second},
	}
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestResilientJobNoTimeout(t *testing.T) {
	r := &ResilientJob{
		Name_:   "no-timeout",
		Execute: func(ctx context.Context) error { return nil },
		Retry:   RetryPolicy{MaxAttempts: 1},
		Timeout: TimeoutPolicy{Timeout: 0},
		Breaker: CircuitBreakerPolicy{Threshold: 3, ResetAfter: time.Second},
	}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
}

func TestResilientJobCircuitBreaker(t *testing.T) {
	var calls atomic.Int32
	r := &ResilientJob{
		Name_: "breaker",
		Execute: func(ctx context.Context) error {
			calls.Add(1)
			return errors.New("fail")
		},
		Retry:   RetryPolicy{MaxAttempts: 1},
		Timeout: TimeoutPolicy{Timeout: 0},
		Breaker: CircuitBreakerPolicy{Threshold: 2, ResetAfter: time.Hour},
	}

	// Two failures should open the breaker.
	for i := 0; i < 2; i++ {
		_ = r.Run(context.Background())
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}

	// Third call should be rejected by circuit breaker.
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected circuit breaker error")
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d after breaker open, want 2", calls.Load())
	}
}

func TestResilientJobCircuitBreakerHalfOpen(t *testing.T) {
	var calls atomic.Int32
	r := &ResilientJob{
		Name_: "half-open",
		Execute: func(ctx context.Context) error {
			if calls.Add(1) == 1 {
				return errors.New("fail")
			}
			return nil
		},
		Retry:   RetryPolicy{MaxAttempts: 1},
		Timeout: TimeoutPolicy{Timeout: 0},
		Breaker: CircuitBreakerPolicy{Threshold: 1, ResetAfter: 50 * time.Millisecond},
	}

	// First run fails → opens breaker.
	_ = r.Run(context.Background())
	if !r.circuitOpen.Load() {
		t.Fatal("breaker should be open after failure")
	}

	// Wait for reset window to expire.
	time.Sleep(100 * time.Millisecond)

	// Should enter half-open state, reset breaker, and succeed.
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() in half-open: %v", err)
	}
	if r.circuitOpen.Load() {
		t.Fatal("breaker should be closed after half-open success")
	}
}

func TestResilientJobContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := &ResilientJob{
		Name_: "ctx-cancel",
		Execute: func(ctx context.Context) error {
			return ctx.Err()
		},
		Retry:   RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
		Timeout: TimeoutPolicy{Timeout: 0},
		Breaker: CircuitBreakerPolicy{Threshold: 10, ResetAfter: time.Second},
	}
	err := r.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRetryDelay(t *testing.T) {
	r := &ResilientJob{
		Retry: RetryPolicy{BaseDelay: 100 * time.Millisecond, MaxDelay: 500 * time.Millisecond},
	}

	if d := r.retryDelay(1); d != 100*time.Millisecond {
		t.Fatalf("retryDelay(1) = %v, want 100ms", d)
	}
	if d := r.retryDelay(2); d != 200*time.Millisecond {
		t.Fatalf("retryDelay(2) = %v, want 200ms", d)
	}
	if d := r.retryDelay(3); d != 400*time.Millisecond {
		t.Fatalf("retryDelay(3) = %v, want 400ms", d)
	}
	// Capped at MaxDelay.
	if d := r.retryDelay(4); d != 500*time.Millisecond {
		t.Fatalf("retryDelay(4) = %v, want 500ms", d)
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.MaxAttempts != 3 {
		t.Fatalf("MaxAttempts = %d, want 3", p.MaxAttempts)
	}
	if p.BaseDelay != 100*time.Millisecond {
		t.Fatalf("BaseDelay = %v, want 100ms", p.BaseDelay)
	}
	if p.MaxDelay != 5*time.Second {
		t.Fatalf("MaxDelay = %v, want 5s", p.MaxDelay)
	}
}

func TestResilientJobRetryContextCanceled(t *testing.T) {
	// Cancel context during retry delay to cover the select/ctx.Done path.
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	r := &ResilientJob{
		Name_: "ctx-retry",
		Execute: func(ctx context.Context) error {
			if calls.Add(1) == 1 {
				return errors.New("transient")
			}
			return nil
		},
		Retry:   RetryPolicy{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second},
		Timeout: TimeoutPolicy{Timeout: 0},
		Breaker: CircuitBreakerPolicy{Threshold: 10, ResetAfter: time.Second},
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := r.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestMainErrorPath(t *testing.T) {
	// Swap newResilientJob to return a failing job, then call main()
	// to exercise the error branch (log.Printf("job failed: ...")).
	orig := newResilientJob
	newResilientJob = func() *ResilientJob {
		job := orig()
		job.Execute = func(ctx context.Context) error {
			return errors.New("injected failure")
		}
		job.Retry = RetryPolicy{MaxAttempts: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
		return job
	}
	defer func() { newResilientJob = orig }()

	main()
}

func TestResilientJobRetryWithZeroAttempts(t *testing.T) {
	var calls atomic.Int32
	r := &ResilientJob{
		Name_: "zero-attempts",
		Execute: func(ctx context.Context) error {
			calls.Add(1)
			return nil
		},
		Retry:   RetryPolicy{MaxAttempts: 0},
		Timeout: TimeoutPolicy{Timeout: 0},
		Breaker: CircuitBreakerPolicy{Threshold: 3, ResetAfter: time.Second},
	}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}
