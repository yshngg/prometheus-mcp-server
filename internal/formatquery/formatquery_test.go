package formatquery

import (
	"context"
	"errors"
	"testing"

	"github.com/yshngg/prometheus-mcp-server/internal/mock"
)

func TestFormatQueryHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		FormatQueryFunc: func(ctx context.Context, query string) (string, error) {
			return "up == 1", nil
		},
	}
	q := NewFormatQuerier(mock)
	_, result, err := q.FormatQueryHandler(context.Background(), nil, &FormatQueryArguments{Query: "up==1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FormattedQuery != "up == 1" {
		t.Fatalf("expected 'up == 1', got %s", result.FormattedQuery)
	}
}

func TestFormatQueryHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		FormatQueryFunc: func(ctx context.Context, query string) (string, error) {
			return "", errors.New("api error")
		},
	}
	q := NewFormatQuerier(mock)
	_, _, err := q.FormatQueryHandler(context.Background(), nil, &FormatQueryArguments{Query: "invalid"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatQueryHandler_DefaultMock(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	q := NewFormatQuerier(mock)
	_, result, err := q.FormatQueryHandler(context.Background(), nil, &FormatQueryArguments{Query: "up == 1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FormattedQuery != "" {
		t.Fatalf("expected empty result from default mock, got %s", result.FormattedQuery)
	}
}
