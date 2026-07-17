package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceGetAssetDetailToolRequiresVisibleAsset(t *testing.T) {
	t.Parallel()

	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{}, &fakeIDGenerator{})

	_, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), ports.AgentToolCall{
		ID:        "tool-call-detail",
		Name:      RealtimeVoiceToolGetAssetDetail,
		Arguments: map[string]any{"assetId": "hidden-asset"},
	}, map[string]struct{}{})
	if err == nil {
		t.Fatalf("expected invisible detail asset to be rejected")
	}
}

func TestRealtimeVoiceGetAssetDetailToolReturnsSafeCheckoutState(t *testing.T) {
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

	result, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), ports.AgentToolCall{
		ID:        "tool-call-detail",
		Name:      RealtimeVoiceToolGetAssetDetail,
		Arguments: map[string]any{"assetId": "drill-1"},
	}, map[string]struct{}{"drill-1": {}})
	if err != nil {
		t.Fatalf("execute detail tool: %v", err)
	}
	for _, forbidden := range []string{
		`"currentCheckout"`,
		`checkout-drill`,
		`checkedOutByPrincipalId`,
		`user-1`,
	} {
		if strings.Contains(result.Content, forbidden) {
			t.Fatalf("expected detail tool result not to include %q, got %s", forbidden, result.Content)
		}
	}
	for _, required := range []string{
		`"tool":"get_asset_detail"`,
		`"assetId":"drill-1"`,
		`"checkoutState":{"state":"open","checkedOut":true`,
		`"checkedOutAt":"2026-06-26T17:30:00Z"`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected detail tool result to include %q, got %s", required, result.Content)
		}
	}
}

func TestRealtimeVoiceGetAssetDetailToolReturnsBoundedVisibleAssetDetail(t *testing.T) {
	t.Parallel()

	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "office-1")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	waterBottle.Description = asset.NewDescription("Blue bottle")
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			office.ID:      office,
			waterBottle.ID: waterBottle,
		},
	}, &fakeIDGenerator{})

	result, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), ports.AgentToolCall{
		ID:        "tool-call-detail",
		Name:      RealtimeVoiceToolGetAssetDetail,
		Arguments: map[string]any{"assetId": "water-bottle-1"},
	}, map[string]struct{}{"water-bottle-1": {}})
	if err != nil {
		t.Fatalf("execute detail tool: %v", err)
	}
	for _, required := range []string{
		`"tool":"get_asset_detail"`,
		`"assetId":"water-bottle-1"`,
		`"title":"Water bottle"`,
		`"description":"Blue bottle"`,
		`"parentTitle":"Office"`,
		`"locationTitle":"Office"`,
		`"containmentPath":["Office","Water bottle"]`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected detail tool result to include %q, got %s", required, result.Content)
		}
	}
}
