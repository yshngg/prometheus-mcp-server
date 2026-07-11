package statusexpose

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/cache"
	"github.com/yshngg/prometheus-mcp-server/internal/prometheus/api"
)

type StatusExposer interface {
	ConfigExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *ConfigExposeParams) (*mcp.CallToolResult, *ConfigExposeResult, error)
	FlagsExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *FlagsExposeParams) (*mcp.CallToolResult, *FlagsExposeResult, error)
	RuntimeInformationExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *RuntimeInformationExposeParams) (*mcp.CallToolResult, *RuntimeInformationExposeResult, error)
	BuildInformationExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *BuildInformationExposeParams) (*mcp.CallToolResult, *BuildInformationExposeResult, error)
	TSDBStatsExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *TSDBStatsExposeParams) (*mcp.CallToolResult, *TSDBStatsExposeResult, error)
	WALReplayStatsExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *WALReplayStatsExposeParams) (*mcp.CallToolResult, *WALReplayStatsExposeResult, error)
}

func NewStatusExposer(api api.PrometheusAPI) StatusExposer {
	return &statusExposer{
		API:   api,
		cache: cache.New(60 * time.Second),
	}
}

type statusExposer struct {
	API   api.PrometheusAPI
	cache *cache.Cache
}

var _ StatusExposer = &statusExposer{}
