package schedulex

import (
	"strconv"
	"time"
)

// MisfireDecision 描述 misfire 后的调度决策。
type MisfireDecision struct {
	Runs    []time.Time
	Skipped []time.Time
	Next    time.Time
	HasNext bool
}

// PlanMisfire 根据 misfire 策略决定哪些错过的时间点需要执行。
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

// ReconcileMisfire calculates due run instants for release golden cases.
func ReconcileMisfire(policy MisfirePolicy, missed []time.Time) []time.Time {
	if len(missed) == 0 || policy == MisfireSkip {
		return nil
	}
	if policy == MisfireRunOnce {
		return []time.Time{missed[len(missed)-1]}
	}
	return append([]time.Time(nil), missed...)
}

func (s *Scheduler) shouldReconcileMisfire(_ *jobState, _, scheduled, now time.Time) bool {
	return now.After(scheduled.Add(misfireGrace))
}

func (s *Scheduler) collectMissed(state *jobState, first, now time.Time) ([]time.Time, bool) {
	missed := make([]time.Time, 0, 4)
	for due := first; !due.After(now); {
		missed = append(missed, due)
		if len(missed) >= maxMisfireCatchUp {
			return missed, true
		}
		next, ok := state.cfg.trigger.Next(due)
		if !ok {
			return missed, false
		}
		due = next
	}
	return missed, false
}

func misfireAttributes(missed []time.Time, decision MisfireDecision, capped bool) map[string]string {
	attributes := map[string]string{
		"missed":  strconv.Itoa(len(missed)),
		"runs":    strconv.Itoa(len(decision.Runs)),
		"skipped": strconv.Itoa(len(decision.Skipped)),
	}
	if capped {
		attributes["capped"] = "true"
	}
	if len(missed) > 0 {
		attributes["first_missed"] = missed[0].Format(time.RFC3339Nano)
		attributes["last_missed"] = missed[len(missed)-1].Format(time.RFC3339Nano)
	}
	return attributes
}
