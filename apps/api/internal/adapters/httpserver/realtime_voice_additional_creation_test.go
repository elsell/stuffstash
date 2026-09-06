package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type additionalItemModel struct{ propose bool }

func (m additionalItemModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if len(input.Messages) == 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	if input.Messages[len(input.Messages)-1].Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find-charger", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Charger"}}}}, nil
	}
	ids, err := httpConversationEvidenceIDs(input, "find-charger")
	if err != nil || ids["charger"] == "" {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	if !m.propose {
		return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "Your Charger is already recorded.", Display: "Your Charger is already recorded.", AssetIDs: []string{ids["charger"]}}}, nil
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "propose-additional", Name: "propose_inventory_change", Arguments: map[string]any{
		"summary": "Create an additional Charger?", "commands": []any{map[string]any{"id": "create-charger", "kind": "create_asset", "summary": "Create additional Charger", "arguments": map[string]any{"title": "Charger", "kind": "item"}}},
	}}}}, nil
}

// Removed with the legacy provider injection signature.
func (additionalItemModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
}

func TestRealtimeAdditionalItemPreservesExistingIdentityAccessAndApproval(t *testing.T) {
	for _, scenario := range []struct {
		name, user, transcript, terminal string
		propose, approve                 bool
	}{
		{"additional owner", "user-1", "I bought another charger", "action.plan.proposed", true, true},
		{"existing item answer", "user-1", "Record my charger", "session.completed", false, false},
		{"unwanted model proposal cancelled", "user-1", "Record my charger", "action.plan.proposed", true, false},
		{"viewer additional", "viewer", "I bought another charger", "session.failed", true, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			store := memory.NewStore()
			authorizer := memory.NewAuthorizer()
			application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, store, authorizer).WithRealtimeVoiceProviders(fakeSpeechToText{transcript: scenario.transcript}, additionalItemModel{propose: scenario.propose}, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}}).WithRealtimeVoiceResponseGenerator(httpTestVoiceResponseGenerator{})
			seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Charger", "")
			if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, "tenant-home", "inventory-home"); err != nil {
				t.Fatal(err)
			}
			original, err := store.ListAssetsByInventory(context.Background(), "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
			if err != nil || len(original) != 1 {
				t.Fatalf("fixture: %v %v", original, err)
			}
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:" + scenario.user}}})
			if err != nil {
				t.Fatal(err)
			}
			defer connection.Close(websocket.StatusNormalClosure, "")
			writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
			started := readRealtimeMessage(t, ctx, connection)
			if started["type"] != "session.started" {
				t.Fatalf("start: %+v", started)
			}
			sessionID := started["sessionId"].(string)
			writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2, "creation-audio")
			events := readRealtimeMessagesUntil(t, ctx, connection, scenario.terminal)
			before, err := store.ListAssetsByInventory(ctx, "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
			if err != nil || len(before) != 1 || before[0].ID != original[0].ID {
				t.Fatal("voice request mutated assets before approval")
			}
			if scenario.terminal != "action.plan.proposed" {
				assertNoRealtimeEventType(t, events, "action.plan.proposed")
				if scenario.terminal == "session.failed" && findRealtimeEvent(t, events, "session.failed")["code"] != "forbidden" {
					t.Fatal("viewer proposal did not fail at authorization")
				}
				return
			}
			proposal := findRealtimeEvent(t, events, "action.plan.proposed")["actionPlan"].(map[string]any)
			if !strings.Contains(proposal["confirmationSummary"].(string), "additional") {
				t.Fatalf("approval hides additional intent: %+v", proposal)
			}
			if !scenario.approve {
				writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "action.plan.cancel", "seq": 4, "sessionId": sessionID, "planId": proposal["planId"]})
				readRealtimeMessagesUntil(t, ctx, connection, "action.plan.cancelled")
				after, err := store.ListAssetsByInventory(ctx, "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
				if err != nil || len(after) != 1 || after[0].ID != original[0].ID {
					t.Fatal("cancelled proposal changed existing inventory")
				}
				return
			}
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "action.plan.approve", "seq": 4, "sessionId": sessionID, "planId": proposal["planId"]})
			readRealtimeMessagesUntil(t, ctx, connection, "action.plan.executed")
			after, err := store.ListAssetsByInventory(ctx, "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
			if err != nil || len(after) != 2 {
				t.Fatalf("approved new instance not created: %v %v", after, err)
			}
			if after[0].ID == after[1].ID || after[0].Title.String() != "Charger" || after[1].Title.String() != "Charger" {
				t.Fatal("additional item overwrote the existing identity")
			}
		})
	}
}
