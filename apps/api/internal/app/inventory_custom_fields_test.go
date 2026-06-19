package app

import (
	"context"
	"errors"
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
		Observer:     observer,
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"tenant-definition", "inventory-definition"}},
		MaxPageLimit: 1,
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

func TestCustomFieldDefinitionsRejectUnauthorizedAndDuplicateKeys(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkTenantErr: ports.ErrForbidden,
		},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"definition-one"}},
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
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"definition-two", "definition-three"}},
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
