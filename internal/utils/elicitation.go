package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
		return true, nil
	}
	return result.Action == "accept", nil
}
