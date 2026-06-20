package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestStorePersistsCustomFieldDefinitionsByScope(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryOneID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryTwoID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryOneID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, inventoryTwoID.String(), tenantID, "Supplies")

	tenantDefinition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	inventoryDefinition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryOneID, customfield.ScopeInventory, "condition", customfield.FieldTypeEnum, []string{"new", "used"})
	if err := saveCustomFieldDefinition(t, ctx, store, tenantDefinition); err != nil {
		t.Fatalf("save tenant definition: %v", err)
	}
	if err := saveCustomFieldDefinition(t, ctx, store, inventoryDefinition); err != nil {
		t.Fatalf("save inventory definition: %v", err)
	}

	tenantDefinitions, err := store.ListTenantCustomFieldDefinitions(ctx, tenantID, ports.CustomFieldDefinitionPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list tenant definitions: %v", err)
	}
	if len(tenantDefinitions) != 1 || tenantDefinitions[0].Key != tenantDefinition.Key {
		t.Fatalf("expected tenant definition only, got %+v", tenantDefinitions)
	}

	effectiveOne, err := store.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryOneID, ports.CustomFieldDefinitionPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list first inventory definitions: %v", err)
	}
	if len(effectiveOne) != 2 || effectiveOne[0].Scope != customfield.ScopeTenant || effectiveOne[1].Scope != customfield.ScopeInventory {
		t.Fatalf("expected tenant then inventory definitions, got %+v", effectiveOne)
	}

	effectiveTwo, err := store.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryTwoID, ports.CustomFieldDefinitionPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list second inventory definitions: %v", err)
	}
	if len(effectiveTwo) != 1 || effectiveTwo[0].Key != tenantDefinition.Key {
		t.Fatalf("expected only inherited tenant definition, got %+v", effectiveTwo)
	}
}

func TestStorePaginatesCustomFieldDefinitions(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	first := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	second := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeEnum, []string{"new", "used"})
	if err := saveCustomFieldDefinition(t, ctx, store, first); err != nil {
		t.Fatalf("save first definition: %v", err)
	}
	if err := saveCustomFieldDefinition(t, ctx, store, second); err != nil {
		t.Fatalf("save second definition: %v", err)
	}

	page, err := store.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{Limit: 1})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(page) != 1 || page[0].ID != first.ID {
		t.Fatalf("expected tenant definition first, got %+v", page)
	}
	nextPage, err := store.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{
		AfterDefinitionKey: page[0].CursorKey(),
		Limit:              1,
	})
	if err != nil {
		t.Fatalf("list next page: %v", err)
	}
	if len(nextPage) != 1 || nextPage[0].ID != second.ID {
		t.Fatalf("expected inventory definition second, got %+v", nextPage)
	}
}

func TestStoreRejectsDuplicateCustomFieldDefinitionKeys(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	first := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	duplicate := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	if err := saveCustomFieldDefinition(t, ctx, store, first); err != nil {
		t.Fatalf("save first definition: %v", err)
	}
	if err := saveCustomFieldDefinition(t, ctx, store, duplicate); err == nil {
		t.Fatalf("expected duplicate tenant key rejection")
	}

	inventoryFirst := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeText, nil)
	inventoryDuplicate := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeText, nil)
	if err := saveCustomFieldDefinition(t, ctx, store, inventoryFirst); err != nil {
		t.Fatalf("save inventory definition: %v", err)
	}
	if err := saveCustomFieldDefinition(t, ctx, store, inventoryDuplicate); err == nil {
		t.Fatalf("expected duplicate inventory key rejection")
	}

	inventoryOnly := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID, inventoryID, customfield.ScopeInventory, "warranty", customfield.FieldTypeText, nil)
	tenantConflict := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantID, "", customfield.ScopeTenant, "warranty", customfield.FieldTypeText, nil)
	if err := saveCustomFieldDefinition(t, ctx, store, inventoryOnly); err != nil {
		t.Fatalf("save inventory-only definition: %v", err)
	}
	if err := saveCustomFieldDefinition(t, ctx, store, tenantConflict); err == nil {
		t.Fatalf("expected tenant key to conflict with existing inventory key")
	}
}

func TestStoreUpdatesCustomFieldDefinitionMetadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Medicine")

	definition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeEnum, []string{"new", "used"})
	if err := saveCustomFieldDefinition(t, ctx, store, definition); err != nil {
		t.Fatalf("save custom field definition: %v", err)
	}
	displayName, ok := customfield.NewDisplayName("Item Condition")
	if !ok {
		t.Fatalf("expected valid display name")
	}
	definition.DisplayName = displayName
	if err := store.UpdateCustomFieldDefinition(ctx, definition, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); err != nil {
		t.Fatalf("update custom field definition: %v", err)
	}

	found, ok, err := store.CustomFieldDefinitionByID(ctx, tenantID, inventoryID, definition.ID)
	if err != nil {
		t.Fatalf("find custom field definition: %v", err)
	}
	if !ok || found.DisplayName != displayName || found.Key != definition.Key || found.Type != definition.Type || len(found.EnumOptions) != 2 || found.EnumOptions[0].String() != "new" || found.EnumOptions[1].String() != "used" {
		t.Fatalf("expected updated custom field definition metadata, got %+v", found)
	}

	mutatedKey := definition
	mutatedKey.Key = customfield.Key("changed")
	if err := store.UpdateCustomFieldDefinition(ctx, mutatedKey, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected immutable key rejection, got %v", err)
	}

	mutatedOptions := definition
	mutatedOptions.EnumOptions = []customfield.Key{customfield.Key("new")}
	if err := store.UpdateCustomFieldDefinition(ctx, mutatedOptions, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected immutable enum options rejection, got %v", err)
	}
}

func TestStoreUpdatesCustomFieldDefinitionSchemaByExpansion(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Medicine")

	medicineType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "medicine")
	supplyType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "supply")
	containerType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAR", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "container")
	if err := saveCustomAssetType(t, ctx, store, medicineType); err != nil {
		t.Fatalf("save medicine type: %v", err)
	}
	if err := saveCustomAssetType(t, ctx, store, supplyType); err != nil {
		t.Fatalf("save supply type: %v", err)
	}
	if err := saveCustomAssetType(t, ctx, store, containerType); err != nil {
		t.Fatalf("save container type: %v", err)
	}

	definition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeEnum, []string{"new", "used"})
	definition.Applicability = customfield.ApplicabilityCustomAssetTypes
	definition.CustomAssetTypeIDs = []customfield.AssetTypeID{medicineType.ID}
	if err := saveCustomFieldDefinition(t, ctx, store, definition); err != nil {
		t.Fatalf("save targeted definition: %v", err)
	}

	expanded := definition
	expanded.EnumOptions = append(expanded.EnumOptions, customfield.Key("expired"))
	expanded.CustomAssetTypeIDs = append(expanded.CustomAssetTypeIDs, supplyType.ID)
	if err := store.UpdateCustomFieldDefinition(ctx, expanded, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); err != nil {
		t.Fatalf("expand definition schema: %v", err)
	}

	found, ok, err := store.CustomFieldDefinitionByID(ctx, tenantID, inventoryID, definition.ID)
	if err != nil {
		t.Fatalf("find expanded definition: %v", err)
	}
	if !ok || len(found.EnumOptions) != 3 || len(found.CustomAssetTypeIDs) != 2 {
		t.Fatalf("expected expanded definition, found=%v definition=%+v", ok, found)
	}

	reorderedExistingTargets := found
	reorderedExistingTargets.CustomAssetTypeIDs = []customfield.AssetTypeID{supplyType.ID, medicineType.ID, containerType.ID}
	if err := store.UpdateCustomFieldDefinition(ctx, reorderedExistingTargets, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "E"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); err != nil {
		t.Fatalf("expand definition schema with reordered existing targets: %v", err)
	}
	found, ok, err = store.CustomFieldDefinitionByID(ctx, tenantID, inventoryID, definition.ID)
	if err != nil {
		t.Fatalf("find reordered expanded definition: %v", err)
	}
	if !ok || len(found.CustomAssetTypeIDs) != 3 {
		t.Fatalf("expected reordered target expansion, found=%v definition=%+v", ok, found)
	}

	allAssets := found
	allAssets.Applicability = customfield.ApplicabilityAllAssets
	allAssets.CustomAssetTypeIDs = nil
	if err := store.UpdateCustomFieldDefinition(ctx, allAssets, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); err != nil {
		t.Fatalf("expand definition to all assets: %v", err)
	}
	found, ok, err = store.CustomFieldDefinitionByID(ctx, tenantID, inventoryID, definition.ID)
	if err != nil {
		t.Fatalf("find all-assets definition: %v", err)
	}
	if !ok || found.Applicability != customfield.ApplicabilityAllAssets || len(found.CustomAssetTypeIDs) != 0 {
		t.Fatalf("expected all-assets definition without targets, found=%v definition=%+v", ok, found)
	}

	narrowed := allAssets
	narrowed.Applicability = customfield.ApplicabilityCustomAssetTypes
	narrowed.CustomAssetTypeIDs = []customfield.AssetTypeID{medicineType.ID}
	if err := store.UpdateCustomFieldDefinition(ctx, narrowed, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenantID, inventoryID, audit.ActionCustomFieldDefinitionUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected narrowing rejection, got %v", err)
	}
}
