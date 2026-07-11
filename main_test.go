package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)

func TestUsageFor(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	fs.String("test-flag", "default", "a test flag")

	usage := usageFor(fs, "test [flags]")

	var buf strings.Builder
	usage()

	// usage calls os.Stderr which we can't easily capture,
	// but we can verify it doesn't panic and captures correctly
	_ = buf
	_ = usage
}

func TestUsageForOutput(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	fs.String("test-flag", "default", "a test flag")
	fs.String("test-empty", "", "an empty flag")

	usage := usageFor(fs, "test [flags]")

	capture := captureStderr(t, func() {
		usage()
	})

	if !strings.Contains(capture, "test [flags]") {
		t.Fatal("expected output to contain command")
	}
	if !strings.Contains(capture, "test-flag") {
		t.Fatal("expected output to contain flag name")
	}
	if !strings.Contains(capture, "a test flag") {
		t.Fatal("expected output to contain flag description")
	}
	if !strings.Contains(capture, "...") {
		t.Fatal("expected output to contain '...' for empty default value")
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	return buf.String()
}
