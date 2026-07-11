package targetdiscover

import (
	"context"
	"errors"
	"testing"

	"github.com/yshngg/prometheus-mcp-server/internal/mock"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func TestTargetDiscoverHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TargetsFunc: func(ctx context.Context) (v1.TargetsResult, error) {
			return v1.TargetsResult{
				Active: []v1.ActiveTarget{
					{Labels: model.LabelSet{"job": "test"}, ScrapePool: "test-pool"},
				},
				Dropped: []v1.DroppedTarget{
					{DiscoveredLabels: map[string]string{"job": "dropped"}},
				},
			}, nil
		},
	}
	d := NewTargetDiscoverer(mock)
	_, result, err := d.TargetDiscoverHandler(context.Background(), nil, &TargetDiscoverParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Active) != 1 {
		t.Fatalf("expected 1 active target, got %d", len(result.Active))
	}
	if len(result.Dropped) != 1 {
		t.Fatalf("expected 1 dropped target, got %d", len(result.Dropped))
	}
}

func TestTargetDiscoverHandler_FilterByScrapePool(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TargetsFunc: func(ctx context.Context) (v1.TargetsResult, error) {
			return v1.TargetsResult{
				Active: []v1.ActiveTarget{
					{Labels: model.LabelSet{"job": "test1"}, ScrapePool: "pool-a"},
					{Labels: model.LabelSet{"job": "test2"}, ScrapePool: "pool-b"},
				},
			}, nil
		},
	}
	d := NewTargetDiscoverer(mock)
	_, result, err := d.TargetDiscoverHandler(context.Background(), nil, &TargetDiscoverParams{
		ScrapePool: "pool-a",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Active) != 1 {
		t.Fatalf("expected 1 active target after filter, got %d", len(result.Active))
	}
}

func TestTargetDiscoverHandler_FilterByActive(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TargetsFunc: func(ctx context.Context) (v1.TargetsResult, error) {
			return v1.TargetsResult{
				Active:  []v1.ActiveTarget{{Labels: model.LabelSet{"job": "test"}}},
				Dropped: []v1.DroppedTarget{{DiscoveredLabels: map[string]string{"job": "dropped"}}},
			}, nil
		},
	}
	d := NewTargetDiscoverer(mock)
	_, result, err := d.TargetDiscoverHandler(context.Background(), nil, &TargetDiscoverParams{
		State: TargetStateActive,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Active) != 1 {
		t.Fatalf("expected 1 active target, got %d", len(result.Active))
	}
	if result.Dropped != nil {
		t.Fatal("expected dropped to be nil for active filter")
	}
}

func TestTargetDiscoverHandler_FilterByDropped(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TargetsFunc: func(ctx context.Context) (v1.TargetsResult, error) {
			return v1.TargetsResult{
				Active:  []v1.ActiveTarget{{Labels: model.LabelSet{"job": "test"}}},
				Dropped: []v1.DroppedTarget{{DiscoveredLabels: map[string]string{"job": "dropped"}}},
			}, nil
		},
	}
	d := NewTargetDiscoverer(mock)
	_, result, err := d.TargetDiscoverHandler(context.Background(), nil, &TargetDiscoverParams{
		State: TargetStateDropped,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Active != nil {
		t.Fatal("expected active to be nil for dropped filter")
	}
}

func TestTargetDiscoverHandler_InvalidState(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TargetsFunc: func(ctx context.Context) (v1.TargetsResult, error) {
			return v1.TargetsResult{}, nil
		},
	}
	d := NewTargetDiscoverer(mock)
	_, _, err := d.TargetDiscoverHandler(context.Background(), nil, &TargetDiscoverParams{
		State: "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
}

func TestTargetDiscoverHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		TargetsFunc: func(ctx context.Context) (v1.TargetsResult, error) {
			return v1.TargetsResult{}, errors.New("api error")
		},
	}
	d := NewTargetDiscoverer(mock)
	_, _, err := d.TargetDiscoverHandler(context.Background(), nil, &TargetDiscoverParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}
