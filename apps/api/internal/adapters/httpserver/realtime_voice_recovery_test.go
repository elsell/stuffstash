package httpserver

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryRejectsMalformedConversationOutput(t *testing.T) {
	for _, tc := range []struct {
		name string
		turn ports.ConversationModelTurn
	}{
		{name: "empty turn"},
		{name: "answer with conflicting tool calls", turn: ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "Done", Display: "Done"}, ToolCalls: []ports.AgentToolCall{{ID: "bad-call", Name: "propose_inventory_change", Arguments: map[string]any{}}}}},
		{name: "oversized answer", turn: ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: strings.Repeat("x", 100000), Display: "Too much speech"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			application := newSeededTestAppWithVoice(t, seededState{
				tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
				inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
				ids:         []string{"voice-session-id", "response-id"},
			}, fakeSpeechToText{transcript: "Where is my water bottle?"}, malformedConversationModel{turn: tc.turn}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			t.Cleanup(server.Close)
			events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.failed")
			failed := findRealtimeEvent(t, events, "session.failed")
			if failed["code"] != "invalid_request" {
				t.Fatalf("malformed model output did not fail safely: %+v", failed)
			}
			assertNoRealtimeEventType(t, events, "action.plan.proposed")
			assertNoRealtimeEventType(t, events, "assistant.response.completed")
			assertNoRealtimeEventType(t, events, "tool.call.started")
		})
	}
}

type malformedConversationModel struct{ turn ports.ConversationModelTurn }

func (m malformedConversationModel) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return m.turn, nil
}

// Removed with the legacy provider injection signature.
func (malformedConversationModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
}
