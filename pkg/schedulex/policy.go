package schedulex

// MisfirePolicy describes how missed trigger times are reconciled.
type MisfirePolicy string

const (
	MisfireSkip    MisfirePolicy = "skip"
	MisfireRunOnce MisfirePolicy = "run_once"
	MisfireCatchUp MisfirePolicy = "catch_up"
)

// OverlapPolicy describes behavior when a job fires while still running.
type OverlapPolicy string

const (
	OverlapSkip     OverlapPolicy = "skip"
	OverlapQueueOne OverlapPolicy = "queue_one"
	OverlapAllow    OverlapPolicy = "allow"
)
