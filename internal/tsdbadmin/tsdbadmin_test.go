package tsdbadmin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yshngg/prometheus-mcp-server/internal/mock"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestSnapshotHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		SnapshotFunc: func(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) {
			return v1.SnapshotResult{Name: "snapshot-20250101"}, nil
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.SnapshotHandler(context.Background(), nil, &SnapshotParams{SkipHead: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "snapshot-20250101" {
		t.Fatalf("expected snapshot name, got %s", result.Name)
	}
}

func TestSnapshotHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		SnapshotFunc: func(ctx context.Context, skipHead bool) (v1.SnapshotResult, error) {
			return v1.SnapshotResult{}, errors.New("api error")
		},
	}
	a := NewTSDBAdmin(mock)
	_, _, err := a.SnapshotHandler(context.Background(), nil, &SnapshotParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteSeriesHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		DeleteSeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time) error {
			return nil
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.DeleteSeriesHandler(context.Background(), nil, &DeleteSeriesParams{
		Match: []string{"up"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestDeleteSeriesHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		DeleteSeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time) error {
			return errors.New("api error")
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.DeleteSeriesHandler(context.Background(), nil, &DeleteSeriesParams{
		Match: []string{"up"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestDeleteSeriesHandler_InvalidTime(t *testing.T) {
	mock := &mock.PrometheusAPI{
		DeleteSeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time) error {
			return nil
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.DeleteSeriesHandler(context.Background(), nil, &DeleteSeriesParams{
		Match: []string{"up"},
		Start: "invalid-time",
		End:   "also-invalid",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success despite invalid time")
	}
}

func TestDeleteSeriesHandler_WithTimeRange(t *testing.T) {
	mock := &mock.PrometheusAPI{
		DeleteSeriesFunc: func(ctx context.Context, matches []string, startTime, endTime time.Time) error {
			if startTime.IsZero() || endTime.IsZero() {
				t.Fatal("expected non-zero time range")
			}
			return nil
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.DeleteSeriesHandler(context.Background(), nil, &DeleteSeriesParams{
		Match: []string{"up"},
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-02T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestDeleteSeriesHandler_NoMatch(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	a := NewTSDBAdmin(mock)
	_, _, err := a.DeleteSeriesHandler(context.Background(), nil, &DeleteSeriesParams{})
	if err == nil {
		t.Fatal("expected error for empty match")
	}
}

func TestCleanTombstonesHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		CleanTombstonesFunc: func(ctx context.Context) error {
			return nil
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.CleanTombstonesHandler(context.Background(), nil, &CleanTombstonesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestCleanTombstonesHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		CleanTombstonesFunc: func(ctx context.Context) error {
			return errors.New("api error")
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.CleanTombstonesHandler(context.Background(), nil, &CleanTombstonesParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestTSDBBlocksHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TSDBBlocksFunc: func(ctx context.Context) (v1.TSDBBlocksResult, error) {
			return v1.TSDBBlocksResult{
				Status: "success",
				Data: v1.TSDBBlocksData{
					Blocks: []v1.TSDBBlocksBlockMetadata{
						{Ulid: "01JQZEXAMPLE"},
					},
				},
			}, nil
		},
	}
	a := NewTSDBAdmin(mock)
	_, result, err := a.TSDBBlocksHandler(context.Background(), nil, &TSDBBlocksParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "success" {
		t.Fatalf("expected success status, got %s", result.Status)
	}
	if len(result.Data.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(result.Data.Blocks))
	}
	if result.Data.Blocks[0].Ulid != "01JQZEXAMPLE" {
		t.Fatalf("expected Ulid 01JQZEXAMPLE, got %s", result.Data.Blocks[0].Ulid)
	}
}

func TestTSDBBlocksHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TSDBBlocksFunc: func(ctx context.Context) (v1.TSDBBlocksResult, error) {
			return v1.TSDBBlocksResult{}, errors.New("api error")
		},
	}
	a := NewTSDBAdmin(mock)
	_, _, err := a.TSDBBlocksHandler(context.Background(), nil, &TSDBBlocksParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}
