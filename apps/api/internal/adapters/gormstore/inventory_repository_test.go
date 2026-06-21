package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestStorePersistsInventories(t *testing.T) {
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

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
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
