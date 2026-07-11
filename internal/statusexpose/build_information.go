package statusexpose

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type BuildInformationExposeParams struct{}

type BuildInformationExposeResult = v1.BuildinfoResult

func (e *statusExposer) BuildInformationExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *BuildInformationExposeParams) (*mcp.CallToolResult, *BuildInformationExposeResult, error) {
	if v, ok := e.cache.Get("buildinfo"); ok {
		result := v.(BuildInformationExposeResult)
		return nil, &result, nil
	}
	result, err := e.API.Buildinfo(ctx)
	if err != nil {
		return nil, nil, err
	}
	e.cache.Set("buildinfo", result)
	return nil, &result, nil
}
