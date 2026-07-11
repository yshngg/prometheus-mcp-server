package statusexpose

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type ConfigExposeParams struct{}

type ConfigExposeResult = v1.ConfigResult

func (e *statusExposer) ConfigExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *ConfigExposeParams) (*mcp.CallToolResult, *ConfigExposeResult, error) {
	if v, ok := e.cache.Get("config"); ok {
		result := v.(ConfigExposeResult)
		return nil, &result, nil
	}
	result, err := e.API.Config(ctx)
	if err != nil {
		return nil, nil, err
	}
	e.cache.Set("config", result)
	return nil, &result, nil
}
