package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceGetAssetDetailToolRequiresVisibleAsset(t *testing.T) {
	t.Parallel()

	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{}, &fakeIDGenerator{})

	_, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-detail",
		Name:      RealtimeVoiceToolGetAssetDetail,
		Arguments: map[string]any{"assetId": "hidden-asset"},
	}, map[string]struct{}{})
	if err == nil {
		t.Fatalf("expected invisible detail asset to be rejected")
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

	result, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
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
