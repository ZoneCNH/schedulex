package schedulex

import (
	"sort"
	"time"
)

// Snapshot captures scheduler state for observability and release evidence.
type Snapshot struct {
	Version  string
	Now      time.Time
	Started  bool
	Running  bool
	Closed   bool
	Shutdown bool
	JobCount int
	Jobs     []JobSnapshot
}

// JobSnapshot captures one job's deterministic schedule state.
type JobSnapshot struct {
	ID            string
	Name          string
	Next          time.Time
	HasNext       bool
	MisfirePolicy MisfirePolicy
	OverlapPolicy OverlapPolicy
	Running       bool
	Queued        bool
}

func (s *Snapshot) sort() {
	sort.Slice(s.Jobs, func(i, j int) bool { return s.Jobs[i].ID < s.Jobs[j].ID })
}
