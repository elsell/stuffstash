package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http/httptest"
	"testing"
)

type expirationCreateModel struct{ date, precision, override string }

func (m *expirationCreateModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	last := in.Messages[len(in.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "vocabulary", Name: app.RealtimeVoiceToolGetInventoryVocabulary, Arguments: map[string]any{}}}}, nil
	}
	typeID := m.override
	for _, result := range last.ToolResults {
		if result.Name != app.RealtimeVoiceToolGetInventoryVocabulary {
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
		var output struct {
			Manifest struct {
				CustomAssetTypes []struct {
					AssetTypeID string
					Key         string
				}
			}
		}
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			return ports.ConversationModelTurn{}, err
		}
		if typeID == "" {
			for _, kind := range output.Manifest.CustomAssetTypes {
				if kind.Key == "medicine" {
					typeID = kind.AssetTypeID
				}
			}
		}
	}
	if typeID == "" {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "proposal", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Add bottle with expiration", "commands": []any{map[string]any{"id": "bottle", "kind": "create_asset", "summary": "Add bottle", "arguments": map[string]any{"title": "Bottle", "customAssetTypeId": typeID, "expiration": map[string]any{"date": m.date, "precision": m.precision}}}}}}}}, nil
}
func TestVoiceExpirationCreateAcrossReviewAndApproval(t *testing.T) {
	for _, precision := range []string{"day", "month"} {
		t.Run(precision, func(t *testing.T) {
			date := "2028-02"
			if precision == "day" {
				date += "-29"
			}
			model := &expirationCreateModel{date: date, precision: precision}
			application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}, ids: []string{"medicine-type", "type-audit", "voice-session-id"}}, fakeSpeechToText{transcript: "Add a bottle with an expiration date"}, model, fakeTextToSpeech{})
			kind, err := application.CreateInventoryCustomAssetType(context.Background(), app.CreateCustomAssetTypeInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", Key: "medicine", DisplayName: "Medicine", ExpirationEnabled: true})
			if err != nil {
				t.Fatal(err)
			}
			ctx, connection, sessionID, planID, proposal := openRealtimeVoiceReviewSessionForApplicationWithProposal(t, application)
			command := proposal["commands"].([]any)[0].(map[string]any)
			value := command["expiration"].(map[string]any)
			if value["date"] != date || value["precision"] != precision {
				t.Fatal("review date changed")
			}
			before, err := application.ListAssets(ctx, app.ListAssetsInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home"})
			if err != nil || len(before.Items) != 0 {
				t.Fatal("item created without approval")
			}
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "action.plan.approve", "seq": 4, "sessionId": sessionID, "planId": planID})
			events := readRealtimeMessagesUntil(t, ctx, connection, "action.plan.executed")
			executed := findRealtimeEvent(t, events, "action.plan.executed")
			result := executed["commandResults"].([]any)[0].(map[string]any)
			item, err := application.GetAsset(ctx, app.GetAssetInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", AssetID: asset.ID(result["assetId"].(string))})
			if err != nil || item.Expiration.Value() != date || string(item.Expiration.Precision()) != precision || item.CustomAssetTypeID.String() != kind.ID.String() {
				t.Fatalf("saved expiration changed: %+v %v", item, err)
			}
		})
	}
}
func TestVoiceExpirationRejectsForeignTypeAtWebSocketBoundary(t *testing.T) {
	model := &expirationCreateModel{date: "2028-02", precision: "month"}
	application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}, {id: "tenant-other", name: "Other", owner: "other"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "other"}}}, fakeSpeechToText{transcript: "Add a dated bottle"}, model, fakeTextToSpeech{})
	kind, err := application.CreateInventoryCustomAssetType(context.Background(), app.CreateCustomAssetTypeInput{Principal: identity.Principal{ID: "other"}, TenantID: "tenant-other", InventoryID: "inventory-other", Key: "medicine", DisplayName: "Private medicine", ExpirationEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	model.override = kind.ID.String()
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)
	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.failed")
	if hasRealtimeEvent(events, "action.plan.proposed") || hasRealtimeEvent(events, "action.plan.executed") {
		t.Fatal("foreign type crossed voice boundary")
	}
	assertSafeRealtimeEvents(t, events, []string{"Private medicine"})
}
