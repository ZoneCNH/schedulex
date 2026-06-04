package schedulex

import "context"

// JobFunc is the work executed for a scheduled job.
type JobFunc func(context.Context) error

// Job describes a schedulable unit and its deterministic trigger.
type Job struct {
	ID            string
	Trigger       Trigger
	Run           JobFunc
	MisfirePolicy MisfirePolicy
	OverlapPolicy OverlapPolicy
	Locker        Locker
	EventSink     EventSink
}

func normalizeJob(job Job) Job {
	if job.MisfirePolicy == "" {
		job.MisfirePolicy = MisfireSkip
	}
	if job.OverlapPolicy == "" {
		job.OverlapPolicy = OverlapSkip
	}
	return job
}
