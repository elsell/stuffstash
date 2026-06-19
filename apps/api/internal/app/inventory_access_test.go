package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestGrantInventoryAccessRequiresShareAndRejectsInvalidGrants(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
	}
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: repository,
		Audit:       &fakeAuditRepository{},
		Outbox:      &fakeOutbox{},
		IDs:         &fakeIDGenerator{ids: []string{"event-one"}},
	})

	_, err := application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized without share permission, got %v", err)
	}
	if len(repository.accessGrants) != 0 {
		t.Fatalf("expected no durable grant without share permission, got %+v", repository.accessGrants)
	}

	allowed := New(Dependencies{
		Observer:    &fakeObserver{},
		Authorizer:  &fakeAuthorizer{},
		Tenants:     &fakeTenantRepository{exists: true},
		Inventories: repository,
		Audit:       &fakeAuditRepository{},
		Outbox:      &fakeOutbox{},
		IDs:         &fakeIDGenerator{ids: []string{"event-two"}},
	})
	for _, item := range []struct {
		name          string
		targetUserID  string
		relationship  string
		expectedError error
	}{
		{name: "self grant", targetUserID: "owner", relationship: "viewer", expectedError: ErrInvalidInput},
		{name: "bad principal", targetUserID: "user/one", relationship: "viewer", expectedError: ErrInvalidInput},
		{name: "bad relationship", targetUserID: "viewer", relationship: "owner", expectedError: ErrInvalidInput},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := allowed.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
				Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				TargetUserID: item.targetUserID,
				Relationship: item.relationship,
			})
			if !errors.Is(err, item.expectedError) {
				t.Fatalf("expected %v, got %v", item.expectedError, err)
			}
		})
	}
}

func TestGrantAndListInventoryAccessGrants(t *testing.T) {
	observer := &fakeObserver{}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
	}
	application := New(Dependencies{
		Observer:     observer,
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  repository,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"event-one", "event-two"}},
		MaxPageLimit: 1,
	})

	_, err := application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("grant viewer: %v", err)
	}
	_, err = application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "editor",
		Relationship: "editor",
	})
	if err != nil {
		t.Fatalf("grant editor: %v", err)
	}

	firstPage, err := application.ListInventoryAccessGrants(context.Background(), ListInventoryAccessGrantsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].PrincipalID != identity.PrincipalID("editor") || !firstPage.HasMore || firstPage.NextCursor == nil {
		t.Fatalf("expected first grant page with editor, got %+v", firstPage)
	}

	secondPage, err := application.ListInventoryAccessGrants(context.Background(), ListInventoryAccessGrantsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
		Cursor:      *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].PrincipalID != identity.PrincipalID("viewer") || secondPage.HasMore {
		t.Fatalf("expected second grant page with viewer, got %+v", secondPage)
	}
	if !observer.hasEvent(ports.EventInventoryAccessGranted) || !observer.hasEvent(ports.EventInventoryAccessListed) {
		t.Fatalf("expected grant/list observability events, got %+v", observer.events)
	}
}
