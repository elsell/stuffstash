package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type nativeHistoryModel struct {
	calls               atomic.Int32
	evidence, signature atomic.Bool
}

func (m *nativeHistoryModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	call := m.calls.Add(1)
	if call == 1 {
		return ports.ConversationModelTurn{ProviderState: []byte("private-native-signature"), ToolCalls: []ports.AgentToolCall{{ID: "search-history", Name: "search_authorized_assets", Arguments: map[string]any{"query": "Drill"}}}}, nil
	}
	if call == 2 {
		return ports.ConversationModelTurn{Text: "I found no matching Drill."}, nil
	}
	for _, message := range input.Messages {
		if string(message.ProviderState) == "private-native-signature" {
			m.signature.Store(true)
		}
		for _, result := range message.ToolResults {
			if result.CallID == "search-history" {
				m.evidence.Store(true)
			}
		}
	}
	return ports.ConversationModelTurn{Text: "The earlier search did not find a Drill."}, nil
}
func TestNativeVoiceWebSocketPreservesPrivateHistoryAcrossUtterances(t *testing.T) {
	model := &nativeHistoryModel{}
	application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model, speech: &scriptedSpeechToText{transcripts: []string{"Do I have a Drill?", "What did you find?"}}})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
	start["conversationContinuity"] = true
	writeRealtimeMessage(t, ctx, connection, start)
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("start: %+v", started)
	}
	sessionID := started["sessionId"].(string)
	for turn := 0; turn < 2; turn++ {
		writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2+turn*2, "utterance-"+string(rune('a'+turn)))
		events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
		assertSafeRealtimeEvents(t, events, []string{"private-native-signature"})
		completed := findRealtimeEvent(t, events, "session.completed")
		if completed["followUpAvailable"] != true {
			t.Fatalf("follow-up closed early: %+v", completed)
		}
	}
	if model.calls.Load() != 3 || !model.evidence.Load() || !model.signature.Load() {
		t.Fatalf("native history lost: calls=%d evidence=%v signature=%v", model.calls.Load(), model.evidence.Load(), model.signature.Load())
	}
}
