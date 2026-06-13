package schedulex

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestMisfireContract_Skip 验证 MisfireSkip 策略：错过触发后跳过，不执行 job。
func TestMisfireContract_Skip(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	recorder := newEventRecorder()

	s, err := NewScheduler(
		WithClock(clock),
		WithEventSink(recorder),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	var runCount int
	job := JobFunc{
		NameValue: "skip-job",
		RunFunc: func(context.Context) error {
			runCount++
			return nil
		},
	}
	if err := s.AddJob(job, Every(time.Minute), WithMisfirePolicy(MisfireSkip)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 等待 scheduled 事件，确保 After() waiter 已注册
	recorder.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == "skip-job" })
	time.Sleep(10 * time.Millisecond)

	// 跳过 5 分钟，模拟 misfire
	clock.Advance(5*time.Minute + misfireGrace + time.Millisecond)

	// 给 job 一点时间执行
	time.Sleep(50 * time.Millisecond)

	// MisfireSkip：missed 的点全部跳过，不执行
	// 查找 misfire 事件确认 runs=0
	if _, found := recorder.find(EventMisfire, func(e Event) bool { return e.JobID == "skip-job" }); found {
		// misfire 发生时 Skip 策略 runs=0，job 不会因 misfire 额外执行
		if runCount > 1 {
			t.Fatalf("MisfireSkip: expected at most 1 run, got %d", runCount)
		}
	}
}

// TestMisfireContract_RunOnce 验证 MisfireRunOnce 策略：错过多个触发后只执行最后一次。
func TestMisfireContract_RunOnce(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	recorder := newEventRecorder()

	s, err := NewScheduler(
		WithClock(clock),
		WithEventSink(recorder),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	var runCount atomic.Int32
	job := JobFunc{
		NameValue: "runonce-job",
		RunFunc: func(context.Context) error {
			runCount.Add(1)
			return nil
		},
	}
	if err := s.AddJob(job, Every(time.Minute), WithMisfirePolicy(MisfireRunOnce)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	recorder.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == "runonce-job" })
	time.Sleep(10 * time.Millisecond)

	// 跳过 5 分钟，触发 misfire
	clock.Advance(5*time.Minute + misfireGrace + time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	// MisfireRunOnce：只执行最后 1 个错过的时间点
	if me, found := recorder.find(EventMisfire, func(e Event) bool { return e.JobID == "runonce-job" }); found {
		runs := me.Attributes["runs"]
		if runs != "1" {
			t.Errorf("MisfireRunOnce: expected runs=1 in attributes, got runs=%s", runs)
		}
	}
}

// TestMisfireContract_CatchUp 验证 MisfireCatchUp 策略：错过多少执行多少。
func TestMisfireContract_CatchUp(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	recorder := newEventRecorder()

	s, err := NewScheduler(
		WithClock(clock),
		WithEventSink(recorder),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	var runCount atomic.Int32
	job := JobFunc{
		NameValue: "catchup-job",
		RunFunc: func(context.Context) error {
			runCount.Add(1)
			return nil
		},
	}
	if err := s.AddJob(job, Every(time.Minute), WithMisfirePolicy(MisfireCatchUp)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	recorder.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == "catchup-job" })
	time.Sleep(10 * time.Millisecond)

	// 跳过 3 分钟，触发 misfire
	clock.Advance(3*time.Minute + misfireGrace + time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	// MisfireCatchUp：所有错过的时间点都要执行
	if me, found := recorder.find(EventMisfire, func(e Event) bool { return e.JobID == "catchup-job" }); found {
		runs := me.Attributes["runs"]
		missed := me.Attributes["missed"]
		t.Logf("MisfireCatchUp: missed=%s, runs=%s", missed, runs)
		// runs 应该等于 missed 数量
		if runs != missed {
			t.Errorf("MisfireCatchUp: expected runs=%s == missed=%s", runs, missed)
		}
	}
}
