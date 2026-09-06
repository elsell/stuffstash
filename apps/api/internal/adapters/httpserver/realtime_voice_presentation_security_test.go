package httpserver

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type invalidPresentationModel struct {
	arguments map[string]any
	calls     atomic.Int32
	rejected  atomic.Bool
}

func (m *invalidPresentationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if m.calls.Add(1) == 1 {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "bad-card", Name: "present_answer", Arguments: m.arguments}}}, nil
	}
	for _, message := range input.Messages {
		for _, result := range message.ToolResults {
			if result.CallID == "bad-card" && strings.Contains(result.Content, `"error"`) {
				m.rejected.Store(true)
			}
		}
	}
	return ports.ConversationModelTurn{Text: "I couldn't prepare those cards."}, nil
}
func TestPresentationToolRejectsScopeArgumentsAndUnobservedReferencesAtWebSocket(t *testing.T) {
	for _, key := range []string{"assetIds", "tenantId", "approved"} {
		t.Run(key, func(t *testing.T) {
			arguments := map[string]any{"spoken": "Private item", "display": "Private item"}
			switch key {
			case "assetIds":
				arguments[key] = []string{"other-tenant-private-asset"}
			case "tenantId":
				arguments[key] = "other-tenant"
			default:
				arguments[key] = true
			}
			model := &invalidPresentationModel{arguments: arguments}
			application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model})
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.completed")
			response := findRealtimeEvent(t, events, "assistant.response.completed")["response"].(map[string]any)
			if !model.rejected.Load() || model.calls.Load() != 2 || response["spokenResponse"] != "I couldn't prepare those cards." {
				t.Fatalf("invalid presentation was not rejected and repaired: %+v", response)
			}
			if artifacts, ok := response["artifacts"].([]any); ok && len(artifacts) != 0 {
				t.Fatalf("unobserved cards escaped: %+v", artifacts)
			}
		})
	}
}
