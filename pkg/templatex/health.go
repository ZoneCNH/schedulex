package templatex

import (
	"context"
	"time"
)

type HealthStatusValue string

const (
	HealthHealthy   HealthStatusValue = "healthy"
	HealthDegraded  HealthStatusValue = "degraded"
	HealthUnhealthy HealthStatusValue = "unhealthy"
)

type HealthStatus struct {
	Name      string            `json:"name"`
	Status    HealthStatusValue `json:"status"`
	Message   string            `json:"message,omitempty"`
	CheckedAt time.Time         `json:"checked_at"`
	LatencyMs int64             `json:"latency_ms"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (c *Client) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()
	name := "templatex"
	var metrics Metrics
	initialized := false
	closed := true
	var timeout time.Duration

	if c != nil {
		c.mu.Lock()
		name = c.cfg.Name
		metrics = c.metrics
		initialized = c.initialized
		closed = c.closed
		timeout = c.cfg.Timeout
		c.mu.Unlock()
		if name == "" {
			name = "templatex"
		}
	}

	if ctx == nil {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   "context is required",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if err := ctx.Err(); err != nil {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   err.Error(),
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if !initialized {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   "client is not initialized",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if closed {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   "client is closed",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if timeout > 0 {
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				message := context.DeadlineExceeded.Error()
				if err := ctx.Err(); err != nil {
					message = err.Error()
				}
				status := HealthStatus{
					Name:      name,
					Status:    HealthUnhealthy,
					Message:   message,
					CheckedAt: time.Now(),
					LatencyMs: time.Since(start).Milliseconds(),
				}
				recordHealthMetric(metrics, status)
				return status
			}
			if remaining < timeout {
				status := HealthStatus{
					Name:      name,
					Status:    HealthDegraded,
					Message:   "context deadline is shorter than client timeout",
					CheckedAt: time.Now(),
					LatencyMs: time.Since(start).Milliseconds(),
					Metadata: map[string]string{
						"reason":  "deadline_below_timeout",
						"timeout": timeout.String(),
					},
				}
				recordHealthMetric(metrics, status)
				return status
			}
		}
	}

	status := HealthStatus{
		Name:      name,
		Status:    HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Now(),
		LatencyMs: time.Since(start).Milliseconds(),
	}
	recordHealthMetric(metrics, status)
	return status
}

func recordHealthMetric(metrics Metrics, status HealthStatus) {
	if metrics == nil {
		return
	}
	labels := map[string]string{
		"name":   status.Name,
		"status": string(status.Status),
	}
	metrics.SetGauge(MetricClientHealthStatus, healthGaugeValue(status.Status), labels)
	metrics.ObserveHistogram(MetricClientHealthLatencyMS, float64(status.LatencyMs), labels)
}

func healthGaugeValue(status HealthStatusValue) float64 {
	if status == HealthHealthy {
		return 1
	}
	return 0
}
