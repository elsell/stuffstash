package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestRealtimeVoiceWebSocketRejectsReplayedAudioChunkAcrossClarificationTurns(t *testing.T) {
	t.Parallel()

	language := &scriptedFinalLanguageModel{}
	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids: []string{
			"first-item-id", "first-item-undo", "first-item-audit",
			"second-item-id", "second-item-undo", "second-item-audit",
			"voice-session-id", "clarification-response-id",
		},
	}, &scriptedSpeechToText{transcripts: []string{"Where should I put it?", "Put it in the office."}}, language, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "First item", "")
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Second item", "")

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

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
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
