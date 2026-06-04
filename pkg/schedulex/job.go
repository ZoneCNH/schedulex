package schedulex

import (
	"context"
	"time"
)

// Job describes a schedulable unit.
type Job interface {
	Name() string
	Run(context.Context) error
}

// JobFunc adapts a function into a Job.
type JobFunc struct {
	NameValue string
	RunFunc   func(context.Context) error
}

// Name returns the stable job name used in snapshots and lock keys.
func (j JobFunc) Name() string { return j.NameValue }

// Run executes the configured function.
func (j JobFunc) Run(ctx context.Context) error {
	if j.RunFunc == nil {
		return nil
	}
	return j.RunFunc(ctx)
}

type jobConfig struct {
	id            string
	job           Job
	trigger       Trigger
	misfirePolicy MisfirePolicy
	overlapPolicy OverlapPolicy
	jitterPolicy  JitterPolicy
	locker        Locker
	lockKey       string
	lockTTL       time.Duration
	eventSink     EventSink
}

// JobOption configures one registered job.
type JobOption func(*jobConfig) error

// WithMisfirePolicy sets how missed fire times are reconciled.
func WithMisfirePolicy(policy MisfirePolicy) JobOption {
	return func(cfg *jobConfig) error {
		if policy == "" {
			policy = MisfireSkip
		}
		if !policy.valid() {
			return ErrInvalidOption
		}
		cfg.misfirePolicy = policy
		return nil
	}
}

// WithOverlapPolicy sets how concurrent firings of the same job are handled.
func WithOverlapPolicy(policy OverlapPolicy) JobOption {
	return func(cfg *jobConfig) error {
		if policy == "" {
			policy = OverlapSkip
		}
		if !policy.valid() {
			return ErrInvalidOption
		}
		cfg.overlapPolicy = policy
		return nil
	}
}

// WithJitter applies deterministic per-job jitter.
func WithJitter(policy JitterPolicy) JobOption {
	return func(cfg *jobConfig) error {
		if policy.Max < 0 {
			return ErrInvalidOption
		}
		cfg.jitterPolicy = policy
		return nil
	}
}

// WithLocker sets the distributed-lock extension point for a job.
func WithLocker(locker Locker) JobOption {
	return func(cfg *jobConfig) error {
		cfg.locker = locker
		return nil
	}
}

// WithLockKey sets the distributed lock key. The job name is used by default.
func WithLockKey(key string) JobOption {
	return func(cfg *jobConfig) error {
		cfg.lockKey = key
		return nil
	}
}

// WithLockTTL sets the requested lock TTL. A conservative default is used when unset.
func WithLockTTL(ttl time.Duration) JobOption {
	return func(cfg *jobConfig) error {
		if ttl < 0 {
			return ErrInvalidOption
		}
		cfg.lockTTL = ttl
		return nil
	}
}

// WithJobEventSink overrides the scheduler event sink for one job.
func WithJobEventSink(sink EventSink) JobOption {
	return func(cfg *jobConfig) error {
		cfg.eventSink = sink
		return nil
	}
}
