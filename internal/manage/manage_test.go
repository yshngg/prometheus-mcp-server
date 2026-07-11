package manage

import (
	"context"
	"errors"
	"testing"

	"github.com/yshngg/prometheus-mcp-server/internal/mock"
)

func TestHealthCheckHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	m := NewManager(mock)
	_, result, err := m.HealthCheckHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestHealthCheckHandler_Failure(t *testing.T) {
	mock := &mock.PrometheusAPI{
		HealthCheckFunc: func(ctx context.Context) error {
			return errors.New("not healthy")
		},
	}
	m := NewManager(mock)
	_, result, err := m.HealthCheckHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestReadinessCheckHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	m := NewManager(mock)
	_, result, err := m.ReadinessCheckHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestReadinessCheckHandler_Failure(t *testing.T) {
	mock := &mock.PrometheusAPI{
		ReadinessCheckFunc: func(ctx context.Context) error {
			return errors.New("not ready")
		},
	}
	m := NewManager(mock)
	_, result, err := m.ReadinessCheckHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestReloadHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	m := NewManager(mock)
	_, result, err := m.ReloadHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestReloadHandler_Failure(t *testing.T) {
	mock := &mock.PrometheusAPI{
		ReloadFunc: func(ctx context.Context) error {
			return errors.New("reload failed")
		},
	}
	m := NewManager(mock)
	_, result, err := m.ReloadHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}

func TestQuitHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{}
	m := NewManager(mock)
	_, result, err := m.QuitHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
}

func TestQuitHandler_Failure(t *testing.T) {
	mock := &mock.PrometheusAPI{
		QuitFunc: func(ctx context.Context) error {
			return errors.New("quit failed")
		},
	}
	m := NewManager(mock)
	_, result, err := m.QuitHandler(context.Background(), nil, struct{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
}
