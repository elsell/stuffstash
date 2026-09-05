package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type continuousVoiceModel struct {
	inputs []ports.LanguageInferenceInput
}

func (m *continuousVoiceModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	m.inputs = append(m.inputs, input)
	return typedVoiceInvestigationTurn(input, voiceReadIntent(agentmodel.OperationLocate, "tools"), nil)
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
	if len(language.inputs) != 6 || len(language.inputs[2].ConversationTurns) != 2 || language.inputs[2].ConversationTurns[1].Kind != "answer" {
		t.Fatalf("answer context not retained: %+v", language.inputs)
	}
	_, _, err = connection.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Fatalf("limit did not close normally: %v", err)
	}
}
