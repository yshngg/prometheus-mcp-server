package tsdbadmin

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type TSDBBlocksParams struct{}
type TSDBBlocksResult = v1.TSDBBlocksResult

func (a *tsdbAdmin) TSDBBlocksHandler(ctx context.Context, request *mcp.CallToolRequest, input *TSDBBlocksParams) (*mcp.CallToolResult, *TSDBBlocksResult, error) {
	result, err := a.API.TSDBBlocks(ctx)
	if err != nil {
		return nil, nil, err
	}
	return nil, &result, nil
}
