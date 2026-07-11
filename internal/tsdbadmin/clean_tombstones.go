package tsdbadmin

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/utils"
)

type CleanTombstonesParams struct{}

type CleanTombstonesResult struct {
	Success bool   `json:"success" jsonschema:"Indicate the result of the management operation, true means success, false means failure"`
	Message string `json:"message,omitempty" jsonschema:"Explanation message when the operation fails."`
}

func (a *tsdbAdmin) CleanTombstonesHandler(ctx context.Context, request *mcp.CallToolRequest, input *CleanTombstonesParams) (*mcp.CallToolResult, *CleanTombstonesResult, error) {
	if request != nil && request.Session != nil {
		confirmed, err := utils.ConfirmDestructive(ctx, request.Session, "Clean Tombstones", "Remove deleted data from disk and clean up tombstones.")
		if err == nil && !confirmed {
			return nil, &CleanTombstonesResult{Success: false, Message: "Clean cancelled by user"}, nil
		}
	}

	result := &CleanTombstonesResult{Success: true}
	if err := a.API.CleanTombstones(ctx); err != nil {
		result.Success = false
		result.Message = err.Error()
	}
	return nil, result, nil
}
