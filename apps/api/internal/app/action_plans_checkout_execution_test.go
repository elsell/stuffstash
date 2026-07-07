package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteActionPlanChecksOutAssetAndMarksExecuted(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindCheckoutAsset, `{"assetId":"asset-1","details":"using at desk"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		item.ID: item,
	}}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"checkout-1", "undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute checkout action plan: %v", err)
	}
	if result.Record.State != actionplan.StateExecuted || result.Record.ExecutedAt.IsZero() {
		t.Fatalf("unexpected executed plan: %+v", result.Record)
	}
	checkout := assets.checkouts[asset.CheckoutID("checkout-1")]
	if checkout.State != asset.CheckoutStateOpen || checkout.AssetID != asset.ID("asset-1") || checkout.CheckedOutByPrincipal != "user-1" || checkout.CheckoutDetails.String() != "using at desk" {
		t.Fatalf("expected open checkout, got %+v", checkout)
	}
	if len(assets.auditRecords) != 1 || assets.auditRecords[0].Action != audit.ActionAssetCheckedOut {
		t.Fatalf("expected checkout audit record, got %+v", assets.auditRecords)
	}
	if len(assets.undoables) != 1 || assets.undoables["undo-1"].OriginalAction != audit.ActionAssetCheckedOut || assets.undoables["undo-1"].AfterCheckout == nil {
		t.Fatalf("expected checkout undoable operation, got %+v", assets.undoables)
	}
	if len(result.CommandResults) != 1 || result.CommandResults[0].AssetID != "asset-1" || result.CommandResults[0].Operation != "checkout" {
		t.Fatalf("expected checkout command result, got %+v", result.CommandResults)
	}
}

func TestExecuteActionPlanReturnsCheckedOutAssetAndMarksExecuted(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindReturnAsset, `{"assetId":"asset-1","details":"back in bin"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	details, _ := asset.NewCheckoutDetails("using at desk")
	checkout := asset.Checkout{
		ID:                    asset.CheckoutID("checkout-1"),
		TenantID:              asset.TenantID("tenant-home"),
		InventoryID:           asset.InventoryID("inventory-home"),
		AssetID:               item.ID,
		State:                 asset.CheckoutStateOpen,
		CheckedOutAt:          time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC),
		CheckedOutByPrincipal: "user-1",
		CheckoutDetails:       details,
		CreatedAt:             time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC),
	}
	assets := &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			item.ID: item,
		},
		checkouts: map[asset.CheckoutID]asset.Checkout{
			checkout.ID: checkout,
		},
	}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err != nil {
		t.Fatalf("execute return action plan: %v", err)
	}
	returned := assets.checkouts[checkout.ID]
	if returned.State != asset.CheckoutStateReturned || returned.ReturnedByPrincipal != "user-1" || returned.ReturnDetails.String() != "back in bin" {
		t.Fatalf("expected returned checkout, got %+v", returned)
	}
	if len(assets.auditRecords) != 1 || assets.auditRecords[0].Action != audit.ActionAssetReturned {
		t.Fatalf("expected return audit record, got %+v", assets.auditRecords)
	}
	if len(assets.undoables) != 1 || assets.undoables["undo-1"].OriginalAction != audit.ActionAssetReturned || assets.undoables["undo-1"].BeforeCheckout == nil || assets.undoables["undo-1"].AfterCheckout == nil {
		t.Fatalf("expected return undoable operation, got %+v", assets.undoables)
	}
	if result.Record.State != actionplan.StateExecuted || len(result.CommandResults) != 1 || result.CommandResults[0].Operation != "return" {
		t.Fatalf("unexpected return execution result: %+v", result)
	}
}

func TestExecuteActionPlanMarksCheckoutPlanFailedWhenExecutionRejectsStaleState(t *testing.T) {
	t.Parallel()

	repository := &fakeActionPlanRepository{
		records: map[string]ports.ActionPlanRecord{
			"plan-1": actionPlanRecordWithCommand("plan-1", actionplan.StateApproved, actionplan.CommandKindCheckoutAsset, `{"assetId":"asset-1"}`),
		},
	}
	item := assetItem("asset-1", "tenant-home", "inventory-home", asset.KindItem, "")
	assets := &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			item.ID: item,
		},
		checkOutAssetErr: ports.ErrForbidden,
	}
	application := newActionPlanExecutionTestApp(repository, assets, &fakeIDGenerator{ids: []string{"checkout-1", "undo-1", "audit-1"}})

	result, err := application.ExecuteActionPlanDetailed(context.Background(), ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	})
	if err == nil {
		t.Fatalf("expected stale checkout execution error")
	}
	if result.Record.State != actionplan.StateFailed || result.Record.FailedAt.IsZero() {
		t.Fatalf("expected failed plan after execution rejection, got %+v", result.Record)
	}
	if repository.records["plan-1"].State != actionplan.StateFailed {
		t.Fatalf("expected repository plan to be failed, got %+v", repository.records["plan-1"])
	}
}
