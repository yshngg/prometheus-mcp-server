package metadataquery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func TestSeriesHandler_WithLimit(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		SeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
			if len(opts) != 1 {
				t.Fatalf("expected 1 option, got %d", len(opts))
			}
			return []model.LabelSet{{"job": "test"}}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.SeriesHandler(context.Background(), nil, &SeriesArguments{
		Match: []string{"up"},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelSets) != 1 {
		t.Fatalf("expected 1 label set, got %d", len(result.LabelSets))
	}
}

func TestSeriesHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		SeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
			return []model.LabelSet{{"job": "test"}}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.SeriesHandler(context.Background(), nil, &SeriesArguments{
		Match: []string{"up"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelSets) != 1 {
		t.Fatalf("expected 1 label set, got %d", len(result.LabelSets))
	}
}

func TestSeriesHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		SeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	q := NewMetadataQuerier(mock)
	_, _, err := q.SeriesHandler(context.Background(), nil, &SeriesArguments{
		Match: []string{"up"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLabelNamesHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return []string{"job", "instance"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelNames) != 2 {
		t.Fatalf("expected 2 label names, got %d", len(result.LabelNames))
	}
}

func TestLabelNamesHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	q := NewMetadataQuerier(mock)
	_, _, err := q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLabelNamesHandler_InvalidTime(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return []string{"job"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{
		Start: "invalid-time",
		End:   "also-invalid",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelNames) != 1 {
		t.Fatalf("expected 1 label name, got %d", len(result.LabelNames))
	}
}

func TestLabelNamesHandler_CacheHit(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			return []string{"job", "instance"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	calls := 0
	mock.LabelNamesFunc = func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
		calls++
		return []string{"job", "instance"}, nil, nil
	}

	_, _, _ = q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{})
	_, _, _ = q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{})
	if calls != 1 {
		t.Fatalf("expected 1 API call (cached), got %d", calls)
	}
}

func TestSeriesHandler_InvalidTime(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		SeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
			return []model.LabelSet{{"job": "test"}}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.SeriesHandler(context.Background(), nil, &SeriesArguments{
		Match: []string{"up"},
		Start: "bad-time",
		End:   "also-bad",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelSets) != 1 {
		t.Fatalf("expected 1 label set, got %d", len(result.LabelSets))
	}
}

func TestLabelValuesHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return model.LabelValues{"value1", "value2"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.LabelValuesHandler(context.Background(), nil, &LabelValuesArguments{
		Label: "job",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelValues) != 2 {
		t.Fatalf("expected 2 label values, got %d", len(result.LabelValues))
	}
}

func TestSeriesHandler_WithTimeAndLimit(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		SeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]model.LabelSet, v1.Warnings, error) {
			if len(opts) != 1 {
				t.Fatalf("expected 1 option, got %d", len(opts))
			}
			return []model.LabelSet{{"job": "test"}}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.SeriesHandler(context.Background(), nil, &SeriesArguments{
		Match: []string{"up"},
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-02T00:00:00Z",
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelSets) != 1 {
		t.Fatalf("expected 1 label set, got %d", len(result.LabelSets))
	}
}

func TestLabelNamesHandler_WithTimeAndLimit(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelNamesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time, opts ...v1.Option) ([]string, v1.Warnings, error) {
			if len(opts) != 1 {
				t.Fatalf("expected 1 option, got %d", len(opts))
			}
			return []string{"job", "instance"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-02T00:00:00Z",
		Limit: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelNames) != 2 {
		t.Fatalf("expected 2 label names, got %d", len(result.LabelNames))
	}
}

func TestLabelValuesHandler_WithTimeAndLimit(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			if len(opts) != 1 {
				t.Fatalf("expected 1 option, got %d", len(opts))
			}
			return model.LabelValues{"value1", "value2"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.LabelValuesHandler(context.Background(), nil, &LabelValuesArguments{
		Label: "job",
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-02T00:00:00Z",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelValues) != 2 {
		t.Fatalf("expected 2 label values, got %d", len(result.LabelValues))
	}
}

func TestLabelValuesHandler_InvalidTime(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return model.LabelValues{"v1"}, nil, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.LabelValuesHandler(context.Background(), nil, &LabelValuesArguments{
		Label: "job",
		Start: "bad",
		End:   "bad",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LabelValues) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result.LabelValues))
	}
}

func TestLabelValuesHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		LabelValuesFunc: func(ctx context.Context, label string, matches []string, startTime, endTime time.Time, opts ...v1.Option) (model.LabelValues, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	q := NewMetadataQuerier(mock)
	_, _, err := q.LabelValuesHandler(context.Background(), nil, &LabelValuesArguments{
		Label: "job",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTargetMetadataQueryHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		TargetsMetadataFunc: func(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error) {
			return []v1.MetricMetadata{{Metric: "up", Help: "test help"}}, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.TargetMetadataQueryHandler(context.Background(), nil, &TargetMetadataQueryParams{
		Metric: "up",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 metadata, got %d", len(result.Data))
	}
}

func TestTargetMetadataQueryHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		TargetsMetadataFunc: func(ctx context.Context, matchTarget, metric, limit string) ([]v1.MetricMetadata, error) {
			return nil, errors.New("api error")
		},
	}
	q := NewMetadataQuerier(mock)
	_, _, err := q.TargetMetadataQueryHandler(context.Background(), nil, &TargetMetadataQueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMetricsMetadataQueryHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		MetadataFunc: func(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error) {
			return map[string][]v1.Metadata{"up": {{Type: "counter", Help: "test"}}}, nil
		},
	}
	q := NewMetadataQuerier(mock)
	_, result, err := q.MetricsMetadataQueryHandler(context.Background(), nil, &MetricsMetadataQueryParams{
		Metric: "up",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(result.Data))
	}
}

func TestMetricsMetadataQueryHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		MetadataFunc: func(ctx context.Context, metric, limit string) (map[string][]v1.Metadata, error) {
			return nil, errors.New("api error")
		},
	}
	q := NewMetadataQuerier(mock)
	_, _, err := q.MetricsMetadataQueryHandler(context.Background(), nil, &MetricsMetadataQueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}
