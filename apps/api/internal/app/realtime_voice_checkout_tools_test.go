package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceListCheckedOutAssetsToolReturnsCurrentCheckoutState(t *testing.T) {
	t.Parallel()

	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	now := time.Date(2026, 6, 26, 17, 30, 0, 0, time.UTC)
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			drill.ID: drill,
		},
		checkouts: map[asset.CheckoutID]asset.Checkout{
			asset.CheckoutID("checkout-drill"): {
				ID:                    asset.CheckoutID("checkout-drill"),
				TenantID:              asset.TenantID("tenant-home"),
				InventoryID:           asset.InventoryID("inventory-home"),
				AssetID:               drill.ID,
				State:                 asset.CheckoutStateOpen,
				CheckedOutAt:          now,
				CheckedOutByPrincipal: "user-1",
			},
		},
	}, &fakeIDGenerator{})

	result, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-checked-out",
		Name:      RealtimeVoiceToolListCheckedOutAssets,
		Arguments: map[string]any{"limit": 5},
	}, map[string]struct{}{})
	if err != nil {
		t.Fatalf("execute checked-out list tool: %v", err)
	}
	for _, required := range []string{
		`"tool":"list_checked_out_assets"`,
		`"assetId":"drill-1"`,
		`"title":"Asset drill-1"`,
		`"currentCheckout":{"id":"checkout-drill"`,
		`"checkedOutByPrincipalId":"user-1"`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected checked-out tool result to include %q, got %s", required, result.Content)
		}
	}
}

func TestRealtimeVoiceListAssetCheckoutHistoryToolRequiresVisibleAsset(t *testing.T) {
	t.Parallel()

	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{}, &fakeIDGenerator{})

	_, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-history",
		Name:      RealtimeVoiceToolListAssetCheckoutHistory,
		Arguments: map[string]any{"assetId": "hidden-asset", "limit": 5},
	}, map[string]struct{}{})
	if err == nil {
		t.Fatalf("expected invisible checkout history asset to be rejected")
	}
}

func TestRealtimeVoiceListAssetCheckoutHistoryToolReturnsBoundedCheckoutRecords(t *testing.T) {
	t.Parallel()

	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	now := time.Date(2026, 6, 26, 17, 30, 0, 0, time.UTC)
	details, _ := asset.NewCheckoutDetails("Loaned to Sam")
	returnDetails, _ := asset.NewCheckoutDetails("Back in the tool bin")
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			drill.ID: drill,
		},
		checkouts: map[asset.CheckoutID]asset.Checkout{
			asset.CheckoutID("checkout-drill"): {
				ID:                    asset.CheckoutID("checkout-drill"),
				TenantID:              asset.TenantID("tenant-home"),
				InventoryID:           asset.InventoryID("inventory-home"),
				AssetID:               drill.ID,
				State:                 asset.CheckoutStateReturned,
				CheckedOutAt:          now.Add(-2 * time.Hour),
				CheckedOutByPrincipal: "user-1",
				CheckoutDetails:       details,
				ReturnedAt:            now,
				ReturnedByPrincipal:   "user-2",
				ReturnDetails:         returnDetails,
			},
		},
	}, &fakeIDGenerator{})

	result, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-history",
		Name:      RealtimeVoiceToolListAssetCheckoutHistory,
		Arguments: map[string]any{"assetId": "drill-1", "limit": 5},
	}, map[string]struct{}{"drill-1": {}})
	if err != nil {
		t.Fatalf("execute checkout history tool: %v", err)
	}
	for _, required := range []string{
		`"tool":"list_asset_checkout_history"`,
		`"assetId":"drill-1"`,
		`"state":"returned"`,
		`"checkoutDetails":"Loaned to Sam"`,
		`"returnDetails":"Back in the tool bin"`,
		`"returnedByPrincipalId":"user-2"`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected checkout history tool result to include %q, got %s", required, result.Content)
		}
	}
}

func checkoutToolSession() RealtimeVoiceSession {
	return RealtimeVoiceSession{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
	}
}
