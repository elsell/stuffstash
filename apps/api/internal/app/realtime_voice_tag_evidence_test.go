package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestRealtimeVoiceSearchPreservesAssignedTagsForAssessment(t *testing.T) {
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	assetID := asset.ID("asset-one")
	item := asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          asset.Title("3–6 months clothes"),
		LifecycleState: asset.LifecycleStateActive,
	}
	tagKey, _ := assettag.NewKey("baby")
	tagName, _ := assettag.NewDisplayName("Baby")
	tagColor, _ := assettag.NewColor("#2f80ed")
	tag, _ := assettag.NewTag("tag-camping", assettag.TenantID(tenantID.String()), assettag.InventoryID(inventoryID.String()), tagKey, tagName, tagColor, time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC))
	clothesKey, _ := assettag.NewKey("clothes")
	clothesName, _ := assettag.NewDisplayName("Clothes")
	clothesTag, _ := assettag.NewTag("tag-clothes", assettag.TenantID(tenantID.String()), assettag.InventoryID(inventoryID.String()), clothesKey, clothesName, tagColor, time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC))
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &visibilityAuthorizer{t: t, tenantID: tenantID, visible: []inventory.InventoryID{inventoryID}},
		Tenants:    &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
		}},
		Search: &recordingAssetSearchRepository{items: []ports.AssetSearchResult{{
			Type:         search.ResultTypeAsset,
			TenantID:     tenantID,
			Inventory:    inventoryItem(inventoryID.String(), tenantID.String(), "Home"),
			Asset:        item,
			AssignedTags: []assettag.Tag{tag, clothesTag},
		}}},
		Audit:            &fakeAuditRepository{},
		DefaultPageLimit: 10,
		MaxPageLimit:     20,
	})

	result, err := application.executeRealtimeVoiceSearchTool(context.Background(), RealtimeVoiceSession{Principal: identity.Principal{ID: "viewer"}, TenantID: tenantID, InventoryID: inventoryID}, ports.AgentToolCall{ID: "search-baby", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "baby clothes", "limit": float64(10)}})
	if err != nil {
		t.Fatal(err)
	}
	var tool struct {
		Items []struct {
			TagNames []string `json:"tagNames"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(result.Content), &tool); err != nil {
		t.Fatal(err)
	}
	if len(tool.Items) != 1 || len(tool.Items[0].TagNames) != 2 || tool.Items[0].TagNames[0] != "Baby" || tool.Items[0].TagNames[1] != "Clothes" {
		t.Fatalf("tag evidence lost in tool result: %s", result.Content)
	}
	observations, err := realtimeVoiceInvestigationObservationsFromToolResult(1, agentmodel.SemanticReferenceKey("subject"), "baby clothes", result)
	if err != nil || len(observations) != 1 {
		t.Fatalf("observations: %+v %v", observations, err)
	}
	raw, err := json.Marshal(observations[0])
	if err != nil {
		t.Fatal(err)
	}
	var observation struct {
		TagNames []string `json:"tagNames"`
	}
	if err := json.Unmarshal(raw, &observation); err != nil {
		t.Fatal(err)
	}
	if len(observation.TagNames) != 2 || observation.TagNames[0] != "Baby" || observation.TagNames[1] != "Clothes" {
		t.Fatalf("tag evidence lost before assessment: %s", raw)
	}
}

func TestDetailObservationRetainsPriorSearchTagEvidence(t *testing.T) {
	prior := agentmodel.CandidateObservation{CandidateID: "clothes", TagNames: []string{"Baby", "Clothes"}, MatchedProbes: []string{"baby clothes"}}
	detail := agentmodel.CandidateObservation{CandidateID: "clothes", Description: "Stored upstairs"}
	merged := mergeRealtimeVoiceInvestigationObservation(prior, detail)
	if len(merged.TagNames) != 2 || merged.TagNames[0] != "Baby" {
		t.Fatalf("detail erased tags: %+v", merged)
	}
	prior.TagNames[0] = "Changed"
	if merged.TagNames[0] != "Baby" {
		t.Fatal("merged tags alias earlier evidence")
	}
}

func TestFreshEmptySearchClearsPriorTagEvidence(t *testing.T) {
	prior := agentmodel.CandidateObservation{CandidateID: "clothes", TagNames: []string{"Baby"}}
	fresh := realtimeVoiceInvestigationObservationFromItem(2, agentmodel.SemanticReferenceSubject, "clothes", realtimeVoiceAssetToolItem{AssetID: "clothes", TagNames: []string{}}, nil)
	merged := mergeRealtimeVoiceInvestigationObservation(prior, fresh)
	if merged.TagNames == nil || len(merged.TagNames) != 0 {
		t.Fatalf("stale tags survived fresh search: %v", merged.TagNames)
	}
}
