package schedulex

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestOverlapContract_Skip 验证 OverlapSkip 策略：job 仍在运行时新触发被跳过。
func TestOverlapContract_Skip(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	recorder := newEventRecorder()

	s, err := NewScheduler(
		WithClock(clock),
		WithMaxConcurrent(4),
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
	block := make(chan struct{})
	job := JobFunc{
		NameValue: "slow-skip-job",
		RunFunc: func(context.Context) error {
			runCount.Add(1)
			<-block // 阻塞直到手动释放
			return nil
		},
	}
	// 每 10 秒触发一次，overlap=skip
	if err := s.AddJob(job, Every(10*time.Second), WithOverlapPolicy(OverlapSkip)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	recorder.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == "slow-skip-job" })
	time.Sleep(10 * time.Millisecond)

	// 推进 10 秒 → 第一次触发，job 开始运行并阻塞
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	// 推进另一个 10 秒 → 第二次触发，应被 skip
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	// 释放阻塞
	close(block)
	time.Sleep(50 * time.Millisecond)

	if got := runCount.Load(); got != 1 {
		t.Fatalf("OverlapSkip: expected 1 run, got %d", got)
	}

	if _, found := recorder.find(EventSkipped, func(e Event) bool { return e.JobID == "slow-skip-job" }); !found {
		t.Error("OverlapSkip: expected at least 1 skip event, got 0")
	}
}

// TestOverlapContract_QueueOne 验证 OverlapQueueOne 策略：job 仍在运行时新触发排队，完成后执行一次。
func TestOverlapContract_QueueOne(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	recorder := newEventRecorder()

	s, err := NewScheduler(
		WithClock(clock),
		WithMaxConcurrent(4),
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
	block := make(chan struct{})
	job := JobFunc{
		NameValue: "queue-one-job",
		RunFunc: func(context.Context) error {
			runCount.Add(1)
			<-block
			return nil
		},
	}
	// 用小间隔触发，overlap=queue_one
	if err := s.AddJob(job, Every(10*time.Second), WithOverlapPolicy(OverlapQueueOne)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	recorder.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == "queue-one-job" })
	time.Sleep(10 * time.Millisecond)

	// 第 1 次触发 (T+10s): job 开始运行并阻塞
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	// 第 2 次触发 (T+20s): 排队（queue_one 只缓存 1 个）
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	// 释放阻塞，让第一个 job 完成。此时 clock 仍在 T+20s 附近，
	// queued dispatch 不会被 misfire 检查拦截。
	close(block)
	time.Sleep(200 * time.Millisecond)

	// 期望：第 1 次正常执行 + 排队的 1 次 = 2 次
	if got := runCount.Load(); got != 2 {
		t.Fatalf("OverlapQueueOne: expected 2 runs, got %d", got)
	}
}

// TestOverlapContract_Allow 验证 OverlapAllow 策略：允许并发执行。
func TestOverlapContract_Allow(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)
	recorder := newEventRecorder()

	s, err := NewScheduler(
		WithClock(clock),
		WithMaxConcurrent(10),
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
	block := make(chan struct{})
	job := JobFunc{
		NameValue: "allow-job",
		RunFunc: func(context.Context) error {
			runCount.Add(1)
			<-block
			return nil
		},
	}
	// 每 10 秒触发，overlap=allow
	if err := s.AddJob(job, Every(10*time.Second), WithOverlapPolicy(OverlapAllow)); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	recorder.waitFor(t, EventScheduled, func(e Event) bool { return e.JobID == "allow-job" })
	time.Sleep(10 * time.Millisecond)

	// 第 1 次触发
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	// 第 2 次触发：Allow 策略应允许并发执行
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	// 第 3 次触发
	clock.Advance(10 * time.Second)
	time.Sleep(50 * time.Millisecond)

	if got := runCount.Load(); got != 3 {
		t.Fatalf("OverlapAllow: expected 3 concurrent runs, got %d", got)
	}

	close(block)
	time.Sleep(100 * time.Millisecond)
}
