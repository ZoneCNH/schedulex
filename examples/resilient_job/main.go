// Package main 展示如何使用 resiliencx pattern 包裹 schedulex job 执行。
//
// 该示例演示：
//   - retry：job 执行失败后自动重试（指数退避）
//   - timeout：单次执行超时保护
//   - circuit breaker：连续失败后熔断，避免无效重试
//
// 这些 pattern 是 job wrapper，不影响 scheduler 的调度逻辑。
// scheduler 仍然负责触发时机、misfire 策略和 overlap 策略。
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// ─── Resiliencx Pattern ───────────────────────────────────────

// RetryPolicy 定义重试策略。
type RetryPolicy struct {
	MaxAttempts int           // 最大尝试次数（含首次）
	BaseDelay   time.Duration // 首次重试延迟
	MaxDelay    time.Duration // 最大重试延迟
}

// DefaultRetryPolicy 返回保守的默认重试策略。
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
	}
}

// TimeoutPolicy 定义单次执行超时。
type TimeoutPolicy struct {
	Timeout time.Duration
}

// CircuitBreakerPolicy 定义熔断策略。
// 当连续失败次数达到 Threshold 时，熔断器打开，后续调用直接返回错误。
// 经过 ResetAfter 后，熔断器进入半开状态，允许一次尝试。
type CircuitBreakerPolicy struct {
	Threshold  int           // 连续失败多少次后熔断
	ResetAfter time.Duration // 熔断后多久恢复半开状态
}

// ResilientJob 是一个 job wrapper，组合 retry + timeout + circuit breaker。
type ResilientJob struct {
	Name_   string
	Execute func(ctx context.Context) error
	Retry   RetryPolicy
	Timeout TimeoutPolicy
	Breaker CircuitBreakerPolicy

	consecutiveFailures atomic.Int32
	lastFailureTime     atomic.Int64
	circuitOpen         atomic.Bool
}

// Name 返回 job 名称。
func (r *ResilientJob) Name() string { return r.Name_ }

// Run 执行 job，应用 retry + timeout + circuit breaker 策略。
func (r *ResilientJob) Run(ctx context.Context) error {
	// 1. Circuit Breaker 检查
	if r.circuitOpen.Load() {
		resetAt := time.Unix(0, r.lastFailureTime.Load()).Add(r.Breaker.ResetAfter)
		if time.Now().Before(resetAt) {
			return fmt.Errorf("circuit breaker open for %s", r.Name_)
		}
		// 半开状态：允许一次尝试
		r.circuitOpen.Store(false)
	}

	// 2. Retry 循环
	var lastErr error
	attempts := r.Retry.MaxAttempts
	if attempts <= 0 {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			delay := r.retryDelay(attempt)
			log.Printf("[%s] retry %d/%d after %v", r.Name_, attempt+1, attempts, delay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		// 3. Timeout 保护
		err := r.executeWithTimeout(ctx)
		if err == nil {
			// 成功：重置熔断计数
			r.consecutiveFailures.Store(0)
			return nil
		}

		lastErr = err

		// 不可重试的错误直接返回
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}

	// 所有重试都失败：更新熔断计数
	failures := r.consecutiveFailures.Add(1)
	r.lastFailureTime.Store(time.Now().UnixNano())
	if int(failures) >= r.Breaker.Threshold {
		r.circuitOpen.Store(true)
		log.Printf("[%s] circuit breaker OPEN after %d consecutive failures", r.Name_, failures)
	}

	return fmt.Errorf("all %d attempts failed for %s: %w", attempts, r.Name_, lastErr)
}

func (r *ResilientJob) executeWithTimeout(ctx context.Context) error {
	timeout := r.Timeout.Timeout
	if timeout <= 0 {
		return r.Execute(ctx)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.Execute(timeoutCtx)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

func (r *ResilientJob) retryDelay(attempt int) time.Duration {
	delay := r.Retry.BaseDelay * time.Duration(1<<uint(attempt-1))
	if delay > r.Retry.MaxDelay {
		delay = r.Retry.MaxDelay
	}
	return delay
}

// ─── 示例用法 ─────────────────────────────────────────────────

func main() {
	// 创建带 resilience 的 job
	settlement := &ResilientJob{
		Name_: "order-settlement",
		Execute: func(ctx context.Context) error {
			log.Println("执行订单结算...")
			// 模拟业务逻辑
			return nil
		},
		Retry:   DefaultRetryPolicy(),
		Timeout: TimeoutPolicy{Timeout: 30 * time.Second},
		Breaker: CircuitBreakerPolicy{
			Threshold:  5,
			ResetAfter: 1 * time.Minute,
		},
	}

	// 在 scheduler 中注册 resilient job：
	//
	//   scheduler.AddJob(settlement, schedulex.Every(time.Minute),
	//       schedulex.WithMisfirePolicy(schedulex.MisfireRunOnce),
	//       schedulex.WithOverlapPolicy(schedulex.OverlapSkip),
	//   )
	//
	// scheduler 负责：触发时机、misfire、overlap
	// ResilientJob 负责：retry、timeout、circuit breaker

	if err := settlement.Run(context.Background()); err != nil {
		log.Printf("job failed: %v", err)
	} else {
		log.Println("job succeeded")
	}
}
