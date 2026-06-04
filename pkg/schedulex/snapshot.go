package schedulex

import (
	"sort"
	"time"
)

// Snapshot captures scheduler state for observability and release evidence.
type Snapshot struct {
	Now      time.Time
	Running  bool
	Shutdown bool
	Jobs     []JobSnapshot
}

// JobSnapshot captures one job's deterministic schedule state.
type JobSnapshot struct {
	ID            string
	Next          time.Time
	HasNext       bool
	MisfirePolicy MisfirePolicy
	OverlapPolicy OverlapPolicy
}

func (s *Snapshot) sort() {
	sort.Slice(s.Jobs, func(i, j int) bool { return s.Jobs[i].ID < s.Jobs[j].ID })
}
