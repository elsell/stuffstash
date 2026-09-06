package httpserver

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type boundedProposalConversation struct {
	count int
	calls int
}

func (m *boundedProposalConversation) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.calls > 1 {
		return ports.ConversationModelTurn{Text: "I could not prepare that many changes."}, nil
	}
	commands := make([]any, 0, m.count)
	for i := 0; i < m.count; i++ {
		commands = append(commands, map[string]any{"id": fmt.Sprintf("create-%d", i), "kind": "create_asset", "summary": "Add an item", "arguments": map[string]any{"title": fmt.Sprintf("Item %d", i), "kind": "item"}})
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "proposal", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Add these items", "commands": commands}}}}, nil
}
func TestConversationProposalCommandLimitAtWebSocketBoundary(t *testing.T) {
	for _, count := range []int{10, 11} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			store := memory.NewStore()
			application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}, ids: []string{"session-id", "proposal-id"}}, store, memory.NewAuthorizer()).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: &boundedProposalConversation{count: count}})
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			terminal := "action.plan.proposed"
			if count > 10 {
				terminal = "session.completed"
			}
			events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", terminal)
			if hasRealtimeEvent(events, "action.plan.executed") {
				t.Fatal("commands executed without approval")
			}
			plan, found, err := store.ActionPlanByID(context.Background(), "tenant-home", "inventory-home", "proposal-id")
			if err != nil {
				t.Fatal(err)
			}
			if count > 10 {
				if found || hasRealtimeEvent(events, "action.plan.proposed") {
					t.Fatal("oversized proposal persisted or exposed")
				}
			} else if !found || len(plan.Commands) != count {
				t.Fatalf("valid bounded proposal lost: found=%v commands=%d", found, len(plan.Commands))
			}
		})
	}
}
