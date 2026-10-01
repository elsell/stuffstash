package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestBrowserConversationAuthenticationBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, origin, authorization, tenant, inventory string
		upgrade, allowed                               bool
	}{
		{"owner", "https://web.example.test", "Bearer dev:user-1", "tenant-home", "inventory-home", true, true},
		{"absent origin", "", "Bearer dev:user-1", "tenant-home", "inventory-home", false, false},
		{"unapproved origin", "https://evil.example.test", "Bearer dev:user-1", "tenant-home", "inventory-home", false, false},
		{"null origin", "null", "Bearer dev:user-1", "tenant-home", "inventory-home", false, false},
		{"missing token", "https://web.example.test", "", "tenant-home", "inventory-home", true, false},
		{"invalid token", "https://web.example.test", "invalid", "tenant-home", "inventory-home", true, false},
		{"wrong tenant", "https://web.example.test", "Bearer dev:user-1", "tenant-other", "inventory-other", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := &nativeBoundaryConversation{}
			application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}, {id: "tenant-other", name: "Other", owner: "user-2"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "user-2"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model, speech: &countingTextSpeech{}})
			server := httptest.NewServer(NewServerWithOptions(":0", application, Options{RateLimitDisabled: true, CORSAllowedOrigins: []string{"https://web.example.test"}}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			headers := http.Header{}
			if tc.origin != "" {
				headers.Set("Origin", tc.origin)
			}
			connection, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+realtimeVoicePath, &websocket.DialOptions{Subprotocols: []string{"stuffstash.browser.v1"}, HTTPHeader: headers})
			if !tc.upgrade {
				if err == nil || response == nil || response.StatusCode != 403 || model.calls.Load() != 0 {
					t.Fatalf("origin accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer connection.CloseNow()
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "session.authenticate", "authorization": tc.authorization})
			_, raw, err := connection.Read(ctx)
			if tc.authorization == "" || tc.authorization == "invalid" {
				if err == nil || model.calls.Load() != 0 {
					t.Fatal("invalid credentials accepted")
				}
				return
			}
			var authenticated map[string]any
			if err != nil || json.Unmarshal(raw, &authenticated) != nil || authenticated["type"] != "session.authenticated" {
				t.Fatalf("authentication: %s %v", raw, err)
			}
			start := realtimeVoiceStartMessage(tc.tenant, tc.inventory)
			start["source"] = "web_text"
			writeRealtimeMessage(t, ctx, connection, start)
			started := readRealtimeMessage(t, ctx, connection)
			if !tc.allowed {
				if started["type"] != "session.failed" || started["code"] != "forbidden" || model.calls.Load() != 0 {
					t.Fatalf("scope bypass: %+v", started)
				}
				return
			}
			if started["type"] != "session.started" {
				t.Fatalf("start: %+v", started)
			}
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "text.input", "seq": 2, "sessionId": started["sessionId"], "text": "What can you help me with?"})
			answered := false
			for {
				event := readRealtimeMessage(t, ctx, connection)
				if event["type"] == "assistant.response.completed" {
					answered = true
				}
				if event["type"] == "session.failed" {
					t.Fatalf("turn: %+v", event)
				}
				if event["type"] == "session.completed" {
					break
				}
			}
			if !answered || model.calls.Load() != 1 {
				t.Fatal("browser text did not reach authorized model")
			}
		})
	}
}

func TestBrowserConversationRejectsInvalidAuthenticationFrames(t *testing.T) {
	for name, frame := range map[string]string{"wrong type": `{"type":"session.start","authorization":"Bearer dev:user-1"}`, "unknown field": `{"type":"session.authenticate","authorization":"Bearer dev:user-1","tenantId":"tenant-home"}`, "oversized": `{"type":"session.authenticate","authorization":"` + strings.Repeat("x", 17*1024) + `"}`} {
		t.Run(name, func(t *testing.T) {
			application := newSeededTestApp(t, seededState{})
			server := httptest.NewServer(NewServerWithOptions(":0", application, Options{CORSAllowedOrigins: []string{"https://web.example.test"}}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+realtimeVoicePath, &websocket.DialOptions{Subprotocols: []string{"stuffstash.browser.v1"}, HTTPHeader: http.Header{"Origin": []string{"https://web.example.test"}}})
			if err != nil {
				t.Fatal(err)
			}
			defer connection.CloseNow()
			_ = connection.Write(ctx, websocket.MessageText, []byte(frame))
			if _, _, err = connection.Read(ctx); websocket.CloseStatus(err) != websocket.StatusPolicyViolation && !(name == "oversized" && websocket.CloseStatus(err) == websocket.StatusMessageTooBig) {
				t.Fatalf("expected policy close, got %v", err)
			}
		})
	}
}

func TestBrowserConversationAuthenticationDeadline(t *testing.T) {
	application := newSeededTestApp(t, seededState{})
	server := httptest.NewServer(NewServerWithOptions(":0", application, Options{CORSAllowedOrigins: []string{"https://web.example.test"}}).Handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+realtimeVoicePath, &websocket.DialOptions{Subprotocols: []string{browserRealtimeProtocol}, HTTPHeader: http.Header{"Origin": []string{"https://web.example.test"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	_, _, err = connection.Read(ctx)
	if err == nil || ctx.Err() != nil {
		t.Fatalf("authentication deadline did not close connection: %v", err)
	}
}
