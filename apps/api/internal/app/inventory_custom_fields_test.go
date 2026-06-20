package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestCreateAndListCustomFieldDefinitions(t *testing.T) {
	observer := &fakeObserver{}
	customFields := &fakeCustomFieldRepository{}
	application := New(Dependencies{
		Observer:              observer,
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          customFields,
		CustomFieldUnitOfWork: customFields,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{ids: []string{"tenant-definition", "inventory-definition"}},
		MaxPageLimit:          1,
	})

	tenantDefinition, err := application.CreateTenantCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if err != nil {
		t.Fatalf("create tenant definition: %v", err)
	}
	if tenantDefinition.Scope != customfield.ScopeTenant || tenantDefinition.Key != customfield.Key("serial") {
		t.Fatalf("unexpected tenant definition: %+v", tenantDefinition)
	}

	inventoryDefinition, err := application.CreateInventoryCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "condition",
		DisplayName: "Condition",
		Type:        "enum",
		EnumOptions: []string{"new", "used"},
	})
	if err != nil {
		t.Fatalf("create inventory definition: %v", err)
	}
	if inventoryDefinition.Scope != customfield.ScopeInventory || len(inventoryDefinition.EnumOptions) != 2 {
		t.Fatalf("unexpected inventory definition: %+v", inventoryDefinition)
	}

	firstPage, err := application.ListInventoryCustomFieldDefinitions(context.Background(), ListCustomFieldDefinitionsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].ID != tenantDefinition.ID || !firstPage.HasMore || firstPage.NextCursor == nil {
		t.Fatalf("expected first page with inherited tenant definition, got %+v", firstPage)
	}

	secondPage, err := application.ListInventoryCustomFieldDefinitions(context.Background(), ListCustomFieldDefinitionsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
		Cursor:      *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != inventoryDefinition.ID || secondPage.HasMore {
		t.Fatalf("expected second page with inventory definition, got %+v", secondPage)
	}
	if !observer.hasEvent(ports.EventCustomFieldDefinitionCreated) || !observer.hasEvent(ports.EventCustomFieldDefinitionsListed) {
		t.Fatalf("expected custom field observability events, got %+v", observer.events)
	}
}

func TestUpdateCustomFieldDefinitionSchemaExpansionRecordsAuditAndObservability(t *testing.T) {
	observer := &fakeObserver{}
	customFields := &fakeCustomFieldRepository{}
	customAssetTypes := &fakeCustomAssetTypeRepository{items: []customfield.AssetType{
		customAssetType(t, "medicine-type", "tenant-one", "inventory-one", customfield.ScopeInventory, "medicine", "Medicine"),
		customAssetType(t, "supply-type", "tenant-one", "inventory-one", customfield.ScopeInventory, "supply", "Supply"),
	}}
	application := New(Dependencies{
		Observer:                  observer,
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Cabinet")}},
		CustomAssetTypes:          customAssetTypes,
		CustomAssetTypeUnitOfWork: customAssetTypes,
		CustomFields:              customFields,
		CustomFieldUnitOfWork:     customFields,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"condition-field", "audit-create", "audit-update"}},
	})

	definition, err := application.CreateInventoryCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:          identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:           tenant.ID("tenant-one"),
		InventoryID:        inventory.InventoryID("inventory-one"),
		Key:                "condition",
		DisplayName:        "Condition",
		Type:               "enum",
		EnumOptions:        []string{"new", "used"},
		Applicability:      "custom_asset_types",
		CustomAssetTypeIDs: []string{"medicine-type"},
	})
	if err != nil {
		t.Fatalf("create custom field definition: %v", err)
	}

	updated, err := application.UpdateInventoryCustomFieldDefinition(context.Background(), UpdateCustomFieldDefinitionInput{
		Principal:          identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:           tenant.ID("tenant-one"),
		InventoryID:        inventory.InventoryID("inventory-one"),
		DefinitionID:       definition.ID,
		DisplayName:        stringPtr("Condition"),
		EnumOptions:        stringSlicePtr("new", "used", "expired"),
		CustomAssetTypeIDs: stringSlicePtr("medicine-type", "supply-type"),
	})
	if err != nil {
		t.Fatalf("update custom field definition: %v", err)
	}
	if len(updated.EnumOptions) != 3 || len(updated.CustomAssetTypeIDs) != 2 {
		t.Fatalf("expected enum and target expansion, got %+v", updated)
	}

	if !observer.hasEvent(ports.EventCustomFieldDefinitionUpdated) {
		t.Fatalf("expected custom field update event, got %+v", observer.events)
	}
	record, ok := customFields.recordForAction(audit.ActionCustomFieldDefinitionUpdated)
	if !ok {
		t.Fatalf("expected custom field update audit, got %+v", customFields.auditRecords)
	}
	if record.Metadata["enum_options_added"] != "1" || record.Metadata["custom_asset_type_targets_added"] != "1" {
		t.Fatalf("expected schema expansion audit metadata, got %+v", record.Metadata)
	}
}

func TestCustomFieldDefinitionsRejectUnauthorizedAndDuplicateKeys(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkTenantErr: ports.ErrForbidden,
		},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          customFields,
		CustomFieldUnitOfWork: customFields,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{ids: []string{"definition-one"}},
	})

	_, err := application.CreateTenantCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized tenant definition create, got %v", err)
	}

	allowed := New(Dependencies{
		Observer:              &fakeObserver{},
		Authorizer:            &fakeAuthorizer{},
		Tenants:               &fakeTenantRepository{exists: true},
		TenantUnitOfWork:      &fakeTenantRepository{exists: true},
		Inventories:           &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields:          customFields,
		CustomFieldUnitOfWork: customFields,
		Audit:                 &fakeAuditRepository{},
		Outbox:                &fakeOutbox{},
		IDs:                   &fakeIDGenerator{ids: []string{"definition-two", "definition-three"}},
	})
	_, err = allowed.CreateTenantCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if err != nil {
		t.Fatalf("create first definition: %v", err)
	}
	_, err = allowed.CreateInventoryCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected duplicate effective key rejection, got %v", err)
	}
}

func customAssetType(t *testing.T, id string, tenantID string, inventoryID string, scope customfield.Scope, key string, displayName string) customfield.AssetType {
	t.Helper()

	assetType, ok := customfield.NewAssetType(
		customfield.AssetTypeID(id),
		customfield.TenantID(tenantID),
		customfield.InventoryID(inventoryID),
		scope,
		customfield.Key(key),
		customfield.DisplayName(displayName),
		customfield.Description(""),
	)
	if !ok {
		t.Fatalf("invalid test custom asset type %q", id)
	}
	return assetType
}

func stringPtr(value string) *string {
	return &value
}

func stringSlicePtr(values ...string) *[]string {
	return &values
}
