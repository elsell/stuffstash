package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestStorePersistsAssetsAndLocationParents(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	location := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	if err := createAsset(t, ctx, store, location); err != nil {
		t.Fatalf("save location asset: %v", err)
	}
	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, location.ID.String())
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("save child asset: %v", err)
	}

	items, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 assets, got %+v", items)
	}
	if items[0].Kind != asset.KindLocation || items[1].ParentAssetID != location.ID {
		t.Fatalf("unexpected assets: %+v", items)
	}
}

func TestStoreRejectsInvalidAssetParents(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryOneID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryTwoID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryOneID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, inventoryTwoID.String(), tenantID, "Supplies")

	itemParent := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryOneID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, itemParent); err != nil {
		t.Fatalf("save item parent: %v", err)
	}
	child := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID.String(), inventoryOneID.String(), asset.KindItem, itemParent.ID.String())
	if err := createAsset(t, ctx, store, child); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected item parent rejection, got %v", err)
	}

	location := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID.String(), inventoryOneID.String(), asset.KindLocation, "")
	if err := createAsset(t, ctx, store, location); err != nil {
		t.Fatalf("save location parent: %v", err)
	}
	crossInventoryChild := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID.String(), inventoryTwoID.String(), asset.KindItem, location.ID.String())
	if err := createAsset(t, ctx, store, crossInventoryChild); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected cross-inventory parent rejection, got %v", err)
	}
}

func TestStoreRejectsRootAssetsOutsideInventoryTenant(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantOneID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantTwoID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenant(t, ctx, store, tenantOneID, "Home")
	saveTenant(t, ctx, store, tenantTwoID, "Cabin")
	saveInventory(t, ctx, store, inventoryID.String(), tenantTwoID, "Supplies")

	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantOneID.String(), inventoryID.String(), asset.KindLocation, "")
	if err := createAsset(t, ctx, store, item); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected tenant/inventory mismatch rejection, got %v", err)
	}
}

func TestStoreRejectsAssetCustomAssetTypesOutsideVisibleScope(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantOneID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantTwoID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryOneID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	inventoryTwoID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAY")
	inventoryThreeID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAZ")
	saveTenant(t, ctx, store, tenantOneID, "Home")
	saveTenant(t, ctx, store, tenantTwoID, "Cabin")
	saveInventory(t, ctx, store, inventoryOneID.String(), tenantOneID, "Tools")
	saveInventory(t, ctx, store, inventoryTwoID.String(), tenantOneID, "Medicine")
	saveInventory(t, ctx, store, inventoryThreeID.String(), tenantTwoID, "Other")

	tenantType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantOneID.String(), "", customfield.ScopeTenant, "medicine")
	if err := saveCustomAssetType(t, ctx, store, tenantType); err != nil {
		t.Fatalf("save tenant custom asset type: %v", err)
	}
	visibleAsset := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantOneID.String(), inventoryOneID.String(), asset.KindItem, "")
	visibleAsset.CustomAssetTypeID = asset.CustomAssetTypeID(tenantType.ID.String())
	if err := createAsset(t, ctx, store, visibleAsset); err != nil {
		t.Fatalf("expected tenant-scoped type to be visible: %v", err)
	}

	siblingType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantOneID.String(), inventoryTwoID.String(), customfield.ScopeInventory, "medicine-inventory")
	if err := saveCustomAssetType(t, ctx, store, siblingType); err != nil {
		t.Fatalf("save sibling inventory custom asset type: %v", err)
	}
	siblingAsset := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB3", tenantOneID.String(), inventoryOneID.String(), asset.KindItem, "")
	siblingAsset.CustomAssetTypeID = asset.CustomAssetTypeID(siblingType.ID.String())
	if err := createAsset(t, ctx, store, siblingAsset); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected sibling inventory custom asset type rejection, got %v", err)
	}

	otherTenantType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FB4", tenantTwoID.String(), inventoryThreeID.String(), customfield.ScopeInventory, "other-medicine")
	if err := saveCustomAssetType(t, ctx, store, otherTenantType); err != nil {
		t.Fatalf("save other tenant custom asset type: %v", err)
	}
	otherTenantAsset := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB5", tenantOneID.String(), inventoryOneID.String(), asset.KindItem, "")
	otherTenantAsset.CustomAssetTypeID = asset.CustomAssetTypeID(otherTenantType.ID.String())
	if err := createAsset(t, ctx, store, otherTenantAsset); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected cross-tenant custom asset type rejection, got %v", err)
	}
}

func TestStoreUpdatesAssetLifecycleAndFiltersListings(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	parent := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	if err := createAsset(t, ctx, store, parent); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	child := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, parent.ID.String())
	if err := createAsset(t, ctx, store, child); err != nil {
		t.Fatalf("create child: %v", err)
	}

	parentArchived := parent
	parentArchived.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(ctx, parentArchived, auditRecord(t, auditIDWithSuffix(parent.ID.String(), "A"), tenantID, inventoryID, audit.ActionAssetArchived), nil); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected active child archive rejection, got %v", err)
	}

	child.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(ctx, child, auditRecord(t, auditIDWithSuffix(child.ID.String(), "A"), tenantID, inventoryID, audit.ActionAssetArchived), nil); err != nil {
		t.Fatalf("archive child: %v", err)
	}
	auditRecords, err := store.ListInventoryAuditRecords(ctx, tenantID, inventoryID, ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list lifecycle audit records: %v", err)
	}
	if !auditRecordsIncludeAction(auditRecords, audit.ActionAssetArchived) {
		t.Fatalf("expected lifecycle audit records to include archive action, got %+v", auditRecords)
	}
	active, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterActive})
	if err != nil {
		t.Fatalf("list active assets: %v", err)
	}
	if len(active) != 1 || active[0].ID != parent.ID {
		t.Fatalf("expected active parent only, got %+v", active)
	}
	archived, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterArchived})
	if err != nil {
		t.Fatalf("list archived assets: %v", err)
	}
	if len(archived) != 1 || archived[0].ID != child.ID {
		t.Fatalf("expected archived child only, got %+v", archived)
	}

	parentArchived.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(ctx, parentArchived, auditRecord(t, auditIDWithSuffix(parent.ID.String(), "A"), tenantID, inventoryID, audit.ActionAssetArchived), nil); err != nil {
		t.Fatalf("archive parent: %v", err)
	}
	child.LifecycleState = asset.LifecycleStateActive
	if err := store.UpdateAssetLifecycle(ctx, child, auditRecord(t, auditIDWithSuffix(child.ID.String(), "R"), tenantID, inventoryID, audit.ActionAssetRestored), nil); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected archived parent restore rejection, got %v", err)
	}
}

func TestStoreRejectsAssetContainmentCycles(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	parent := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	if err := createAsset(t, ctx, store, parent); err != nil {
		t.Fatalf("save parent: %v", err)
	}
	child := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindContainer, parent.ID.String())
	if err := createAsset(t, ctx, store, child); err != nil {
		t.Fatalf("save child: %v", err)
	}

	parent.ParentAssetID = child.ID
	if err := createAsset(t, ctx, store, parent); err == nil {
		t.Fatalf("expected duplicate asset rejection")
	}
}

func TestStorePaginatesAssetsAndRejectsDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	first := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	second := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	if err := createAsset(t, ctx, store, first); err != nil {
		t.Fatalf("create first asset: %v", err)
	}
	if err := createAsset(t, ctx, store, second); err != nil {
		t.Fatalf("create second asset: %v", err)
	}

	page, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{Limit: 1})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(page) != 1 || page[0].ID != first.ID {
		t.Fatalf("expected first page with first asset, got %+v", page)
	}
	nextPage, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{AfterAssetID: first.ID, Limit: 1})
	if err != nil {
		t.Fatalf("list next page: %v", err)
	}
	if len(nextPage) != 1 || nextPage[0].ID != second.ID {
		t.Fatalf("expected next page with second asset, got %+v", nextPage)
	}
	if err := createAsset(t, ctx, store, first); err == nil {
		t.Fatalf("expected duplicate asset create to fail")
	}
}

func TestStoreListsAssetsByUpdatedDescending(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	firstUpdatedAt := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	secondUpdatedAt := time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC)
	tieUpdatedAt := time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC)
	oldest := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	oldest.CreatedAt = firstUpdatedAt
	oldest.UpdatedAt = firstUpdatedAt
	middle := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	middle.CreatedAt = secondUpdatedAt
	middle.UpdatedAt = secondUpdatedAt
	tieLowID := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	tieLowID.CreatedAt = tieUpdatedAt
	tieLowID.UpdatedAt = tieUpdatedAt
	tieHighID := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	tieHighID.CreatedAt = tieUpdatedAt
	tieHighID.UpdatedAt = tieUpdatedAt
	for _, item := range []asset.Asset{oldest, middle, tieLowID, tieHighID} {
		if err := createAsset(t, ctx, store, item); err != nil {
			t.Fatalf("create asset %s: %v", item.ID, err)
		}
	}

	firstPage, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{
		Limit:           2,
		LifecycleFilter: ports.AssetLifecycleFilterAll,
		Sort:            ports.AssetListSortUpdatedDesc,
	})
	if err != nil {
		t.Fatalf("list first updated-desc page: %v", err)
	}
	if len(firstPage) != 2 || firstPage[0].ID != tieHighID.ID || firstPage[1].ID != tieLowID.ID {
		t.Fatalf("expected timestamp tie to sort by descending id, got %+v", firstPage)
	}

	nextPage, err := store.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{
		AfterAssetID:    firstPage[1].ID,
		AfterUpdatedAt:  firstPage[1].UpdatedAt,
		Limit:           2,
		LifecycleFilter: ports.AssetLifecycleFilterAll,
		Sort:            ports.AssetListSortUpdatedDesc,
	})
	if err != nil {
		t.Fatalf("list second updated-desc page: %v", err)
	}
	if len(nextPage) != 2 || nextPage[0].ID != middle.ID || nextPage[1].ID != oldest.ID {
		t.Fatalf("expected cursor to continue updated-desc order, got %+v", nextPage)
	}
}

func TestStoreUpdatesAssetsAndMovesContainersWithChildren(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	garage := assetItem("garage", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	shelf := assetItem("shelf", tenantID.String(), inventoryID.String(), asset.KindLocation, "garage")
	box := assetItem("box", tenantID.String(), inventoryID.String(), asset.KindContainer, "shelf")
	wrench := assetItem("wrench", tenantID.String(), inventoryID.String(), asset.KindItem, "box")
	for _, item := range []asset.Asset{garage, shelf, box, wrench} {
		if err := createAsset(t, ctx, store, item); err != nil {
			t.Fatalf("create asset %s: %v", item.ID, err)
		}
	}

	box.ParentAssetID = garage.ID
	title, ok := asset.NewTitle("Moved Box")
	if !ok {
		t.Fatalf("expected valid title")
	}
	box.Title = title
	customFields, ok := asset.NewCustomFields(map[string]any{"serial": "abc"})
	if !ok {
		t.Fatalf("expected valid custom fields")
	}
	box.CustomFields = customFields
	if err := updateAsset(t, ctx, store, box); err != nil {
		t.Fatalf("update box: %v", err)
	}

	foundBox, ok, err := store.AssetByID(ctx, tenantID, inventoryID, box.ID)
	if err != nil {
		t.Fatalf("find box: %v", err)
	}
	if !ok || foundBox.ParentAssetID != garage.ID || foundBox.Title.String() != "Moved Box" || foundBox.CustomFields.Values()["serial"] != "abc" {
		t.Fatalf("expected moved box with updated fields, found=%t %+v", ok, foundBox)
	}
	foundWrench, ok, err := store.AssetByID(ctx, tenantID, inventoryID, wrench.ID)
	if err != nil {
		t.Fatalf("find wrench: %v", err)
	}
	if !ok || foundWrench.ParentAssetID != box.ID {
		t.Fatalf("expected child to remain inside moved box, found=%t %+v", ok, foundWrench)
	}

	box.ParentAssetID = asset.ID("")
	if err := updateAsset(t, ctx, store, box); err != nil {
		t.Fatalf("move box to root: %v", err)
	}
	rootBox, ok, err := store.AssetByID(ctx, tenantID, inventoryID, box.ID)
	if err != nil {
		t.Fatalf("find root box: %v", err)
	}
	if !ok || rootBox.ParentAssetID.String() != "" {
		t.Fatalf("expected box at root, found=%t %+v", ok, rootBox)
	}
}

func TestStoreRejectsInvalidAssetUpdates(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("tenant-one")
	inventoryID := inventory.InventoryID("inventory-one")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	garage := assetItem("garage", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	shelf := assetItem("shelf", tenantID.String(), inventoryID.String(), asset.KindLocation, "garage")
	box := assetItem("box", tenantID.String(), inventoryID.String(), asset.KindContainer, "shelf")
	itemParent := assetItem("wrench", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	for _, item := range []asset.Asset{garage, shelf, box, itemParent} {
		if err := createAsset(t, ctx, store, item); err != nil {
			t.Fatalf("create asset %s: %v", item.ID, err)
		}
	}

	garage.ParentAssetID = box.ID
	if err := updateAsset(t, ctx, store, garage); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected cycle rejection, got %v", err)
	}
	box.ParentAssetID = box.ID
	if err := updateAsset(t, ctx, store, box); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected self-parent rejection, got %v", err)
	}
	box.ParentAssetID = itemParent.ID
	if err := updateAsset(t, ctx, store, box); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected item-parent rejection, got %v", err)
	}
	box.Kind = asset.KindItem
	if err := updateAsset(t, ctx, store, box); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected kind-change rejection, got %v", err)
	}
}

func TestStoreRoundTripsAssetCustomFields(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	customFields, ok := asset.NewCustomFields(map[string]any{"serial": "abc"})
	if !ok {
		t.Fatalf("expected valid custom fields")
	}
	item.CustomFields = customFields

	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("create asset: %v", err)
	}
	found, ok, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}
	if !ok || found.CustomFields.Values()["serial"] != "abc" {
		t.Fatalf("expected custom fields to round-trip, got found=%t %+v", ok, found.CustomFields.Values())
	}
}
