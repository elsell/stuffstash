package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type continuousVoiceModel struct {
	inputs []ports.ConversationModelInput
}

func (m *continuousVoiceModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.inputs = append(m.inputs, input)
	return httpConversationRead(input, app.RealtimeVoiceToolSearchAuthorizedAssets, map[string]any{"query": "tools"}, nil)
}

func TestRealtimeContinuityRetainsAnswerContextAndEndsAtLimit(t *testing.T) {
	language := &continuousVoiceModel{}
	application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{transcript: "Where are my tools?"}, language, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}})
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
	for turn := 0; turn < 3; turn++ {
		writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2+turn*2, "audio-"+string(rune('a'+turn)))
		events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
		completed := findRealtimeEvent(t, events, "session.completed")
		if completed["followUpAvailable"] != (turn < 2) {
			t.Fatalf("turn %d availability: %+v", turn, completed)
		}
		response := findRealtimeEvent(t, events, "assistant.response.completed")
		payload := response["response"].(map[string]any)
		if payload["kind"] != "answer" {
			t.Fatalf("normal answer not completed: %+v", payload)
		}
	}
	if len(language.inputs) != 6 {
		t.Fatalf("expected one read and answer per turn, got %d model calls", len(language.inputs))
	}
	// The next audio turn must receive the actual prior answer and tool evidence,
	// not a reconstructed intent or a synthetic clarification classification.
	followUp := language.inputs[2].Messages
	if len(followUp) < 2 {
		t.Fatalf("missing prior conversation: %+v", followUp)
	}
	var priorAnswer, priorEvidence bool
	for _, message := range followUp[:len(followUp)-1] {
		priorAnswer = priorAnswer || (message.Role == ports.ConversationRoleAssistant && message.Text == "I couldn't find matching belongings in this inventory.")
		priorEvidence = priorEvidence || (message.Role == ports.ConversationRoleTool && len(message.ToolResults) == 1)
	}
	if !priorAnswer || !priorEvidence || followUp[len(followUp)-1].Role != ports.ConversationRoleUser {
		t.Fatalf("answer and tool context not retained before follow-up: %+v", followUp)
	}

	_, _, err = connection.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Fatalf("limit did not close normally: %v", err)
	}
}
