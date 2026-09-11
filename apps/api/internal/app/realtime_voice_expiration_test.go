package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestVoiceAssetDetailPreservesExpirationFacts(t *testing.T) {
	item := assetItem("bottle", "tenant-home", "inventory-home", asset.KindItem, "")
	item.CustomAssetTypeID = "medicine"
	item.Expiration, _ = expirationdate.ParseDate("2026-09", expirationdate.Month)
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}}, &fakeIDGenerator{})
	kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
	kind.ExpirationEnabled = true
	application.customAssetTypes = &fakeCustomAssetTypeRepository{items: []customfield.AssetType{kind}}
	application.clock = fakeClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)}
	application.notificationService = notificationapp.New(notificationapp.Dependencies{Authorizer: application.authorizer, Inventories: application.inventories, Preferences: memory.NewStore(), Audit: application.audit, IDs: application.ids, Clock: application.clock})
	result, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), ports.AgentToolCall{ID: "detail", Name: RealtimeVoiceToolGetAssetDetail, Arguments: map[string]any{"assetId": "bottle"}}, map[string]struct{}{"bottle": {}})
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		Items []struct {
			Expiration *struct {
				Date            string
				Precision       string
				State           string
				TrackingEnabled bool
			}
		}
	}
	if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Items) != 1 || output.Items[0].Expiration == nil {
		t.Fatal("expiration missing from tool")
	}
	value := output.Items[0].Expiration
	if value.Date != "2026-09" || value.Precision != "month" || value.State != "upcoming" || !value.TrackingEnabled {
		t.Fatalf("incorrect facts: %+v", value)
	}
}
func TestVoiceVocabularyExposesExpirationCapability(t *testing.T) {
	kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
	kind.ExpirationEnabled = true
	manifest, catalog, err := projectRealtimeVoiceVocabulary([]customfield.AssetType{kind}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(manifest.CustomAssetTypes[0])
	var value map[string]any
	_ = json.Unmarshal(encoded, &value)
	if value["expirationEnabled"] != true || value["assetTypeId"] != "medicine" {
		t.Fatal("manifest omitted capability")
	}
	for _, definition := range catalog.definitions {
		encoded, _ = json.Marshal(definition)
		value = map[string]any{}
		_ = json.Unmarshal(encoded, &value)
		if value["expirationEnabled"] != true || value["assetTypeId"] != "medicine" {
			t.Fatal("definition omitted capability")
		}
	}
}
