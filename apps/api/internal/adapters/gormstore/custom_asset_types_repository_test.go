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
)

func TestStoreUpdatesCustomAssetTypeMetadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Medicine")

	assetType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "medicine")
	if err := saveCustomAssetType(t, ctx, store, assetType); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}
	displayName, ok := customfield.NewDisplayName("Medicine and Vitamins")
	if !ok {
		t.Fatalf("expected valid display name")
	}
	description, ok := customfield.NewDescription("Medication and supplement supplies")
	if !ok {
		t.Fatalf("expected valid description")
	}
	assetType.DisplayName = displayName
	assetType.Description = description
	if err := store.UpdateCustomAssetType(ctx, assetType, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "T"), tenantID, inventoryID, audit.ActionCustomAssetTypeUpdated)); err != nil {
		t.Fatalf("update custom asset type: %v", err)
	}

	found, ok, err := store.CustomAssetTypeByID(ctx, tenantID, inventoryID, assetType.ID)
	if err != nil {
		t.Fatalf("find custom asset type: %v", err)
	}
	if !ok || found.DisplayName != displayName || found.Description != description || found.Key != assetType.Key {
		t.Fatalf("expected updated custom asset type metadata, got %+v", found)
	}

	mutatedKey := assetType
	mutatedKey.Key = customfield.Key("changed")
	if err := store.UpdateCustomAssetType(ctx, mutatedKey, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "T"), tenantID, inventoryID, audit.ActionCustomAssetTypeUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected immutable key rejection, got %v", err)
	}
}

func TestStoreArchivesCustomAssetTypeWithoutDeletingReferences(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Medicine")

	assetType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "medicine")
	if err := saveCustomAssetType(t, ctx, store, assetType); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}

	definition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, inventoryID, customfield.ScopeInventory, "expires-on", customfield.FieldTypeDate, nil)
	definition.Applicability = customfield.ApplicabilityCustomAssetTypes
	definition.CustomAssetTypeIDs = []customfield.AssetTypeID{assetType.ID}
	if err := saveCustomFieldDefinition(t, ctx, store, definition); err != nil {
		t.Fatalf("save targeted custom field definition: %v", err)
	}

	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	item.CustomAssetTypeID = asset.CustomAssetTypeID(assetType.ID.String())
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatalf("create typed asset: %v", err)
	}

	archived, ok := assetType.Archive()
	if !ok {
		t.Fatalf("expected active custom asset type to archive")
	}
	if err := store.ArchiveCustomAssetType(ctx, assetType, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "A"), tenantID, inventoryID, audit.ActionCustomAssetTypeArchived)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected archive repository to reject active lifecycle state, got %v", err)
	}
	if err := store.ArchiveCustomAssetType(ctx, archived, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "A"), tenantID, inventoryID, audit.ActionCustomAssetTypeArchived)); err != nil {
		t.Fatalf("archive custom asset type: %v", err)
	}

	var archivedRow customAssetTypeModel
	if err := store.db.WithContext(ctx).Where(&customAssetTypeModel{ID: assetType.ID.String()}).First(&archivedRow).Error; err != nil {
		t.Fatalf("find archived custom asset type row: %v", err)
	}
	if archivedRow.LifecycleState != customfield.AssetTypeLifecycleArchived.String() {
		t.Fatalf("expected archived lifecycle state, got %q", archivedRow.LifecycleState)
	}

	listed, err := store.ListInventoryCustomAssetTypes(ctx, tenantID, inventoryID, ports.CustomAssetTypePageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list custom asset types: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected archived custom asset type to be hidden from list, got %+v", listed)
	}
	archivedListed, err := store.ListInventoryCustomAssetTypes(ctx, tenantID, inventoryID, ports.CustomAssetTypePageRequest{Limit: 10, Lifecycle: ports.CustomizationLifecycleArchived})
	if err != nil {
		t.Fatalf("list archived custom asset types: %v", err)
	}
	if len(archivedListed) != 1 || archivedListed[0].ID != assetType.ID {
		t.Fatalf("expected lifecycle-filtered archived custom asset type, got %+v", archivedListed)
	}

	lookedUp, err := store.CustomAssetTypesByID(ctx, tenantID, inventoryID, []customfield.AssetTypeID{assetType.ID})
	if err != nil {
		t.Fatalf("lookup custom asset type by id: %v", err)
	}
	if len(lookedUp) != 0 {
		t.Fatalf("expected archived custom asset type to be unavailable for new use, got %+v", lookedUp)
	}

	preservedAsset, found, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("find asset with archived type reference: %v", err)
	}
	if !found || preservedAsset.CustomAssetTypeID != item.CustomAssetTypeID {
		t.Fatalf("expected existing asset reference to be preserved, got found=%v asset=%+v", found, preservedAsset)
	}

	definitions, err := store.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list custom field definitions: %v", err)
	}
	if len(definitions) != 1 || len(definitions[0].CustomAssetTypeIDs) != 1 || definitions[0].CustomAssetTypeIDs[0] != assetType.ID {
		t.Fatalf("expected existing target reference to be preserved, got %+v", definitions)
	}

	newAsset := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	newAsset.CustomAssetTypeID = asset.CustomAssetTypeID(assetType.ID.String())
	if err := createAsset(t, ctx, store, newAsset); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected archived custom asset type assignment rejection, got %v", err)
	}

	newDefinition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID, inventoryID, customfield.ScopeInventory, "dose", customfield.FieldTypeText, nil)
	newDefinition.Applicability = customfield.ApplicabilityCustomAssetTypes
	newDefinition.CustomAssetTypeIDs = []customfield.AssetTypeID{assetType.ID}
	if err := saveCustomFieldDefinition(t, ctx, store, newDefinition); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected archived custom asset type target rejection, got %v", err)
	}
}

func TestStoreListsCustomAssetTypesByLifecycleAndScope(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Medicine")
	tenantType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), "", customfield.ScopeTenant, "tenant-type")
	inventoryType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "inventory-type")
	for _, assetType := range []customfield.AssetType{tenantType, inventoryType} {
		if err := saveCustomAssetType(t, ctx, store, assetType); err != nil {
			t.Fatalf("save custom asset type: %v", err)
		}
		archived, ok := assetType.Archive()
		if !ok {
			t.Fatalf("archive custom asset type %s", assetType.ID)
		}
		if err := store.ArchiveCustomAssetType(ctx, archived, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "A"), tenantID, inventoryID, audit.ActionCustomAssetTypeArchived)); err != nil {
			t.Fatalf("persist archive: %v", err)
		}
	}
	for _, filter := range []ports.CustomizationLifecycleFilter{"", ports.CustomizationLifecycleActive} {
		items, err := store.ListInventoryCustomAssetTypes(ctx, tenantID, inventoryID, ports.CustomAssetTypePageRequest{Limit: 10, Lifecycle: filter})
		if err != nil || len(items) != 0 {
			t.Fatalf("expected %q to hide archived types, items=%+v err=%v", filter, items, err)
		}
	}
	for _, filter := range []ports.CustomizationLifecycleFilter{ports.CustomizationLifecycleArchived, ports.CustomizationLifecycleAll} {
		items, err := store.ListInventoryCustomAssetTypes(ctx, tenantID, inventoryID, ports.CustomAssetTypePageRequest{Limit: 10, Lifecycle: filter})
		if err != nil || len(items) != 2 || items[0].ID != tenantType.ID || items[1].ID != inventoryType.ID {
			t.Fatalf("expected %q types in scope order, items=%+v err=%v", filter, items, err)
		}
	}
	tenantItems, err := store.ListTenantCustomAssetTypes(ctx, tenantID, ports.CustomAssetTypePageRequest{Limit: 10, Lifecycle: ports.CustomizationLifecycleArchived})
	if err != nil || len(tenantItems) != 1 || tenantItems[0].ID != tenantType.ID {
		t.Fatalf("expected tenant archived type only, items=%+v err=%v", tenantItems, err)
	}
}
