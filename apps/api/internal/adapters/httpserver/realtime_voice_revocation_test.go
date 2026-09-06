package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceWebSocketRechecksRevokedInventoryAccessBeforeProviderDisclosure(t *testing.T) {
	t.Parallel()

	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	providers := &revocationProbeVoiceProviders{}
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "owner-user"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "owner-user"}},
		ids:         []string{"voice-session-id"},
	}, store, authorizer).WithRealtimeVoiceProviders(providers, providers, providers)
	if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer-user"}, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home")); err != nil {
		t.Fatalf("grant viewer: %v", err)
	}

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:viewer-user"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("expected authorized session start, got %+v", started)
	}
	sessionID, _ := started["sessionId"].(string)
	if err := authorizer.RevokeInventoryViewer(ctx, identity.Principal{ID: "viewer-user"}, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home")); err != nil {
		t.Fatalf("revoke viewer: %v", err)
	}

	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2, "revoked-turn")
	events := readRealtimeMessagesUntil(t, ctx, connection, "session.failed")
	assertNoRealtimeEventType(t, events, "transcript.final")
	assertNoRealtimeEventType(t, events, "agent.diagnostic")
	if providers.sttCalls != 0 || providers.languageCalls != 0 || providers.ttsCalls != 0 {
		t.Fatalf("revoked turn reached voice providers: %+v", providers)
	}
}

func TestRealtimeVoiceContinuationRechecksRevokedAccessBeforeDisclosure(t *testing.T) {
	t.Parallel()

	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	providers := &revocationProbeVoiceProviders{allowAnswer: true}
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "owner-user"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "owner-user"}},
		ids:         []string{"voice-session-id"},
	}, store, authorizer).WithRealtimeVoiceProviders(providers, providers, providers)
	if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer-user"}, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home")); err != nil {
		t.Fatalf("grant viewer: %v", err)
	}

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:viewer-user"}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
	start["conversationContinuity"] = true
	writeRealtimeMessage(t, ctx, connection, start)
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("expected authorized session start, got %+v", started)
	}
	sessionID, _ := started["sessionId"].(string)
	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2, "first-answer")
	first := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
	if findRealtimeEvent(t, first, "session.completed")["followUpAvailable"] != true {
		t.Fatal("answer follow-up unavailable")
	}
	if err := authorizer.RevokeInventoryViewer(ctx, identity.Principal{ID: "viewer-user"}, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home")); err != nil {
		t.Fatalf("revoke viewer: %v", err)
	}

	writeRealtimeAudioTurn(t, ctx, connection, sessionID, 4, "revoked-turn")
	events := readRealtimeMessagesUntil(t, ctx, connection, "session.failed")
	assertNoRealtimeEventType(t, events, "transcript.final")
	assertNoRealtimeEventType(t, events, "agent.diagnostic")
	if providers.sttCalls != 1 || providers.languageCalls != 2 || providers.ttsCalls != 1 {
		t.Fatalf("revoked turn reached voice providers: %+v", providers)
	}
}

type revocationProbeVoiceProviders struct {
	allowAnswer   bool
	sttCalls      int
	languageCalls int
	ttsCalls      int
}

func (p *revocationProbeVoiceProviders) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	p.sttCalls++
	return ports.SpeechToTextResult{Transcript: "Where are the secret documents?"}, nil
}

func (p *revocationProbeVoiceProviders) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	p.languageCalls++
	if p.allowAnswer {
		return httpConversationRead(input, app.RealtimeVoiceToolSearchAuthorizedAssets, map[string]any{"query": "tools"}, nil)
	}
	return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
}

func (p *revocationProbeVoiceProviders) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	p.ttsCalls++
	if p.allowAnswer {
		return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{[]byte("audio")}}, nil
	}
	return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
}
