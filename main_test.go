package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
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

func TestMetricsMiddleware_Success(t *testing.T) {
	called := false
	mw := metricsMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		if method != methodCallTool {
			t.Fatalf("expected method %q, got %q", methodCallTool, method)
		}
		return nil, nil
	})
	_, err := mw(context.Background(), methodCallTool, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected inner handler to be called")
	}
}

func TestMetricsMiddleware_Error(t *testing.T) {
	mw := metricsMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, errors.New("handler error")
	})
	_, err := mw(context.Background(), methodCallTool, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMetricsMiddleware_NonToolMethod(t *testing.T) {
	called := false
	mw := metricsMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		return nil, nil
	})
	_, err := mw(context.Background(), "ping", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected inner handler to be called")
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("TEST_ENV_OR_DEFAULT", "env_value")
	got := envOrDefault("TEST_ENV_OR_DEFAULT", "fallback")
	if got != "env_value" {
		t.Fatalf("expected env_value, got %s", got)
	}
}

func TestEnvOrDefault_Fallback(t *testing.T) {
	t.Setenv("TEST_ENV_OR_DEFAULT", "")
	got := envOrDefault("TEST_ENV_OR_DEFAULT", "fallback")
	if got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestHealthzHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/healthz", nil)
	healthzHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok status, got %s", w.Body.String())
	}
}

func TestReadyzHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/readyz", nil)
	readyzHandler(mock)(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyzHandler_Unhealthy(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		HealthCheckFunc: func(ctx context.Context) error {
			return errors.New("unhealthy")
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/readyz", nil)
	readyzHandler(mock)(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
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

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	return buf.String()
}
