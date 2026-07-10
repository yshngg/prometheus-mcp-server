package bindingblocks

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (b *binder) addPrompts() {
	b.server.AddPrompt(&mcp.Prompt{
		Name:        "all-available-metrics",
		Description: "List all available metrics in the Prometheus instance.",
	}, func(ctx context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role: "assistant",
					Content: &mcp.TextContent{
						Text: "List All Available Metrics is Equivalent to List Values of __name__ Label",
					},
				},
			},
		}, nil
	})
}
