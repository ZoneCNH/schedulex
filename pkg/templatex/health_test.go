package templatex

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestHealthCheckHealthy(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{Name: "templatex"}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy status, got %q", status.Status)
	}
	if status.Name != "templatex" {
		t.Fatalf("expected templatex health name, got %q", status.Name)
	}
	if status.LatencyMs < 0 {
		t.Fatalf("expected non-negative latency, got %d", status.LatencyMs)
	}
	if !metrics.hasGauge(MetricClientHealthStatus) {
		t.Fatalf("expected health status gauge, got %#v", metrics.gauges)
	}
	if !metrics.hasHistogram(MetricClientHealthLatencyMS) {
		t.Fatalf("expected health latency histogram, got %#v", metrics.histograms)
	}
}

func TestHealthCheckClosedClientUnhealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
}

func TestHealthCheckCanceledContextUnhealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	status := client.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if !strings.Contains(status.Message, context.Canceled.Error()) {
		t.Fatalf("expected canceled message, got %q", status.Message)
	}
}

func TestHealthCheckNilContextUnhealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.HealthCheck(nil) //nolint:staticcheck // verifies the defensive nil-context branch.
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if status.Message != "context is required" {
		t.Fatalf("expected nil context message, got %q", status.Message)
	}
}

func TestHealthCheckDeadlineBelowTimeoutDegraded(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{
		Name:    "templatex",
		Timeout: time.Hour,
	}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status := client.HealthCheck(ctx)
	if status.Status != HealthDegraded {
		t.Fatalf("expected degraded status, got %q", status.Status)
	}
	if status.Message != "context deadline is shorter than client timeout" {
		t.Fatalf("expected degraded message, got %q", status.Message)
	}
	if status.Metadata["reason"] != "deadline_below_timeout" {
		t.Fatalf("expected degraded reason metadata, got %#v", status.Metadata)
	}
	if status.Metadata["timeout"] != time.Hour.String() {
		t.Fatalf("expected timeout metadata, got %#v", status.Metadata)
	}

	payload, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("marshal degraded health status: %v", err)
	}
	encoded := string(payload)
	for _, field := range []string{"name", "status", "checked_at", "latency_ms", "metadata"} {
		if !strings.Contains(encoded, `"`+field+`"`) {
			t.Fatalf("expected JSON field %q in %s", field, encoded)
		}
	}
	if !strings.Contains(encoded, `"status":"degraded"`) {
		t.Fatalf("expected degraded JSON status, got %s", encoded)
	}
	if strings.Contains(encoded, "CheckedAt") || strings.Contains(encoded, "LatencyMs") {
		t.Fatalf("expected snake_case JSON fields, got %s", encoded)
	}

	labels := map[string]string{
		"name":   "templatex",
		"status": string(HealthDegraded),
	}
	if !metrics.gaugeWithLabels(MetricClientHealthStatus, 0, labels) {
		t.Fatalf("expected degraded health status gauge, got %#v", metrics.gauges)
	}
	if !metrics.histogramWithLabels(MetricClientHealthLatencyMS, labels) {
		t.Fatalf("expected degraded health latency histogram, got %#v", metrics.histograms)
	}
}

func TestHealthCheckTimeoutWithoutDeadlineHealthy(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{
		Name:    "templatex",
		Timeout: time.Minute,
	}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy status, got %q", status.Status)
	}
	if status.Metadata != nil {
		t.Fatalf("expected no health metadata, got %#v", status.Metadata)
	}

	labels := map[string]string{
		"name":   "templatex",
		"status": string(HealthHealthy),
	}
	if !metrics.gaugeWithLabels(MetricClientHealthStatus, 1, labels) {
		t.Fatalf("expected healthy health status gauge, got %#v", metrics.gauges)
	}
	if !metrics.histogramWithLabels(MetricClientHealthLatencyMS, labels) {
		t.Fatalf("expected healthy health latency histogram, got %#v", metrics.histograms)
	}
}

func TestHealthCheckDeadlineAboveTimeoutHealthy(t *testing.T) {
	client, err := New(context.Background(), Config{
		Name:    "templatex",
		Timeout: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	status := client.HealthCheck(ctx)
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy status, got %q", status.Status)
	}
	if status.Metadata["reason"] == "deadline_below_timeout" {
		t.Fatalf("expected no degraded reason metadata, got %#v", status.Metadata)
	}
}

func TestHealthCheckNilClientUnhealthy(t *testing.T) {
	var client *Client

	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if status.Name != "templatex" {
		t.Fatalf("expected fallback health name, got %q", status.Name)
	}
}

func TestHealthCheckElapsedDeadlineWithoutContextErrorUnhealthy(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{
		Name:    "templatex",
		Timeout: time.Hour,
	}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.HealthCheck(deadlineOnlyContext{
		Context:  context.Background(),
		deadline: time.Now().Add(-time.Second),
	})
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if status.Message != context.DeadlineExceeded.Error() {
		t.Fatalf("expected deadline message, got %q", status.Message)
	}

	labels := map[string]string{
		"name":   "templatex",
		"status": string(HealthUnhealthy),
	}
	if !metrics.gaugeWithLabels(MetricClientHealthStatus, 0, labels) {
		t.Fatalf("expected unhealthy health status gauge, got %#v", metrics.gauges)
	}
}

func TestHealthCheckElapsedDeadlineUsesCurrentContextError(t *testing.T) {
	client, err := New(context.Background(), Config{
		Name:    "templatex",
		Timeout: time.Hour,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx := &changingErrDeadlineContext{
		Context:  context.Background(),
		deadline: time.Now().Add(-time.Second),
		err:      context.Canceled,
	}
	status := client.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if status.Message != context.Canceled.Error() {
		t.Fatalf("expected current context error message, got %q", status.Message)
	}
}

func TestHealthCheckZeroValueClientUnhealthy(t *testing.T) {
	var client Client

	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if status.Name != "templatex" {
		t.Fatalf("expected fallback health name, got %q", status.Name)
	}
}

func TestHealthStatusJSONContract(t *testing.T) {
	payload, err := json.Marshal(HealthStatus{
		Name:      "templatex",
		Status:    HealthHealthy,
		LatencyMs: 7,
	})
	if err != nil {
		t.Fatalf("marshal health status: %v", err)
	}
	encoded := string(payload)
	for _, field := range []string{"name", "status", "checked_at", "latency_ms"} {
		if !strings.Contains(encoded, `"`+field+`"`) {
			t.Fatalf("expected JSON field %q in %s", field, encoded)
		}
	}
	if strings.Contains(encoded, "CheckedAt") || strings.Contains(encoded, "LatencyMs") {
		t.Fatalf("expected snake_case JSON fields, got %s", encoded)
	}
}

type deadlineOnlyContext struct {
	context.Context
	deadline time.Time
}

func (ctx deadlineOnlyContext) Deadline() (time.Time, bool) {
	return ctx.deadline, true
}

func (ctx deadlineOnlyContext) Err() error {
	return nil
}

type changingErrDeadlineContext struct {
	context.Context
	deadline time.Time
	err      error
	errCalls int
}

func (ctx *changingErrDeadlineContext) Deadline() (time.Time, bool) {
	return ctx.deadline, true
}

func (ctx *changingErrDeadlineContext) Err() error {
	ctx.errCalls++
	if ctx.errCalls == 1 {
		return nil
	}
	return ctx.err
}
