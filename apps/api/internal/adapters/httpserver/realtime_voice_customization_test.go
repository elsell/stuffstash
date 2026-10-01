package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type customizationProposalModel struct{ kind string }

func (m customizationProposalModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if in.Messages[len(in.Messages)-1].Role != ports.ConversationRoleUser {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	args := map[string]any{"key": "warranty", "displayName": "Warranty"}
	if m.kind == "create_custom_field_definition" {
		args["fieldType"] = "date"
		args["applicability"] = "all_assets"
	} else {
		args["description"] = "Items with a warranty"
		args["expirationEnabled"] = true
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "configure", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Create Warranty", "commands": []any{map[string]any{"id": "configure", "kind": m.kind, "summary": "Create Warranty", "arguments": args}}}}}}, nil
}

type configurationRevocationAuthorizer struct {
	ports.Authorizer
	denied atomic.Bool
}

func (a *configurationRevocationAuthorizer) CheckInventory(ctx context.Context, p identity.Principal, permission ports.InventoryPermission, id inventory.InventoryID) error {
	if permission == ports.InventoryPermissionConfigure && a.denied.Load() {
		return ports.ErrForbidden
	}
	return a.Authorizer.CheckInventory(ctx, p, permission, id)
}

func TestConversationCustomizationRequiresConfigureAndExplicitApproval(t *testing.T) {
	for _, kind := range []string{"create_custom_asset_type", "create_custom_field_definition"} {
		for _, mode := range []string{"approve", "cancel", "editor", "revoked"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				authorizer := &configurationRevocationAuthorizer{Authorizer: memory.NewAuthorizer()}
				application := newSeededTestAppWithAuthorizer(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, authorizer).WithRealtimeVoiceProviders(fakeSpeechToText{}, customizationProposalModel{kind: kind}, fakeTextToSpeech{})
				owner := identity.Principal{ID: "user-1"}
				user := "user-1"
				if mode == "editor" {
					user = "user-2"
					if err := authorizer.GrantInventoryEditor(context.Background(), identity.Principal{ID: "user-2"}, "tenant-home", "inventory-home"); err != nil {
						t.Fatal(err)
					}
				}
				count := func() int {
					t.Helper()
					if kind == "create_custom_asset_type" {
						result, err := application.ListInventoryCustomAssetTypes(context.Background(), app.ListCustomAssetTypesInput{Principal: owner, TenantID: "tenant-home", InventoryID: "inventory-home"})
						if err != nil {
							t.Fatal(err)
						}
						return len(result.Items)
					}
					result, err := application.ListInventoryCustomFieldDefinitions(context.Background(), app.ListCustomFieldDefinitionsInput{Principal: owner, TenantID: "tenant-home", InventoryID: "inventory-home"})
					if err != nil {
						t.Fatal(err)
					}
					return len(result.Items)
				}
				before := count()
				server := httptest.NewServer(NewServerWithOptions(":0", application, Options{RateLimitDisabled: true}).Handler)
				defer server.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+realtimeVoicePath, &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:" + user}}})
				if err != nil {
					t.Fatal(err)
				}
				defer conn.CloseNow()
				writeRealtimeMessage(t, ctx, conn, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
				session := readRealtimeMessage(t, ctx, conn)["sessionId"]
				writeRealtimeMessage(t, ctx, conn, map[string]any{"type": "text.input", "seq": 2, "sessionId": session, "text": "Create a Warranty definition"})
				terminal := "action.plan.proposed"
				if mode == "editor" {
					terminal = "session.failed"
				}
				events := readRealtimeMessagesUntil(t, ctx, conn, terminal)
				if count() != before {
					t.Fatal("schema changed before approval")
				}
				if mode == "editor" {
					if hasRealtimeEvent(events, "action.plan.proposed") {
						t.Fatal("editor offered configuration approval")
					}
					return
				}
				plan := findRealtimeEvent(t, events, "action.plan.proposed")["actionPlan"].(map[string]any)
				command := plan["commands"].([]any)[0].(map[string]any)
				if command["operation"] != "configure" || len(command["changes"].([]any)) < 3 {
					t.Fatal("schema details not disclosed")
				}
				action := "action.plan.approve"
				done := "action.plan.executed"
				if mode == "cancel" {
					action = "action.plan.cancel"
					done = "action.plan.cancelled"
				}
				if mode == "revoked" {
					authorizer.denied.Store(true)
					done = "action.plan.failed"
				}
				writeRealtimeMessage(t, ctx, conn, map[string]any{"type": action, "seq": 3, "sessionId": session, "planId": plan["planId"]})
				readRealtimeMessagesUntil(t, ctx, conn, done)
				authorizer.denied.Store(false)
				want := before
				if mode == "approve" {
					want++
				}
				if count() != want {
					t.Fatal("incorrect schema mutation count")
				}
			})
		}
	}
}
