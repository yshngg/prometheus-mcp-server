package metadataquery

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/cache"
	"github.com/yshngg/prometheus-mcp-server/internal/prometheus/api"
)

type MetadataQuerier interface {
	SeriesHandler(ctx context.Context, request *mcp.CallToolRequest, input *SeriesArguments) (*mcp.CallToolResult, *SeriesResult, error)
	LabelNamesHandler(ctx context.Context, request *mcp.CallToolRequest, input *LabelNamesArguments) (*mcp.CallToolResult, *LabelNamesResult, error)
	LabelValuesHandler(ctx context.Context, request *mcp.CallToolRequest, input *LabelValuesArguments) (*mcp.CallToolResult, *LabelValuesResult, error)

	TargetMetadataQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *TargetMetadataQueryParams) (*mcp.CallToolResult, *TargetMetadataQueryResult, error)
	MetricsMetadataQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *MetricsMetadataQueryParams) (*mcp.CallToolResult, *MetricsMetadataQueryResult, error)
}

func NewMetadataQuerier(api api.PrometheusAPI) MetadataQuerier {
	return &metadataQuerier{
		API:   api,
		cache: cache.New(30 * time.Second),
	}
}

type metadataQuerier struct {
	API   api.PrometheusAPI
	cache *cache.Cache
}

var _ MetadataQuerier = &metadataQuerier{}
