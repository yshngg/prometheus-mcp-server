package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
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
	"github.com/yshngg/prometheus-mcp-server/internal/binding"
	"github.com/yshngg/prometheus-mcp-server/internal/mock"
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

func TestUsageFor_NoFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	capture := captureStderr(t, func() {
		usageFor(fs, "test")()
	})
	if !strings.Contains(capture, "VERSION") {
		t.Fatal("expected output to contain version")
	}
	if !strings.Contains(capture, "test") {
		t.Fatal("expected output to contain command name")
	}
}

func TestUsageFor_EmptyDefault(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.PanicOnError)
	fs.String("test-empty", "", "an empty default flag")
	capture := captureStderr(t, func() {
		usageFor(fs, "test")()
	})
	if !strings.Contains(capture, "VERSION") {
		t.Fatal("expected output to contain version")
	}
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
	mock := &mock.PrometheusAPI{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/readyz", nil)
	readyzHandler(mock)(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyzHandler_Unhealthy(t *testing.T) {
	mock := &mock.PrometheusAPI{
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
	mock := &mock.PrometheusAPI{
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

	result, err := binding.HandleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "instance" {
		t.Fatalf("expected [instance], got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_ResourceLabelValuesEmptyPrefix(t *testing.T) {
	mock := &mock.PrometheusAPI{
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

	result, err := binding.HandleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(result.Completion.Values))
	}
}

func TestHandleCompletion_ResourceQueryMetricNames(t *testing.T) {
	mock := &mock.PrometheusAPI{
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

	result, err := binding.HandleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "node_cpu_seconds_total" {
		t.Fatalf("expected [node_cpu_seconds_total], got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_PromptMetricNames(t *testing.T) {
	mock := &mock.PrometheusAPI{
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

	result, err := binding.HandleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "http_requests_total" {
		t.Fatalf("expected [http_requests_total], got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
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

	_, err := binding.HandleCompletion(context.Background(), req, mock)
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestHandleCompletion_APIErrorQuery(t *testing.T) {
	mock := &mock.PrometheusAPI{
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

	_, err := binding.HandleCompletion(context.Background(), req, mock)
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestHandleCompletion_NonMatchingPrompt(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	req := &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/prompt",
				Name: "unknown-prompt",
			},
			Argument: mcp.CompleteParamsArgument{Name: "x", Value: "y"},
		},
	}
	result, err := binding.HandleCompletion(context.Background(), req, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 0 {
		t.Fatalf("expected empty values, got %v", result.Completion.Values)
	}
}

func TestHandleCompletion_UnknownResource(t *testing.T) {
	mock := &mock.PrometheusAPI{}
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

	result, err := binding.HandleCompletion(context.Background(), req, mock)
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
	err := runStdio(ctx, server)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestNewServer_ValidPromAddr(t *testing.T) {
	// newServer creates an HTTP client but does not connect to Prometheus,
	// so even an unreachable address works here (connection happens lazily
	// on the first API call).
	server, promCli, err := newServer("http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if promCli == nil {
		t.Fatal("expected non-nil client")
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
	defer func() {
		if err := ss.Close(); err != nil {
			t.Logf("close server session: %v", err)
		}
	}()

	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() {
		if err := cs.Close(); err != nil {
			t.Logf("close client session: %v", err)
		}
	}()

	_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "delete-series"})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !handlerCalled {
		t.Fatal("expected handler to be called after elicitation confirmed")
	}
}

func TestRunHTTP_Ping(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{}
	binder := binding.NewBinder(server, mock)
	binder.Bind()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "localhost:9876"
	go func() {
		if err := runHTTP(ctx, server, mock, addr, ""); err != nil {
			t.Logf("runHTTP exited: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/ping")
	if err != nil {
		t.Fatalf("ping request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Fatalf("expected 'pong', got %q", string(body))
	}

	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		resp, err = http.Get("http://" + addr + path)
		if err != nil {
			t.Fatalf("%s request failed: %v", path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from %s, got %d", path, resp.StatusCode)
		}
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
}

// TestEndToEnd verifies the full MCP stack: server creation with middleware,
// tool/resource/prompt registration, and client communication via in-memory
// transport. Validates that sending middleware sets cache TTLs and that the
// destructive middleware does not block non-destructive tools.
func TestEndToEnd(t *testing.T) {
	ctx := context.Background()
	mock := &mock.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return &model.Vector{
				{Metric: model.Metric{"__name__": "up", "job": "test"}, Value: model.SampleValue(1)},
			}, nil, nil
		},
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return model.LabelValues{"up", "node_cpu"}, nil, nil
		},
		ConfigFunc: func(ctx context.Context) (v1.ConfigResult, error) {
			return v1.ConfigResult{YAML: "global:\n  scrape_interval: 15s"}, nil
		},
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "prometheus-mcp-server", Version: "1.0"}, &mcp.ServerOptions{
		SchemaCache: mcp.NewSchemaCache(),
		PageSize:    50,
	})
	server.AddReceivingMiddleware(metricsMiddleware)
	server.AddReceivingMiddleware(destructiveToolMiddleware)
	server.AddSendingMiddleware(cacheHintMiddleware)

	binder := binding.NewBinder(server, mock)
	binder.Bind()

	st, ct := mcp.NewInMemoryTransports()
	_, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "1.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() {
		if err := cs.Close(); err != nil {
			t.Logf("close client session: %v", err)
		}
	}()

	// ListTools
	toolsResult, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(toolsResult.Tools) == 0 {
		t.Fatal("expected at least 1 tool")
	}

	// CallTool — instant query
	callResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "instant-query",
		Arguments: map[string]any{"query": "up"},
	})
	if err != nil {
		t.Fatalf("call tool instant-query: %v", err)
	}
	if callResult.IsError {
		t.Fatal("expected instant-query to succeed")
	}

	// CallTool — non-destructive tool passes through middleware
	healthResult, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "health-check"})
	if err != nil {
		t.Fatalf("call tool health-check: %v", err)
	}
	if healthResult.IsError {
		t.Fatal("expected health-check to succeed")
	}

	// CallTool — destructive tool with no client elicitation support
	// should fall through (ConfirmDestructive returns (true, nil) on error)
	delResult, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "delete-series", Arguments: map[string]any{"match[]": []string{"up"}}})
	if err != nil {
		t.Fatalf("call tool delete-series: %v", err)
	}
	if delResult.IsError {
		t.Fatal("expected delete-series to fall through when elicitation unsupported")
	}

	// ListResources
	resourcesResult, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	if len(resourcesResult.Resources) == 0 {
		t.Fatal("expected at least 1 resource")
	}

	// ReadResource — config
	configResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///config"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(configResult.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(configResult.Contents))
	}
	if !strings.Contains(configResult.Contents[0].Text, "scrape_interval") {
		t.Fatalf("expected config content, got: %s", configResult.Contents[0].Text)
	}

	// ListPrompts
	promptsResult, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("list prompts: %v", err)
	}
	if len(promptsResult.Prompts) == 0 {
		t.Fatal("expected at least 1 prompt")
	}

	// GetPrompt — all-available-metrics
	promptResult, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "all-available-metrics"})
	if err != nil {
		t.Fatalf("get prompt: %v", err)
	}
	if len(promptResult.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(promptResult.Messages))
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

	if err := w.Close(); err != nil {
		t.Logf("close pipe: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read: %v", err)
	}
	return buf.String()
}
