package schedulex

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestShutdownLeak_NoGoroutineLeak 验证 Scheduler 关闭后无 goroutine 泄漏。
// 创建 100 个 job，启动后停止，验证 goroutine 数恢复到启动前水平。
func TestShutdownLeak_NoGoroutineLeak(t *testing.T) {
	// 确保 GC 完成，获得稳定基线
	runtime.GC()
	runtime.Gosched()
	time.Sleep(50 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)

	s, err := NewScheduler(
		WithClock(clock),
		WithMaxConcurrent(200),
	)
	if err != nil {
		t.Fatal(err)
	}

	// 注册 100 个 job，每个用长间隔触发（不会在测试期间真正执行）
	for i := 0; i < 100; i++ {
		name := "leak-job-" + itoa(i)
		job := JobFunc{
			NameValue: name,
			RunFunc:   func(context.Context) error { return nil },
		}
		if err := s.AddJob(job, Every(time.Hour)); err != nil {
			t.Fatalf("AddJob %s: %v", name, err)
		}
	}

	ctx := context.Background()
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// 给 goroutine 启动的时间
	time.Sleep(100 * time.Millisecond)

	// 关闭 scheduler
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// 等待 goroutine 退出
	runtime.GC()
	runtime.Gosched()
	time.Sleep(200 * time.Millisecond)

	after := runtime.NumGoroutine()

	// 允许少量浮动（±5），因为 runtime 本身可能有 goroutine 泄漏噪音
	leaked := after - baseline
	if leaked > 5 {
		t.Fatalf("goroutine leak detected: baseline=%d, after=%d, leaked=%d", baseline, after, leaked)
	}
	t.Logf("goroutine check: baseline=%d, after=%d, leaked=%d", baseline, after, leaked)
}

// TestShutdownLeak_WithRunningJobs 验证有 job 正在执行时关闭也不泄漏。
func TestShutdownLeak_WithRunningJobs(t *testing.T) {
	runtime.GC()
	runtime.Gosched()
	time.Sleep(50 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	start := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	clock := NewStaticClock(start)

	s, err := NewScheduler(
		WithClock(clock),
		WithMaxConcurrent(200),
	)
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	blockers := make([]chan struct{}, 0, 50)
	for i := 0; i < 50; i++ {
		name := "running-job-" + itoa(i)
		ch := make(chan struct{})
		mu.Lock()
		blockers = append(blockers, ch)
		mu.Unlock()
		block := ch
		job := JobFunc{
			NameValue: name,
			RunFunc: func(context.Context) error {
				<-block
				return nil
			},
		}
		if err := s.AddJob(job, Every(time.Hour)); err != nil {
			t.Fatalf("AddJob %s: %v", name, err)
		}
	}

	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 推进时钟让所有 job 触发并开始运行
	clock.Advance(time.Hour)
	time.Sleep(100 * time.Millisecond)

	// 开始关闭（会等待 running job 完成）
	var done atomic.Bool
	go func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.Shutdown(shutdownCtx)
		done.Store(true)
	}()

	// 释放所有阻塞的 job
	mu.Lock()
	for _, ch := range blockers {
		close(ch)
	}
	mu.Unlock()

	// 等待 shutdown 完成
	time.Sleep(500 * time.Millisecond)
	if !done.Load() {
		t.Fatal("Shutdown did not complete after releasing jobs")
	}

	runtime.GC()
	runtime.Gosched()
	time.Sleep(200 * time.Millisecond)

	after := runtime.NumGoroutine()
	leaked := after - baseline
	if leaked > 5 {
		t.Fatalf("goroutine leak with running jobs: baseline=%d, after=%d, leaked=%d", baseline, after, leaked)
	}
	t.Logf("goroutine check (running jobs): baseline=%d, after=%d, leaked=%d", baseline, after, leaked)
}

// itoa 是一个简单的 int→string 转换，避免引入 strconv 依赖。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 8)
	for n > 0 {
		b = append(b, byte('0'+n%10))
		n /= 10
	}
	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
