package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

func (m *manager) QuitHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *promapi.Result, error) {
	if err := m.api.Quit(ctx); err != nil {
		return nil, &promapi.Result{Success: false, Message: err.Error()}, nil
	}
	return nil, &promapi.Result{Success: true}, nil
}
