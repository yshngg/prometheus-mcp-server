package statusexpose

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type RuntimeInformationExposeParams struct{}

type RuntimeInformationExposeResult = v1.RuntimeinfoResult

func (e *statusExposer) RuntimeInformationExposeHandler(ctx context.Context, request *mcp.CallToolRequest, input *RuntimeInformationExposeParams) (*mcp.CallToolResult, *RuntimeInformationExposeResult, error) {
	if v, ok := e.cache.Get("runtimeinfo"); ok {
		result := v.(RuntimeInformationExposeResult)
		return nil, &result, nil
	}
	result, err := e.API.Runtimeinfo(ctx)
	if err != nil {
		return nil, nil, err
	}
	e.cache.Set("runtimeinfo", result)
	return nil, &result, nil
}
