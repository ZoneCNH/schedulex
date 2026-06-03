package templatex

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewRejectsInvalidConfig(t *testing.T) {
	metrics := &recordingMetrics{}

	_, err := New(context.Background(), Config{Timeout: time.Second}, WithMetrics(metrics))
	if err == nil {
		t.Fatal("expected invalid config to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
	if !metrics.counterWithLabel(MetricClientErrorsTotal, "kind", string(ErrorKindValidation)) {
		t.Fatalf("expected validation error metric, got %#v", metrics.counters)
	}
}

func TestNewRejectsNilContext(t *testing.T) {
	metrics := &recordingMetrics{}

	_, err := New(nil, Config{Name: "templatex"}, WithMetrics(metrics)) //nolint:staticcheck // verifies the defensive nil-context branch.
	if err == nil {
		t.Fatal("expected nil context to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
	if !metrics.counterWithLabel(MetricClientErrorsTotal, "kind", string(ErrorKindValidation)) {
		t.Fatalf("expected validation error metric, got %#v", metrics.counters)
	}
}

func TestNewRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(ctx, Config{Name: "templatex"})
	if err == nil {
		t.Fatal("expected canceled context to fail")
	}
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable error, got %T %[1]v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled cause, got %v", err)
	}
}

func TestNewRejectsExpiredContext(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := New(ctx, Config{Name: "templatex"})
	if err == nil {
		t.Fatal("expected expired context to fail")
	}
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout error, got %T %[1]v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline cause, got %v", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{Name: "templatex"}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if !metrics.hasCounter(MetricClientCreatedTotal) {
		t.Fatalf("expected client creation metric, got %#v", metrics.counters)
	}

	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if !metrics.hasCounter(MetricClientClosedTotal) {
		t.Fatalf("expected client close metric, got %#v", metrics.counters)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestCloseRejectsNilClient(t *testing.T) {
	var client *Client

	err := client.Close(context.Background())
	if err == nil {
		t.Fatal("expected nil client to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
}

func TestCloseRejectsNilContext(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{Name: "templatex"}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	err = client.Close(nil) //nolint:staticcheck // verifies the defensive nil-context branch.
	if err == nil {
		t.Fatal("expected nil close context to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
	if !metrics.counterWithLabel(MetricClientErrorsTotal, "kind", string(ErrorKindValidation)) {
		t.Fatalf("expected validation error metric, got %#v", metrics.counters)
	}
}

func TestCloseRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client, err := New(context.Background(), Config{Name: "templatex"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	err = client.Close(ctx)
	if err == nil {
		t.Fatal("expected canceled close context to fail")
	}
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable error, got %T %[1]v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled cause, got %v", err)
	}
}

func TestCloseRejectsExpiredContext(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{Name: "templatex"}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err = client.Close(ctx)
	if err == nil {
		t.Fatal("expected expired close context to fail")
	}
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout error, got %T %[1]v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline cause, got %v", err)
	}
	if !metrics.counterWithLabel(MetricClientErrorsTotal, "kind", string(ErrorKindTimeout)) {
		t.Fatalf("expected timeout error metric, got %#v", metrics.counters)
	}
}

func TestCloseRejectsZeroValueClient(t *testing.T) {
	var client Client

	err := client.Close(context.Background())
	if err == nil {
		t.Fatal("expected zero-value client to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %T %[1]v", err)
	}
}
