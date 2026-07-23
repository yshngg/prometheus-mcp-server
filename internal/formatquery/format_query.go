package formatquery

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FormatQueryArguments struct {
	Query string `json:"query" jsonschema:"<string>: Prometheus expression query string."`
}

type FormatQueryResult struct {
	FormattedQuery string `json:"formatted_query" jsonschema:"<string>: The formatted/beautified PromQL expression."`
}

func (q *formatQuerier) FormatQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *FormatQueryArguments) (*mcp.CallToolResult, *FormatQueryResult, error) {
	result, err := q.API.FormatQuery(ctx, input.Query)
	if err != nil {
		return nil, nil, err
	}
	return nil, &FormatQueryResult{FormattedQuery: result}, nil
}
