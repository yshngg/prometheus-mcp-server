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

	q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{})
	q.LabelNamesHandler(context.Background(), nil, &LabelNamesArguments{})
	if calls != 1 {
		t.Fatalf("expected 1 API call (cached), got %d", calls)
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
