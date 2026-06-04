package schedulex

import (
	"hash/fnv"
	"time"
)

// JitterPolicy configures deterministic per-job jitter.
type JitterPolicy struct {
	Max  time.Duration
	Seed int64
}

func ApplyDeterministicJitter(base time.Time, p JitterPolicy, jobID string, run int64) time.Time {
	if p.Max <= 0 {
		return base
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(jobID))
	var b [16]byte
	u := uint64(p.Seed) ^ uint64(run)
	for i := range b {
		b[i] = byte(u >> (uint(i%8) * 8))
	}
	_, _ = h.Write(b[:])
	delta := time.Duration(h.Sum64() % uint64(p.Max+1))
	return base.Add(delta)
}

type MisfireDecision struct {
	Runs    []time.Time
	Skipped []time.Time
	Next    time.Time
	HasNext bool
}

func PlanMisfire(policy MisfirePolicy, missed []time.Time, next time.Time, hasNext bool) MisfireDecision {
	d := MisfireDecision{Next: next, HasNext: hasNext}
	if len(missed) == 0 {
		return d
	}
	switch policy {
	case MisfireRunOnce:
		d.Runs = []time.Time{missed[len(missed)-1]}
		d.Skipped = append(d.Skipped, missed[:len(missed)-1]...)
	case MisfireCatchUp:
		d.Runs = append(d.Runs, missed...)
	default:
		d.Skipped = append(d.Skipped, missed...)
	}
	return d
}
