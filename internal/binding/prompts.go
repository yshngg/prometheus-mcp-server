package binding

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (b *binder) addPrompts() {
	b.server.AddPrompt(&mcp.Prompt{
		Name:        "all-available-metrics",
		Description: "List all available metrics in the Prometheus instance.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "prefix",
				Description: "Optional prefix to filter metric names.",
				Required:    false,
			},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prefix := req.Params.Arguments["prefix"]

		values, warnings, err := b.api.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
		if err != nil {
			return nil, fmt.Errorf("list metric names: %w", err)
		}

		var metrics []string
		for _, v := range values {
			name := string(v)
			if prefix == "" || strings.HasPrefix(name, prefix) {
				metrics = append(metrics, name)
			}
		}

		text := fmt.Sprintf("Total available metrics: %d", len(values))
		if prefix != "" {
			text = fmt.Sprintf("Metrics matching prefix %q: %d (showing up to 100)", prefix, len(metrics))
		}
		if len(metrics) > 0 {
			if len(metrics) > 100 {
				metrics = metrics[:100]
			}
			text += ":\n" + strings.Join(metrics, "\n")
		}
		if len(warnings) > 0 {
			text += fmt.Sprintf("\n\nWarnings: %v", warnings)
		}

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role: "assistant",
					Content: &mcp.TextContent{
						Text: text,
					},
				},
			},
		}, nil
	})
}
