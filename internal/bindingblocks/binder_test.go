package bindingblocks

import (
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

func TestAddResources_RegistersResources(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mockapi.PrometheusAPI{}
	b := &binder{server: server, api: mock}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("addResources panicked: %v", r)
		}
	}()
	b.addResources()
}

func TestPrompts_NoPanic(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)
	mock := &mockapi.PrometheusAPI{}
	b := &binder{server: server, api: mock}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("addPrompts panicked: %v", r)
		}
	}()
	b.addPrompts()
}
