package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
	"net/http/httptest"
	"nhooyr.io/websocket"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTypedConversationWebSocketAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name, token, tenant, inventory string
		allowed                        bool
	}{
		{"owner", "dev:user-1", "tenant-home", "inventory-home", true},
		{"outsider", "dev:user-2", "tenant-home", "inventory-home", false},
		{"cross tenant", "dev:user-1", "tenant-other", "inventory-other", false},
		{"wrong inventory", "dev:user-1", "tenant-home", "inventory-other", false},
		{"missing token", "", "tenant-home", "inventory-home", false},
		{"malformed token", "invalid", "tenant-home", "inventory-home", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := &nativeBoundaryConversation{}
			speech := &countingTextSpeech{}
			application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}, {id: "tenant-other", name: "Other", owner: "user-2"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "user-2"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model, speech: speech})
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			headers := http.Header{}
			if tc.token != "" {
				headers.Set("Authorization", "Bearer "+tc.token)
			}
			conn, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: headers})
			if err != nil {
				if tc.allowed || response == nil || response.StatusCode != http.StatusUnauthorized {
					t.Fatalf("unexpected handshake: %v", err)
				}
				if model.calls.Load() != 0 {
					t.Fatal("unauthenticated model invocation")
				}
				return
			}
			defer conn.Close(websocket.StatusNormalClosure, "")
			writeRealtimeMessage(t, ctx, conn, realtimeVoiceStartMessage(tc.tenant, tc.inventory))
			started := readRealtimeMessage(t, ctx, conn)
			if !tc.allowed {
				if started["type"] != "session.failed" || started["code"] != "forbidden" || model.calls.Load() != 0 {
					t.Fatalf("authorization bypass: %+v", started)
				}
				return
			}
			if started["type"] != "session.started" {
				t.Fatalf("native-only provider rejected: %+v", started)
			}
			writeRealtimeMessage(t, ctx, conn, map[string]any{"type": "text.input", "seq": 2, "sessionId": started["sessionId"], "text": "What can you help me with?"})
			answered := false
			for {
				event := readRealtimeMessage(t, ctx, conn)
				if event["type"] == "assistant.response.completed" {
					answered = true
				}
				if event["type"] == "session.failed" {
					t.Fatalf("native conversation failed: %+v", event)
				}
				if event["type"] == "session.completed" {
					break
				}
			}
			if !answered || model.calls.Load() != 1 || speech.calls.Load() != 0 {
				t.Fatal("authorized native conversation did not complete")
			}
		})
	}
}

func TestTypedConversationRejectsInvalidFrames(t *testing.T) {
	for _, tc := range []struct {
		name         string
		text         string
		sequence     int
		wrongSession bool
		extraScope   bool
		audioFirst   bool
	}{
		{name: "blank", text: "  ", sequence: 2},
		{name: "oversize", text: strings.Repeat("界", 8001), sequence: 2},
		{name: "replayed sequence", text: "Find tools", sequence: 1},
		{name: "wrong session", text: "Find tools", sequence: 2, wrongSession: true},
		{name: "scope injection", text: "Find tools", sequence: 2, extraScope: true},
		{name: "mixed audio text", text: "Find tools", sequence: 3, audioFirst: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := &nativeBoundaryConversation{}
			application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model})
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}}})
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close(websocket.StatusNormalClosure, "")
			writeRealtimeMessage(t, ctx, conn, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
			started := readRealtimeMessage(t, ctx, conn)
			id := started["sessionId"]
			if tc.audioFirst {
				writeRealtimeMessage(t, ctx, conn, map[string]any{"type": "audio.chunk", "seq": 2, "sessionId": id, "chunkId": "a", "audioBase64": "YQ==", "isFinalChunk": true})
			}
			if tc.wrongSession {
				id = "other-session"
			}
			frame := map[string]any{"type": "text.input", "seq": tc.sequence, "sessionId": id, "text": tc.text}
			if tc.extraScope {
				frame["inventoryId"] = "inventory-other"
			}
			writeRealtimeMessage(t, ctx, conn, frame)
			failed := readRealtimeMessage(t, ctx, conn)
			if failed["type"] != "session.failed" || model.calls.Load() != 0 {
				t.Fatalf("invalid input reached model: %+v", failed)
			}
		})
	}
}

type countingTextSpeech struct{ calls atomic.Int32 }

func (speech *countingTextSpeech) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	speech.calls.Add(1)
	return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
}

func TestTypedViewerCannotProposeInventoryChanges(t *testing.T) {
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "owner"}},
		ids:         []string{"session-id", "unexpected-plan-id"},
	}, store, authorizer).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: nativeTitleProposalModel{}})
	if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, "tenant-home", "inventory-home"); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:viewer"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
	started := readRealtimeMessage(t, ctx, connection)
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "text.input", "seq": 2, "sessionId": started["sessionId"], "text": "Add a credential holder"})
	events := readRealtimeMessagesUntil(t, ctx, connection, "session.failed")
	failed := findRealtimeEvent(t, events, "session.failed")
	if failed["code"] != "forbidden" {
		t.Fatalf("expected write permission denial: %+v", failed)
	}
	if hasRealtimeEvent(events, "action.plan.proposed") || hasRealtimeEvent(events, "action.plan.executed") {
		t.Fatal("viewer proposed or executed changes")
	}
	if _, found, err := store.ActionPlanByID(context.Background(), "tenant-home", "inventory-home", "unexpected-plan-id"); err != nil || found {
		t.Fatalf("unauthorized draft persisted: found=%v err=%v", found, err)
	}
}
