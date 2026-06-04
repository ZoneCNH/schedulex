package schedulex

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// loop 是单个任务的主调度循环。
func (s *Scheduler) loop(state *jobState) {
	defer s.wg.Done()
	after := s.clock.Now()
	for {
		next, ok := state.cfg.trigger.Next(after)
		if !ok {
			return
		}
		attempt := int(state.attempt.Add(1))
		scheduled := ApplyDeterministicJitter(next, state.cfg.jitterPolicy, state.cfg.id, int64(attempt))
		s.emit(s.ctx, s.event(state, EventScheduled, scheduled, withAttempt(attempt)))
		wait := scheduled.Sub(s.clock.Now())
		if wait < 0 {
			wait = 0
		}
		select {
		case <-s.ctx.Done():
			return
		case <-s.clock.After(wait):
		}

		now := s.clock.Now()
		if s.shouldReconcileMisfire(state, next, scheduled, now) {
			missed, capped := s.collectMissed(state, next, now)
			if len(missed) == 0 {
				after = next
				continue
			}
			decision := PlanMisfire(state.cfg.misfirePolicy, missed, time.Time{}, false)
			s.emit(s.ctx, s.event(
				state,
				EventMisfire,
				scheduled,
				withAttempt(attempt),
				withReason(string(state.cfg.misfirePolicy)),
				withAttributes(misfireAttributes(missed, decision, capped)),
			))
			for _, missedAt := range decision.Runs {
				runAttempt := attempt
				if !missedAt.Equal(next) {
					runAttempt = int(state.attempt.Add(1))
				}
				runScheduled := ApplyDeterministicJitter(missedAt, state.cfg.jitterPolicy, state.cfg.id, int64(runAttempt))
				s.dispatchReady(state, runScheduled, runAttempt)
			}
			after = missed[len(missed)-1]
			continue
		}

		s.dispatch(state, scheduled, attempt)
		after = next
	}
}

func (s *Scheduler) dispatch(state *jobState, scheduled time.Time, attempt int) {
	s.dispatchRun(state, scheduled, attempt, true)
}

func (s *Scheduler) dispatchReady(state *jobState, scheduled time.Time, attempt int) {
	s.dispatchRun(state, scheduled, attempt, false)
}

func (s *Scheduler) dispatchRun(state *jobState, scheduled time.Time, attempt int, reconcileMisfire bool) {
	runnable, reason := s.reserveOverlap(state, scheduled, attempt)
	if !runnable {
		if reason != "queued" {
			s.emit(s.ctx, s.event(state, EventSkipped, scheduled, withAttempt(attempt), withReason(reason)))
		}
		return
	}

	releaseSlot, ok := s.acquireSlot()
	if !ok {
		return
	}

	if reconcileMisfire {
		var shouldRun bool
		scheduled, shouldRun = s.reconcilePendingMisfire(state, scheduled, attempt)
		if !shouldRun {
			releaseSlot()
			return
		}
	}

	runnable, reason = s.markRunStarted(state, scheduled, attempt)
	if !runnable {
		releaseSlot()
		if reason != "queued" {
			s.emit(s.ctx, s.event(state, EventSkipped, scheduled, withAttempt(attempt), withReason(reason)))
		}
		return
	}

	s.startRun(state, scheduled, attempt, releaseSlot)
}

func (s *Scheduler) reserveOverlap(state *jobState, scheduled time.Time, attempt int) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	busy := state.running > 0 || state.queuedDispatching
	if !busy || state.cfg.overlapPolicy == OverlapAllow {
		return true, ""
	}
	if state.cfg.overlapPolicy == OverlapQueueOne {
		if state.queuedDispatching {
			return false, "overlap_queue_full"
		}
		if !state.queued {
			state.queued = true
			state.queuedScheduled = scheduled
			state.queuedAttempt = attempt
			return false, "queued"
		}
		return false, "overlap_queue_full"
	}
	return false, "overlap"
}

func (s *Scheduler) acquireSlot() (func(), bool) {
	select {
	case s.sem <- struct{}{}:
		return func() { <-s.sem }, true
	case <-s.ctx.Done():
		return nil, false
	}
}

func (s *Scheduler) reconcilePendingMisfire(state *jobState, scheduled time.Time, attempt int) (time.Time, bool) {
	now := s.clock.Now()
	if !s.shouldReconcileMisfire(state, scheduled, scheduled, now) {
		return scheduled, true
	}
	decision := PlanMisfire(state.cfg.misfirePolicy, []time.Time{scheduled}, time.Time{}, false)
	s.emit(s.ctx, s.event(
		state,
		EventMisfire,
		scheduled,
		withAttempt(attempt),
		withReason(string(state.cfg.misfirePolicy)),
		withAttributes(misfireAttributes([]time.Time{scheduled}, decision, false)),
	))
	if len(decision.Runs) == 0 {
		return time.Time{}, false
	}
	return decision.Runs[len(decision.Runs)-1], true
}

func (s *Scheduler) markRunStarted(state *jobState, scheduled time.Time, attempt int) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if state.cfg.overlapPolicy == OverlapAllow {
		state.running++
		return true, ""
	}
	if state.running == 0 && !state.queuedDispatching {
		state.running = 1
		return true, ""
	}
	if state.cfg.overlapPolicy == OverlapQueueOne {
		if state.queuedDispatching {
			return false, "overlap_queue_full"
		}
		if !state.queued {
			state.queued = true
			state.queuedScheduled = scheduled
			state.queuedAttempt = attempt
			return false, "queued"
		}
		return false, "overlap_queue_full"
	}
	return false, "overlap"
}

func (s *Scheduler) startRun(state *jobState, scheduled time.Time, attempt int, releaseSlot func()) {
	s.wg.Add(1)
	go s.run(state, scheduled, attempt, releaseSlot)
}

func (s *Scheduler) finishRun(state *jobState) {
	var queuedScheduled time.Time
	var queuedAttempt int
	var startQueued bool

	s.mu.Lock()
	if state.running > 0 {
		state.running--
	}
	canStartQueued := !s.closed && s.ctx != nil && s.ctx.Err() == nil
	if state.running == 0 && state.queued && canStartQueued {
		queuedScheduled = state.queuedScheduled
		queuedAttempt = state.queuedAttempt
		state.queued = false
		state.queuedDispatching = true
		state.queuedScheduled = time.Time{}
		state.queuedAttempt = 0
		startQueued = true
	} else if state.running == 0 && !canStartQueued {
		state.queued = false
		state.queuedDispatching = false
		state.queuedScheduled = time.Time{}
		state.queuedAttempt = 0
	}
	s.mu.Unlock()

	if startQueued {
		s.dispatchQueued(state, queuedScheduled, queuedAttempt)
	}
}

func (s *Scheduler) dispatchQueued(state *jobState, scheduled time.Time, attempt int) {
	releaseSlot, ok := s.acquireSlot()
	if !ok {
		s.clearQueuedDispatching(state)
		return
	}

	var shouldRun bool
	scheduled, shouldRun = s.reconcilePendingMisfire(state, scheduled, attempt)
	if !shouldRun {
		releaseSlot()
		s.clearQueuedDispatching(state)
		return
	}

	if !s.markQueuedRunStarted(state) {
		releaseSlot()
		return
	}
	s.startRun(state, scheduled, attempt, releaseSlot)
}

func (s *Scheduler) clearQueuedDispatching(state *jobState) {
	s.mu.Lock()
	state.queuedDispatching = false
	s.mu.Unlock()
}

func (s *Scheduler) markQueuedRunStarted(state *jobState) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !state.queuedDispatching || s.closed || s.ctx == nil || s.ctx.Err() != nil {
		state.queuedDispatching = false
		return false
	}
	state.queuedDispatching = false
	state.running++
	return true
}

func (s *Scheduler) run(state *jobState, scheduled time.Time, attempt int, releaseSlot func()) {
	defer s.wg.Done()
	defer s.finishRun(state)
	defer releaseSlot()

	var lease Lease
	if state.cfg.locker != nil {
		l, err := state.cfg.locker.TryLock(s.ctx, state.cfg.lockKey, state.cfg.lockTTL)
		if err != nil {
			eventType := EventLockFailed
			if errors.Is(err, ErrLockUnavailable) {
				eventType = EventLockSkipped
			}
			s.emit(s.ctx, s.event(state, eventType, scheduled, withAttempt(attempt), withErr(err)))
			return
		}
		lease = l
		defer func() { _ = lease.Release(context.Background()) }()
	}
	started := s.clock.Now()
	s.emit(s.ctx, s.event(state, EventStarted, scheduled, withAttempt(attempt), withStarted(started)))
	if err := runJob(s.ctx, state.cfg.job); err != nil {
		s.emit(s.ctx, s.event(state, EventFailed, scheduled, withAttempt(attempt), withStarted(started), withFinished(s.clock.Now()), withErr(err)))
		return
	}
	s.emit(s.ctx, s.event(state, EventSucceeded, scheduled, withAttempt(attempt), withStarted(started), withFinished(s.clock.Now())))
}

func runJob(ctx context.Context, job Job) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("schedulex: job panic: %v", recovered)
		}
	}()
	return job.Run(ctx)
}
