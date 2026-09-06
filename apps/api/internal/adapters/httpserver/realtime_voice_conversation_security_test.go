package httpserver

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type nativeBoundaryConversation struct{ calls atomic.Int32 }

func (m *nativeBoundaryConversation) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls.Add(1)
	return ports.ConversationModelTurn{Text: "I can help you find and organize your belongings."}, nil
}

type nativeBoundaryResolver struct{ model ports.ConversationModel }

func (r nativeBoundaryResolver) ResolveRealtimeVoiceProviders(context.Context, ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	return ports.RealtimeVoiceProviderSet{ConversationModel: r.model, SpeechToText: fakeSpeechToText{transcript: "What can you help me with?"}, TextToSpeech: fakeTextToSpeech{chunks: [][]byte{[]byte("speech")}}}, nil
}
func TestModelLedConversationWebSocketAuthorization(t *testing.T) {
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
			application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}, {id: "tenant-other", name: "Other", owner: "user-2"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "user-2"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model})
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
			writeRealtimeMessage(t, ctx, conn, map[string]any{"type": "audio.chunk", "seq": 2, "sessionId": started["sessionId"], "chunkId": "audio-1", "audioBase64": base64.StdEncoding.EncodeToString([]byte("audio")), "isFinalChunk": true})
			writeRealtimeMessage(t, ctx, conn, map[string]any{"type": "audio.end", "seq": 3, "sessionId": started["sessionId"]})
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
			if !answered || model.calls.Load() != 1 {
				t.Fatal("authorized native conversation did not complete")
			}
		})
	}
}

// This model proposes a normal inventory title that happens to contain a word
// also used in provider configuration. Typed command fields remain the boundary.
type nativeTitleProposalModel struct{}

func (nativeTitleProposalModel) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "proposal", Name: "propose_inventory_change", Arguments: map[string]any{
		"summary": "Add the credential holder?", "commands": []any{map[string]any{"id": "create-holder", "kind": "create_asset", "summary": "Add holder", "arguments": map[string]any{"title": "Credential holder", "kind": "item"}}},
	}}}}, nil
}
func TestModelLedProposalPreservesOrdinaryTitleAtWebSocketBoundary(t *testing.T) {
	application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: nativeTitleProposalModel{}})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "action.plan.proposed")
	proposed := findRealtimeEvent(t, events, "action.plan.proposed")
	plan, ok := proposed["actionPlan"].(map[string]any)
	if !ok {
		t.Fatalf("missing review: %+v", proposed)
	}
	commands, ok := plan["commands"].([]any)
	if !ok || len(commands) != 1 {
		t.Fatalf("invalid review: %+v", plan)
	}
	command, ok := commands[0].(map[string]any)
	if !ok || command["title"] != "Credential holder" {
		t.Fatalf("ordinary title lost: %+v", commands)
	}
	if hasRealtimeEvent(events, "action.plan.executed") {
		t.Fatal("proposal executed without approval")
	}
}
