package rulequery

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type RuleQuerier interface {
	RuleQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *RuleQueryArguments) (*mcp.CallToolResult, *RuleQueryResult, error)
}

// NewRuleQuerier returns a RuleQuerier implementation that uses the provided PrometheusAPI to execute rule queries.
func NewRuleQuerier(api promapi.PrometheusAPI) RuleQuerier {
	return &ruleQuerier{API: api}
}

type ruleQuerier struct {
	API promapi.PrometheusAPI
}

var _ RuleQuerier = &ruleQuerier{}
