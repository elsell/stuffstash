package agentmodel

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestConversationMemoryBoundsOpaqueStateAndDoesNotResumeTruncatedHistory(t *testing.T) {
	scope := ConversationScope{SessionID: "session", PrincipalID: "owner", TenantID: "home", InventoryID: "inventory"}
	memory := NewConversationMemory(scope, 512)
	if _, err := memory.Acquire(context.Background(), scope); err != nil {
		t.Fatal(err)
	}
	err := memory.Commit([]ports.ConversationMessage{{Role: ports.ConversationRoleAssistant, Text: "An answer", ProviderState: make([]byte, 513)}})
	memory.Release()
	if !errors.Is(err, ErrConversationContextExhausted) {
		t.Fatalf("opaque state escaped the cap: %v", err)
	}
	if _, err := memory.Acquire(context.Background(), scope); !errors.Is(err, ErrConversationContextExhausted) {
		t.Fatalf("truncated history became reusable: %v", err)
	}
}
func TestConversationMemoryWaitCanBeCancelledWithoutReleasingAnotherTurn(t *testing.T) {
	scope := ConversationScope{SessionID: "session"}
	memory := NewConversationMemory(scope, 512)
	if _, err := memory.Acquire(context.Background(), scope); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := memory.Acquire(ctx, scope); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled turn acquired busy history: %v", err)
	}
	if err := memory.Commit([]ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Find my Drill"}}); err != nil {
		t.Fatal(err)
	}
	memory.Release()
	messages, err := memory.Acquire(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	defer memory.Release()
	if len(messages) != 1 || messages[0].Text != "Find my Drill" {
		t.Fatal("cancelled waiter disturbed the active turn")
	}
}
func TestConversationContextLimitStopsProviderDisclosure(t *testing.T) {
	model := &conversationProposalModel{}
	_, err := RunConversation(context.Background(), model, &conversationProposalTools{}, ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Find a Drill"}}}, ConversationLimits{ModelCalls: 2, ToolCalls: 2, ContextBytes: 1})
	if !errors.Is(err, ErrConversationContextExhausted) || model.calls != 0 {
		t.Fatalf("oversized context reached provider: calls=%d err=%v", model.calls, err)
	}
}
