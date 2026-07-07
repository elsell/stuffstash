package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceWebSocketRejectsReplayedAudioChunkAcrossClarificationTurns(t *testing.T) {
	t.Parallel()

	language := &scriptedFinalLanguageModel{finals: []ports.StructuredAgentResponse{{
		Kind:            ports.StructuredAgentResponseKindClarification,
		SpokenResponse:  "Which item should I update?",
		DisplayResponse: "Which item should I update?",
	}}}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "clarification-response-id"},
	}, &scriptedSpeechToText{transcripts: []string{"Where should I put it?", "Put it in the office."}}, language, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    "tenant-home",
		"inventoryId": "inventory-home",
		"source":      "mobile_voice",
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	})
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)
	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2, "replayed-chunk")
	firstTurn := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	firstResponse := findRealtimeEvent(t, firstTurn, "assistant.response.completed")
	firstPayload, _ := firstResponse["response"].(map[string]any)
	if firstPayload["kind"] != "clarification" {
		t.Fatalf("expected first turn clarification, got %+v", firstPayload)
	}

	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 4, "replayed-chunk")
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected replayed follow-up chunk to fail safely, got %+v", failed)
	}
	assertSafeRealtimeEvents(t, []map[string]any{failed}, []string{"replayed-chunk", "fake-audio", "apiKey", "Bearer", "provider_session_id"})
}
