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

func TestExecuteActionPlanDetailedReturnsCreatedAssetCommandResult(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindCreateAsset, `{"title":"Water bottle","kind":"item","description":"Blue bottle"}`),
		},
	}
	repository.records["plan-1"].Commands[0].ID = "cmd-water-bottle"
	application := newActionPlanExecutionTestApp(repository, &fakeAssetRepository{}, &fakeIDGenerator{ids: []string{"asset-1", "undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if result.Record.State != actionplan.StateExecuted {
		t.Fatalf("expected executed plan, got %+v", result.Record)
	}
	if len(result.CommandResults) != 1 {
		t.Fatalf("expected one command result, got %+v", result.CommandResults)
	}
	if got := result.CommandResults[0]; got.CommandID != "cmd-water-bottle" || got.AssetID != "asset-1" || got.Operation != "create" || got.AssetKind != "item" {
		t.Fatalf("unexpected command result: %+v", got)
	}
}

func TestExecuteActionPlanDetailedReturnsMovedAssetCommandResult(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindMoveAsset, `{"assetId":"asset-1","parentAssetId":"location-1"}`),
		},
	}
	repository.records["plan-1"].Commands[0].ID = "cmd-move-shed"
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	location := assetItem("location-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID:     item,
		location.ID: location,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if result.Record.State != actionplan.StateExecuted {
		t.Fatalf("expected executed plan, got %+v", result.Record)
	}
	if len(result.CommandResults) != 1 {
		t.Fatalf("expected one command result, got %+v", result.CommandResults)
	}
	if got := result.CommandResults[0]; got.CommandID != "cmd-move-shed" || got.AssetID != "asset-1" || got.Operation != "move" || got.AssetKind != "item" {
		t.Fatalf("unexpected command result: %+v", got)
	}
}

func TestExecuteActionPlanDetailedReturnsArchivedAssetCommandResult(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`),
		},
	}
	repository.records["plan-1"].Commands[0].ID = "cmd-archive-drill"
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if result.Record.State != actionplan.StateExecuted {
		t.Fatalf("expected executed plan, got %+v", result.Record)
	}
	if len(result.CommandResults) != 1 {
		t.Fatalf("expected one command result, got %+v", result.CommandResults)
	}
	if got := result.CommandResults[0]; got.CommandID != "cmd-archive-drill" || got.AssetID != "asset-1" || got.Operation != "archive" || got.AssetKind != "item" {
		t.Fatalf("unexpected command result: %+v", got)
	}
}

func TestExecuteActionPlanDetailedReturnsRestoredAssetCommandResult(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindRestoreAsset, `{"assetId":"asset-1"}`),
		},
	}
	repository.records["plan-1"].Commands[0].ID = "cmd-restore-drill"
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindContainer, "")
	item.LifecycleState = asset.LifecycleStateArchived
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if result.Record.State != actionplan.StateExecuted {
		t.Fatalf("expected executed plan, got %+v", result.Record)
	}
	if len(result.CommandResults) != 1 {
		t.Fatalf("expected one command result, got %+v", result.CommandResults)
	}
	if got := result.CommandResults[0]; got.CommandID != "cmd-restore-drill" || got.AssetID != "asset-1" || got.Operation != "restore" || got.AssetKind != "container" {
		t.Fatalf("unexpected command result: %+v", got)
	}
}
