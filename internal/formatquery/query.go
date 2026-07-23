package formatquery

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type FormatQuerier interface {
	FormatQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *FormatQueryArguments) (*mcp.CallToolResult, *FormatQueryResult, error)
}

func NewFormatQuerier(api promapi.PrometheusAPI) FormatQuerier {
	return &formatQuerier{API: api}
}

type formatQuerier struct {
	API promapi.PrometheusAPI
}

var _ FormatQuerier = &formatQuerier{}
