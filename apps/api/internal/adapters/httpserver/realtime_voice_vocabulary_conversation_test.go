package httpserver

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const vocabularyToolName = "get_inventory_vocabulary"

type vocabularyConversationModel struct {
	arguments  map[string]any
	beforeTool func(context.Context) error
	calls      int
	advertised bool
	results    []string
}

func (m *vocabularyConversationModel) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.calls == 1 {
		for _, tool := range input.Tools {
			if tool.Name == vocabularyToolName {
				m.advertised = true
			}
		}
		if m.beforeTool != nil {
			if err := m.beforeTool(ctx); err != nil {
				return ports.ConversationModelTurn{}, err
			}
		}
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "vocabulary", Name: vocabularyToolName, Arguments: m.arguments}}}, nil
	}
	for _, message := range input.Messages {
		for _, result := range message.ToolResults {
			if result.Name == vocabularyToolName {
				m.results = append(m.results, result.Content)
			}
		}
	}
	return ports.ConversationModelTurn{Text: "I can use your inventory's item types, fields and tags to help organize it."}, nil
}

func seedConversationVocabulary(t *testing.T, store *memory.Store, tenantID, inventoryID, prefix string) {
	t.Helper()
	ctx := context.Background()
	typ, ok := customfield.NewAssetType(customfield.AssetTypeID(prefix+"-type-id"), customfield.TenantID(tenantID), customfield.InventoryID(inventoryID), customfield.ScopeInventory, customfield.Key(prefix+"-type"), "Storage type", "Household storage")
	if !ok {
		t.Fatal("invalid type fixture")
	}
	if err := store.SaveCustomAssetType(ctx, typ, audit.Record{ID: audit.ID(prefix + "-type-audit")}); err != nil {
		t.Fatal(err)
	}
	field, ok := customfield.NewDefinition(customfield.ID(prefix+"-field-id"), customfield.TenantID(tenantID), customfield.InventoryID(inventoryID), customfield.ScopeInventory, customfield.Key(prefix+"-field"), "Condition", customfield.FieldTypeText, nil, customfield.ApplicabilityCustomAssetTypes, []customfield.AssetTypeID{typ.ID})
	if !ok {
		t.Fatal("invalid field fixture")
	}
	if err := store.SaveCustomFieldDefinition(ctx, field, audit.Record{ID: audit.ID(prefix + "-field-audit")}); err != nil {
		t.Fatal(err)
	}
	tag, ok := assettag.NewTag(assettag.ID(prefix+"-tag-id"), assettag.TenantID(tenantID), assettag.InventoryID(inventoryID), assettag.Key(prefix+"-tag"), "Storage", "", time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC))
	if !ok {
		t.Fatal("invalid tag fixture")
	}
	if err := store.CreateAssetTag(ctx, tag, audit.Record{ID: audit.ID(prefix + "-tag-audit")}); err != nil {
		t.Fatal(err)
	}
}

func TestModelLedVocabularyReadAtWebSocketBoundary(t *testing.T) {
	for _, tc := range []struct {
		name              string
		arguments         map[string]any
		wantError, revoke bool
	}{
		{name: "viewer manifest", arguments: map[string]any{}},
		{name: "targeted field", arguments: map[string]any{"definitions": []any{map[string]any{"kind": "custom_field", "key": "home-field"}}}},
		{name: "cross tenant scope injection", arguments: map[string]any{"tenantId": "tenant-other"}, wantError: true},
		{name: "sibling inventory scope injection", arguments: map[string]any{"inventoryId": "inventory-private"}, wantError: true},
		{name: "hidden definition", arguments: map[string]any{"definitions": []any{map[string]any{"kind": "custom_field", "key": "private-field"}}}, wantError: true},
		{name: "cross tenant definition", arguments: map[string]any{"definitions": []any{map[string]any{"kind": "custom_field", "key": "other-field"}}}, wantError: true},
		{name: "revoked before read", arguments: map[string]any{}, revoke: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, authorizer := memory.NewStore(), memory.NewAuthorizer()
			model := &vocabularyConversationModel{arguments: tc.arguments}
			application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
				tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "owner"}, {id: "tenant-other", name: "Other", owner: "other-owner"}},
				inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "owner"}, {id: "inventory-private", tenantID: "tenant-home", name: "Private", owner: "owner"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "other-owner"}},
			}, store, authorizer).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: model})
			seedConversationVocabulary(t, store, "tenant-home", "inventory-home", "home")
			seedConversationVocabulary(t, store, "tenant-home", "inventory-private", "private")
			seedConversationVocabulary(t, store, "tenant-other", "inventory-other", "other")
			viewer := identity.Principal{ID: "viewer"}
			if err := authorizer.GrantInventoryViewer(context.Background(), viewer, "tenant-home", "inventory-home"); err != nil {
				t.Fatal(err)
			}
			if tc.revoke {
				model.beforeTool = func(ctx context.Context) error {
					return authorizer.RevokeInventoryViewer(ctx, viewer, "tenant-home", "inventory-home")
				}
			}
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			terminal := "session.completed"
			if tc.revoke {
				terminal = "session.failed"
			}
			events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "viewer", terminal)
			if tc.revoke {
				failed := findRealtimeEvent(t, events, terminal)
				if failed["code"] != "forbidden" || model.calls != 1 || len(model.results) != 0 {
					t.Fatalf("revoked vocabulary disclosed: %+v model=%+v", failed, model)
				}
				return
			}
			if !model.advertised || model.calls != 2 || len(model.results) != 1 {
				t.Fatalf("vocabulary unavailable to model: %+v", model)
			}
			var result struct {
				Error    string `json:"error"`
				Manifest struct {
					CustomAssetTypes []struct {
						Key string `json:"key"`
					} `json:"customAssetTypes"`
					CustomFields []struct {
						Key string `json:"key"`
					} `json:"customFields"`
					Tags []struct {
						Key string `json:"key"`
					} `json:"tags"`
				} `json:"manifest"`
				Definitions []struct {
					Key        string   `json:"key"`
					Applicable []string `json:"applicableCustomAssetTypeKeys"`
				} `json:"definitions"`
			}
			if err := json.Unmarshal([]byte(model.results[0]), &result); err != nil {
				t.Fatal(err)
			}
			if (result.Error != "") != tc.wantError {
				t.Fatalf("unexpected tool outcome: %s", model.results[0])
			}
			for _, hidden := range []string{"private-field", "other-field", "private-type", "other-type", "private-tag", "other-tag", "home-type-id", "home-field-id", "home-tag-id"} {
				if strings.Contains(model.results[0], hidden) {
					t.Fatalf("vocabulary disclosed %s", hidden)
				}
			}
			if !tc.wantError {
				if len(result.Manifest.CustomAssetTypes) != 1 || result.Manifest.CustomAssetTypes[0].Key != "home-type" || len(result.Manifest.CustomFields) != 1 || result.Manifest.CustomFields[0].Key != "home-field" || len(result.Manifest.Tags) != 1 || result.Manifest.Tags[0].Key != "home-tag" {
					t.Fatalf("scoped vocabulary missing: %s", model.results[0])
				}
				records, err := store.ListInventoryAuditRecords(context.Background(), "tenant-home", "inventory-home", ports.AuditRecordPageRequest{Limit: 100})
				if err != nil {
					t.Fatal(err)
				}
				readActions := map[audit.Action]bool{}
				for _, record := range records {
					if record.Source == audit.SourceConversation {
						readActions[record.Action] = true
					}
				}
				for _, action := range []audit.Action{audit.ActionCustomAssetTypeListed, audit.ActionCustomFieldDefinitionListed, audit.ActionAssetTagListed} {
					if !readActions[action] {
						t.Fatalf("missing conversation read audit %s", action)
					}
				}

				if tc.name == "targeted field" && (len(result.Definitions) != 1 || result.Definitions[0].Key != "home-field" || len(result.Definitions[0].Applicable) != 1 || result.Definitions[0].Applicable[0] != "home-type") {
					t.Fatalf("field applicability missing: %s", model.results[0])
				}
			}
			findRealtimeEvent(t, events, "assistant.response.completed")
		})
	}
}
