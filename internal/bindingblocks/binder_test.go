package bindingblocks

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
)

func TestNewBinder(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mockapi.PrometheusAPI{}
	b := NewBinder(server, mock)
	if b == nil {
		t.Fatal("expected non-nil binder")
	}
}

func TestBind_NoPanic(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mockapi.PrometheusAPI{}
	b := NewBinder(server, mock)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Bind() panicked: %v", r)
		}
	}()
	b.Bind()
}

func TestAddResources_NoOp(t *testing.T) {
	b := &binder{}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("addResources panicked: %v", r)
		}
	}()
	b.addResources()
}

func TestWithMetrics_Success(t *testing.T) {
	handler := withMetrics[any, any](func(ctx context.Context, request *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		return nil, "ok", nil
	})
	result, out, err := handler(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("expected 'ok', got %v", out)
	}
	_ = result
}

func TestWithMetrics_Error(t *testing.T) {
	handler := withMetrics[any, any](func(ctx context.Context, request *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		return nil, nil, errors.New("handler error")
	})
	_, _, err := handler(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAllAvailableMetricsHandler(t *testing.T) {
	result, err := allAvailableMetricsHandler(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0].Role != "assistant" {
		t.Fatalf("expected assistant role, got %s", result.Messages[0].Role)
	}
}
