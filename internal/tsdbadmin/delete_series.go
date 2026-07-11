package tsdbadmin

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
	"github.com/yshngg/prometheus-mcp-server/internal/timeutil"
	"k8s.io/klog/v2"
)

type DeleteSeriesParams struct {
	Match []string `json:"match[]" jsonschema:"<series_selector>: Repeated label matcher argument that selects the series to delete. At least one match[] argument must be provided."`
	Start string   `json:"start,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: Start timestamp. Optional and defaults to minimum possible time."`
	End   string   `json:"end,omitzero" jsonschema:"<rfc3339 | unix_timestamp>: End timestamp. Optional and defaults to maximum possible time."`
}

func (a *tsdbAdmin) DeleteSeriesHandler(ctx context.Context, request *mcp.CallToolRequest, input *DeleteSeriesParams) (*mcp.CallToolResult, *promapi.Result, error) {
	var (
		start, end time.Time
		err        error
	)
	if input.Start != "" {
		if start, err = timeutil.ParseTime(input.Start); err != nil {
			klog.InfoS("parse start time", "err", err)
		}
	}
	if input.End != "" {
		if end, err = timeutil.ParseTime(input.End); err != nil {
			klog.InfoS("parse end time", "err", err)
		}
	}

	if len(input.Match) == 0 {
		return nil, nil, fmt.Errorf("at least one match[] selector is required")
	}

	return nil, promapi.ResultOf(a.API.DeleteSeries(ctx, input.Match, start, end)), nil
}
