package gormstore

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testAuditRecordSequence uint64

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
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, "", audit.ActionTenantCreated))
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
	}, identity.Principal{ID: identity.PrincipalID("user-two")}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", newTenantID, "", audit.ActionTenantCreated))
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

	err := store.SaveInventoryAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAX", item, tenantID, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, item.ID, audit.ActionInventoryCreated))
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
	}, tenantID, identity.Principal{ID: identity.PrincipalID("user-two")}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAY"), audit.ActionInventoryCreated))
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
	if err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAX", grant); err != nil {
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
	if err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAX", grant); err != nil {
		t.Fatalf("save initial grant: %v", err)
	}
	if err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAY", grant); err != nil {
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
		if err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, item.eventID, grant); err != nil {
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
	if err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAX", viewerGrant); err != nil {
		t.Fatalf("save viewer grant: %v", err)
	}
	if err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAY", editorGrant); err != nil {
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

	err := saveInventoryAccessGrantAndEnqueue(t, ctx, store, "01ARZ3NDEKTSV4RRFFQ69G5FAY", ports.InventoryAccessGrant{
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
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID, "", audit.ActionTenantCreated)); err != nil {
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

func TestStorePersistsAndScopesAuditRecords(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	tenantOne := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantTwo := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryOne := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	inventoryTwo := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAY")
	inventoryThree := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAZ")
	saveTenant(t, ctx, store, tenantOne, "Home")
	saveTenant(t, ctx, store, tenantTwo, "Cabin")
	saveInventory(t, ctx, store, inventoryOne.String(), tenantOne, "Tools")
	saveInventory(t, ctx, store, inventoryTwo.String(), tenantOne, "Supplies")
	saveInventory(t, ctx, store, inventoryThree.String(), tenantTwo, "Cabin Tools")
	occurredAt := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)

	for _, record := range []audit.Record{
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantOne, inventoryOne, audit.ActionAssetUpdated, occurredAt),
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantOne, inventoryOne, audit.ActionAssetCreated, occurredAt),
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantOne, inventoryTwo, audit.ActionAssetMoved, occurredAt.Add(time.Second)),
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB3", tenantTwo, inventoryThree, audit.ActionAssetCreated, occurredAt.Add(2*time.Second)),
	} {
		if err := store.SaveAuditRecord(ctx, record); err != nil {
			t.Fatalf("save audit record %s: %v", record.ID, err)
		}
	}

	firstPage, err := store.ListInventoryAuditRecords(ctx, tenantOne, inventoryOne, ports.AuditRecordPageRequest{Limit: 1})
	if err != nil {
		t.Fatalf("list inventory audit records: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].ID != audit.ID("01ARZ3NDEKTSV4RRFFQ69G5FB0") || firstPage[0].Metadata["note"] != "safe" {
		t.Fatalf("unexpected first audit page: %+v", firstPage)
	}
	secondPage, err := store.ListInventoryAuditRecords(ctx, tenantOne, inventoryOne, ports.AuditRecordPageRequest{
		AfterOccurredAt: firstPage[0].OccurredAt,
		AfterRecordID:   firstPage[0].ID,
		Limit:           10,
	})
	if err != nil {
		t.Fatalf("list second audit page: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].ID != audit.ID("01ARZ3NDEKTSV4RRFFQ69G5FB1") {
		t.Fatalf("unexpected second audit page: %+v", secondPage)
	}

	tenantPage, err := store.ListTenantAuditRecords(ctx, tenantOne, ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list tenant audit records: %v", err)
	}
	if len(tenantPage) != 3 {
		t.Fatalf("expected tenant page to include only tenant one records, got %+v", tenantPage)
	}
}

func TestStoreRollsBackTenantAndOutboxWhenAuditInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	existingTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	newTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	auditID := "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	saveTenant(t, ctx, store, existingTenantID, "Existing")
	if err := store.SaveAuditRecord(ctx, auditRecord(t, auditID, existingTenantID, "", audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save existing audit record: %v", err)
	}

	tenantName, ok := tenant.NewName("Rollback Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenant.Tenant{
		ID:   newTenantID,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, auditID, newTenantID, "", audit.ActionTenantCreated))
	if err == nil {
		t.Fatalf("expected duplicate audit ID to fail")
	}

	exists, err := store.TenantExists(ctx, newTenantID)
	if err != nil {
		t.Fatalf("check tenant exists: %v", err)
	}
	if exists {
		t.Fatalf("expected tenant write to roll back after audit failure")
	}
	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected outbox write to roll back after audit failure, got %+v", events)
	}
}

func TestStoreRollsBackAssetWhenAuditInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	auditID := "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")
	if err := store.SaveAuditRecord(ctx, auditRecord(t, auditID, tenantID, inventoryID, audit.ActionAssetCreated)); err != nil {
		t.Fatalf("save existing audit record: %v", err)
	}

	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	err := store.CreateAsset(ctx, item, auditRecord(t, auditID, tenantID, inventoryID, audit.ActionAssetCreated))
	if err == nil {
		t.Fatalf("expected duplicate audit ID to fail")
	}
	_, found, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}
	if found {
		t.Fatalf("expected asset write to roll back after audit failure")
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
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, eventID, id, "", audit.ActionTenantCreated)); err != nil {
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
	if err := store.SaveInventoryAndEnqueueOwnerGrant(ctx, eventID, item, tenantID, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, eventID, tenantID, item.ID, audit.ActionInventoryCreated)); err != nil {
		t.Fatalf("save inventory with outbox: %v", err)
	}
}

func createAsset(t *testing.T, ctx context.Context, store Store, item asset.Asset) error {
	t.Helper()

	return store.CreateAsset(ctx, item, auditRecord(t, auditIDWithSuffix(item.ID.String(), "C"), tenant.ID(item.TenantID.String()), inventory.InventoryID(item.InventoryID.String()), audit.ActionAssetCreated))
}

func updateAsset(t *testing.T, ctx context.Context, store Store, item asset.Asset) error {
	t.Helper()

	return store.UpdateAsset(ctx, item, []audit.Record{
		auditRecord(t, auditIDWithSuffix(item.ID.String(), "U"), tenant.ID(item.TenantID.String()), inventory.InventoryID(item.InventoryID.String()), audit.ActionAssetUpdated),
	})
}

func saveCustomFieldDefinition(t *testing.T, ctx context.Context, store Store, definition customfield.Definition) error {
	t.Helper()

	return store.SaveCustomFieldDefinition(ctx, definition, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenant.ID(definition.TenantID.String()), inventory.InventoryID(definition.InventoryID.String()), audit.ActionCustomFieldDefinitionCreated))
}

func saveCustomAssetType(t *testing.T, ctx context.Context, store Store, assetType customfield.AssetType) error {
	t.Helper()

	return store.SaveCustomAssetType(ctx, assetType, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "T"), tenant.ID(assetType.TenantID.String()), inventory.InventoryID(assetType.InventoryID.String()), audit.ActionCustomAssetTypeCreated))
}

func saveInventoryAccessGrantAndEnqueue(t *testing.T, ctx context.Context, store Store, eventID string, grant ports.InventoryAccessGrant) error {
	t.Helper()

	return store.SaveInventoryAccessGrantAndEnqueue(ctx, eventID, grant, auditRecord(t, eventID, grant.TenantID, grant.InventoryID, audit.ActionInventoryAccessGranted))
}

func auditIDWithSuffix(id string, suffix string) string {
	sequence := atomic.AddUint64(&testAuditRecordSequence, 1)
	return id + "-" + suffix + "-" + strconv.FormatUint(sequence, 10)
}

func auditRecord(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action) audit.Record {
	t.Helper()
	return auditRecordAt(t, id, tenantID, inventoryID, action, time.Now())
}

func auditRecordAt(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action, occurredAt time.Time) audit.Record {
	t.Helper()

	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID.String()),
		audit.InventoryID(inventoryID.String()),
		audit.PrincipalID("user-one"),
		action,
		audit.SourceAPI,
		audit.TargetAsset,
		id+"-target",
		occurredAt,
		"",
		map[string]string{"note": "safe"},
	)
	if !ok {
		t.Fatalf("expected valid audit record")
	}
	return record
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
		customfield.ApplicabilityAllAssets,
		nil,
	)
	if !ok {
		t.Fatalf("expected valid custom field definition")
	}
	return definition
}

func customAssetType(t *testing.T, id string, tenantID string, inventoryID string, scope customfield.Scope, keyValue string) customfield.AssetType {
	t.Helper()

	assetTypeID, ok := customfield.NewAssetTypeID(id)
	if !ok {
		t.Fatalf("expected valid custom asset type id")
	}
	key, ok := customfield.NewKey(keyValue)
	if !ok {
		t.Fatalf("expected valid custom asset type key")
	}
	displayName, ok := customfield.NewDisplayName("Type " + keyValue)
	if !ok {
		t.Fatalf("expected valid custom asset type display name")
	}
	description, ok := customfield.NewDescription("")
	if !ok {
		t.Fatalf("expected valid custom asset type description")
	}
	assetType, ok := customfield.NewAssetType(assetTypeID, customfield.TenantID(tenantID), customfield.InventoryID(inventoryID), scope, key, displayName, description)
	if !ok {
		t.Fatalf("expected valid custom asset type")
	}
	return assetType
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
