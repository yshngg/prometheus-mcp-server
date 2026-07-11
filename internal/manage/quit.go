package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/utils"
)

func (m *manager) QuitHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error) {
	if request != nil && request.Session != nil {
		confirmed, err := utils.ConfirmDestructive(ctx, request.Session, "Shutdown Prometheus", "Trigger a graceful shutdown of the Prometheus server.")
		if err == nil && !confirmed {
			return nil, &ManagementResult{Success: false, Message: "Shutdown cancelled by user"}, nil
		}
	}

	if err := m.api.Quit(ctx); err != nil {
		return nil, &ManagementResult{Success: false, Message: err.Error()}, nil
	}
	return nil, &ManagementResult{Success: true}, nil
}
