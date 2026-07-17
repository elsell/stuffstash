package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceListAuthorizedAssetsToolCanListRootLevelAssets(t *testing.T) {
	t.Parallel()

	rootBin := assetItem("root-bin-1", "tenant-home", "inventory-home", asset.KindContainer, "")
	rootBinTitle, _ := asset.NewTitle("Root bin")
	rootBin.Title = rootBinTitle
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	nestedCable := assetItem("nested-cable-1", "tenant-home", "inventory-home", asset.KindItem, "root-bin-1")
	nestedCableTitle, _ := asset.NewTitle("Nested cable")
	nestedCable.Title = nestedCableTitle
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			rootBin.ID:     rootBin,
			office.ID:      office,
			nestedCable.ID: nestedCable,
		},
	}, &fakeIDGenerator{})

	result, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), ports.AgentToolCall{
		ID:   "tool-call-root-list",
		Name: RealtimeVoiceToolListAuthorizedAssets,
		Arguments: map[string]any{
			"parentScope": "root",
			"limit":       10,
		},
	}, map[string]struct{}{})
	if err != nil {
		t.Fatalf("execute root list tool: %v", err)
	}
	for _, required := range []string{
		`"tool":"list_authorized_assets"`,
		`"parentScope":"root"`,
		`"title":"Root bin"`,
		`"title":"Office"`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected root list result to include %q, got %s", required, result.Content)
		}
	}
	if strings.Contains(result.Content, "Nested cable") {
		t.Fatalf("expected root list result to exclude nested asset, got %s", result.Content)
	}
}

func TestRealtimeVoiceListAuthorizedAssetsRejectsRootScopeWithParentFilters(t *testing.T) {
	t.Parallel()

	for _, arguments := range []map[string]any{
		{"parentScope": "root", "parentTitle": "Office"},
		{"parentScope": "root", "locationTitle": "Office"},
		{"parentScope": "root", "parentAssetId": "office-1"},
		{"parentAssetId": "office-1", "parentTitle": "Office"},
		{"parentAssetId": "office-1", "locationTitle": "Office"},
	} {
		if _, err := parseRealtimeVoiceListArgs(arguments); err == nil {
			t.Fatalf("expected root parent scope with parent/location filter to be rejected for arguments %+v", arguments)
		}
	}
}

func TestRealtimeVoiceListAuthorizedAssetsRequiresParentIDVisibility(t *testing.T) {
	t.Parallel()

	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{}, &fakeIDGenerator{})
	_, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), ports.AgentToolCall{
		ID:        "tool-call-hidden-parent",
		Name:      RealtimeVoiceToolListAuthorizedAssets,
		Arguments: map[string]any{"parentAssetId": "hidden-parent", "limit": 10},
	}, map[string]struct{}{})
	if err == nil {
		t.Fatal("expected an unobserved parent asset ID to be rejected")
	}
}
