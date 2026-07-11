package expressionquery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func TestInstantQueryHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			if query != "up" {
				t.Fatalf("expected query 'up', got %q", query)
			}
			return &model.Scalar{}, nil, nil
		},
	}
	q := NewExpressionQuerier(mock)
	_, result, err := q.InstantQueryHandler(context.Background(), nil, &InstantQueryArguments{
		Query: "up",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstantQueryHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	q := NewExpressionQuerier(mock)
	_, _, err := q.InstantQueryHandler(context.Background(), nil, &InstantQueryArguments{
		Query: "up",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstantQueryHandler_WithTimeout(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			if len(opts) != 1 {
				t.Fatalf("expected 1 option, got %d", len(opts))
			}
			return &model.Scalar{}, nil, nil
		},
	}
	q := NewExpressionQuerier(mock)
	_, _, err := q.InstantQueryHandler(context.Background(), nil, &InstantQueryArguments{
		Query:   "up",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRangeQueryHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			if query != "up" {
				t.Fatalf("expected query 'up', got %q", query)
			}
			if r.Step != 15*time.Second {
				t.Fatalf("expected step 15s, got %v", r.Step)
			}
			return &model.Matrix{}, nil, nil
		},
	}
	q := NewExpressionQuerier(mock)
	_, result, err := q.RangeQueryHandler(context.Background(), nil, &RangeQueryArguments{
		Query: "up",
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-01T01:00:00Z",
		Step:  15,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestRangeQueryHandler_ZeroStep(t *testing.T) {
	mock := &mockapi.PrometheusAPI{}
	q := NewExpressionQuerier(mock)
	_, _, err := q.RangeQueryHandler(context.Background(), nil, &RangeQueryArguments{
		Query: "up",
		Step:  0,
	})
	if err == nil {
		t.Fatal("expected error for zero step")
	}
}

func TestInstantQueryHandler_InvalidTime(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryFunc: func(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return &model.Scalar{}, nil, nil
		},
	}
	q := NewExpressionQuerier(mock)
	_, _, err := q.InstantQueryHandler(context.Background(), nil, &InstantQueryArguments{
		Query: "up",
		Time:  "invalid-time",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRangeQueryHandler_InvalidTime(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return &model.Matrix{}, nil, nil
		},
	}
	q := NewExpressionQuerier(mock)
	_, _, err := q.RangeQueryHandler(context.Background(), nil, &RangeQueryArguments{
		Query: "up",
		Start: "invalid-start",
		End:   "invalid-end",
		Step:  15,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRangeQueryHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		QueryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return nil, nil, errors.New("api error")
		},
	}
	q := NewExpressionQuerier(mock)
	_, _, err := q.RangeQueryHandler(context.Background(), nil, &RangeQueryArguments{
		Query: "up",
		Start: "2025-01-01T00:00:00Z",
		End:   "2025-01-01T01:00:00Z",
		Step:  15,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
