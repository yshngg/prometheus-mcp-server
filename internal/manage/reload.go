package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

func (m *manager) ReloadHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error) {
	return nil, promapi.ResultOf(m.api.Reload(ctx)), nil
}
