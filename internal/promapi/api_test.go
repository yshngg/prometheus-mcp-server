package promapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
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

type mockAPI struct {
	PrometheusAPI
	labelNamesCallCount int
	LabelValuesFunc     func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error)
}

func NewMockAPI() *mockAPI {
	return &mockAPI{labelNamesCallCount: 0}
}

func (m *mockAPI) LabelNames(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
	m.labelNamesCallCount++
	return []string{"job"}, nil, nil
}

func (m *mockAPI) LabelValues(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
	if m.LabelValuesFunc != nil {
		return m.LabelValuesFunc(ctx, label, matches, startTime, endTime, opts...)
	}
	return nil, nil, nil
}

func (m *mockAPI) Config(ctx context.Context) (v1.ConfigResult, error) { return v1.ConfigResult{}, nil }
func (m *mockAPI) Flags(ctx context.Context) (v1.FlagsResult, error) { return v1.FlagsResult{}, nil }
func (m *mockAPI) HealthCheck(ctx context.Context) error { return nil }
func (m *mockAPI) ReadinessCheck(ctx context.Context) error { return nil }
func (m *mockAPI) Reload(ctx context.Context) error { return nil }
func (m *mockAPI) Quit(ctx context.Context) error { return nil }
func (m *mockAPI) Alerts(ctx context.Context) (v1.AlertsResult, error) { return v1.AlertsResult{}, nil }
func (m *mockAPI) AlertManagers(ctx context.Context) (v1.AlertManagersResult, error) { return v1.AlertManagersResult{}, nil }
func (m *mockAPI) Rules(ctx context.Context) (v1.RulesResult, error) { return v1.RulesResult{}, nil }
func (m *mockAPI) Targets(ctx context.Context) (v1.TargetsResult, error) { return v1.TargetsResult{}, nil }
func (m *mockAPI) TargetsMetadata(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error) { return nil, nil }
func (m *mockAPI) Metadata(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error) { return nil, nil }
func (m *mockAPI) CleanTombstones(ctx context.Context) error { return nil }
func (m *mockAPI) DeleteSeries(ctx context.Context, matches []string, startTime, endTime time.Time) error { return nil }
func (m *mockAPI) Snapshot(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) { return v1.SnapshotResult{}, nil }
func (m *mockAPI) TSDB(ctx context.Context, opts ...v1.Option) (v1.TSDBResult, error) { return v1.TSDBResult{}, nil }
func (m *mockAPI) WalReplay(ctx context.Context) (v1.WalReplayStatus, error) { return v1.WalReplayStatus{}, nil }
func (m *mockAPI) Buildinfo(ctx context.Context) (v1.BuildinfoResult, error) { return v1.BuildinfoResult{}, nil }
func (m *mockAPI) Runtimeinfo(ctx context.Context) (v1.RuntimeinfoResult, error) { return v1.RuntimeinfoResult{}, nil }
func (m *mockAPI) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) { return nil, nil, nil }
func (m *mockAPI) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) { return nil, nil, nil }
func (m *mockAPI) Series(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) { return nil, nil, nil }
func (m *mockAPI) QueryExemplars(ctx context.Context, query string, startTime, endTime time.Time) ([]v1.ExemplarQueryResult, error) { return nil, nil }

func TestCachingAPI_LabelNames(t *testing.T) {
	inner := NewMockAPI()
	caching := NewCachingAPI(inner, time.Hour).(*CachingPrometheusAPI)

	ctx := context.Background()
	names1, _, _ := caching.LabelNames(ctx, nil, time.Time{}, time.Time{})
	if len(names1) != 1 {
		t.Fatal("expected 1 label name")
	}
	names2, _, _ := caching.LabelNames(ctx, nil, time.Time{}, time.Time{})
	if len(names2) != 1 {
		t.Fatal("expected 1 label name")
	}
	if inner.labelNamesCallCount != 1 {
		t.Fatalf("expected 1 API call (cached), got %d", inner.labelNamesCallCount)
	}
}

func TestCachingAPI_LabelNamesWithFilter(t *testing.T) {
	inner := NewMockAPI()
	caching := NewCachingAPI(inner, time.Hour).(*CachingPrometheusAPI)

	ctx := context.Background()
	// With match filter — should bypass cache
	_, _, _ = caching.LabelNames(ctx, []string{"up"}, time.Time{}, time.Time{})
	_, _, _ = caching.LabelNames(ctx, []string{"up"}, time.Time{}, time.Time{})
	if inner.labelNamesCallCount != 2 {
		t.Fatalf("expected 2 API calls (no cache with filter), got %d", inner.labelNamesCallCount)
	}
}

func TestResultOf_Success(t *testing.T) {
	r := ResultOf(nil)
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Message != "" {
		t.Fatalf("expected empty message, got %s", r.Message)
	}
}

func TestResultOf_Error(t *testing.T) {
	r := ResultOf(errors.New("test error"))
	if r.Success {
		t.Fatal("expected failure")
	}
	if r.Message != "test error" {
		t.Fatalf("expected 'test error', got %s", r.Message)
	}
}

func TestCachingAPI_LabelValues(t *testing.T) {
	calls := 0
	inner := &mockAPI{}
	inner.LabelValuesFunc = func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
		calls++
		return model.LabelValues{"v1"}, nil, nil
	}
	caching := NewCachingAPI(inner, time.Hour).(*CachingPrometheusAPI)

	ctx := context.Background()
	_, _, _ = caching.LabelValues(ctx, "job", nil, time.Time{}, time.Time{})
	_, _, _ = caching.LabelValues(ctx, "job", nil, time.Time{}, time.Time{})
	if calls != 1 {
		t.Fatalf("expected 1 API call (cached), got %d", calls)
	}
}

func TestCachingAPI_PassthroughExtra(t *testing.T) {
	inner := NewMockAPI()
	caching := NewCachingAPI(inner, time.Hour).(*CachingPrometheusAPI)

	ctx := context.Background()
	if _, err := caching.Config(ctx); err != nil {
		t.Fatalf("Config: %v", err)
	}
	if _, err := caching.Flags(ctx); err != nil {
		t.Fatalf("Flags: %v", err)
	}
	if _, err := caching.TSDB(ctx); err != nil {
		t.Fatalf("TSDB: %v", err)
	}
	if _, err := caching.WalReplay(ctx); err != nil {
		t.Fatalf("WalReplay: %v", err)
	}
	if _, err := caching.Buildinfo(ctx); err != nil {
		t.Fatalf("Buildinfo: %v", err)
	}
	if _, err := caching.Runtimeinfo(ctx); err != nil {
		t.Fatalf("Runtimeinfo: %v", err)
	}
	if _, err := caching.Snapshot(ctx, false); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if err := caching.DeleteSeries(ctx, nil, time.Time{}, time.Time{}); err != nil {
		t.Fatalf("DeleteSeries: %v", err)
	}
	if _, _, err := caching.Query(ctx, "", time.Time{}); err != nil {
		t.Fatalf("Query: %v", err)
	}
	if _, _, err := caching.QueryRange(ctx, "", v1.Range{}, []v1.Option{}...); err != nil {
		t.Fatalf("QueryRange: %v", err)
	}
}

func TestCachingAPI_Passthrough(t *testing.T) {
	inner := NewMockAPI()
	caching := NewCachingAPI(inner, time.Hour).(*CachingPrometheusAPI)

	ctx := context.Background()
	// Verify non-cached methods pass through without error
	if _, err := caching.Alerts(ctx); err != nil {
		t.Fatalf("Alerts: %v", err)
	}
	if _, err := caching.AlertManagers(ctx); err != nil {
		t.Fatalf("AlertManagers: %v", err)
	}
	if _, err := caching.Rules(ctx); err != nil {
		t.Fatalf("Rules: %v", err)
	}
	if _, err := caching.Targets(ctx); err != nil {
		t.Fatalf("Targets: %v", err)
	}
	if _, err := caching.TargetsMetadata(ctx, "", "", ""); err != nil {
		t.Fatalf("TargetsMetadata: %v", err)
	}
	if _, err := caching.Metadata(ctx, "", ""); err != nil {
		t.Fatalf("Metadata: %v", err)
	}
	if err := caching.CleanTombstones(ctx); err != nil {
		t.Fatalf("CleanTombstones: %v", err)
	}
	if err := caching.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
	if err := caching.ReadinessCheck(ctx); err != nil {
		t.Fatalf("ReadinessCheck: %v", err)
	}
	if err := caching.Reload(ctx); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if err := caching.Quit(ctx); err != nil {
		t.Fatalf("Quit: %v", err)
	}
}
