package testkit

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestConfigBuildsValidFixture(t *testing.T) {
	cfg := Config("fixture")
	if cfg.Name != "fixture" {
		t.Fatalf("unexpected name: %q", cfg.Name)
	}
	if cfg.Timeout != time.Second {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
	RequireNoError(t, cfg.Validate())
}

func TestRequireNoErrorAcceptsNil(t *testing.T) {
	RequireNoError(t, nil)
}

func TestRequireNoErrorReportsError(t *testing.T) {
	tb := newRecordingTB()

	expectFatal(t, func() {
		requireNoError(tb, errors.New("boom"))
	})

	if !tb.helperCalled {
		t.Fatal("expected Helper to be called")
	}
	if tb.message != "expected no error, got boom" {
		t.Fatalf("unexpected fatal message: %q", tb.message)
	}
}

type recordingTB struct {
	helperCalled bool
	message      string
}

func newRecordingTB() *recordingTB {
	return &recordingTB{}
}

func (tb *recordingTB) Helper() {
	tb.helperCalled = true
}

func (tb *recordingTB) Fatalf(format string, args ...any) {
	tb.message = fmt.Sprintf(format, args...)
	panic(tb.message)
}

func expectFatal(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected fatal panic")
		}
	}()

	fn()
}
