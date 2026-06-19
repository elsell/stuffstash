package gormstore

import (
	"context"
	"errors"
	"testing"
	"time"

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

	items, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()))
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

	items, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantOne.String()))
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

	items, err := store.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()))
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(items) != 1 || items[0].ID != inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW") {
		t.Fatalf("expected inventory write to roll back when outbox insert fails, got %+v", items)
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
