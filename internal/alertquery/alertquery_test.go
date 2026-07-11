package alertquery

import (
	"context"
	"errors"
	"testing"

	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestAlertQueryHandler_Success(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		AlertsFunc: func(ctx context.Context) (v1.AlertsResult, error) {
			return v1.AlertsResult{Alerts: []v1.Alert{{Labels: nil}}}, nil
		},
	}
	q := NewAlertQuerier(mock)
	_, result, err := q.AlertQueryHandler(context.Background(), nil, &AlertQueryParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(result.Alerts))
	}
}

func TestAlertQueryHandler_APIError(t *testing.T) {
	mock := &mockapi.PrometheusAPI{
		AlertsFunc: func(ctx context.Context) (v1.AlertsResult, error) {
			return v1.AlertsResult{}, errors.New("api error")
		},
	}
	q := NewAlertQuerier(mock)
	_, _, err := q.AlertQueryHandler(context.Background(), nil, &AlertQueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}
