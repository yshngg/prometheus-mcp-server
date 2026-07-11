package main

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
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
	go runHTTP(ctx, server, mock, addr)

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

	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestRunSSE_Ping(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mockapi.PrometheusAPI{}
	binder := bindingblocks.NewBinder(server, mock)
	binder.Bind()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "localhost:9877"
	go runSSE(ctx, server, addr)

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

	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestRunSSE_Endpoint(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	mock := &mockapi.PrometheusAPI{}
	binder := bindingblocks.NewBinder(server, mock)
	binder.Bind()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := "localhost:9878"
	go runSSE(ctx, server, addr)

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/mcp")
	if err != nil {
		t.Fatalf("mcp request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "endpoint") {
			return
		}
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
}


