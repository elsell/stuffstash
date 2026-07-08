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

func TestRealtimeVoiceQueryReportsClientCancellationBeforeAudioEnd(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id"},
	}, fakeSpeechToText{}, scriptedLanguageModel{}, fakeTextToSpeech{})
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

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "session.cancel",
		"seq":       2,
		"sessionId": sessionID,
		"reason":    "user_cancelled",
	})
	cancelled := readRealtimeMessage(t, ctx, connection)
	if cancelled["type"] != "session.cancelled" {
		t.Fatalf("expected session.cancelled, got %+v", cancelled)
	}
	if cancelled["sessionId"] != sessionID {
		t.Fatalf("expected cancelled session %q, got %+v", sessionID, cancelled)
	}
}

func TestRealtimeVoiceQueryReportsProviderContextCancellationAsCancelled(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id"},
	}, fakeSpeechToText{err: context.Canceled}, scriptedLanguageModel{}, fakeTextToSpeech{})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.cancelled")
	if len(events) != 1 || events[0]["type"] != "session.cancelled" {
		t.Fatalf("expected terminal session.cancelled, got %+v", events)
	}
}
