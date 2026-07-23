package rulequery

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type RuleType string

const (
	RuleTypeAlert  RuleType = "alert"
	RuleTypeRecord RuleType = "record"
	RuleTypeAny    RuleType = "any"
)

// URL query parameters:
// type=alert|record: return only the alerting rules (e.g. type=alert) or the recording rules (e.g. type=record). When the parameter is absent or empty, no filtering is done.
// rule_name[]=<string>: only return rules with the given rule name. If the parameter is repeated, rules with any of the provided names are returned. If we've filtered out all the rules of a group, the group is not returned. When the parameter is absent or empty, no filtering is done.
// rule_group[]=<string>: only return rules with the given rule group name. If the parameter is repeated, rules with any of the provided rule group names are returned. When the parameter is absent or empty, no filtering is done.
// file[]=<string>: only return rules with the given filepath. If the parameter is repeated, rules with any of the provided filepaths are returned. When the parameter is absent or empty, no filtering is done.
// exclude_alerts=<bool>: only return rules, do not return active alerts.
// match[]=<label_selector>: only return rules that have configured labels that satisfy the label selectors. If the parameter is repeated, rules that match any of the sets of label selectors are returned. Note that matching is on the labels in the definition of each rule, not on the values after template expansion (for alerting rules). Optional.
// group_limit=<number>: The group_limit parameter allows you to specify a limit for the number of rule groups that is returned in a single response. If the total number of rule groups exceeds the specified group_limit value, the response will include a groupNextToken property. You can use the value of this groupNextToken property in subsequent requests in the group_next_token parameter to paginate over the remaining rule groups. The groupNextToken property will not be present in the final response, indicating that you have retrieved all the available rule groups. Please note that there are no guarantees regarding the consistency of the response if the rule groups are being modified during the pagination process.
// group_next_token: the pagination token that was returned in previous request when the group_limit property is set. The pagination token is used to iteratively paginate over a large number of rule groups. To use the group_next_token parameter, the group_limit parameter also need to be present. If a rule group that coincides with the next token is removed while you are paginating over the rule groups, a response with status code 400 will be returned.
// curl http://localhost:9090/api/v1/rules
type RuleQueryArguments struct {
	Type           RuleType `json:"type,omitzero" jsonschema:"alert|record: return only the alerting rules (e.g. type=alert) or the recording rules (e.g. type=record). When the parameter is absent or empty, no filtering is done."`
	RuleName       []string `json:"rule_name[],omitzero" jsonschema:"<string>: only return rules with the given rule name. If the parameter is repeated, rules with any of the provided names are returned. If we've filtered out all the rules of a group, the group is not returned. When the parameter is absent or empty, no filtering is done."`
	RuleGroup      []string `json:"rule_group[],omitzero" jsonschema:"<string>: only return rules with the given rule group name. If the parameter is repeated, rules with any of the provided rule group names are returned. When the parameter is absent or empty, no filtering is done."`
	File           []string `json:"file[],omitzero" jsonschema:"<string>: only return rules with the given filepath. If the parameter is repeated, rules with any of the provided filepaths are returned. When the parameter is absent or empty, no filtering is done."`
	ExcludeAlerts  bool     `json:"exclude_alerts,omitzero" jsonschema:"<bool>: only return rules, do not return active alerts."`
	Match          []string `json:"match[],omitzero" jsonschema:"<label_selector>: only return rules that have configured labels that satisfy the label selectors. If the parameter is repeated, rules that match any of the sets of label selectors are returned. Note that matching is on the labels in the definition of each rule, not on the values after template expansion (for alerting rules). Optional."`
	GroupLimit     int64    `json:"group_limit,omitzero" jsonschema:"<number>: The group_limit parameter allows you to specify a limit for the number of rule groups that is returned in a single response. If the total number of rule groups exceeds the specified group_limit value, the response will include a groupNextToken property. You can use the value of this groupNextToken property in subsequent requests in the group_next_token parameter to paginate over the remaining rule groups. The groupNextToken property will not be present in the final response, indicating that you have retrieved all the available rule groups. Please note that there are no guarantees regarding the consistency of the response if the rule groups are being modified during the pagination process."`
	GroupNextToken string   `json:"group_next_token,omitzero" jsonschema:"the pagination token that was returned in previous request when the group_limit property is set. The pagination token is used to iteratively paginate over a large number of rule groups. To use the group_next_token parameter, the group_limit parameter also need to be present. If a rule group that coincides with the next token is removed while you are paginating over the rule groups, a response with status code 400 will be returned."`
}

type RuleQueryResult = v1.RulesResult

func (q *ruleQuerier) RuleQueryHandler(ctx context.Context, request *mcp.CallToolRequest, input *RuleQueryArguments) (*mcp.CallToolResult, *RuleQueryResult, error) {
	// TODO: Implement filtering and pagination based on input (type, rule_name[], rule_group[], file[], exclude_alerts, group_limit, group_next_token)
	result, err := q.API.Rules(ctx, input.Match)
	if err != nil {
		return nil, nil, err
	}
	return nil, &result, nil
}
