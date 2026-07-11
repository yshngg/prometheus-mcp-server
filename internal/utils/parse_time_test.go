package utils

import (
	"testing"
	"time"
)

func TestParseTime_RFC3339(t *testing.T) {
	ts, err := ParseTime("2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !ts.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, ts)
	}
}

func TestParseTime_Unix(t *testing.T) {
	ts, err := ParseTime("1704067200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Unix(1704067200, 0)
	if !ts.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, ts)
	}
}

func TestParseTime_Empty(t *testing.T) {
	_, err := ParseTime("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestParseTime_Invalid(t *testing.T) {
	_, err := ParseTime("not-a-time")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}
