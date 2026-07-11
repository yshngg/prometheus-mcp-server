package elicitation

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
	defer func() {
		if err := ss.Close(); err != nil {
			t.Logf("close server session: %v", err)
		}
	}()

	c := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "1.0"}, nil)
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() {
		if err := cs.Close(); err != nil {
			t.Logf("close client session: %v", err)
		}
	}()

	confirmed, err := ConfirmDestructive(ctx, ss, "test-action", "test detail")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmed=true when elicitation not supported")
	}
}
