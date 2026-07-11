package statusexpose

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type FlagsExposeParams struct{}

type FlagsExposeResult = v1.FlagsResult

func (e *statusExposer) FlagsExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *FlagsExposeParams) (*mcp.CallToolResult, *FlagsExposeResult, error) {
	if v, ok := e.cache.Get("flags"); ok {
		result := v.(FlagsExposeResult)
		return nil, &result, nil
	}
	result, err := e.API.Flags(ctx)
	if err != nil {
		return nil, nil, err
	}
	e.cache.Set("flags", result)
	return nil, &result, nil
}
