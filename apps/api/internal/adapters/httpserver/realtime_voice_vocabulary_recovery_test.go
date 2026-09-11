package httpserver

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type vocabularyRecoveryModel struct {
	result    string
	calls     int
	arguments map[string]any
}

func (m *vocabularyRecoveryModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if m.calls == 1 {
		if m.arguments == nil {
			m.arguments = map[string]any{"definitions": []any{map[string]any{"kind": "custom_asset_type", "key": "Medicine"}}}
		}
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "vocab", Name: app.RealtimeVoiceToolGetInventoryVocabulary, Arguments: m.arguments}}}, nil
	}
	for _, message := range input.Messages {
		for _, result := range message.ToolResults {
			if result.Name == app.RealtimeVoiceToolGetInventoryVocabulary {
				m.result = result.Content
			}
		}
	}
	return ports.ConversationModelTurn{Text: "I checked the inventory types."}, nil
}
func TestVoiceVocabularyMistypedKeyPreservesAuthorizedManifest(t *testing.T) {
	model := &vocabularyRecoveryModel{}
	application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{transcript: "Add medicine"}, model, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}})
	kind, err := application.CreateInventoryCustomAssetType(context.Background(), app.CreateCustomAssetTypeInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", Key: "medicine", DisplayName: "Medicine", ExpirationEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	var result struct {
		Error                      string
		UnavailableDefinitionCount int
		Manifest                   struct {
			CustomAssetTypes []struct{ AssetTypeID string }
		}
		Definitions []any
	}
	if err := json.Unmarshal([]byte(model.result), &result); err != nil {
		t.Fatal(err)
	}
	if result.Error != "" || result.UnavailableDefinitionCount != 1 || len(result.Definitions) != 0 || len(result.Manifest.CustomAssetTypes) != 1 || result.Manifest.CustomAssetTypes[0].AssetTypeID != kind.ID.String() {
		t.Fatalf("bad lookup hid manifest or resolved guessed key: %s", model.result)
	}
}

func TestVoiceVocabularyRecoveryKeepsMalformedRequestsInvalid(t *testing.T) {
	for _, args := range []map[string]any{
		{"definitions": []any{map[string]any{"kind": "unknown", "key": "Medicine"}}},
		{"definitions": []any{map[string]any{"kind": "custom_asset_type", "key": ""}}},
		{"definitions": []any{map[string]any{"kind": "custom_asset_type", "key": strings.Repeat("x", 81)}}},
		{"tenantId": "tenant-other", "definitions": []any{map[string]any{"kind": "custom_asset_type", "key": "Medicine"}}},
	} {
		model := &vocabularyRecoveryModel{arguments: args}
		application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{transcript: "Discover types"}, model, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}})
		server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
		runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
		server.Close()
		var output map[string]any
		if err := json.Unmarshal([]byte(model.result), &output); err != nil {
			t.Fatal(err)
		}
		if output["error"] == nil || output["manifest"] != nil {
			t.Fatalf("malformed request returned manifest: %s", model.result)
		}
	}
}
