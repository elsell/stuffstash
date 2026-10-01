package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

// A controlled provider proposes real domain edits over the public websocket.
type detailEditModel struct{ fields map[string]any }

func (m detailEditModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	last := in.Messages[len(in.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Bottle"}}}}, nil
	}
	var output struct{ Items []struct{ AssetID string } }
	if len(last.ToolResults) != 1 || json.Unmarshal([]byte(last.ToolResults[0].Content), &output) != nil || len(output.Items) != 1 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "edit", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Update bottle", "commands": []any{map[string]any{"id": "edit", "kind": "update_asset", "summary": "Provider summary omits details", "arguments": map[string]any{"assetId": output.Items[0].AssetID, "title": "Blue bottle", "description": "", "customFields": m.fields}}}}}}}, nil
}

func TestConversationDetailPatchReviewAndApprovalBoundary(t *testing.T) {
	for _, mode := range []string{"approve", "cancel", "unknown-field", "wrong-type", "unknown-clear"} {
		t.Run(mode, func(t *testing.T) {
			fields := map[string]any{"serial": "NEW", "capacity": nil}
			if mode == "unknown-field" {
				fields["hidden"] = "value"
			}
			if mode == "unknown-clear" {
				fields["hidden"] = nil
			}
			if mode == "wrong-type" {
				fields["capacity"] = "many"
			}
			principal := identity.Principal{ID: "user-1"}
			application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{}, detailEditModel{fields: fields}, fakeTextToSpeech{})
			for _, field := range []struct{ key, label, kind string }{{"serial", "Serial number", "text"}, {"capacity", "Capacity", "number"}, {"maker", "Maker", "text"}} {
				_, err := application.CreateInventoryCustomFieldDefinition(context.Background(), app.CreateCustomFieldDefinitionInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Key: field.key, DisplayName: field.label, Type: field.kind, Applicability: "all_assets"})
				if err != nil {
					t.Fatal(err)
				}
			}
			created, err := application.CreateAssetWithOperation(context.Background(), app.CreateAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: "item", Title: "Bottle", Description: "Original description", CustomFields: map[string]any{"serial": "OLD", "capacity": float64(5), "maker": "Keep"}})
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(NewServerWithOptions(":0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+realtimeVoicePath, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}}})
			if err != nil {
				t.Fatal(err)
			}
			defer conn.CloseNow()
			writeRealtimeMessage(t, ctx, conn, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
			started := readRealtimeMessage(t, ctx, conn)
			session := started["sessionId"]
			writeRealtimeMessage(t, ctx, conn, map[string]any{"type": "text.input", "seq": 2, "sessionId": session, "text": "Rename Bottle to Blue bottle, clear description and capacity, set serial to NEW"})
			terminal := "action.plan.proposed"
			valid := mode == "approve" || mode == "cancel"
			if !valid {
				terminal = "session.failed"
			}
			events := readRealtimeMessagesUntil(t, ctx, conn, terminal)
			get := func() app.GetAssetInput {
				return app.GetAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", AssetID: created.Asset.ID}
			}
			before, err := application.GetAsset(ctx, get())
			if err != nil || before.Title.String() != "Bottle" || before.CustomFields.Values()["serial"] != "OLD" {
				t.Fatal("proposal changed data")
			}
			if !valid {
				if hasRealtimeEvent(events, "action.plan.proposed") {
					t.Fatal("invalid patch offered for approval")
				}
				return
			}
			proposal := findRealtimeEvent(t, events, "action.plan.proposed")["actionPlan"].(map[string]any)
			command := proposal["commands"].([]any)[0].(map[string]any)
			encodedChanges, _ := json.Marshal(command["changes"])
			summary := string(encodedChanges)
			for _, value := range []string{"Blue bottle", "Description", "Serial number", "NEW", "Capacity"} {
				if !strings.Contains(summary, value) {
					t.Fatalf("review hides %q: %s", value, summary)
				}
			}
			if command["expirationCleared"] == true {
				t.Fatal("unrelated expiration appears cleared")
			}
			action := "action.plan.approve"
			done := "action.plan.executed"
			if mode == "cancel" {
				action = "action.plan.cancel"
				done = "action.plan.cancelled"
			}
			writeRealtimeMessage(t, ctx, conn, map[string]any{"type": action, "seq": 3, "sessionId": session, "planId": proposal["planId"]})
			readRealtimeMessagesUntil(t, ctx, conn, done)
			saved, err := application.GetAsset(ctx, get())
			if err != nil {
				t.Fatal(err)
			}
			if mode == "cancel" {
				if saved.Title != before.Title || !saved.CustomFields.Equal(before.CustomFields) {
					t.Fatal("cancel wrote data")
				}
				return
			}
			values := saved.CustomFields.Values()
			if saved.Title.String() != "Blue bottle" || saved.Description.String() != "" || values["serial"] != "NEW" || values["capacity"] != nil || values["maker"] != "Keep" {
				t.Fatalf("patch lost data or changes: %+v", saved)
			}
		})
	}
}
