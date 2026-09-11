package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/app"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http/httptest"
	"strings"
	"testing"
)

type expirationQueryModel struct {
	arguments map[string]any
	result    string
	calls     int
}

func (m *expirationQueryModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.calls == 1 {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "expiration-query", Name: app.RealtimeVoiceToolQueryExpiringAssets, Arguments: m.arguments}}}, nil
	}
	for _, message := range in.Messages {
		for _, result := range message.ToolResults {
			if result.Name == app.RealtimeVoiceToolQueryExpiringAssets {
				m.result = result.Content
			}
		}
	}
	return ports.ConversationModelTurn{Text: "I checked the recorded expiration dates."}, nil
}
func TestExpirationQueryAtWebSocketBoundaryIsScoped(t *testing.T) {
	for _, forged := range []bool{false, true} {
		t.Run(map[bool]string{false: "authorized", true: "forged scope"}[forged], func(t *testing.T) {
			model := &expirationQueryModel{arguments: map[string]any{"status": "all"}}
			if forged {
				model.arguments["tenantId"] = "tenant-other"
			}
			application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}, {id: "tenant-other", name: "Other", owner: "other"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "other"}}}, fakeSpeechToText{transcript: "What expires?"}, model, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}})
			for _, scope := range []struct{ tenant, inventory, principal, title string }{{"tenant-home", "inventory-home", "user-1", "Home bottle"}, {"tenant-other", "inventory-other", "other", "Private bottle"}} {
				principal := identity.Principal{ID: identity.PrincipalID(scope.principal)}
				kind, err := application.CreateInventoryCustomAssetType(context.Background(), app.CreateCustomAssetTypeInput{Principal: principal, TenantID: tenant.ID(scope.tenant), InventoryID: inventory.InventoryID(scope.inventory), Key: "medicine", DisplayName: "Medicine", ExpirationEnabled: true})
				if err != nil {
					t.Fatal(err)
				}
				_, err = application.CreateAssetWithOperation(context.Background(), app.CreateAssetInput{Principal: principal, TenantID: tenant.ID(scope.tenant), InventoryID: inventory.InventoryID(scope.inventory), Kind: "item", Title: scope.title, CustomAssetTypeID: kind.ID.String(), Expiration: &assetapp.ExpirationInput{Date: "2028-02", Precision: "month"}})
				if err != nil {
					t.Fatal(err)
				}
			}
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			t.Cleanup(server.Close)
			runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
			if model.result == "" || strings.Contains(model.result, "Private bottle") {
				t.Fatalf("unscoped query: %s", model.result)
			}
			var output struct {
				Error string
				Count int
				Items []struct{ Title string }
			}
			if err := json.Unmarshal([]byte(model.result), &output); err != nil {
				t.Fatal(err)
			}
			if forged {
				if output.Error == "" || output.Count != 0 {
					t.Fatal("forged scope accepted")
				}
				return
			}
			if output.Error != "" || output.Count != 1 || output.Items[0].Title != "Home bottle" {
				t.Fatalf("authorized query failed: %s", model.result)
			}
		})
	}
}
