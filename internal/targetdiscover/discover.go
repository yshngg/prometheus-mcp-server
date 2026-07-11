package targetdiscover

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)


type TargetDiscoverer interface {
	TargetDiscoverHandler(ctx context.Context, request *mcp.CallToolRequest, input *TargetDiscoverParams) (*mcp.CallToolResult, *TargetDiscoverResult, error)
}

// NewTargetDiscoverer returns a TargetDiscoverer that uses the provided PrometheusAPI to perform target discovery.
func NewTargetDiscoverer(api promapi.PrometheusAPI) TargetDiscoverer {
	return &targetDiscoverer{API: api}
}

type targetDiscoverer struct {
	API promapi.PrometheusAPI
}

var _ TargetDiscoverer = &targetDiscoverer{}
