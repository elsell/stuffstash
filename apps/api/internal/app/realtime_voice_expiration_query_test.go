package app

import (
	"context"
	"encoding/json"
	"fmt"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestExpirationQueryContinuesPastEmptyScanWithoutLosingMatches(t *testing.T) {
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	application.customAssetTypes = store
	application.notificationService = notificationapp.New(notificationapp.Dependencies{Authorizer: application.authorizer, Inventories: store, Preferences: store, Audit: store, Clock: application.clock, IDs: application.ids})
	kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
	kind.ExpirationEnabled = true
	if err := store.SaveCustomAssetType(context.Background(), kind, audit.Record{ID: "seed-type"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 201; i++ {
		item := realtimeVoiceInvestigationAsset(fmt.Sprintf("a-%03d", i), "Undated", asset.KindItem, "")
		seedRealtimeVoiceLoopAsset(t, store, item, fmt.Sprintf("seed-%d", i))
	}
	item := realtimeVoiceInvestigationAsset("z-target", "Bottle", asset.KindItem, "")
	item.CustomAssetTypeID = "medicine"
	item.Expiration, _ = expirationdate.ParseDate("2026-07", expirationdate.Month)
	seedRealtimeVoiceLoopAsset(t, store, item, "seed-target")
	call := ports.AgentToolCall{ID: "query", Name: RealtimeVoiceToolQueryExpiringAssets, Arguments: map[string]any{"status": "all"}}
	first, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, map[string]struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	var output realtimeVoiceAssetToolOutput
	if err := json.Unmarshal([]byte(first.Content), &output); err != nil {
		t.Fatal(err)
	}
	if output.Count != 0 || !output.HasMore || output.NextCursor == "" {
		t.Fatalf("empty scan claimed complete: %s", first.Content)
	}
	call.Arguments["cursor"] = output.NextCursor
	second, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, map[string]struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	output = realtimeVoiceAssetToolOutput{}
	if err := json.Unmarshal([]byte(second.Content), &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Items) != 1 || output.Items[0].AssetID != "z-target" || output.HasMore {
		t.Fatalf("continuation lost target: %s", second.Content)
	}
	call.Arguments["status"] = "expired"
	if _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), call, map[string]struct{}{}); err == nil {
		t.Fatal("cursor reused with different filters")
	}
}
