package binding

import (
	"context"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

// HandleCompletion provides autocomplete suggestions for resource URI templates
// and prompt arguments. The MCP server calls this when the client requests
// completion/complete for a resource template variable or prompt argument.
func HandleCompletion(ctx context.Context, req *mcp.CompleteRequest, client promapi.PrometheusAPI) (*mcp.CompleteResult, error) {
	val := req.Params.Argument.Value

	switch req.Params.Ref.Type {
	case "ref/resource":
		uri := req.Params.Ref.URI
		switch {
		case strings.Contains(uri, "label/") && strings.Contains(uri, "/values"):
			names, _, err := client.LabelNames(ctx, nil, time.Time{}, time.Time{})
			if err != nil {
				return nil, err
			}
			var matches []string
			for _, n := range names {
				s := string(n)
				if val == "" || strings.HasPrefix(s, val) {
					matches = append(matches, s)
				}
			}
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values:  matches,
					HasMore: len(matches) > 20,
					Total:   len(matches),
				},
			}, nil

		case strings.Contains(uri, "query?query={promql}"):
			values, _, err := client.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
			if err != nil {
				return nil, err
			}
			var matches []string
			for _, v := range values {
				s := string(v)
				if val == "" || strings.HasPrefix(s, val) {
					matches = append(matches, s)
				}
			}
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values:  matches,
					HasMore: len(matches) > 20,
					Total:   len(matches),
				},
			}, nil
		}

	case "ref/prompt":
		if req.Params.Ref.Name == "all-available-metrics" && req.Params.Argument.Name == "prefix" {
			values, _, err := client.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
			if err != nil {
				return nil, err
			}
			var matches []string
			for _, v := range values {
				s := string(v)
				if val == "" || strings.HasPrefix(s, val) {
					matches = append(matches, s)
				}
			}
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values:  matches,
					HasMore: len(matches) > 20,
					Total:   len(matches),
				},
			}, nil
		}
	}

	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{Values: []string{}},
	}, nil
}
