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
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
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

func TestAuthMiddleware_NoToken(t *testing.T) {
	mw := authMiddleware("")
	innerCalled := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/mcp", nil)
	handler.ServeHTTP(w, r)

	if !innerCalled {
		t.Fatal("expected inner handler to be called when no auth token configured")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mw := authMiddleware("valid-token")
	innerCalled := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/mcp", nil)
	r.Header.Set("Authorization", "Bearer valid-token")
	handler.ServeHTTP(w, r)

	if !innerCalled {
		t.Fatal("expected inner handler to be called with valid token")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mw := authMiddleware("valid-token")
	innerCalled := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/mcp", nil)
	r.Header.Set("Authorization", "Bearer wrong-token")
	handler.ServeHTTP(w, r)

	if innerCalled {
		t.Fatal("expected inner handler NOT to be called with invalid token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mw := authMiddleware("valid-token")
	innerCalled := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/mcp", nil)
	handler.ServeHTTP(w, r)

	if innerCalled {
		t.Fatal("expected inner handler NOT to be called with missing token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleCompletion_ResourceLabelValues(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return []string{"__name__", "job", "instance", "namespace"}, nil, nil
		},
	}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "prom:///api/v1/label/{name}/values",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "name",
				Value: "ins",
			},
		},
	}

	result, err := handleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "instance" {
		t.Fatalf("expected [instance], got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_ResourceLabelValuesEmptyPrefix(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return []string{"__name__", "job", "instance"}, nil, nil
		},
	}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "prom:///api/v1/label/{name}/values",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "name",
				Value: "",
			},
		},
	}

	result, err := handleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(result.Completion.Values))
	}
}

func TestHandleCompletion_ResourceQueryMetricNames(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			if label == "__name__" {
				return model.LabelValues{"up", "node_cpu_seconds_total", "http_requests_total"}, nil, nil
			}
			return nil, nil, nil
		},
	}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "prom:///api/v1/query?query={promql}",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "promql",
				Value: "node",
			},
		},
	}

	result, err := handleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "node_cpu_seconds_total" {
		t.Fatalf("expected [node_cpu_seconds_total], got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_PromptMetricNames(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			if label == "__name__" {
				return model.LabelValues{"up", "node_cpu_seconds_total", "http_requests_total"}, nil, nil
			}
			return nil, nil, nil
		},
	}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/prompt",
				Name: "all-available-metrics",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "prefix",
				Value: "http",
			},
		},
	}

	result, err := handleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "http_requests_total" {
		t.Fatalf("expected [http_requests_total], got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_UnknownResource(t *testing.T) {
	mock := &mockapi.PrometheusAPI{}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "prom:///unknown",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "x",
				Value: "y",
			},
		},
	}

	result, err := handleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 0 {
		t.Fatalf("expected empty values, got %v", result.Completion.Values)
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
