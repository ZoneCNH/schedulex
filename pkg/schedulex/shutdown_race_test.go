package schedulex

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestShutdownRace_ConcurrentStopAndFire 验证并发关闭和 job 触发的安全性。
// 使用 go test -race 验证无竞态。
func TestShutdownRace_ConcurrentStopAndFire(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)

	for i := 0; i < 50; i++ {
		s, err := NewScheduler(
			WithClock(clock),
			WithMaxConcurrent(10),
		)
		if err != nil {
			t.Fatal(err)
		}

		for j := 0; j < 5; j++ {
			job := JobFunc{
				NameValue: "race-job-" + itoa(i) + "-" + itoa(j),
				RunFunc:   func(context.Context) error { return nil },
			}
			if err := s.AddJob(job, Every(time.Second)); err != nil {
				t.Fatal(err)
			}
		}

		if err := s.Start(context.Background()); err != nil {
			t.Fatal(err)
		}

		// 立即触发并同时关闭
		clock.Advance(time.Second)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = s.Shutdown(ctx)
		}()

		go func() {
			defer wg.Done()
			// 再次推进时钟触发更多事件
			clock.Advance(time.Second)
		}()

		wg.Wait()
	}
}

// TestShutdownRace_MultipleStopCalls 验证多次并发调用 Stop 的安全性。
func TestShutdownRace_MultipleStopCalls(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)

	for i := 0; i < 50; i++ {
		s, err := NewScheduler(
			WithClock(clock),
			WithMaxConcurrent(10),
		)
		if err != nil {
			t.Fatal(err)
		}

		job := JobFunc{
			NameValue: "multi-stop-job",
			RunFunc:   func(context.Context) error { return nil },
		}
		if err := s.AddJob(job, Every(time.Second)); err != nil {
			t.Fatal(err)
		}
		if err := s.Start(context.Background()); err != nil {
			t.Fatal(err)
		}

		// 并发调用 5 次 Shutdown
		var wg sync.WaitGroup
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = s.Shutdown(ctx)
			}()
		}

		clock.Advance(time.Second)
		wg.Wait()
	}
}

// TestShutdownRace_AddJobDuringShutdown 验证在关闭过程中 AddJob 的安全性。
func TestShutdownRace_AddJobDuringShutdown(t *testing.T) {
	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)

	for i := 0; i < 50; i++ {
		s, err := NewScheduler(
			WithClock(clock),
			WithMaxConcurrent(10),
		)
		if err != nil {
			t.Fatal(err)
		}

		job := JobFunc{
			NameValue: "pre-job",
			RunFunc:   func(context.Context) error { return nil },
		}
		if err := s.AddJob(job, Every(time.Second)); err != nil {
			t.Fatal(err)
		}
		if err := s.Start(context.Background()); err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = s.Shutdown(ctx)
		}()

		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				newJob := JobFunc{
					NameValue: "late-job-" + itoa(j),
					RunFunc:   func(context.Context) error { return nil },
				}
				_ = s.AddJob(newJob, Every(time.Minute))
			}
		}()

		clock.Advance(time.Second)
		wg.Wait()
	}
}
