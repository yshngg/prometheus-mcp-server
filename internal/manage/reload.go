package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (m *manager) ReloadHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error) {
	if err := m.api.Reload(ctx); err != nil {
		return nil, &ManagementResult{Success: false, Message: err.Error()}, nil
	}
	return nil, &ManagementResult{Success: true}, nil
}
