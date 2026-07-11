package elicitation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/klog/v2"
)

// ConfirmDestructive asks the client to confirm a destructive action via
// elicitation. It returns true if the action should proceed (client
// confirms, doesn't support elicitation, or an error occurs). Returns
// false if the client explicitly declines.
func ConfirmDestructive(ctx context.Context, session *mcp.ServerSession, action, detail string) (bool, error) {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"confirm": {
				"type": "boolean",
				"description": "Confirm this destructive action"
			}
		},
		"required": ["confirm"]
	}`)

	result, err := session.Elicit(ctx, &mcp.ElicitParams{
		Mode:    "form",
		Message: fmt.Sprintf("Confirm destructive action: %s\n\n%s", action, detail),
		RequestedSchema: schema,
	})
	if err != nil {
		klog.InfoS("elicitation not supported by client, proceeding with action", "action", action, "err", err)
		return true, nil
	}
	return result.Action == "accept", nil
}
