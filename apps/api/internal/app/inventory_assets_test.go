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

func TestCreateAssetRejectsItemParentAndCustomFields(t *testing.T) {
	itemParent := assetItem("asset-parent", "tenant-one", "inventory-one", asset.KindItem, "")
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
		Assets: &fakeAssetRepository{
			items: map[asset.ID]asset.Asset{itemParent.ID: itemParent},
		},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
		IDs:    &fakeIDGenerator{ids: []string{"asset-one"}},
	})

	_, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:      tenant.ID("tenant-one"),
		InventoryID:   inventory.InventoryID("inventory-one"),
		Kind:          "item",
		Title:         "Bit set",
		ParentAssetID: itemParent.ID.String(),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for item parent, got %v", err)
	}

	_, err = application.CreateAsset(context.Background(), CreateAssetInput{
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
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"asset-one"}},
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
		Observer:     observer,
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{},
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

func TestUpdateAssetRejectsInvalidMovement(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"):   assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):    assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("box"):      assetItem("box", "tenant-one", "inventory-one", asset.KindContainer, "shelf"),
		asset.ID("wrench"):   assetItem("wrench", "tenant-one", "inventory-one", asset.KindItem, "box"),
		asset.ID("supplies"): assetItem("supplies", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: &fakeCustomFieldRepository{},
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{},
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
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: &fakeCustomFieldRepository{},
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
