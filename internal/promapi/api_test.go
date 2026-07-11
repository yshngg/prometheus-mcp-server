package promapi

import (
	"net/http"
	"testing"
)

func TestNew_InvalidAddr(t *testing.T) {
	_, err := New("http://[::1]:namedport", http.DefaultClient, nil)
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
}

func TestNew_ValidAddr(t *testing.T) {
	// This should succeed, but the returned client may not be usable for real queries.
	_, err := New("http://localhost:9090", http.DefaultClient, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
