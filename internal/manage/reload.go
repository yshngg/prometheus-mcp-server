package manage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/utils"
)

func (m *manager) ReloadHandler(ctx context.Context, request *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ManagementResult, error) {
	if request != nil && request.Session != nil {
		confirmed, err := utils.ConfirmDestructive(ctx, request.Session, "Reload Prometheus", "Trigger a reload of configuration and rule files.")
		if err == nil && !confirmed {
			return nil, &ManagementResult{Success: false, Message: "Reload cancelled by user"}, nil
		}
	}

	if err := m.api.Reload(ctx); err != nil {
		return nil, &ManagementResult{Success: false, Message: err.Error()}, nil
	}
	return nil, &ManagementResult{Success: true}, nil
}
