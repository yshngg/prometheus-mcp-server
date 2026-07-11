package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

func (m *manager) HealthCheckHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error) {
	result := &promapi.Result{
		Success: true,
	}

	err := m.api.HealthCheck(ctx)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
	}
	return nil, result, nil
}
