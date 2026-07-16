package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
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

	locationResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "location",
		Title:       "Garage",
	})
	if err != nil {
		t.Fatalf("create location asset: %v", err)
	}
	location := locationResult.Asset
	if location.Kind != asset.KindLocation || location.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("unexpected location asset: %+v", location)
	}

	itemResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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
	item := itemResult.Asset
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

func assetTag(id, tenantID, inventoryID, keyValue, displayNameValue, color string) assettag.Tag {
	key, ok := assettag.NewKey(keyValue)
	if !ok {
		panic("invalid test tag key")
	}
	displayName, ok := assettag.NewDisplayName(displayNameValue)
	if !ok {
		panic("invalid test tag display name")
	}
	tagColor, ok := assettag.NewColor(color)
	if !ok {
		panic("invalid test tag color")
	}
	tag, ok := assettag.NewTag(assettag.ID(id), assettag.TenantID(tenantID), assettag.InventoryID(inventoryID), key, displayName, tagColor, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC))
	if !ok {
		panic("invalid test tag")
	}
	return tag
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
	itemResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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
	item := itemResult.Asset

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

	_, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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

func TestUpdateAssetPersistsFieldsAndTagsAsOneUndoableChangeAndSkipsNoOp(t *testing.T) {
	item := assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, "")
	workshop := assetTag("tag-workshop", "tenant-one", "inventory-one", "workshop", "Workshop", "#2f80ed")
	camping := assetTag("tag-camping", "tenant-one", "inventory-one", "camping", "Camping", "#27ae60")
	assets := &fakeAssetRepository{
		items:         map[asset.ID]asset.Asset{item.ID: item},
		assetTags:     map[assettag.ID]assettag.Tag{workshop.ID: workshop, camping.ID: camping},
		assetTagLinks: map[asset.ID]map[assettag.ID]struct{}{item.ID: {workshop.ID: {}}},
	}
	application := New(Dependencies{
		Observer: &fakeObserver{}, Authorizer: &fakeAuthorizer{}, Tenants: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		Assets:      assets, AssetTags: assets, AssetUnitOfWork: assets, AssetTagUnitOfWork: assets, AssetEditUnitOfWork: assets, Undoables: assets,
		Audit: &fakeAuditRepository{}, Outbox: &fakeOutbox{}, IDs: &fakeIDGenerator{ids: []string{"operation-update", "audit-update", "audit-undo"}},
	})

	title := "Cordless drill"
	tagIDs := []string{camping.ID.String()}
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal: identity.Principal{ID: identity.PrincipalID("editor")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: item.ID, Title: &title, TagIDs: &tagIDs,
	}); err != nil {
		t.Fatalf("update fields and tags: %v", err)
	}
	if len(assets.auditRecords) != 1 || assets.auditRecords[0].Action != audit.ActionAssetUpdated {
		t.Fatalf("expected one coherent update audit, got %+v", assets.auditRecords)
	}
	record := assets.auditRecords[0]
	if record.Metadata["previous_title"] != item.Title.String() || record.Metadata["updated_title"] != title || record.Metadata["previous_tag_count"] != "1" || record.Metadata["updated_tag_count"] != "1" {
		t.Fatalf("expected safe field and tag metadata, got %+v", record.Metadata)
	}
	operation := assets.undoables[record.Metadata["operation_id"]]
	if operation.ID == "" || len(operation.BeforeTagIDs) != 1 || operation.BeforeTagIDs[0] != workshop.ID || len(operation.AfterTagIDs) != 1 || operation.AfterTagIDs[0] != camping.ID {
		t.Fatalf("expected coherent tag snapshots on undo, got %+v", operation)
	}

	auditCount := len(assets.auditRecords)
	operationCount := len(assets.undoables)
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal: identity.Principal{ID: identity.PrincipalID("editor")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), AssetID: item.ID, Title: &title, TagIDs: &tagIDs,
	}); err != nil {
		t.Fatalf("repeat no-op edit: %v", err)
	}
	if len(assets.auditRecords) != auditCount || len(assets.undoables) != operationCount {
		t.Fatalf("no-op edit created history: audits=%d operations=%d", len(assets.auditRecords), len(assets.undoables))
	}

	if _, err := application.UndoOperation(context.Background(), ApplyUndoableOperationInput{
		Principal: identity.Principal{ID: identity.PrincipalID("editor")}, TenantID: tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"), OperationID: operation.ID,
	}); err != nil {
		t.Fatalf("undo coherent edit: %v", err)
	}
	restoredTags, err := assets.AssetTagsByAsset(context.Background(), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), item.ID)
	if err != nil {
		t.Fatalf("read restored tags: %v", err)
	}
	if assets.items[item.ID].Title != item.Title || len(restoredTags) != 1 || restoredTags[0].ID != workshop.ID {
		t.Fatalf("expected undo to restore fields and tags, item=%+v tags=%+v", assets.items[item.ID], restoredTags)
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

	childResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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
	child := childResult.Asset
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

	_, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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

	itemResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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
	item := itemResult.Asset
	if item.CustomFields.Values()["serial"] != "abc" || item.CustomFields.Values()["condition"] != "used" {
		t.Fatalf("expected custom fields to be saved, got %+v", item.CustomFields.Values())
	}

	_, err = application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
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

func appAuditRecordsIncludeAction(records []audit.Record, action audit.Action) bool {
	for _, record := range records {
		if record.Action == action {
			return true
		}
	}
	return false
}
