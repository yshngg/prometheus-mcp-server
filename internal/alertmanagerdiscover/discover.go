package alertmanagerdiscover

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type AlertmanagerDiscoverer interface {
	AlertmanagerDiscoverHandler(ctx context.Context, request *mcp.CallToolRequest, input *AlertmanagerDiscoverParams) (*mcp.CallToolResult, *AlertmanagerDiscoverResult, error)
}

// NewAlertmanagerDiscoverer returns an AlertmanagerDiscoverer backed by the provided PrometheusAPI.
func NewAlertmanagerDiscoverer(api promapi.PrometheusAPI) AlertmanagerDiscoverer {
	return &alertmanagerDiscoverer{API: api}
}

type alertmanagerDiscoverer struct {
	API promapi.PrometheusAPI
}

var _ AlertmanagerDiscoverer = &alertmanagerDiscoverer{}
