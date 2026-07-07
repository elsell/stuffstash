package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestUndoCheckoutMarksCheckoutUndone(t *testing.T) {
	application, repository := checkoutUndoApplication()
	checkout := checkoutUndoRecord(asset.CheckoutStateOpen)
	repository.checkouts = map[asset.CheckoutID]asset.Checkout{checkout.ID: checkout}
	repository.undoables = map[string]ports.UndoableOperation{
		"operation-one": checkoutUndoableOperation("operation-one", audit.ActionAssetCheckedOut, ports.UndoableOperationAvailable, nil, &checkout),
	}

	item, err := application.UndoOperation(context.Background(), checkoutUndoInput("operation-one"))
	if err != nil {
		t.Fatalf("undo checkout: %v", err)
	}
	if item.ID != checkout.AssetID {
		t.Fatalf("expected affected asset, got %+v", item)
	}
	result := repository.checkouts[checkout.ID]
	if result.State != asset.CheckoutStateUndone || !result.ReturnedAt.IsZero() || result.ReturnedByPrincipal != "" || result.ReturnDetails.String() != "" {
		t.Fatalf("expected undone checkout without return fields, got %+v", result)
	}
	operation := repository.undoables["operation-one"]
	if operation.Status != ports.UndoableOperationUndone || operation.UndoAuditRecordID.String() == "" {
		t.Fatalf("expected undone operation with audit link, got %+v", operation)
	}
}

func TestUndoReturnReopensCheckout(t *testing.T) {
	application, repository := checkoutUndoApplication()
	open := checkoutUndoRecord(asset.CheckoutStateOpen)
	returned := open
	returned.State = asset.CheckoutStateReturned
	returned.ReturnedAt = open.CheckedOutAt.Add(time.Hour)
	returned.ReturnedByPrincipal = "editor-two"
	returned.ReturnDetails, _ = asset.NewCheckoutDetails("back")
	returned.UpdatedAt = returned.ReturnedAt
	repository.checkouts = map[asset.CheckoutID]asset.Checkout{returned.ID: returned}
	repository.undoables = map[string]ports.UndoableOperation{
		"operation-return": checkoutUndoableOperation("operation-return", audit.ActionAssetReturned, ports.UndoableOperationAvailable, &open, &returned),
	}

	_, err := application.UndoOperation(context.Background(), checkoutUndoInput("operation-return"))
	if err != nil {
		t.Fatalf("undo return: %v", err)
	}
	result := repository.checkouts[open.ID]
	if result.State != asset.CheckoutStateOpen || !result.ReturnedAt.IsZero() || result.ReturnedByPrincipal != "" || result.ReturnDetails.String() != "" {
		t.Fatalf("expected reopened checkout, got %+v", result)
	}
}

func TestRedoReturnReappliesReturn(t *testing.T) {
	application, repository := checkoutUndoApplication()
	open := checkoutUndoRecord(asset.CheckoutStateOpen)
	returned := open
	returned.State = asset.CheckoutStateReturned
	returned.ReturnedAt = open.CheckedOutAt.Add(time.Hour)
	returned.ReturnedByPrincipal = "editor-two"
	returned.ReturnDetails, _ = asset.NewCheckoutDetails("back")
	returned.UpdatedAt = returned.ReturnedAt
	repository.checkouts = map[asset.CheckoutID]asset.Checkout{open.ID: open}
	repository.undoables = map[string]ports.UndoableOperation{
		"operation-return": checkoutUndoableOperation("operation-return", audit.ActionAssetReturned, ports.UndoableOperationUndone, &open, &returned),
	}

	_, err := application.RedoOperation(context.Background(), checkoutUndoInput("operation-return"))
	if err != nil {
		t.Fatalf("redo return: %v", err)
	}
	result := repository.checkouts[open.ID]
	if result.State != asset.CheckoutStateReturned || result.ReturnedByPrincipal != "editor-two" || result.ReturnDetails.String() != "back" {
		t.Fatalf("expected returned checkout, got %+v", result)
	}
}

func TestUndoCheckoutRejectsLaterCheckoutHistory(t *testing.T) {
	application, repository := checkoutUndoApplication()
	checkout := checkoutUndoRecord(asset.CheckoutStateOpen)
	later := checkout
	later.ID = asset.CheckoutID("checkout-later")
	later.State = asset.CheckoutStateReturned
	later.CheckedOutAt = checkout.CheckedOutAt.Add(time.Hour)
	later.CreatedAt = later.CheckedOutAt
	later.UpdatedAt = later.CheckedOutAt
	later.ReturnedAt = later.CheckedOutAt.Add(time.Hour)
	later.ReturnedByPrincipal = "editor-one"
	repository.checkouts = map[asset.CheckoutID]asset.Checkout{
		checkout.ID: checkout,
		later.ID:    later,
	}
	repository.undoables = map[string]ports.UndoableOperation{
		"operation-one": checkoutUndoableOperation("operation-one", audit.ActionAssetCheckedOut, ports.UndoableOperationAvailable, nil, &checkout),
	}

	_, err := application.UndoOperation(context.Background(), checkoutUndoInput("operation-one"))
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for stale checkout undo, got %v", err)
	}
}

func checkoutUndoApplication() (App, *fakeAssetRepository) {
	item := assetItem("asset-one", "tenant-one", "inventory-one", asset.KindItem, "")
	repository := &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{item.ID: item},
	}
	return New(Dependencies{
		Observer:        &fakeObserver{},
		Authorizer:      &fakeAuthorizer{},
		Tenants:         &fakeTenantRepository{exists: true},
		Inventories:     &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Home")}},
		Assets:          repository,
		Checkouts:       repository,
		AssetUnitOfWork: repository,
		Undoables:       repository,
		Audit:           &fakeAuditRepository{},
		IDs:             &fakeIDGenerator{ids: []string{"audit-undo", "audit-redo"}},
	}), repository
}

func checkoutUndoInput(operationID string) ApplyUndoableOperationInput {
	return ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: operationID,
	}
}

func checkoutUndoRecord(state asset.CheckoutState) asset.Checkout {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	return asset.Checkout{
		ID:                    asset.CheckoutID("checkout-one"),
		TenantID:              asset.TenantID("tenant-one"),
		InventoryID:           asset.InventoryID("inventory-one"),
		AssetID:               asset.ID("asset-one"),
		State:                 state,
		CheckedOutAt:          now,
		CheckedOutByPrincipal: "editor-one",
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func checkoutUndoableOperation(id string, action audit.Action, status ports.UndoableOperationStatus, before *asset.Checkout, after *asset.Checkout) ports.UndoableOperation {
	return ports.UndoableOperation{
		ID:             id,
		TenantID:       tenant.ID("tenant-one"),
		InventoryID:    inventory.InventoryID("inventory-one"),
		PrincipalID:    identity.PrincipalID("editor-one"),
		Source:         audit.SourceAPI,
		TargetType:     audit.TargetAsset,
		TargetID:       "asset-one",
		OriginalAction: action,
		Status:         status,
		CreatedAt:      time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		BeforeCheckout: before,
		AfterCheckout:  after,
	}
}
