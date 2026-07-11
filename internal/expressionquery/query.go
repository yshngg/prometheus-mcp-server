package expressionquery

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type ExpressionQuerier interface {
	InstantQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *InstantQueryArguments) (*mcp.CallToolResult, *InstantQueryResult, error)
	RangeQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *RangeQueryArguments) (*mcp.CallToolResult, *RangeQueryResult, error)
}

// NewExpressionQuerier creates and returns an ExpressionQuerier backed by the provided PrometheusAPI.
// The returned implementation delegates queries to the given API.
func NewExpressionQuerier(api promapi.PrometheusAPI) ExpressionQuerier {
	return &expressionQuerier{API: api}
}

type expressionQuerier struct {
	API promapi.PrometheusAPI
}

var _ ExpressionQuerier = &expressionQuerier{}
