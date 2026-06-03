package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ZoneCNH/xlib-standard/pkg/templatex"
)

func TestMainPrintsModuleName(t *testing.T) {
	output := captureStdout(t, main)
	if output != "github.com/ZoneCNH/xlib-standard\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunReportsInvalidConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer

	run(&stdout, &stderr, templatex.Config{})

	if stdout.String() != "" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if stderr.String() != "create client: validation: Config.Validate: name is required\n" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return buf.String()
}
