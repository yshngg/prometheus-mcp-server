package statusexpose

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestConfigExposeHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		ConfigFunc: func(ctx context.Context) (v1.ConfigResult, error) {
			return v1.ConfigResult{YAML: "global:\n  scrape_interval: 15s"}, nil
		},
	}
	e := NewStatusExposer(mock)
	_, result, err := e.ConfigExposeHandler(context.Background(), nil, &ConfigExposeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.YAML == "" {
		t.Fatal("expected non-empty YAML")
	}
}

func TestConfigExposeHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		ConfigFunc: func(ctx context.Context) (v1.ConfigResult, error) {
			return v1.ConfigResult{}, errors.New("api error")
		},
	}
	e := NewStatusExposer(mock)
	_, _, err := e.ConfigExposeHandler(context.Background(), nil, &ConfigExposeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFlagsExposeHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		FlagsFunc: func(ctx context.Context) (v1.FlagsResult, error) {
			return v1.FlagsResult{"flag1": "value1"}, nil
		},
	}
	e := NewStatusExposer(mock)
	_, result, err := e.FlagsExposeHandler(context.Background(), nil, &FlagsExposeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFlagsExposeHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		FlagsFunc: func(ctx context.Context) (v1.FlagsResult, error) {
			return nil, errors.New("api error")
		},
	}
	e := NewStatusExposer(mock)
	_, _, err := e.FlagsExposeHandler(context.Background(), nil, &FlagsExposeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildInformationExposeHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		BuildinfoFunc: func(ctx context.Context) (v1.BuildinfoResult, error) {
			return v1.BuildinfoResult{Version: "2.45.0"}, nil
		},
	}
	e := NewStatusExposer(mock)
	_, result, err := e.BuildInformationExposeHandler(context.Background(), nil, &BuildInformationExposeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "2.45.0" {
		t.Fatalf("expected version 2.45.0, got %s", result.Version)
	}
}

func TestBuildInformationExposeHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		BuildinfoFunc: func(ctx context.Context) (v1.BuildinfoResult, error) {
			return v1.BuildinfoResult{}, errors.New("api error")
		},
	}
	e := NewStatusExposer(mock)
	_, _, err := e.BuildInformationExposeHandler(context.Background(), nil, &BuildInformationExposeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRuntimeInformationExposeHandler_Success(t *testing.T) {
	startTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockapi.PrometheusAPI{
		RuntimeinfoFunc: func(ctx context.Context) (v1.RuntimeinfoResult, error) {
			return v1.RuntimeinfoResult{StartTime: startTime, CWD: "/prometheus"}, nil
		},
	}
	e := NewStatusExposer(mock)
	_, result, err := e.RuntimeInformationExposeHandler(context.Background(), nil, &RuntimeInformationExposeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CWD != "/prometheus" {
		t.Fatalf("expected CWD /prometheus, got %s", result.CWD)
	}
}

func TestRuntimeInformationExposeHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		RuntimeinfoFunc: func(ctx context.Context) (v1.RuntimeinfoResult, error) {
			return v1.RuntimeinfoResult{}, errors.New("api error")
		},
	}
	e := NewStatusExposer(mock)
	_, _, err := e.RuntimeInformationExposeHandler(context.Background(), nil, &RuntimeInformationExposeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTSDBStatsExposeHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		TSDBFunc: func(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error) {
			return v1.TSDBResult{
				HeadStats: v1.TSDBHeadStats{NumSeries: 100},
			}, nil
		},
	}
	e := NewStatusExposer(mock)
	_, result, err := e.TSDBStatsExposeHandler(context.Background(), nil, &TSDBStatsExposeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HeadStats.NumSeries != 100 {
		t.Fatalf("expected 100 series, got %d", result.HeadStats.NumSeries)
	}
}

func TestTSDBStatsExposeHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		TSDBFunc: func(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error) {
			return v1.TSDBResult{}, errors.New("api error")
		},
	}
	e := NewStatusExposer(mock)
	_, _, err := e.TSDBStatsExposeHandler(context.Background(), nil, &TSDBStatsExposeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWALReplayStatsExposeHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		WalReplayFunc: func(ctx context.Context) (v1.WalReplayStatus, error) {
			return v1.WalReplayStatus{Current: 1000}, nil
		},
	}
	e := NewStatusExposer(mock)
	_, result, err := e.WALReplayStatsExposeHandler(context.Background(), nil, &WALReplayStatsExposeParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Current != 1000 {
		t.Fatalf("expected 1000 current, got %d", result.Current)
	}
}

func TestWALReplayStatsExposeHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		WalReplayFunc: func(ctx context.Context) (v1.WalReplayStatus, error) {
			return v1.WalReplayStatus{}, errors.New("api error")
		},
	}
	e := NewStatusExposer(mock)
	_, _, err := e.WALReplayStatsExposeHandler(context.Background(), nil, &WALReplayStatsExposeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigExposeHandler_Cache(t *testing.T) {
	calls := 0
	mock := &mockapi.PrometheusAPI{
		ConfigFunc: func(ctx context.Context) (v1.ConfigResult, error) {
			calls++
			return v1.ConfigResult{YAML: "test"}, nil
		},
	}
	e := NewStatusExposer(mock)

	e.ConfigExposeHandler(context.Background(), nil, &ConfigExposeParams{})
	e.ConfigExposeHandler(context.Background(), nil, &ConfigExposeParams{})
	if calls != 1 {
		t.Fatalf("expected 1 API call (cached), got %d", calls)
	}
}
