package tsdbadmin

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/utils"
)

type DeleteSeriesParams struct {
	Match []string `json:"match[]" jsonschema:"<series_selector>: Repeated label matcher argument that selects the series to delete. At least one match[] argument must be provided."`
	Start string   `json:"start,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: Start timestamp. Optional and defaults to minimum possible time."`
	End   string   `json:"end,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: End timestamp. Optional and defaults to maximum possible time."`
}

type DeleteSeriesResult struct {
	Success bool   `json:"success" jsonschema:"Indicate the result of the management operation, true means success, false means failure"`
	Message string `json:"message,omitempty" jsonschema:"Explanation message when the operation fails."`
}

func (a *tsdbAdmin) DeleteSeriesHandler(ctx context.Context, request *mcp.CallToolRequest, input *DeleteSeriesParams) (*mcp.CallToolResult, *DeleteSeriesResult, error) {
	var (
		start, end time.Time
		err        error
	)
	if input.Start != "" {
		if start, err = utils.ParseTime(input.Start); err != nil {
			slog.Warn("parse start time", "err", err)
		}
	}
	if input.End != "" {
		if end, err = utils.ParseTime(input.End); err != nil {
			slog.Warn("parse end time", "err", err)
		}
	}

	if len(input.Match) == 0 {
		return nil, nil, fmt.Errorf("at least one match[] selector is required")
	}

	result := &DeleteSeriesResult{Success: true}
	if err = a.API.DeleteSeries(ctx, input.Match, start, end); err != nil {
		result.Success = false
		result.Message = err.Error()
	}
	return nil, result, nil
}
