package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteActionPlanCreatesDependentHierarchyAndMarksExecuted(t *testing.T) {
	t.Parallel()

	record := actionPlanRecord("plan-1", actionplan.StateApproved)
	record.Commands = []ports.ActionPlanCommandRecord{
		{
			ID:            "cmd-box",
			Kind:          actionplan.CommandKindCreateAsset,
			Summary:       "Create Box underneath the TV in Living room",
			ArgumentsJSON: []byte(`{"title":"Box underneath the TV","kind":"container","parentAssetId":"location-1"}`),
		},
		{
			ID:            "cmd-remote",
			Kind:          actionplan.CommandKindCreateAsset,
			Summary:       "Create Apple TV remote inside Box underneath the TV",
			ArgumentsJSON: []byte(`{"title":"Apple TV remote","kind":"item","parentCommandId":"cmd-box"}`),
		},
	}
	location := assetItem("location-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{"plan-1": record},
	}
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{location.ID: location}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"box-1", "undo-box", "audit-box", "remote-1", "undo-remote", "audit-remote"}})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("expected executed plan, got %+v", executed)
	}
	box := assets.items[asset.ID("box-1")]
	if box.Kind != asset.KindContainer || box.Title.String() != "Box underneath the TV" || box.ParentAssetID != location.ID {
		t.Fatalf("expected created box inside location, got %+v", box)
	}
	remote := assets.items[asset.ID("remote-1")]
	if remote.Kind != asset.KindItem || remote.Title.String() != "Apple TV remote" || remote.ParentAssetID != box.ID {
		t.Fatalf("expected created remote inside new box, got %+v", remote)
	}
	if len(assets.auditRecords) != 2 || assets.auditRecords[0].Action != audit.ActionAssetCreated || assets.auditRecords[1].Action != audit.ActionAssetCreated {
		t.Fatalf("expected two audited creates, got %+v", assets.auditRecords)
	}
}

func TestExecuteActionPlanCreatesMissingLocationThenMovesVisibleAsset(t *testing.T) {
	t.Parallel()

	record := actionPlanRecord("plan-1", actionplan.StateApproved)
	record.Commands = []ports.ActionPlanCommandRecord{
		{
			ID:            "cmd-kitchen",
			Kind:          actionplan.CommandKindCreateLocation,
			Summary:       "Create Kitchen",
			ArgumentsJSON: []byte(`{"title":"Kitchen","kind":"location"}`),
		},
		{
			ID:            "cmd-move-water-bottle",
			Kind:          actionplan.CommandKindMoveAsset,
			Summary:       "Move Water bottle to Kitchen",
			ArgumentsJSON: []byte(`{"assetId":"water-bottle","parentCommandId":"cmd-kitchen"}`),
		},
	}
	office := assetItem("office", "tenant-home", "inventory-home", asset.KindLocation, "")
	waterBottle := assetItem("water-bottle", "tenant-home", "inventory-home", asset.KindItem, office.ID.String())
	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{"plan-1": record},
	}
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		office.ID:      office,
		waterBottle.ID: waterBottle,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"kitchen-id", "undo-kitchen", "audit-kitchen", "move-undo", "move-audit"}})

	executed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute action plan: %v", err)
	}
	if executed.State != actionplan.StateExecuted || executed.ExecutedAt.IsZero() {
		t.Fatalf("expected executed plan, got %+v", executed)
	}
	kitchen := assets.items[asset.ID("kitchen-id")]
	if kitchen.Kind != asset.KindLocation || kitchen.Title.String() != "Kitchen" {
		t.Fatalf("expected created kitchen location, got %+v", kitchen)
	}
	moved := assets.items[waterBottle.ID]
	if moved.ParentAssetID != kitchen.ID {
		t.Fatalf("expected water bottle moved into kitchen, got %+v", moved)
	}
	if len(assets.auditRecords) != 2 || assets.auditRecords[0].Action != audit.ActionAssetCreated || assets.auditRecords[1].Action != audit.ActionAssetMoved {
		t.Fatalf("expected create and move audits, got %+v", assets.auditRecords)
	}
}

func TestExecuteActionPlanRejectsMoveIntoNewDescendant(t *testing.T) {
	t.Parallel()

	record := actionPlanRecord("plan-1", actionplan.StateApproved)
	record.Commands = []ports.ActionPlanCommandRecord{
		{
			ID:            "cmd-shelf",
			Kind:          actionplan.CommandKindCreateAsset,
			Summary:       "Create Shelf inside Garage",
			ArgumentsJSON: []byte(`{"title":"Shelf","kind":"container","parentAssetId":"garage"}`),
		},
		{
			ID:            "cmd-move-garage",
			Kind:          actionplan.CommandKindMoveAsset,
			Summary:       "Move Garage into Shelf",
			ArgumentsJSON: []byte(`{"assetId":"garage","parentCommandId":"cmd-shelf"}`),
		},
	}
	garage := assetItem("garage", "tenant-home", "inventory-home", asset.KindLocation, "")
	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{"plan-1": record},
	}
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{garage.ID: garage}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"shelf-id", "undo-shelf", "audit-shelf"}})

	_, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, exists := assets.items[asset.ID("shelf-id")]; exists {
		t.Fatalf("expected rejected plan not to create shelf")
	}
}

func TestCreateActionPlanRejectsForwardDependentCreateReference(t *testing.T) {
	t.Parallel()

	application := newActionPlanTestApp(&fakeActionPlanRepository{}, &fakeIDGenerator{ids: []string{"plan-1"}}, nil)
	_, err := application.CreateActionPlan(context.Background(), CreateActionPlanInput{
		Principal:           identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:            tenant.ID("tenant-home"),
		InventoryID:         inventory.InventoryID("inventory-home"),
		Source:              "mobile_voice",
		ConfirmationSummary: "Create Apple TV remote in a new box?",
		Commands: []ActionPlanCommandInput{
			{
				ID:      "cmd-remote",
				Kind:    actionplan.CommandKindCreateAsset,
				Summary: "Create Apple TV remote",
				Arguments: map[string]any{
					"title":           "Apple TV remote",
					"kind":            "item",
					"parentCommandId": "cmd-box",
				},
			},
			{
				ID:      "cmd-box",
				Kind:    actionplan.CommandKindCreateAsset,
				Summary: "Create box",
				Arguments: map[string]any{
					"title": "Box underneath the TV",
					"kind":  "container",
				},
			},
		},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateActionPlanRejectsMoveToForwardDependentCreateReference(t *testing.T) {
	t.Parallel()

	application := newActionPlanTestApp(&fakeActionPlanRepository{}, &fakeIDGenerator{ids: []string{"plan-1"}}, nil)
	_, err := application.CreateActionPlan(context.Background(), CreateActionPlanInput{
		Principal:           identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:            tenant.ID("tenant-home"),
		InventoryID:         inventory.InventoryID("inventory-home"),
		Source:              "mobile_voice",
		ConfirmationSummary: "Move Water bottle to Kitchen?",
		Commands: []ActionPlanCommandInput{
			{
				ID:      "cmd-move-water-bottle",
				Kind:    actionplan.CommandKindMoveAsset,
				Summary: "Move Water bottle to Kitchen",
				Arguments: map[string]any{
					"assetId":         "water-bottle",
					"parentCommandId": "cmd-kitchen",
				},
			},
			{
				ID:      "cmd-kitchen",
				Kind:    actionplan.CommandKindCreateLocation,
				Summary: "Create Kitchen",
				Arguments: map[string]any{
					"title": "Kitchen",
					"kind":  "location",
				},
			},
		},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestExecuteActionPlanRevalidatesPersistedDependentCreateCommands(t *testing.T) {
	t.Parallel()

	record := actionPlanRecord("plan-1", actionplan.StateApproved)
	record.Commands = []ports.ActionPlanCommandRecord{
		{
			ID:            "cmd-duplicate",
			Kind:          actionplan.CommandKindCreateAsset,
			Summary:       "Create Box underneath the TV",
			ArgumentsJSON: []byte(`{"title":"Box underneath the TV","kind":"container"}`),
		},
		{
			ID:            "cmd-duplicate",
			Kind:          actionplan.CommandKindCreateAsset,
			Summary:       "Create Apple TV remote",
			ArgumentsJSON: []byte(`{"title":"Apple TV remote","kind":"item","parentCommandId":"cmd-duplicate"}`),
		},
	}
	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{"plan-1": record},
	}
	assets := &fakeAssetRepository{}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"box-1", "undo-box", "audit-box", "remote-1", "undo-remote", "audit-remote"}})

	failed, err := application.ExecuteActionPlan(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if failed.State != actionplan.StateFailed || failed.FailedAt.IsZero() {
		t.Fatalf("expected failed action plan, got %+v", failed)
	}
	if len(assets.items) != 0 || len(assets.auditRecords) != 0 {
		t.Fatalf("expected no partial mutations, assets=%+v audit=%+v", assets.items, assets.auditRecords)
	}
}
