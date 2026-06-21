package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCreateTenantEnqueuesAndDrainsOwnerGrant(t *testing.T) {
	authorizer := &fakeAuthorizer{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:            &fakeObserver{},
		Authorizer:          authorizer,
		Tenants:             &fakeTenantRepository{},
		TenantUnitOfWork:    &fakeTenantRepository{},
		Inventories:         &fakeInventoryRepository{},
		InventoryUnitOfWork: &fakeInventoryRepository{},
		Audit:               &fakeAuditRepository{},
		Outbox:              outbox,
		IDs:                 &fakeIDGenerator{ids: []string{"tenant-one", "audit-one", "event-one"}},
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
		Inventories:         &fakeInventoryRepository{},
		InventoryUnitOfWork: &fakeInventoryRepository{},
		Audit:               &fakeAuditRepository{},
		Outbox:              outbox,
		IDs:                 &fakeIDGenerator{ids: []string{"inventory-one", "audit-one", "event-one"}},
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
		Observer:            observer,
		Authorizer:          authorizer,
		Tenants:             &fakeTenantRepository{},
		TenantUnitOfWork:    &fakeTenantRepository{},
		Inventories:         &fakeInventoryRepository{},
		InventoryUnitOfWork: &fakeInventoryRepository{},
		Audit:               &fakeAuditRepository{},
		Outbox:              outbox,
		IDs:                 &fakeIDGenerator{ids: []string{"tenant-one", "audit-one", "event-one"}},
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
		Observer:            &fakeObserver{},
		Authorizer:          authorizer,
		Tenants:             &fakeTenantRepository{},
		TenantUnitOfWork:    &fakeTenantRepository{},
		Inventories:         &fakeInventoryRepository{},
		InventoryUnitOfWork: &fakeInventoryRepository{},
		Audit:               &fakeAuditRepository{},
		Outbox:              outbox,
		IDs:                 &fakeIDGenerator{},
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

func TestDrainAuthorizationOutboxDeadLettersUnrecoverableEventAndContinues(t *testing.T) {
	observer := &fakeObserver{}
	outbox := &fakeOutbox{
		events: []ports.AuthorizationOutboxEvent{
			{
				ID:          "bad-event",
				Kind:        ports.AuthorizationOutboxEventKind("unknown"),
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
		Observer:            observer,
		Authorizer:          &fakeAuthorizer{},
		Tenants:             &fakeTenantRepository{},
		TenantUnitOfWork:    &fakeTenantRepository{},
		Inventories:         &fakeInventoryRepository{},
		InventoryUnitOfWork: &fakeInventoryRepository{},
		Audit:               &fakeAuditRepository{},
		Outbox:              outbox,
		IDs:                 &fakeIDGenerator{},
	})

	if err := application.DrainAuthorizationOutbox(context.Background(), 10); err != nil {
		t.Fatalf("dead-lettering unrecoverable events should not fail the batch: %v", err)
	}
	if !slices.Contains(outbox.deadLettered, "bad-event") {
		t.Fatalf("expected bad event to be dead-lettered, got %+v", outbox.deadLettered)
	}
	if !slices.Contains(outbox.processed, "inventory-event") {
		t.Fatalf("expected inventory event to process after dead-letter, got %+v", outbox.processed)
	}
	if !observer.hasEvent(ports.EventAuthorizationOutboxDeadLettered) {
		t.Fatalf("expected outbox dead-letter observability event, got %+v", observer.events)
	}

	events, err := outbox.ClaimPendingAuthorizationOutboxEvents(context.Background(), "claim-two", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected dead-lettered event to stay out of pending claims, got %+v", events)
	}
}

func TestDrainAuthorizationOutboxDeadLettersDurableInvalidEvent(t *testing.T) {
	ctx := context.Background()
	store := newAppTestGORMStore(t, ctx)
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "event-one", tenant.Tenant{
		ID:   tenant.ID("tenant-one"),
		Name: tenantName,
	}, identity.Principal{}, auditRecord("audit-one", "tenant-one", "", audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save tenant and enqueue invalid owner grant: %v", err)
	}

	observer := &fakeObserver{}
	authorizer := &fakeAuthorizer{}
	application := New(Dependencies{
		Observer:            observer,
		Authorizer:          authorizer,
		Tenants:             store,
		TenantUnitOfWork:    store,
		Inventories:         store,
		InventoryUnitOfWork: store,
		Audit:               store,
		Outbox:              store,
		IDs:                 &fakeIDGenerator{ids: []string{"claim-one"}},
	})

	if err := application.DrainAuthorizationOutbox(ctx, 10); err != nil {
		t.Fatalf("dead-lettering durable invalid events should not fail the batch: %v", err)
	}
	if len(authorizer.tenantOwnerGrants) != 0 {
		t.Fatalf("expected invalid event not to reach authorizer, got %+v", authorizer.tenantOwnerGrants)
	}
	if !observer.hasEvent(ports.EventAuthorizationOutboxDeadLettered) {
		t.Fatalf("expected outbox dead-letter observability event, got %+v", observer.events)
	}
	deadLetterEvent, ok := observer.eventNamed(ports.EventAuthorizationOutboxDeadLettered)
	if !ok || !strings.Contains(deadLetterEvent.Fields["reason"], "principal id") {
		t.Fatalf("expected actionable dead-letter reason, got %+v", deadLetterEvent)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected durable dead-lettered event to stay out of pending claims, got %+v", events)
	}
}
