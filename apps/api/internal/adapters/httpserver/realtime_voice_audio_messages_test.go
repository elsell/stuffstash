package httpserver

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestRealtimeVoiceQueryRejectsScopeChangingAudioMessages(t *testing.T) {
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
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"tenantId":     "tenant-other",
		"inventoryId":  "inventory-other",
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	failed := readRealtimeMessage(t, ctx, connection)
	if failed["type"] != "session.failed" {
		t.Fatalf("expected session.failed, got %+v", failed)
	}
	if failed["code"] != "invalid_request" {
		t.Fatalf("expected invalid request, got %+v", failed)
	}
	assertSafeRealtimeEvents(t, []map[string]any{failed}, []string{"tenant-other", "inventory-other"})
}

func TestRealtimeVoiceQueryRejectsMalformedAudioFinalChunkMarkers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		marker any
		omit   bool
	}{
		{name: "missing", omit: true},
		{name: "wrong type", marker: "true"},
		{name: "null", marker: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			message := map[string]any{
				"type":        "audio.chunk",
				"seq":         2,
				"sessionId":   sessionID,
				"chunkId":     "chunk-1",
				"audioBase64": base64.StdEncoding.EncodeToString([]byte("fake-audio")),
			}
			if !tt.omit {
				message["isFinalChunk"] = tt.marker
			}
			writeRealtimeMessage(t, ctx, connection, message)
			failed := readRealtimeMessage(t, ctx, connection)
			if failed["type"] != "session.failed" {
				t.Fatalf("expected session.failed, got %+v", failed)
			}
			if failed["code"] != "invalid_request" {
				t.Fatalf("expected invalid request, got %+v", failed)
			}
		})
	}
}
