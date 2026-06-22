package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestStorePersistsTenants(t *testing.T) {
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

func TestStoreListsActiveTenantsByCursor(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	saveTenantForList(t, ctx, store, tenant.ID("tenant-a"), "Tenant A", tenant.LifecycleStateActive)
	saveTenantForList(t, ctx, store, tenant.ID("tenant-b"), "Tenant B", tenant.LifecycleStateActive)
	saveTenantForList(t, ctx, store, tenant.ID("tenant-c"), "Tenant C", tenant.LifecycleStateArchived)
	saveTenantForList(t, ctx, store, tenant.ID("tenant-d"), "Tenant D", tenant.LifecycleStateActive)

	items, err := store.ListTenants(ctx, ports.TenantListPageRequest{
		AfterTenantID: tenant.ID("tenant-a"),
		Limit:         2,
	})
	if err != nil {
		t.Fatalf("list tenants: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(items))
	}
	if items[0].ID != tenant.ID("tenant-b") || items[1].ID != tenant.ID("tenant-d") {
		t.Fatalf("expected tenant-b and tenant-d, got %+v", items)
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

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
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

func saveTenantForList(t *testing.T, ctx context.Context, store Store, id tenant.ID, name string, lifecycleState tenant.LifecycleState) {
	t.Helper()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: id, Name: tenantName, LifecycleState: lifecycleState}); err != nil {
		t.Fatalf("save tenant %s: %v", id, err)
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
