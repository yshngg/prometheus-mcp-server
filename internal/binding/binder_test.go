package binding

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/mock"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func connectTestClient(t *testing.T, server *mcp.Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
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
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func TestNewBinder(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mock.PrometheusAPI{}
	b := NewBinder(server, mock)
	if b == nil {
		t.Fatal("expected non-nil binder")
	}
}

func TestBind_NoPanic(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mock.PrometheusAPI{}
	b := NewBinder(server, mock)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Bind() panicked: %v", r)
		}
	}()
	b.Bind()
}

func TestAddResources_RegistersResources(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mock.PrometheusAPI{}
	b := &binder{server: server, api: mock}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("addResources panicked: %v", r)
		}
	}()
	b.addResources()
}

func TestPrompts_NoPanic(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mock.PrometheusAPI{}
	b := &binder{server: server, api: mock}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("addPrompts panicked: %v", r)
		}
	}()
	b.addPrompts()
}

func TestPrompts_AllAvailableMetrics(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return model.LabelValues{"up", "node_cpu", "http_requests"}, nil, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addPrompts()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	result, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "all-available-metrics"})
	if err != nil {
		t.Fatalf("get prompt: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	tc, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Messages[0].Content)
	}
	text := tc.Text
	if !strings.Contains(text, "up") || !strings.Contains(text, "http_requests") {
		t.Fatalf("expected metric names in prompt, got: %s", text)
	}
}

func TestPrompts_AllAvailableMetricsWithPrefix(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return model.LabelValues{"up", "node_cpu", "node_memory", "http_requests"}, nil, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addPrompts()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	result, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "all-available-metrics",
		Arguments: map[string]string{"prefix": "node"},
	})
	if err != nil {
		t.Fatalf("get prompt: %v", err)
	}
	tc, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Messages[0].Content)
	}
	text := tc.Text
	if !strings.Contains(text, "node_cpu") {
		t.Fatalf("expected node_cpu in metrics, got: %s", text)
	}
	if !strings.Contains(text, "node_memory") {
		t.Fatalf("expected node_memory in metrics, got: %s", text)
	}
	if strings.Contains(text, "Total available metrics") {
		t.Fatalf("expected filtered output, not total, got: %s", text)
	}
}

func TestResources_ReadConfig(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		ConfigFunc: func(ctx context.Context) (v1.ConfigResult, error) {
			return v1.ConfigResult{YAML: "global:\n  scrape_interval: 15s"}, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///config"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "scrape_interval") {
		t.Fatalf("expected config content, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadFlags(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		FlagsFunc: func(ctx context.Context) (v1.FlagsResult, error) {
			return v1.FlagsResult{"storage.tsdb.retention.time": "15d"}, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///flags"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "15d") {
		t.Fatalf("expected flags content, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadBuildInfo(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		BuildinfoFunc: func(ctx context.Context) (v1.BuildinfoResult, error) {
			return v1.BuildinfoResult{Version: "2.45.0", Revision: "abc123"}, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///build-info"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "2.45.0") {
		t.Fatalf("expected build info content, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadResourceTemplateQuery_NoQuery(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///api/v1/query"})
	if err == nil {
		t.Fatal("expected error for missing query parameter")
	}
}

func TestResources_ReadResourceTemplateQuery_APIFailure(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return nil, nil, errors.New("query failed")
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///api/v1/query?query=up"})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestPrompts_AllAvailableMetricsAPIError(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	b := &binder{server: server, api: mock}
	b.addPrompts()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	_, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "all-available-metrics"})
	if err == nil {
		t.Fatal("expected prompt error on API failure")
	}
}

func TestResources_ReadResourceTemplateLabelValues_InvalidURI(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///api/v1/label//values"})
	if err == nil {
		t.Fatal("expected error for invalid label name URI")
	}
}

func TestResources_ReadResourceTemplateLabelValues_APIFailure(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///api/v1/label/job/values"})
	if err == nil {
		t.Fatal("expected error from API failure")
	}
}

func TestResources_ReadResourceTemplateQuery(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return &model.Vector{
				{Metric: model.Metric{"__name__": "up", "job": "test"}, Value: model.SampleValue(1), Timestamp: model.Now()},
			}, nil, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///api/v1/query?query=up"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "up") {
		t.Fatalf("expected query result, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadRuntimeInfo(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		RuntimeinfoFunc: func(ctx context.Context) (v1.RuntimeinfoResult, error) {
			return v1.RuntimeinfoResult{CWD: "/prometheus", StartTime: time.Now()}, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///runtime-info"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "/prometheus") {
		t.Fatalf("expected runtime info content, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadTSDBStats(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		TSDBFunc: func(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error) {
			return v1.TSDBResult{HeadStats: v1.TSDBHeadStats{NumSeries: 100}}, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///tsdb-stats"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "100") {
		t.Fatalf("expected tsdb stats content, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadWALReplayStats(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		WalReplayFunc: func(ctx context.Context) (v1.WalReplayStatus, error) {
			return v1.WalReplayStatus{Current: 500, Max: 1000}, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///wal-replay-stats"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "500") {
		t.Fatalf("expected wal replay stats, got: %s", res.Contents[0].Text)
	}
}

func TestResources_ReadResourceTemplateLabelValues(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mock.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return model.LabelValues{"job1", "job2"}, nil, nil
		},
	}
	b := &binder{server: server, api: mock}
	b.addResources()

	ctx := context.Background()
	cs := connectTestClient(t, server)

	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "prom:///api/v1/label/job/values"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(res.Contents))
	}
	if !strings.Contains(res.Contents[0].Text, "job1") {
		t.Fatalf("expected label values, got: %s", res.Contents[0].Text)
	}
}
