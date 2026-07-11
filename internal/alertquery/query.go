package alertquery

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type AlertQuerier interface {
	AlertQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *AlertQueryParams) (*mcp.CallToolResult, *AlertQueryResult, error)
}

// NewAlertQuerier returns an AlertQuerier backed by the provided PrometheusAPI.
// The concrete implementation is an *alertQuerier configured to use the given API.
func NewAlertQuerier(api promapi.PrometheusAPI) AlertQuerier {
	return &alertQuerier{API: api}
}

type alertQuerier struct {
	API promapi.PrometheusAPI
}

var _ AlertQuerier = &alertQuerier{}
