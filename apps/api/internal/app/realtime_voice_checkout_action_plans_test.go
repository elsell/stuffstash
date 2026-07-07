package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceCheckoutActionPlanArgsValidateVisibleAsset(t *testing.T) {
	t.Parallel()

	parsed, err := parseRealtimeVoiceActionPlanArgs(map[string]any{
		"commandKind":                "checkout_asset",
		"intentSummary":              "Check out the drill.",
		"modelInterpretationSummary": "The user wants to mark the visible drill as checked out.",
		"confirmationSummary":        "Check out drill?",
		"commandSummary":             "Check out drill",
		"arguments": map[string]any{
			"assetId": "drill-1",
			"details": "using at bench",
		},
	}, "Check out the drill. I'm using it at the bench.")
	if err != nil {
		t.Fatalf("parse checkout action-plan args: %v", err)
	}
	if len(parsed.Commands) != 1 || parsed.Commands[0].Kind != actionplan.CommandKindCheckoutAsset {
		t.Fatalf("unexpected checkout commands: %+v", parsed.Commands)
	}
	if err := validateRealtimeVoiceActionPlanVisibleIDs(parsed.Commands, map[string]struct{}{"drill-1": {}}); err != nil {
		t.Fatalf("validate visible checkout asset: %v", err)
	}
}

func TestRealtimeVoiceReturnActionPlanArgsValidateVisibleAsset(t *testing.T) {
	t.Parallel()

	parsed, err := parseRealtimeVoiceActionPlanArgs(map[string]any{
		"intentSummary":              "Return the drill.",
		"modelInterpretationSummary": "The user wants to mark the visible checked-out drill as returned.",
		"confirmationSummary":        "Return drill?",
		"commands": []any{map[string]any{
			"id":      "cmd-return-drill",
			"kind":    "return_asset",
			"summary": "Return drill",
			"arguments": map[string]any{
				"assetId": "drill-1",
				"details": "back in tool bin",
			},
		}},
	}, "The drill is back in the tool bin.")
	if err != nil {
		t.Fatalf("parse return action-plan args: %v", err)
	}
	if len(parsed.Commands) != 1 || parsed.Commands[0].Kind != actionplan.CommandKindReturnAsset {
		t.Fatalf("unexpected return commands: %+v", parsed.Commands)
	}
	if err := validateRealtimeVoiceActionPlanVisibleIDs(parsed.Commands, map[string]struct{}{"drill-1": {}}); err != nil {
		t.Fatalf("validate visible return asset: %v", err)
	}
}

func TestRealtimeVoiceCheckoutProposalCommandIncludesAssetDisplayContext(t *testing.T) {
	t.Parallel()

	item := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			item.ID: item,
		},
	}, &fakeIDGenerator{})

	command := ports.ActionPlanCommandRecord{
		ID:            "cmd-checkout-drill",
		Kind:          actionplan.CommandKindCheckoutAsset,
		Summary:       "Check out drill",
		ArgumentsJSON: []byte(`{"assetId":"drill-1","details":"using at bench"}`),
	}
	proposal, err := application.realtimeVoiceActionPlanCommand(context.Background(), RealtimeVoiceSession{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
	}, command)
	if err != nil {
		t.Fatalf("build checkout proposal command: %v", err)
	}
	if proposal.Kind != string(actionplan.CommandKindCheckoutAsset) || proposal.Operation != "checkout" || proposal.Title != "Asset drill-1" || proposal.AssetKind != asset.KindItem.String() {
		t.Fatalf("unexpected checkout proposal command: %+v", proposal)
	}
}
