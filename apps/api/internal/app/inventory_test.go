package app

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestCreateTenantEnqueuesAndDrainsOwnerGrant(t *testing.T) {
	authorizer := &fakeAuthorizer{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:    &fakeObserver{},
		Authorizer:  authorizer,
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"tenant-one", "event-one"}},
	})

	item, err := application.CreateTenant(context.Background(), CreateTenantInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		Name:      "Home",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if item.ID != tenant.ID("tenant-one") {
		t.Fatalf("expected tenant-one, got %q", item.ID)
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "event-one" {
		t.Fatalf("expected event-one processed, got %+v", outbox.processed)
	}
	if !slices.Contains(authorizer.tenantOwnerGrants, "user-one:tenant-one") {
		t.Fatalf("expected tenant owner grant, got %+v", authorizer.tenantOwnerGrants)
	}
}

func TestCreateInventoryEnqueuesAndDrainsOwnerGrant(t *testing.T) {
	authorizer := &fakeAuthorizer{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: authorizer,
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"inventory-one", "event-one"}},
	})

	item, err := application.CreateInventory(context.Background(), CreateInventoryInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
		Name:      "Tools",
	})
	if err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	if item.ID != inventory.InventoryID("inventory-one") {
		t.Fatalf("expected inventory-one, got %q", item.ID)
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "event-one" {
		t.Fatalf("expected event-one processed, got %+v", outbox.processed)
	}
	if !slices.Contains(authorizer.inventoryOwnerGrants, "user-one:tenant-one:inventory-one") {
		t.Fatalf("expected inventory owner grant, got %+v", authorizer.inventoryOwnerGrants)
	}
}

func TestCreateTenantKeepsOutboxEventPendingWhenGrantFails(t *testing.T) {
	expected := errors.New("spicedb unavailable")
	authorizer := &fakeAuthorizer{grantTenantOwnerErr: expected}
	observer := &fakeObserver{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:    observer,
		Authorizer:  authorizer,
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"tenant-one", "event-one"}},
	})

	_, err := application.CreateTenant(context.Background(), CreateTenantInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		Name:      "Home",
	})
	if err != nil {
		t.Fatalf("create tenant should persist outbox despite grant failure: %v", err)
	}
	if len(outbox.processed) != 0 {
		t.Fatalf("expected no processed events, got %+v", outbox.processed)
	}
	if len(outbox.failed) != 1 || outbox.failed[0] != "event-one" {
		t.Fatalf("expected event-one failure recorded, got %+v", outbox.failed)
	}
	if len(outbox.events) != 1 {
		t.Fatalf("expected event to remain pending, got %+v", outbox.events)
	}
	if !observer.hasEvent(ports.EventAuthorizationOutboxFailed) {
		t.Fatalf("expected outbox failure observability event, got %+v", observer.events)
	}
}

func TestDrainAuthorizationOutboxContinuesAfterFailedEvent(t *testing.T) {
	expected := errors.New("spicedb unavailable")
	authorizer := &fakeAuthorizer{grantTenantOwnerErr: expected}
	outbox := &fakeOutbox{
		events: []ports.AuthorizationOutboxEvent{
			{
				ID:          "tenant-event",
				Kind:        ports.AuthorizationOutboxGrantTenantOwner,
				PrincipalID: identity.PrincipalID("user-one"),
				TenantID:    tenant.ID("tenant-one"),
			},
			{
				ID:          "inventory-event",
				Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
				PrincipalID: identity.PrincipalID("user-one"),
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
			},
		},
	}
	application := New(Dependencies{
		Observer:    &fakeObserver{},
		Authorizer:  authorizer,
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{},
	})

	err := application.DrainAuthorizationOutbox(context.Background(), 10)
	if !errors.Is(err, expected) {
		t.Fatalf("expected drain error %v, got %v", expected, err)
	}
	if !slices.Contains(outbox.failed, "tenant-event") {
		t.Fatalf("expected tenant event failure, got %+v", outbox.failed)
	}
	if !slices.Contains(outbox.processed, "inventory-event") {
		t.Fatalf("expected inventory event to process after failed tenant event, got %+v", outbox.processed)
	}
}

func TestListInventoriesReturnsAuthorizationBackendFailures(t *testing.T) {
	expected := errors.New("authorization backend unavailable")
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: expected,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Outbox: &fakeOutbox{},
	})

	_, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected backend error, got %v", err)
	}
}

func TestListInventoriesSkipsForbiddenInventories(t *testing.T) {
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Outbox: &fakeOutbox{},
	})

	items, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected forbidden inventory to be hidden, got %+v", items)
	}
}

type fakeAuthorizer struct {
	checkInventoryErr    error
	grantTenantOwnerErr  error
	tenantOwnerGrants    []string
	inventoryOwnerGrants []string
}

func (f *fakeAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (f *fakeAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return f.checkInventoryErr
}

func (f *fakeAuthorizer) GrantTenantOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID) error {
	if f.grantTenantOwnerErr != nil {
		return f.grantTenantOwnerErr
	}
	f.tenantOwnerGrants = append(f.tenantOwnerGrants, principal.ID.String()+":"+tenantID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryOwnerGrants = append(f.inventoryOwnerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

type fakeTenantRepository struct {
	exists bool
}

func (f *fakeTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return f.exists, nil
}

type fakeInventoryRepository struct {
	items []inventory.Inventory
}

type fakeOutbox struct {
	events    []ports.AuthorizationOutboxEvent
	processed []string
	failed    []string
}

func (f *fakeOutbox) SaveTenantAndEnqueueOwnerGrant(_ context.Context, eventID string, item tenant.Tenant, principal identity.Principal) error {
	f.events = append(f.events, ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantTenantOwner,
		PrincipalID: principal.ID,
		TenantID:    item.ID,
	})
	return nil
}

func (f *fakeOutbox) SaveInventoryAndEnqueueOwnerGrant(_ context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal) error {
	f.events = append(f.events, ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
		PrincipalID: principal.ID,
		TenantID:    tenantID,
		InventoryID: item.ID,
	})
	return nil
}

func (f *fakeOutbox) ClaimPendingAuthorizationOutboxEvents(_ context.Context, claimID string, _ int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	events := make([]ports.AuthorizationOutboxEvent, 0, len(f.events))
	for index, event := range f.events {
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		f.events[index] = event
		events = append(events, event)
	}
	return events, nil
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventProcessed(_ context.Context, eventID string, claimID string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.processed = append(f.processed, eventID)
			f.events = append(f.events[:index], f.events[index+1:]...)
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventFailed(_ context.Context, eventID string, claimID string, _ string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.failed = append(f.failed, eventID)
			event.ClaimID = ""
			event.ClaimedUntil = time.Time{}
			f.events[index] = event
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeInventoryRepository) SaveInventory(context.Context, inventory.Inventory) error {
	return nil
}

func (f *fakeInventoryRepository) ListInventoriesByTenant(context.Context, inventory.TenantID) ([]inventory.Inventory, error) {
	return f.items, nil
}

type fakeObserver struct {
	events []ports.Event
}

func (f *fakeObserver) Record(_ context.Context, event ports.Event) {
	f.events = append(f.events, event)
}

func (f *fakeObserver) hasEvent(name ports.EventName) bool {
	for _, event := range f.events {
		if event.Name == name {
			return true
		}
	}
	return false
}

type fakeIDGenerator struct {
	ids []string
}

func (f *fakeIDGenerator) NewID() string {
	if len(f.ids) == 0 {
		return "fixed-id"
	}
	id := f.ids[0]
	f.ids = f.ids[1:]
	return id
}

func inventoryItem(id string, tenantID string, name string) inventory.Inventory {
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		panic("invalid test inventory name")
	}
	return inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID),
		Name:     inventoryName,
	}
}
