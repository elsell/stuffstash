package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStorePersistsTenantsAndInventories(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenantID, Name: tenantName}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}

	exists, err := store.TenantExists(ctx, tenantID)
	if err != nil {
		t.Fatalf("check tenant exists: %v", err)
	}
	if !exists {
		t.Fatalf("expected tenant to exist")
	}

	inventoryName, ok := inventory.NewName("Tools")
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	item := inventory.Inventory{
		ID:       inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW"),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}
	if err := store.SaveInventory(ctx, item); err != nil {
		t.Fatalf("save inventory: %v", err)
	}

	items, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 inventory, got %d", len(items))
	}
	if items[0].ID != item.ID || items[0].TenantID != item.TenantID || items[0].Name != item.Name {
		t.Fatalf("unexpected inventory: %+v", items[0])
	}
}

func TestStoreKeepsInventoriesScopedToTenant(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	tenantOne := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantTwo := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantOne, "Home")
	saveTenant(t, ctx, store, tenantTwo, "Cabin")
	saveInventory(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantOne, "Tools")
	saveInventory(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantTwo, "Supplies")

	items, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantOne.String()), ports.InventoryListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 inventory, got %d", len(items))
	}
	if items[0].TenantID != inventory.TenantID(tenantOne.String()) {
		t.Fatalf("expected tenant %q, got %q", tenantOne, items[0].TenantID)
	}
}

func TestTenantExistsReturnsFalseForMissingTenant(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	exists, err := store.TenantExists(ctx, tenant.ID("missing"))
	if err != nil {
		t.Fatalf("check tenant exists: %v", err)
	}
	if exists {
		t.Fatalf("expected missing tenant")
	}
}

func TestStoreSavesTenantAndOutboxEventAtomically(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}

	err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenant.Tenant{
		ID:   tenantID,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")})
	if err != nil {
		t.Fatalf("save tenant and enqueue owner grant: %v", err)
	}

	exists, err := store.TenantExists(ctx, tenantID)
	if err != nil {
		t.Fatalf("check tenant exists: %v", err)
	}
	if !exists {
		t.Fatalf("expected tenant to exist")
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(events))
	}
	if events[0].Kind != ports.AuthorizationOutboxGrantTenantOwner || events[0].TenantID != tenantID || events[0].PrincipalID != "user-one" {
		t.Fatalf("unexpected outbox event: %+v", events[0])
	}
}

func TestStoreRollsBackTenantWhenOutboxInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	eventID := "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	existingTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	newTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenantWithOutbox(t, ctx, store, eventID, existingTenantID, "Home")

	tenantName, ok := tenant.NewName("Cabin")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	err := store.SaveTenantAndEnqueueOwnerGrant(ctx, eventID, tenant.Tenant{
		ID:   newTenantID,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-two")})
	if err == nil {
		t.Fatalf("expected duplicate outbox event to fail")
	}

	exists, err := store.TenantExists(ctx, newTenantID)
	if err != nil {
		t.Fatalf("check tenant exists: %v", err)
	}
	if exists {
		t.Fatalf("expected tenant write to roll back when outbox insert fails")
	}
}

func TestStoreSavesInventoryAndOutboxEventAtomically(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenant(t, ctx, store, tenantID, "Home")

	inventoryName, ok := inventory.NewName("Tools")
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	item := inventory.Inventory{
		ID:       inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW"),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}

	err := store.SaveInventoryAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAX", item, tenantID, identity.Principal{ID: identity.PrincipalID("user-one")})
	if err != nil {
		t.Fatalf("save inventory and enqueue owner grant: %v", err)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", len(events))
	}
	if events[0].Kind != ports.AuthorizationOutboxGrantInventoryOwner || events[0].TenantID != tenantID || events[0].InventoryID != item.ID || events[0].PrincipalID != "user-one" {
		t.Fatalf("unexpected outbox event: %+v", events[0])
	}
}

func TestStoreRollsBackInventoryWhenOutboxInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	eventID := "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventoryWithOutbox(t, ctx, store, eventID, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Tools")

	inventoryName, ok := inventory.NewName("Supplies")
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	err := store.SaveInventoryAndEnqueueOwnerGrant(ctx, eventID, inventory.Inventory{
		ID:       inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAY"),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}, tenantID, identity.Principal{ID: identity.PrincipalID("user-two")})
	if err == nil {
		t.Fatalf("expected duplicate outbox event to fail")
	}

	items, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(items) != 1 || items[0].ID != inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW") {
		t.Fatalf("expected inventory write to roll back when outbox insert fails, got %+v", items)
	}
}

func TestStoreSavesInventoryAccessGrantAndOutboxEventAtomically(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	grant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  identity.PrincipalID("viewer-user"),
		Relationship: ports.InventoryAccessViewer,
	}
	if err := store.SaveInventoryAccessGrantAndEnqueue(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAX", grant); err != nil {
		t.Fatalf("save inventory access grant: %v", err)
	}

	grants, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryID, ports.InventoryAccessGrantPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list inventory access grants: %v", err)
	}
	if len(grants) != 1 || grants[0] != grant {
		t.Fatalf("expected saved grant, got %+v", grants)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 outbox event, got %+v", events)
	}
	if events[0].Kind != ports.AuthorizationOutboxGrantInventoryViewer || events[0].TenantID != tenantID || events[0].InventoryID != inventoryID || events[0].PrincipalID != "viewer-user" {
		t.Fatalf("unexpected outbox event: %+v", events[0])
	}
}

func TestStoreInventoryAccessGrantIsIdempotentWithoutDuplicateOutboxEvent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	grant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  identity.PrincipalID("viewer-user"),
		Relationship: ports.InventoryAccessViewer,
	}
	if err := store.SaveInventoryAccessGrantAndEnqueue(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAX", grant); err != nil {
		t.Fatalf("save initial grant: %v", err)
	}
	if err := store.SaveInventoryAccessGrantAndEnqueue(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY", grant); err != nil {
		t.Fatalf("save duplicate grant: %v", err)
	}

	grants, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryID, ports.InventoryAccessGrantPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list inventory access grants: %v", err)
	}
	if len(grants) != 1 || grants[0] != grant {
		t.Fatalf("expected one idempotent grant, got %+v", grants)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].ID != "01ARZ3NDEKTSV4RRFFQ69G5FAX" {
		t.Fatalf("expected one outbox event from first grant, got %+v", events)
	}
}

func TestStoreScopesInventoryAccessGrantsToInventory(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryOneID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryTwoID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryOneID.String(), tenantID, "Tools")
	saveInventory(t, ctx, store, inventoryTwoID.String(), tenantID, "Supplies")

	for _, item := range []struct {
		eventID     string
		inventoryID inventory.InventoryID
	}{
		{eventID: "01ARZ3NDEKTSV4RRFFQ69G5FAY", inventoryID: inventoryOneID},
		{eventID: "01ARZ3NDEKTSV4RRFFQ69G5FAZ", inventoryID: inventoryTwoID},
	} {
		grant := ports.InventoryAccessGrant{
			TenantID:     tenantID,
			InventoryID:  item.inventoryID,
			PrincipalID:  identity.PrincipalID("same-user"),
			Relationship: ports.InventoryAccessViewer,
		}
		if err := store.SaveInventoryAccessGrantAndEnqueue(ctx, item.eventID, grant); err != nil {
			t.Fatalf("save scoped grant: %v", err)
		}
	}

	firstInventoryGrants, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryOneID, ports.InventoryAccessGrantPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list first inventory grants: %v", err)
	}
	if len(firstInventoryGrants) != 1 || firstInventoryGrants[0].InventoryID != inventoryOneID {
		t.Fatalf("expected only first inventory grant, got %+v", firstInventoryGrants)
	}

	secondInventoryGrants, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryTwoID, ports.InventoryAccessGrantPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list second inventory grants: %v", err)
	}
	if len(secondInventoryGrants) != 1 || secondInventoryGrants[0].InventoryID != inventoryTwoID {
		t.Fatalf("expected only second inventory grant, got %+v", secondInventoryGrants)
	}
}

func TestStorePaginatesInventoryAccessGrants(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	editorGrant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  identity.PrincipalID("editor-user"),
		Relationship: ports.InventoryAccessEditor,
	}
	viewerGrant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  identity.PrincipalID("viewer-user"),
		Relationship: ports.InventoryAccessViewer,
	}
	if err := store.SaveInventoryAccessGrantAndEnqueue(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAX", viewerGrant); err != nil {
		t.Fatalf("save viewer grant: %v", err)
	}
	if err := store.SaveInventoryAccessGrantAndEnqueue(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY", editorGrant); err != nil {
		t.Fatalf("save editor grant: %v", err)
	}

	page, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryID, ports.InventoryAccessGrantPageRequest{Limit: 1})
	if err != nil {
		t.Fatalf("list first grant page: %v", err)
	}
	if len(page) != 1 || page[0] != editorGrant {
		t.Fatalf("expected editor first by cursor key, got %+v", page)
	}

	nextPage, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryID, ports.InventoryAccessGrantPageRequest{
		AfterGrantKey: "editor-user:editor",
		Limit:         1,
	})
	if err != nil {
		t.Fatalf("list second grant page: %v", err)
	}
	if len(nextPage) != 1 || nextPage[0] != viewerGrant {
		t.Fatalf("expected viewer second by cursor key, got %+v", nextPage)
	}
}

func TestStoreRejectsInventoryAccessGrantOutsideInventoryTenant(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantOneID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantTwoID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	saveTenant(t, ctx, store, tenantOneID, "Home")
	saveTenant(t, ctx, store, tenantTwoID, "Cabin")
	saveInventory(t, ctx, store, inventoryID.String(), tenantTwoID, "Supplies")

	err := store.SaveInventoryAccessGrantAndEnqueue(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY", ports.InventoryAccessGrant{
		TenantID:     tenantOneID,
		InventoryID:  inventoryID,
		PrincipalID:  identity.PrincipalID("viewer-user"),
		Relationship: ports.InventoryAccessViewer,
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected tenant/inventory mismatch rejection, got %v", err)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no outbox event for rejected grant, got %+v", events)
	}
}

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
	if err := store.SaveCustomFieldDefinition(ctx, tenantDefinition); err != nil {
		t.Fatalf("save tenant definition: %v", err)
	}
	if err := store.SaveCustomFieldDefinition(ctx, inventoryDefinition); err != nil {
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
	if err := store.SaveCustomFieldDefinition(ctx, first); err != nil {
		t.Fatalf("save first definition: %v", err)
	}
	if err := store.SaveCustomFieldDefinition(ctx, second); err != nil {
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
	if err := store.SaveCustomFieldDefinition(ctx, first); err != nil {
		t.Fatalf("save first definition: %v", err)
	}
	if err := store.SaveCustomFieldDefinition(ctx, duplicate); err == nil {
		t.Fatalf("expected duplicate tenant key rejection")
	}

	inventoryFirst := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeText, nil)
	inventoryDuplicate := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID, inventoryID, customfield.ScopeInventory, "condition", customfield.FieldTypeText, nil)
	if err := store.SaveCustomFieldDefinition(ctx, inventoryFirst); err != nil {
		t.Fatalf("save inventory definition: %v", err)
	}
	if err := store.SaveCustomFieldDefinition(ctx, inventoryDuplicate); err == nil {
		t.Fatalf("expected duplicate inventory key rejection")
	}

	inventoryOnly := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID, inventoryID, customfield.ScopeInventory, "warranty", customfield.FieldTypeText, nil)
	tenantConflict := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantID, "", customfield.ScopeTenant, "warranty", customfield.FieldTypeText, nil)
	if err := store.SaveCustomFieldDefinition(ctx, inventoryOnly); err != nil {
		t.Fatalf("save inventory-only definition: %v", err)
	}
	if err := store.SaveCustomFieldDefinition(ctx, tenantConflict); err == nil {
		t.Fatalf("expected tenant key to conflict with existing inventory key")
	}
}

func TestStorePaginatesInventoriesByTenant(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "First")
	saveInventory(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, "Second")

	page, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{Limit: 1})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(page) != 1 || page[0].ID != inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW") {
		t.Fatalf("expected first inventory page, got %+v", page)
	}

	nextPage, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{
		AfterInventoryID: page[0].ID,
		Limit:            1,
	})
	if err != nil {
		t.Fatalf("list next page: %v", err)
	}
	if len(nextPage) != 1 || nextPage[0].ID != inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX") {
		t.Fatalf("expected second inventory page, got %+v", nextPage)
	}
}

func TestStoreMarksOutboxEventsProcessedAndFailed(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenant.Tenant{
		ID:   tenantID,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")}); err != nil {
		t.Fatalf("save tenant and enqueue owner grant: %v", err)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 claimed event, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventFailed(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "claim-one", "spicedb unavailable"); err != nil {
		t.Fatalf("mark outbox failed: %v", err)
	}
	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].Attempts != 1 || events[0].LastError != "spicedb unavailable" {
		t.Fatalf("expected failed event to remain pending, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventProcessed(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "wrong-claim"); !errors.Is(err, ports.ErrAuthorizationOutboxClaimLost) {
		t.Fatalf("expected claim lost error, got %v", err)
	}
	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-three", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected active claim to hide event from wrong processor, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventProcessed(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "claim-two"); err != nil {
		t.Fatalf("mark outbox processed: %v", err)
	}
	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-three", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected processed event hidden from pending list, got %+v", events)
	}
}

func TestStoreMarksOutboxEventsDeadLettered(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenantWithOutbox(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Home")

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 claimed event, got %+v", events)
	}

	if err := store.MarkAuthorizationOutboxEventDeadLettered(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "wrong-claim", "invalid event"); !errors.Is(err, ports.ErrAuthorizationOutboxClaimLost) {
		t.Fatalf("expected claim lost error, got %v", err)
	}
	if err := store.MarkAuthorizationOutboxEventDeadLettered(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAW", "claim-one", "invalid event"); err != nil {
		t.Fatalf("mark outbox dead-lettered: %v", err)
	}

	var model authorizationOutboxEventModel
	if err := store.db.WithContext(ctx).Where(&authorizationOutboxEventModel{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAW"}).First(&model).Error; err != nil {
		t.Fatalf("load outbox event: %v", err)
	}
	if model.DeadLetteredAt == nil || model.DeadLetterReason != "invalid event" {
		t.Fatalf("expected dead-letter details, got %+v", model)
	}

	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected dead-lettered event hidden from pending list, got %+v", events)
	}
}

func TestStoreClaimsHideEventsUntilLeaseExpires(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenantWithOutbox(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Home")

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].ClaimID != "claim-one" {
		t.Fatalf("expected claim-one to own event, got %+v", events)
	}

	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected active lease to hide event, got %+v", events)
	}
}

func TestStoreReclaimsEventsAfterLeaseExpires(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	saveTenantWithOutbox(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID, "Home")

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected claim-one to claim event, got %+v", events)
	}

	events, err = store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].ClaimID != "claim-two" {
		t.Fatalf("expected expired lease to be reclaimed, got %+v", events)
	}
}

func TestStorePersistsAssetsAndLocationParents(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	location := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), asset.KindLocation, "")
	if err := store.CreateAsset(ctx, location); err != nil {
		t.Fatalf("save location asset: %v", err)
	}
	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, location.ID.String())
	if err := store.CreateAsset(ctx, item); err != nil {
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
	if err := store.CreateAsset(ctx, itemParent); err != nil {
		t.Fatalf("save item parent: %v", err)
	}
	child := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID.String(), inventoryOneID.String(), asset.KindItem, itemParent.ID.String())
	if err := store.CreateAsset(ctx, child); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected item parent rejection, got %v", err)
	}

	location := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID.String(), inventoryOneID.String(), asset.KindLocation, "")
	if err := store.CreateAsset(ctx, location); err != nil {
		t.Fatalf("save location parent: %v", err)
	}
	crossInventoryChild := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID.String(), inventoryTwoID.String(), asset.KindItem, location.ID.String())
	if err := store.CreateAsset(ctx, crossInventoryChild); !errors.Is(err, ports.ErrForbidden) {
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
	if err := store.CreateAsset(ctx, item); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected tenant/inventory mismatch rejection, got %v", err)
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
	if err := store.CreateAsset(ctx, parent); err != nil {
		t.Fatalf("save parent: %v", err)
	}
	child := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindContainer, parent.ID.String())
	if err := store.CreateAsset(ctx, child); err != nil {
		t.Fatalf("save child: %v", err)
	}

	parent.ParentAssetID = child.ID
	if err := store.CreateAsset(ctx, parent); err == nil {
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
	if err := store.CreateAsset(ctx, first); err != nil {
		t.Fatalf("create first asset: %v", err)
	}
	if err := store.CreateAsset(ctx, second); err != nil {
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
	if err := store.CreateAsset(ctx, first); err == nil {
		t.Fatalf("expected duplicate asset create to fail")
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
		if err := store.CreateAsset(ctx, item); err != nil {
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
	if err := store.UpdateAsset(ctx, box); err != nil {
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
	if err := store.UpdateAsset(ctx, box); err != nil {
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
		if err := store.CreateAsset(ctx, item); err != nil {
			t.Fatalf("create asset %s: %v", item.ID, err)
		}
	}

	garage.ParentAssetID = box.ID
	if err := store.UpdateAsset(ctx, garage); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected cycle rejection, got %v", err)
	}
	box.ParentAssetID = box.ID
	if err := store.UpdateAsset(ctx, box); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected self-parent rejection, got %v", err)
	}
	box.ParentAssetID = itemParent.ID
	if err := store.UpdateAsset(ctx, box); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected item-parent rejection, got %v", err)
	}
	box.Kind = asset.KindItem
	if err := store.UpdateAsset(ctx, box); !errors.Is(err, ports.ErrForbidden) {
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

	if err := store.CreateAsset(ctx, item); err != nil {
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

func newTestStore(t *testing.T, ctx context.Context) Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite fake: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate sqlite fake: %v", err)
	}

	return NewStore(db)
}

func saveTenant(t *testing.T, ctx context.Context, store Store, id tenant.ID, name string) {
	t.Helper()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: id, Name: tenantName}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}
}

func saveTenantWithOutbox(t *testing.T, ctx context.Context, store Store, eventID string, id tenant.ID, name string) {
	t.Helper()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(ctx, eventID, tenant.Tenant{
		ID:   id,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")}); err != nil {
		t.Fatalf("save tenant with outbox: %v", err)
	}
}

func saveInventory(t *testing.T, ctx context.Context, store Store, id string, tenantID tenant.ID, name string) {
	t.Helper()

	inventoryName, ok := inventory.NewName(name)
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	item := inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}
	if err := store.SaveInventory(ctx, item); err != nil {
		t.Fatalf("save inventory: %v", err)
	}
}

func saveInventoryWithOutbox(t *testing.T, ctx context.Context, store Store, eventID string, id string, tenantID tenant.ID, name string) {
	t.Helper()

	inventoryName, ok := inventory.NewName(name)
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	item := inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}
	if err := store.SaveInventoryAndEnqueueOwnerGrant(ctx, eventID, item, tenantID, identity.Principal{ID: identity.PrincipalID("user-one")}); err != nil {
		t.Fatalf("save inventory with outbox: %v", err)
	}
}

func customFieldDefinition(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, scope customfield.Scope, keyValue string, fieldType customfield.FieldType, rawOptions []string) customfield.Definition {
	t.Helper()

	definitionID, ok := customfield.NewID(id)
	if !ok {
		t.Fatalf("expected valid definition id")
	}
	key, ok := customfield.NewKey(keyValue)
	if !ok {
		t.Fatalf("expected valid custom field key")
	}
	displayName, ok := customfield.NewDisplayName("Field " + keyValue)
	if !ok {
		t.Fatalf("expected valid display name")
	}
	options := make([]customfield.Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := customfield.NewKey(raw)
		if !ok {
			t.Fatalf("expected valid enum option")
		}
		options = append(options, option)
	}
	definition, ok := customfield.NewDefinition(
		definitionID,
		customfield.TenantID(tenantID.String()),
		customfield.InventoryID(inventoryID.String()),
		scope,
		key,
		displayName,
		fieldType,
		options,
	)
	if !ok {
		t.Fatalf("expected valid custom field definition")
	}
	return definition
}

func assetItem(id string, tenantID string, inventoryID string, kind asset.Kind, parentID string) asset.Asset {
	title, ok := asset.NewTitle("Asset " + id)
	if !ok {
		panic("invalid test asset title")
	}
	parent := asset.ID("")
	if parentID != "" {
		var parentOK bool
		parent, parentOK = asset.NewID(parentID)
		if !parentOK {
			panic("invalid parent id")
		}
	}
	return asset.Asset{
		ID:             asset.ID(id),
		TenantID:       asset.TenantID(tenantID),
		InventoryID:    asset.InventoryID(inventoryID),
		ParentAssetID:  parent,
		Kind:           kind,
		Title:          title,
		CustomFields:   asset.NewEmptyCustomFields(),
		LifecycleState: asset.LifecycleStateActive,
	}
}
