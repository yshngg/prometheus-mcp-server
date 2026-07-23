package rulequery

import (
	"context"
	"errors"
	"testing"

	"github.com/yshngg/prometheus-mcp-server/internal/mock"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestRuleQueryHandler_Success(t *testing.T) {
	mock := &mock.PrometheusAPI{
		RulesFunc: func(ctx context.Context, matches []string) (v1.RulesResult, error) {
			if matches != nil {
				t.Fatalf("expected nil matches, got %v", matches)
			}
			return v1.RulesResult{Groups: []v1.RuleGroup{{Name: "test"}}}, nil
		},
	}
	q := NewRuleQuerier(mock)
	_, result, err := q.RuleQueryHandler(context.Background(), nil, &RuleQueryArguments{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Groups) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(result.Groups))
	}
}

func TestRuleQueryHandler_WithMatch(t *testing.T) {
	mock := &mock.PrometheusAPI{
		RulesFunc: func(ctx context.Context, matches []string) (v1.RulesResult, error) {
			if len(matches) != 1 || matches[0] != `job="prometheus"` {
				t.Fatalf("unexpected matches: %v", matches)
			}
			return v1.RulesResult{Groups: []v1.RuleGroup{{Name: "test"}}}, nil
		},
	}
	q := NewRuleQuerier(mock)
	_, result, err := q.RuleQueryHandler(context.Background(), nil, &RuleQueryArguments{
		Match: []string{`job="prometheus"`},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Groups) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(result.Groups))
	}
}

func TestRuleQueryHandler_APIError(t *testing.T) {
	mock := &mock.PrometheusAPI{
		RulesFunc: func(ctx context.Context, matches []string) (v1.RulesResult, error) {
			if matches != nil {
				t.Fatalf("expected nil matches, got %v", matches)
			}
			return v1.RulesResult{}, errors.New("api error")
		},
	}
	q := NewRuleQuerier(mock)
	_, _, err := q.RuleQueryHandler(context.Background(), nil, &RuleQueryArguments{})
	if err == nil {
		t.Fatal("expected error")
	}
}
