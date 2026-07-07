package assets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCheckoutAssetCreatesOpenCheckoutAuditAndUndoableOperation(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	assetID := asset.ID("asset-one")
	principal := identity.Principal{ID: identity.PrincipalID("editor-one")}
	repository := newFakeCheckoutAssetRepository(activeCheckoutTestAsset(tenantID, inventoryID, assetID))
	unitOfWork := &fakeCheckoutUnitOfWork{repository: repository}
	service := New(Dependencies{
		Observer:        noopObserver{},
		Authorizer:      allowAuthorizer{},
		Tenants:         tenantExistsRepository{},
		Inventories:     inventoryRepository{item: activeCheckoutTestInventory(tenantID, inventoryID)},
		Assets:          repository,
		Checkouts:       repository,
		AssetUnitOfWork: unitOfWork,
		Undoables:       fakeCheckoutUndoables{},
		Audit:           auditRepository{},
		IDs:             &sequenceIDGenerator{ids: []string{"checkout-one", "operation-one", "audit-one"}},
		Clock:           staticClock{now: now},
	})

	checkout, err := service.CheckoutAsset(ctx, CheckoutAssetInput{
		Principal:   principal,
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		AssetID:     assetID,
		Details:     "  using at my desk  ",
	})
	if err != nil {
		t.Fatalf("CheckoutAsset returned error: %v", err)
	}
	if checkout.ID != asset.CheckoutID("checkout-one") || checkout.State != asset.CheckoutStateOpen || checkout.CheckoutDetails.String() != "using at my desk" {
		t.Fatalf("unexpected checkout: %+v", checkout)
	}
	if len(unitOfWork.auditRecords) != 1 || unitOfWork.auditRecords[0].Action != audit.ActionAssetCheckedOut || unitOfWork.auditRecords[0].Metadata["operation_id"] != "operation-one" {
		t.Fatalf("expected checkout audit with operation metadata, got %+v", unitOfWork.auditRecords)
	}
	if len(unitOfWork.undoables) != 1 || unitOfWork.undoables[0].OriginalAction != audit.ActionAssetCheckedOut || unitOfWork.undoables[0].AfterCheckout == nil {
		t.Fatalf("expected checkout undoable operation, got %+v", unitOfWork.undoables)
	}
}

func TestCheckoutAssetRejectsDuplicateOpenCheckout(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	assetID := asset.ID("asset-one")
	repository := newFakeCheckoutAssetRepository(activeCheckoutTestAsset(tenantID, inventoryID, assetID))
	repository.current = &asset.Checkout{
		ID:          asset.CheckoutID("checkout-existing"),
		TenantID:    asset.TenantID(tenantID.String()),
		InventoryID: asset.InventoryID(inventoryID.String()),
		AssetID:     assetID,
		State:       asset.CheckoutStateOpen,
	}
	service := New(Dependencies{
		Observer:        noopObserver{},
		Authorizer:      allowAuthorizer{},
		Tenants:         tenantExistsRepository{},
		Inventories:     inventoryRepository{item: activeCheckoutTestInventory(tenantID, inventoryID)},
		Assets:          repository,
		Checkouts:       repository,
		AssetUnitOfWork: &fakeCheckoutUnitOfWork{repository: repository},
		Undoables:       fakeCheckoutUndoables{},
		Audit:           auditRepository{},
		IDs:             &sequenceIDGenerator{ids: []string{"checkout-one", "operation-one", "audit-one"}},
		Clock:           staticClock{now: time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)},
	})

	_, err := service.CheckoutAsset(ctx, CheckoutAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor-one")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		AssetID:     assetID,
	})
	if err == nil {
		t.Fatal("expected duplicate checkout to fail")
	}
}

func TestReturnAssetClosesOpenCheckoutByDifferentEditor(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	assetID := asset.ID("asset-one")
	checkout := asset.Checkout{
		ID:                    asset.CheckoutID("checkout-one"),
		TenantID:              asset.TenantID(tenantID.String()),
		InventoryID:           asset.InventoryID(inventoryID.String()),
		AssetID:               assetID,
		State:                 asset.CheckoutStateOpen,
		CheckedOutAt:          now.Add(-time.Hour),
		CheckedOutByPrincipal: "editor-one",
		CreatedAt:             now.Add(-time.Hour),
		UpdatedAt:             now.Add(-time.Hour),
	}
	repository := newFakeCheckoutAssetRepository(activeCheckoutTestAsset(tenantID, inventoryID, assetID))
	repository.current = &checkout
	unitOfWork := &fakeCheckoutUnitOfWork{repository: repository}
	service := New(Dependencies{
		Observer:        noopObserver{},
		Authorizer:      allowAuthorizer{},
		Tenants:         tenantExistsRepository{},
		Inventories:     inventoryRepository{item: activeCheckoutTestInventory(tenantID, inventoryID)},
		Assets:          repository,
		Checkouts:       repository,
		AssetUnitOfWork: unitOfWork,
		Undoables:       fakeCheckoutUndoables{},
		Audit:           auditRepository{},
		IDs:             &sequenceIDGenerator{ids: []string{"operation-return", "audit-return"}},
		Clock:           staticClock{now: now},
	})

	returned, err := service.ReturnAsset(ctx, ReturnAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor-two")},
		Source:      audit.SourceAPI,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		AssetID:     assetID,
		Details:     "back on shelf",
	})
	if err != nil {
		t.Fatalf("ReturnAsset returned error: %v", err)
	}
	if returned.State != asset.CheckoutStateReturned || returned.ReturnedByPrincipal != "editor-two" || returned.ReturnDetails.String() != "back on shelf" {
		t.Fatalf("unexpected returned checkout: %+v", returned)
	}
	if len(unitOfWork.auditRecords) != 1 || unitOfWork.auditRecords[0].Action != audit.ActionAssetReturned || unitOfWork.auditRecords[0].Metadata["checked_out_by_principal_id"] != "editor-one" {
		t.Fatalf("expected return audit with checkout actor metadata, got %+v", unitOfWork.auditRecords)
	}
	if len(unitOfWork.undoables) != 1 || unitOfWork.undoables[0].OriginalAction != audit.ActionAssetReturned || unitOfWork.undoables[0].BeforeCheckout == nil || unitOfWork.undoables[0].AfterCheckout == nil {
		t.Fatalf("expected return undoable operation, got %+v", unitOfWork.undoables)
	}
}

func TestListAssetCheckoutHistoryRejectsUnknownAsset(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	repository := newFakeCheckoutAssetRepository(activeCheckoutTestAsset(tenantID, inventoryID, asset.ID("asset-one")))
	service := New(Dependencies{
		Observer:    noopObserver{},
		Authorizer:  allowAuthorizer{},
		Tenants:     tenantExistsRepository{},
		Inventories: inventoryRepository{item: activeCheckoutTestInventory(tenantID, inventoryID)},
		Assets:      repository,
		Checkouts:   repository,
	})

	_, err := service.ListAssetCheckoutHistory(ctx, ListAssetCheckoutHistoryInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer-one")},
		TenantID:    tenantID,
		InventoryID: inventoryID,
		AssetID:     asset.ID("missing-asset"),
	})
	if err == nil {
		t.Fatal("expected unknown asset checkout history to fail")
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestCheckoutDetailsRejectsOverLimitText(t *testing.T) {
	value := make([]rune, asset.MaxCheckoutDetailsLength+1)
	for index := range value {
		value[index] = 'x'
	}
	_, ok := asset.NewCheckoutDetails(string(value))
	if ok {
		t.Fatal("expected over-limit checkout details to be rejected")
	}
}

func activeCheckoutTestAsset(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) asset.Asset {
	return asset.Asset{
		ID:             assetID,
		TenantID:       asset.TenantID(tenantID.String()),
		InventoryID:    asset.InventoryID(inventoryID.String()),
		Kind:           asset.KindItem,
		Title:          asset.Title("Socket set"),
		LifecycleState: asset.LifecycleStateActive,
	}
}

func activeCheckoutTestInventory(tenantID tenant.ID, inventoryID inventory.InventoryID) inventory.Inventory {
	return inventory.Inventory{
		ID:             inventoryID,
		TenantID:       inventory.TenantID(tenantID.String()),
		Name:           inventory.Name("Home"),
		LifecycleState: inventory.LifecycleStateActive,
	}
}

type fakeCheckoutAssetRepository struct {
	item    asset.Asset
	current *asset.Checkout
	history []asset.Checkout
}

func newFakeCheckoutAssetRepository(item asset.Asset) *fakeCheckoutAssetRepository {
	return &fakeCheckoutAssetRepository{item: item}
}

func (r *fakeCheckoutAssetRepository) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	if r.item.ID == assetID && r.item.TenantID.String() == tenantID.String() && r.item.InventoryID.String() == inventoryID.String() {
		return r.item, true, nil
	}
	return asset.Asset{}, false, nil
}

func (*fakeCheckoutAssetRepository) AssetHasActiveChildren(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (bool, error) {
	return false, nil
}

func (r *fakeCheckoutAssetRepository) ListAssetsByInventory(context.Context, tenant.ID, inventory.InventoryID, ports.AssetListPageRequest) ([]asset.Asset, error) {
	return []asset.Asset{r.item}, nil
}

func (r *fakeCheckoutAssetRepository) CurrentAssetCheckout(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (asset.Checkout, bool, error) {
	if r.current == nil {
		return asset.Checkout{}, false, nil
	}
	return *r.current, true, nil
}

func (r *fakeCheckoutAssetRepository) AssetCheckoutByID(_ context.Context, _ tenant.ID, _ inventory.InventoryID, checkoutID asset.CheckoutID) (asset.Checkout, bool, error) {
	if r.current != nil && r.current.ID == checkoutID {
		return *r.current, true, nil
	}
	for _, checkout := range r.history {
		if checkout.ID == checkoutID {
			return checkout, true, nil
		}
	}
	return asset.Checkout{}, false, nil
}

func (r *fakeCheckoutAssetRepository) ListAssetCheckoutHistory(context.Context, tenant.ID, inventory.InventoryID, asset.ID, ports.AssetCheckoutHistoryPageRequest) ([]asset.Checkout, error) {
	return r.history, nil
}

func (r *fakeCheckoutAssetRepository) ListCheckedOutAssets(context.Context, tenant.ID, inventory.InventoryID, ports.CheckedOutAssetsPageRequest) ([]ports.CheckedOutAsset, error) {
	if r.current == nil {
		return nil, nil
	}
	return []ports.CheckedOutAsset{{Asset: r.item, Checkout: *r.current}}, nil
}

func (*fakeCheckoutAssetRepository) HasLaterCheckout(context.Context, asset.Checkout) (bool, error) {
	return false, nil
}

type fakeCheckoutUnitOfWork struct {
	repository   *fakeCheckoutAssetRepository
	auditRecords []audit.Record
	undoables    []ports.UndoableOperation
}

func (*fakeCheckoutUnitOfWork) CreateAsset(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (*fakeCheckoutUnitOfWork) CreateAssetWithParentPromotion(context.Context, asset.Asset, audit.Record, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (*fakeCheckoutUnitOfWork) UpdateAsset(context.Context, asset.Asset, []audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (*fakeCheckoutUnitOfWork) UpdateAssetLifecycle(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error {
	return nil
}

func (u *fakeCheckoutUnitOfWork) CheckOutAsset(_ context.Context, checkout asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	u.repository.current = &checkout
	u.auditRecords = append(u.auditRecords, auditRecord)
	if undoableOperation != nil {
		u.undoables = append(u.undoables, *undoableOperation)
	}
	return nil
}

func (u *fakeCheckoutUnitOfWork) ReturnAsset(_ context.Context, _ asset.Checkout, returned asset.Checkout, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	u.repository.current = &returned
	u.auditRecords = append(u.auditRecords, auditRecord)
	if undoableOperation != nil {
		u.undoables = append(u.undoables, *undoableOperation)
	}
	return nil
}

func (*fakeCheckoutUnitOfWork) DeleteAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID, audit.Record) error {
	return nil
}

type fakeCheckoutUndoables struct{}

func (fakeCheckoutUndoables) UndoableOperationByID(context.Context, tenant.ID, inventory.InventoryID, string) (ports.UndoableOperation, bool, error) {
	return ports.UndoableOperation{}, false, nil
}

func (fakeCheckoutUndoables) ApplyAssetUndoableOperation(context.Context, string, ports.UndoableOperationDirection, asset.Asset, asset.Asset, audit.Record) (ports.UndoableOperation, asset.Asset, error) {
	return ports.UndoableOperation{}, asset.Asset{}, apperrors.ErrInvalidInput
}

func (fakeCheckoutUndoables) ApplyAssetCheckoutUndoableOperation(context.Context, string, ports.UndoableOperationDirection, asset.Checkout, asset.Checkout, audit.Record) (ports.UndoableOperation, asset.Checkout, error) {
	return ports.UndoableOperation{}, asset.Checkout{}, apperrors.ErrInvalidInput
}

type sequenceIDGenerator struct {
	ids []string
}

func (g *sequenceIDGenerator) NewID() string {
	if len(g.ids) == 0 {
		return ""
	}
	next := g.ids[0]
	g.ids = g.ids[1:]
	return next
}

type staticClock struct {
	now time.Time
}

func (c staticClock) Now() time.Time {
	return c.now
}
