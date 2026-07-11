package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type ManagementResult struct {
	Success bool   `json:"success" jsonschema:"Indicate the result of the management operation, true means success, false means failure"`
	Message string `json:"message,omitempty" jsonschema:"Explanation message when the operation fails."`
}

type Manager interface {
	HealthCheckHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error)
	ReadinessCheckHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error)
	ReloadHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error)
	QuitHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error)
}

type manager struct {
	api promapi.PrometheusAPI
}

func NewManager(api promapi.PrometheusAPI) Manager {
	return &manager{api: api}
}

var _ Manager = &manager{}
