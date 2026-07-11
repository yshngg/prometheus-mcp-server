package utils

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestConfirmDestructive_ElicitNotSupported(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, nil)
	ctx := context.Background()

	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer func() { _ = ss.Close() }()

	c := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "1.0"}, nil)
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = cs.Close() }()

	confirmed, err := ConfirmDestructive(ctx, ss, "test-action", "test detail")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmed=true when elicitation not supported")
	}
}
