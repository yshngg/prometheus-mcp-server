package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type Manager interface {
	HealthCheckHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error)
	ReadinessCheckHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error)
	ReloadHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error)
	QuitHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error)
}

type manager struct {
	api promapi.PrometheusAPI
}

func NewManager(api promapi.PrometheusAPI) Manager {
	return &manager{api: api}
}

var _ Manager = &manager{}
