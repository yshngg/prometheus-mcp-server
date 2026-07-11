package tsdbadmin

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type CleanTombstonesParams struct{}

func (a *tsdbAdmin) CleanTombstonesHandler(ctx context.Context, request *mcp.CallToolRequest, input *CleanTombstonesParams) (*mcp.CallToolResult, *promapi.Result, error) {
	return nil, promapi.ResultOf(a.API.CleanTombstones(ctx)), nil
}
