package alertmanagerdiscover

import (
	"context"
	"errors"
	"testing"

	"github.com/yshngg/prometheus-mcp-server/internal/mock"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestAlertmanagerDiscoverHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		AlertManagersFunc: func(ctx context.Context) (v1.AlertManagersResult, error) {
			return v1.AlertManagersResult{Active: []v1.AlertManager{{URL: "http://alertmanager:9093"}}}, nil
		},
	}
	d := NewAlertmanagerDiscoverer(mock)
	_, result, err := d.AlertmanagerDiscoverHandler(context.Background(), nil, &AlertmanagerDiscoverParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Active) != 1 {
		t.Fatalf("expected 1 active alertmanager, got %d", len(result.Active))
	}
}

func TestAlertmanagerDiscoverHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		AlertManagersFunc: func(ctx context.Context) (v1.AlertManagersResult, error) {
			return v1.AlertManagersResult{}, errors.New("api error")
		},
	}
	d := NewAlertmanagerDiscoverer(mock)
	_, _, err := d.AlertmanagerDiscoverHandler(context.Background(), nil, &AlertmanagerDiscoverParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}
