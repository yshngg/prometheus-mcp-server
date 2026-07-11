package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (m *manager) QuitHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error) {
	if err := m.api.Quit(ctx); err != nil {
		return nil, &ManagementResult{Success: false, Message: err.Error()}, nil
	}
	return nil, &ManagementResult{Success: true}, nil
}
