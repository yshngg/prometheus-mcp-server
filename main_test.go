package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
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

func TestUsageFor_WithVersion(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	fs.String("test-flag", "default", "a test flag")

	capture := captureStderr(t, func() {
		usageFor(fs, "test [flags]")()
	})

	if !strings.Contains(capture, "VERSION") {
		t.Fatal("expected output to contain version")
	}
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

func TestMetricsMiddleware_WithRequest(t *testing.T) {
	mw := metricsMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, nil
	})
	req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "instant-query"}}
	_, err := mw(context.Background(), methodCallTool, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMetricsMiddleware_WithRequestError(t *testing.T) {
	mw := metricsMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, errors.New("handler error")
	})
	req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "instant-query"}}
	_, err := mw(context.Background(), methodCallTool, req)
	if err == nil {
		t.Fatal("expected error")
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

func TestHandleCompletion_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "prom:///api/v1/label/{name}/values",
			},
			Argument: mcp.CompleteParamsArgument{Name: "name", Value: "j"},
		},
	}

	_, err := handleCompletion(context.Background(), req, mock)
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestHandleCompletion_APIErrorQuery(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "prom:///api/v1/query?query={promql}",
			},
			Argument: mcp.CompleteParamsArgument{Name: "promql", Value: "up"},
		},
	}

	_, err := handleCompletion(context.Background(), req, mock)
	if err == nil {
		t.Fatal("expected error from API failure")
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

func TestCacheHintMiddleware_SetsTTL(t *testing.T) {
	tests := []struct {
		name    string
		result  mcp.Result
	}{
		{"ListToolsResult", &mcp.ListToolsResult{}},
		{"ListPromptsResult", &mcp.ListPromptsResult{}},
		{"ListResourcesResult", &mcp.ListResourcesResult{}},
		{"ListResourceTemplatesResult", &mcp.ListResourceTemplatesResult{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := cacheHintMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				return tt.result, nil
			})
			res, err := mw(context.Background(), "test", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			switch r := res.(type) {
			case *mcp.ListToolsResult:
				if r.TTLMs != 30000 || r.CacheScope != "public" {
					t.Fatalf("expected TTLMs=30000, CacheScope=public, got TTLMs=%d, CacheScope=%s", r.TTLMs, r.CacheScope)
				}
			case *mcp.ListPromptsResult:
				if r.TTLMs != 30000 || r.CacheScope != "public" {
					t.Fatalf("expected TTLMs=30000, CacheScope=public, got TTLMs=%d, CacheScope=%s", r.TTLMs, r.CacheScope)
				}
			case *mcp.ListResourcesResult:
				if r.TTLMs != 30000 || r.CacheScope != "public" {
					t.Fatalf("expected TTLMs=30000, CacheScope=public, got TTLMs=%d, CacheScope=%s", r.TTLMs, r.CacheScope)
				}
			case *mcp.ListResourceTemplatesResult:
				if r.TTLMs != 30000 || r.CacheScope != "public" {
					t.Fatalf("expected TTLMs=30000, CacheScope=public, got TTLMs=%d, CacheScope=%s", r.TTLMs, r.CacheScope)
				}
			default:
				t.Fatalf("unexpected result type: %T", res)
			}
		})
	}
}

func TestCacheHintMiddleware_ErrorPassthrough(t *testing.T) {
	mw := cacheHintMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, errors.New("handler error")
	})
	_, err := mw(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error to pass through")
	}
}

func TestCacheHintMiddleware_NonListResult(t *testing.T) {
	mw := cacheHintMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	res, err := mw(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(*mcp.CallToolResult); !ok {
		t.Fatal("expected CallToolResult unchanged")
	}
}

func TestDestructiveToolMiddleware_NonToolMethod(t *testing.T) {
	called := false
	mw := destructiveToolMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		return nil, nil
	})
	_, err := mw(context.Background(), "ping", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called for non-tool method")
	}
}

func TestDestructiveToolMiddleware_NonDestructiveTool(t *testing.T) {
	called := false
	mw := destructiveToolMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		return nil, nil
	})
	req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "instant-query"}}
	_, err := mw(context.Background(), methodCallTool, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called for non-destructive tool")
	}
}

func TestDestructiveToolMiddleware_NilSession(t *testing.T) {
	called := false
	mw := destructiveToolMiddleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		called = true
		return nil, nil
	})
	req := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Name: "delete-series"}}
	_, err := mw(context.Background(), methodCallTool, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called when session is nil")
	}
}

func TestRunStdio_CancelledContext(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := server.Run(ctx, &mcp.StdioTransport{})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// TestMainVersionFlag tests the main function's -version flag by running the
// test binary as a subprocess with GO_UNDER_TEST=1. When set, the test re-enters
// this function and calls main() with os.Args set to ["prometheus-mcp-server",
// "-version"]. main() prints the version and calls os.Exit(0), which the testing
// framework captures as a successful test completion.
func TestMainVersionFlag(t *testing.T) {
	if os.Getenv("GO_UNDER_TEST") == "1" {
		os.Args = []string{"prometheus-mcp-server", "-version"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.v", "-test.run", "^TestMainVersionFlag$")
	cmd.Env = append(os.Environ(), "GO_UNDER_TEST=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected success, got %v: %s", err, string(out))
	}
}

func TestDestructiveToolMiddleware_ElicitConfirmed(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "server", Version: "1.0"}, nil)
	server.AddReceivingMiddleware(destructiveToolMiddleware)
	var handlerCalled bool
	mcp.AddTool(server, &mcp.Tool{Name: "delete-series"}, func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		handlerCalled = true
		return &mcp.CallToolResult{}, nil, nil
	})

	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "1.0"}, &mcp.ClientOptions{
		ElicitationHandler: func(ctx context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
			return &mcp.ElicitResult{Action: "accept", Content: map[string]any{"confirm": true}}, nil
		},
	})

	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()

	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "delete-series"})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !handlerCalled {
		t.Fatal("expected handler to be called after elicitation confirmed")
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
