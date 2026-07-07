package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestCreateAndListAssets(t *testing.T) {
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Assets:           assets,
		AssetUnitOfWork:  assets,
		Undoables:        assets,
		Audit:            &fakeAuditRepository{},
		Outbox:           &fakeOutbox{},
		IDs:              &fakeIDGenerator{ids: []string{"asset-one", "asset-two"}},
		DefaultPageLimit: 1,
		MaxPageLimit:     2,
	})

	location, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "location",
		Title:       "Garage",
	})
	if err != nil {
		t.Fatalf("create location asset: %v", err)
	}
	if location.Kind != asset.KindLocation || location.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("unexpected location asset: %+v", location)
	}

	item, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:      tenant.ID("tenant-one"),
		InventoryID:   inventory.InventoryID("inventory-one"),
		Kind:          "item",
		Title:         "Drill",
		Description:   "Cordless",
		ParentAssetID: location.ID.String(),
	})
	if err != nil {
		t.Fatalf("create item asset: %v", err)
	}
	if item.ParentAssetID != location.ID {
		t.Fatalf("expected parent %q, got %q", location.ID, item.ParentAssetID)
	}

	result, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(result.Items) != 1 || !result.HasMore || result.NextCursor == nil || result.Limit != 1 {
		t.Fatalf("expected paginated first page, got %+v", result)
	}

	nextPage, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Cursor:      *result.NextCursor,
	})
	if err != nil {
		t.Fatalf("list next assets page: %v", err)
	}
	if len(nextPage.Items) != 1 || nextPage.HasMore || nextPage.Items[0].ID != item.ID {
		t.Fatalf("expected second page with item, got %+v", nextPage)
	}
}

func TestCreateUpdateAndReadAssetTags(t *testing.T) {
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:           &fakeObserver{},
		Authorizer:         &fakeAuthorizer{},
		Tenants:            &fakeTenantRepository{exists: true},
		Inventories:        &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		Assets:             assets,
		AssetTags:          assets,
		AssetUnitOfWork:    assets,
		AssetTagUnitOfWork: assets,
		Undoables:          assets,
		Audit:              &fakeAuditRepository{},
		Outbox:             &fakeOutbox{},
		IDs:                &fakeIDGenerator{ids: []string{"tag-one", "audit-tag-one", "asset-one", "op-asset-one", "audit-asset-one", "audit-asset-tags", "audit-clear-tags"}},
	})

	tag, err := application.CreateAssetTag(context.Background(), CreateAssetTagInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		DisplayName: "Workshop",
		Color:       "#2f80ed",
	})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	item, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Drill",
		TagIDs:      []string{tag.ID.String()},
	})
	if err != nil {
		t.Fatalf("create asset with tag: %v", err)
	}

	detail, err := application.GetAssetDetail(context.Background(), GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     item.ID,
	})
	if err != nil {
		t.Fatalf("get asset detail: %v", err)
	}
	if len(detail.Tags) != 1 || detail.Tags[0].ID != tag.ID {
		t.Fatalf("expected detail tag %q, got %+v", tag.ID, detail.Tags)
	}

	list, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(list.Tags[item.ID]) != 1 || list.Tags[item.ID][0].Key.String() != "workshop" {
		t.Fatalf("expected list tags for asset, got %+v", list.Tags)
	}

	clearTags := []string{}
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     item.ID,
		TagIDs:      &clearTags,
	}); err != nil {
		t.Fatalf("clear asset tags: %v", err)
	}
	cleared, err := application.GetAssetDetail(context.Background(), GetAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     item.ID,
	})
	if err != nil {
		t.Fatalf("get cleared asset detail: %v", err)
	}
	if len(cleared.Tags) != 0 {
		t.Fatalf("expected cleared tags, got %+v", cleared.Tags)
	}
}

func TestAssetWritesValidateTagsBeforePersistence(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:           &fakeObserver{},
		Authorizer:         &fakeAuthorizer{},
		Tenants:            &fakeTenantRepository{exists: true},
		Inventories:        &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		Assets:             assets,
		AssetTags:          assets,
		AssetUnitOfWork:    assets,
		AssetTagUnitOfWork: assets,
		Undoables:          assets,
		Audit:              &fakeAuditRepository{},
		Outbox:             &fakeOutbox{},
		IDs:                &fakeIDGenerator{ids: []string{"asset-one", "op-asset-one", "audit-asset-one", "op-update", "audit-update"}},
	})

	_, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Saw",
		TagIDs:      []string{"missing-tag"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid tag create rejection, got %v", err)
	}
	if _, exists := assets.items[asset.ID("asset-one")]; exists {
		t.Fatalf("asset must not be created when tag validation fails")
	}

	title := "Renamed drill"
	invalidTags := []string{"missing-tag"}
	_, err = application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
		TagIDs:      &invalidTags,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid tag update rejection, got %v", err)
	}
	if assets.items[asset.ID("drill")].Title.String() == title {
		t.Fatalf("asset must not be updated when tag validation fails")
	}
}

func TestCreateAssetPromotesItemParentToContainer(t *testing.T) {
	itemParent := assetItem("asset-parent", "tenant-one", "inventory-one", asset.KindItem, "")
	assets := &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{itemParent.ID: itemParent},
	}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Assets:          assets,
		AssetUnitOfWork: assets,
		Undoables:       assets,
		Audit:           &fakeAuditRepository{},
		Outbox:          &fakeOutbox{},
		IDs:             &fakeIDGenerator{ids: []string{"asset-one"}},
	})

	child, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:      tenant.ID("tenant-one"),
		InventoryID:   inventory.InventoryID("inventory-one"),
		Kind:          "item",
		Title:         "Bit set",
		ParentAssetID: itemParent.ID.String(),
	})
	if err != nil {
		t.Fatalf("create asset under item parent: %v", err)
	}
	if child.ParentAssetID != itemParent.ID {
		t.Fatalf("expected child parent %q, got %q", itemParent.ID, child.ParentAssetID)
	}
	promotedParent := assets.items[itemParent.ID]
	if promotedParent.Kind != asset.KindContainer {
		t.Fatalf("expected parent promotion to container, got %+v", promotedParent)
	}
	if len(assets.auditRecords) != 2 || assets.auditRecords[0].Action != audit.ActionAssetUpdated || assets.auditRecords[1].Action != audit.ActionAssetCreated {
		t.Fatalf("expected parent promotion and child creation audit records, got %+v", assets.auditRecords)
	}
}

func TestCreateAssetRejectsCustomFields(t *testing.T) {
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Assets:          assets,
		AssetUnitOfWork: assets,
		Undoables:       assets,
		Audit:           &fakeAuditRepository{},
		Outbox:          &fakeOutbox{},
		IDs:             &fakeIDGenerator{ids: []string{"asset-one"}},
	})

	_, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Kind:         "item",
		Title:        "Bit set",
		CustomFields: map[string]any{"serial": "abc"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for custom fields, got %v", err)
	}
}

func TestCreateAssetValidatesCustomFieldsAgainstDefinitions(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	serialDefinition := customFieldDefinition("serial-definition", "tenant-one", "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	conditionDefinition := customFieldDefinition("condition-definition", "tenant-one", "inventory-one", customfield.ScopeInventory, "condition", customfield.FieldTypeEnum, []string{"new", "used"})
	if err := customFields.SaveCustomFieldDefinition(context.Background(), serialDefinition, auditRecord("audit-serial", "tenant-one", "", audit.ActionCustomFieldDefinitionCreated)); err != nil {
		t.Fatalf("save serial definition: %v", err)
	}
	if err := customFields.SaveCustomFieldDefinition(context.Background(), conditionDefinition, auditRecord("audit-condition", "tenant-one", "inventory-one", audit.ActionCustomFieldDefinitionCreated)); err != nil {
		t.Fatalf("save condition definition: %v", err)
	}
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          customFields,
		CustomFieldUnitOfWork: customFields,
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{ids: []string{"asset-one"}},
	})

	item, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Drill",
		CustomFields: map[string]any{
			"serial":    "abc",
			"condition": "used",
		},
	})
	if err != nil {
		t.Fatalf("create asset with custom fields: %v", err)
	}
	if item.CustomFields.Values()["serial"] != "abc" || item.CustomFields.Values()["condition"] != "used" {
		t.Fatalf("expected custom fields to be saved, got %+v", item.CustomFields.Values())
	}

	_, err = application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Bad Drill",
		CustomFields: map[string]any{
			"condition": "broken",
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid enum value rejection, got %v", err)
	}
}

func TestUpdateAssetMovesAndValidatesCustomFields(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	serialDefinition := customFieldDefinition("serial-definition", "tenant-one", "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	if err := customFields.SaveCustomFieldDefinition(context.Background(), serialDefinition, auditRecord("audit-serial", "tenant-one", "", audit.ActionCustomFieldDefinitionCreated)); err != nil {
		t.Fatalf("save serial definition: %v", err)
	}
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"): assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):  assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("drill"):  assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, "garage"),
	}}
	observer := &fakeObserver{}
	application := New(Dependencies{
		Observer:              observer,
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          customFields,
		CustomFieldUnitOfWork: customFields,
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	description := "Blue case"
	updated, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
		Description: &description,
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   "shelf",
		},
		CustomFields: map[string]any{"serial": "abc"},
	})
	if err != nil {
		t.Fatalf("update asset: %v", err)
	}
	if updated.Title.String() != title || updated.Description.String() != description || updated.ParentAssetID != asset.ID("shelf") {
		t.Fatalf("unexpected updated asset: %+v", updated)
	}
	if updated.CustomFields.Values()["serial"] != "abc" {
		t.Fatalf("expected updated custom fields, got %+v", updated.CustomFields.Values())
	}
	if assets.items[asset.ID("drill")].ParentAssetID != asset.ID("shelf") {
		t.Fatalf("expected persisted parent shelf, got %+v", assets.items[asset.ID("drill")])
	}
	if !observer.hasEvent(ports.EventAssetUpdated) {
		t.Fatalf("expected asset updated observability event, got %+v", observer.events)
	}

	_, err = application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("drill"),
		CustomFields: map[string]any{"serial": 42},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid custom field update rejection, got %v", err)
	}
}

func TestUndoAndRedoAssetUpdate(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"): assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):  assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("drill"):  assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, "garage"),
	}}
	application := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	updated, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   "shelf",
		},
	})
	if err != nil {
		t.Fatalf("update asset: %v", err)
	}

	operationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]
	if operationID == "" {
		t.Fatalf("expected update audit record to expose operation_id, got %+v", assets.auditRecords)
	}
	undone, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: operationID,
	})
	if err != nil {
		t.Fatalf("undo asset update: %v", err)
	}
	if undone.Title.String() == title || undone.ParentAssetID != asset.ID("garage") {
		t.Fatalf("expected undo to restore prior asset state, got %+v", undone)
	}

	redone, err := application.RedoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: operationID,
	})
	if err != nil {
		t.Fatalf("redo asset update: %v", err)
	}
	if redone.Title != updated.Title || redone.ParentAssetID != updated.ParentAssetID {
		t.Fatalf("expected redo to reapply updated state, got %+v", redone)
	}
	if !appAuditRecordsIncludeAction(assets.auditRecords, audit.ActionUndoableOperationUndone) || !appAuditRecordsIncludeAction(assets.auditRecords, audit.ActionUndoableOperationRedone) {
		t.Fatalf("expected undo and redo audit records, got %+v", assets.auditRecords)
	}
}

func TestRedoRejectsStaleAssetState(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
	}); err != nil {
		t.Fatalf("update asset: %v", err)
	}
	operationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]
	if _, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: operationID,
	}); err != nil {
		t.Fatalf("undo asset update: %v", err)
	}

	staleTitle := "Changed after undo"
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &staleTitle,
	}); err != nil {
		t.Fatalf("make asset stale: %v", err)
	}
	_, err := application.RedoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: operationID,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected stale redo rejection, got %v", err)
	}
}

func TestUndoAllowsArchivedExistingCustomAssetType(t *testing.T) {
	medicineType := customAssetType(t, "medicine-type", "tenant-one", "inventory-one", customfield.ScopeInventory, "medicine", "Medicine")
	medicineType, ok := medicineType.Archive()
	if !ok {
		t.Fatalf("archive custom asset type")
	}
	drill := assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, "")
	drill.CustomAssetTypeID = asset.CustomAssetTypeID(medicineType.ID.String())
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		drill.ID: drill,
	}}
	application := New(Dependencies{
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomAssetTypes:          &fakeCustomAssetTypeRepository{items: []customfield.AssetType{medicineType}},
		CustomAssetTypeUnitOfWork: &fakeCustomAssetTypeRepository{items: []customfield.AssetType{medicineType}},
		CustomFields:              &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork:     &fakeCustomFieldRepository{},
		Assets:                    assets,
		AssetUnitOfWork:           assets,
		Undoables:                 assets,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     drill.ID,
		Title:       &title,
	}); err != nil {
		t.Fatalf("update asset with archived existing custom asset type: %v", err)
	}
	operationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]

	undone, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: operationID,
	})
	if err != nil {
		t.Fatalf("undo asset with archived existing custom asset type: %v", err)
	}
	if undone.CustomAssetTypeID != drill.CustomAssetTypeID || undone.Title != drill.Title {
		t.Fatalf("expected undo to preserve archived custom asset type and title, got %+v", undone)
	}
}

func TestUndoAndRedoAssetLifecycleOperations(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{},
	})

	archived, err := application.ArchiveAsset(context.Background(), UpdateAssetLifecycleInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
	})
	if err != nil {
		t.Fatalf("archive asset: %v", err)
	}
	if archived.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected archived asset, got %+v", archived)
	}
	archiveOperationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]

	undoArchive, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: archiveOperationID,
	})
	if err != nil {
		t.Fatalf("undo archive: %v", err)
	}
	if undoArchive.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected undo archive to restore active state, got %+v", undoArchive)
	}

	redoArchive, err := application.RedoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: archiveOperationID,
	})
	if err != nil {
		t.Fatalf("redo archive: %v", err)
	}
	if redoArchive.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected redo archive to archive asset, got %+v", redoArchive)
	}

	restored, err := application.RestoreAsset(context.Background(), UpdateAssetLifecycleInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
	})
	if err != nil {
		t.Fatalf("restore asset: %v", err)
	}
	if restored.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("expected restored asset, got %+v", restored)
	}
	restoreOperationID := assets.auditRecords[len(assets.auditRecords)-1].Metadata["operation_id"]

	undoRestore, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		OperationID: restoreOperationID,
	})
	if err != nil {
		t.Fatalf("undo restore: %v", err)
	}
	if undoRestore.LifecycleState != asset.LifecycleStateArchived {
		t.Fatalf("expected undo restore to archive asset, got %+v", undoRestore)
	}
}

func appAuditRecordsIncludeAction(records []audit.Record, action audit.Action) bool {
	for _, record := range records {
		if record.Action == action {
			return true
		}
	}
	return false
}

func TestUpdateAssetRejectsInvalidMovement(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"):   assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):    assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("box"):      assetItem("box", "tenant-one", "inventory-one", asset.KindContainer, "shelf"),
		asset.ID("wrench"):   assetItem("wrench", "tenant-one", "inventory-one", asset.KindItem, "box"),
		asset.ID("supplies"): assetItem("supplies", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets:                assets,
		AssetUnitOfWork:       assets,
		Undoables:             assets,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{},
	})

	for _, item := range []struct {
		name    string
		assetID asset.ID
		parent  string
	}{
		{name: "self parent", assetID: asset.ID("box"), parent: "box"},
		{name: "cycle through descendant", assetID: asset.ID("garage"), parent: "box"},
		{name: "item parent", assetID: asset.ID("wrench"), parent: "supplies"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
				Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
				AssetID:     item.assetID,
				ParentAssetID: AssetParentUpdate{
					Present: true,
					Value:   item.parent,
				},
			})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid movement rejection, got %v", err)
			}
		})
	}

	updated, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("box"),
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Null:    true,
		},
	})
	if err != nil {
		t.Fatalf("move container to root: %v", err)
	}
	if updated.ParentAssetID.String() != "" || assets.items[asset.ID("wrench")].ParentAssetID != asset.ID("box") {
		t.Fatalf("expected box moved to root with child preserved, box=%+v wrench=%+v", updated, assets.items[asset.ID("wrench")])
	}

	_, err = application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("box"),
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   " ",
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected blank parent rejection, got %v", err)
	}
}

func TestUpdateAssetRequiresEditPermission(t *testing.T) {
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          &fakeCustomFieldRepository{},
		CustomFieldUnitOfWork: &fakeCustomFieldRepository{},
		Assets: &fakeAssetRepository{items: map[asset.ID]asset.Asset{
			asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
		}},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
		IDs:    &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	_, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized update, got %v", err)
	}
}
