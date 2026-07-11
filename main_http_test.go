package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/bindingblocks"
	"github.com/yshngg/prometheus-mcp-server/internal/mockapi"
)

func TestRunHTTP_Ping(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mockapi.PrometheusAPI{}
	binder := bindingblocks.NewBinder(server, mock)
	binder.Bind()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "localhost:9876"
	go func() {
		if err := runHTTP(ctx, server, mock, addr, ""); err != nil {
			t.Logf("runHTTP exited: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/ping")
	if err != nil {
		t.Fatalf("ping request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Fatalf("expected 'pong', got %q", string(body))
	}

	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		resp, err = http.Get("http://" + addr + path)
		if err != nil {
			t.Fatalf("%s request failed: %v", path, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from %s, got %d", path, resp.StatusCode)
		}
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
}