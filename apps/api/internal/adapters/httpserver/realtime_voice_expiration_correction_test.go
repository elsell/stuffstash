package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/app"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http/httptest"
	"testing"
)

type expirationCorrectionModel struct {
	clear    bool
	override string
}

func (m expirationCorrectionModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	last := in.Messages[len(in.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Bottle"}}}}, nil
	}
	var output struct{ Items []struct{ AssetID string } }
	if len(last.ToolResults) != 1 || json.Unmarshal([]byte(last.ToolResults[0].Content), &output) != nil || len(output.Items) != 1 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	if m.override != "" {
		output.Items[0].AssetID = m.override
	}
	var date any = map[string]any{"date": "2028-02", "precision": "month"}
	if m.clear {
		date = nil
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "correct", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Correct bottle expiration", "commands": []any{map[string]any{"id": "correction", "kind": "update_asset", "summary": "Correct expiration", "arguments": map[string]any{"assetId": output.Items[0].AssetID, "expiration": date}}}}}}}, nil
}
func TestVoiceExpirationCorrectionAcrossReviewAndApproval(t *testing.T) {
	for _, clear := range []bool{false, true} {
		t.Run(map[bool]string{false: "set", true: "clear"}[clear], func(t *testing.T) {
			principal := identity.Principal{ID: "user-1"}
			application := newSeededTestAppWithVoice(t, seededState{ids: []string{"type", "type-audit", "bottle", "undo", "asset-audit", "voice-session-id"}, tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{transcript: "Correct bottle expiration"}, expirationCorrectionModel{clear: clear}, fakeTextToSpeech{})
			kind, err := application.CreateInventoryCustomAssetType(context.Background(), app.CreateCustomAssetTypeInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Key: "medicine", DisplayName: "Medicine", ExpirationEnabled: true})
			if err != nil {
				t.Fatal(err)
			}
			created, err := application.CreateAssetWithOperation(context.Background(), app.CreateAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: "item", Title: "Bottle", CustomAssetTypeID: kind.ID.String(), Expiration: &assetapp.ExpirationInput{Date: "2027-01", Precision: "month"}})
			if err != nil {
				t.Fatal(err)
			}
			ctx, connection, sessionID, planID, proposal := openRealtimeVoiceReviewSessionForApplicationWithProposal(t, application)
			command := proposal["commands"].([]any)[0].(map[string]any)
			if clear {
				if command["expirationCleared"] != true || command["expiration"] != nil {
					t.Fatal("clear review missing")
				}
			} else if command["expiration"].(map[string]any)["date"] != "2028-02" {
				t.Fatal("date review missing")
			}
			before, err := application.GetAsset(ctx, app.GetAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", AssetID: created.Asset.ID})
			if err != nil || before.Expiration.Value() != "2027-01" {
				t.Fatal("changed before approval")
			}
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "action.plan.approve", "seq": 4, "sessionId": sessionID, "planId": planID})
			readRealtimeMessagesUntil(t, ctx, connection, "action.plan.executed")
			after, err := application.GetAsset(ctx, app.GetAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", AssetID: created.Asset.ID})
			want := "2028-02"
			if clear {
				want = ""
			}
			if err != nil || after.Expiration.Value() != want {
				t.Fatalf("wrong saved correction: %v", err)
			}
		})
	}
}

func TestVoiceExpirationCorrectionRejectsUnreturnedAsset(t *testing.T) {
	principal := identity.Principal{ID: "user-1"}
	model := &expirationCorrectionModel{clear: true}
	application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-private", tenantID: "tenant-home", name: "Private", owner: "user-1"}}}, fakeSpeechToText{transcript: "Remove expiration"}, model, fakeTextToSpeech{})
	_, err := application.CreateAssetWithOperation(context.Background(), app.CreateAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: "item", Title: "Bottle"})
	if err != nil {
		t.Fatal(err)
	}
	private, err := application.CreateAssetWithOperation(context.Background(), app.CreateAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-private", Kind: "item", Title: "Private bottle"})
	if err != nil {
		t.Fatal(err)
	}
	model.override = private.Asset.ID.String()
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)
	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.failed")
	if hasRealtimeEvent(events, "action.plan.proposed") || hasRealtimeEvent(events, "action.plan.executed") {
		t.Fatal("unreturned asset correction was proposed")
	}
}
