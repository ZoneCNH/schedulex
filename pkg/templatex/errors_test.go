package templatex

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewErrorFormatsKindOpAndMessage(t *testing.T) {
	err := NewError(ErrorKindValidation, "templatex.Test", "bad input", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Kind != ErrorKindValidation {
		t.Fatalf("expected validation kind, got %q", err.Kind)
	}
	if err.Retryable {
		t.Fatal("expected non-retryable error")
	}
	if got := err.Error(); !strings.Contains(got, "validation: templatex.Test: bad input") {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestWrapErrorPreservesCauseAndKind(t *testing.T) {
	cause := context.DeadlineExceeded
	err := WrapError(ErrorKindTimeout, "templatex.Test", "", true, cause)

	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout kind, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause, got %v", err)
	}
	if !err.Retryable {
		t.Fatal("expected retryable error")
	}
}

func TestErrorHandlesNilReceiverAndCauseOnlyMessage(t *testing.T) {
	var nilError *Error
	if got := nilError.Error(); got != "" {
		t.Fatalf("nil Error.Error() = %q; want empty string", got)
	}
	if got := nilError.Unwrap(); got != nil {
		t.Fatalf("nil Error.Unwrap() = %v; want nil", got)
	}

	cause := errors.New("root cause")
	err := &Error{Kind: ErrorKindInternal, Op: "templatex.Test", Cause: cause}
	if got := err.Error(); got != "internal: templatex.Test: root cause" {
		t.Fatalf("cause-only Error() = %q", got)
	}
}

func TestErrorKindFallbacksForPlainErrors(t *testing.T) {
	err := errors.New("plain")
	if IsKind(err, ErrorKindInternal) {
		t.Fatal("plain errors should not match templatex error kinds")
	}
	if got := errorKind(err); got != ErrorKindInternal {
		t.Fatalf("errorKind(plain) = %q; want %q", got, ErrorKindInternal)
	}
}

func TestContextErrorClassifiesDeadlineAsRetryableTimeout(t *testing.T) {
	err := contextError("templatex.Test", context.DeadlineExceeded)
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout kind, got %v", err)
	}
	if !err.Retryable {
		t.Fatal("expected deadline errors to be retryable")
	}
}
