package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strconv"
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

func TestRevokeInventoryAccessRequiresShareAndIsIdempotent(t *testing.T) {
	observer := &fakeObserver{}
	outbox := &fakeOutbox{}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		accessGrants: []ports.InventoryAccessGrant{
			{
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				PrincipalID:  identity.PrincipalID("viewer"),
				Relationship: ports.InventoryAccessViewer,
			},
		},
		outbox: outbox,
	}
	application := New(Dependencies{
		Observer:    observer,
		Authorizer:  &fakeAuthorizer{},
		Tenants:     &fakeTenantRepository{exists: true},
		Inventories: repository,
		Audit:       &fakeAuditRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"audit-revoke", "event-revoke", "claim-revoke", "audit-missing", "event-missing", "claim-missing"}},
	})

	removed, err := application.RevokeInventoryAccess(context.Background(), RevokeInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("revoke viewer: %v", err)
	}
	if !removed || len(repository.accessGrants) != 0 {
		t.Fatalf("expected direct grant removed, removed=%t grants=%+v", removed, repository.accessGrants)
	}
	if len(repository.auditRecords) != 1 || repository.auditRecords[0].Action != audit.ActionInventoryAccessRevoked {
		t.Fatalf("expected revocation audit record, got %+v", repository.auditRecords)
	}
	if !observer.hasEvent(ports.EventInventoryAccessRevoked) {
		t.Fatalf("expected revocation observability event, got %+v", observer.events)
	}

	removed, err = application.RevokeInventoryAccess(context.Background(), RevokeInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("idempotent revoke: %v", err)
	}
	if removed || len(repository.auditRecords) != 1 {
		t.Fatalf("expected missing revoke to be idempotent without audit, removed=%t audit=%+v", removed, repository.auditRecords)
	}

	unauthorized := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants:     &fakeTenantRepository{exists: true},
		Inventories: repository,
		Audit:       &fakeAuditRepository{},
		Outbox:      &fakeOutbox{},
		IDs:         &fakeIDGenerator{ids: []string{"audit-unauthorized", "event-unauthorized"}},
	})
	_, err = unauthorized.RevokeInventoryAccess(context.Background(), RevokeInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "owner",
		Relationship: "viewer",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized revoke rejection, got %v", err)
	}
}

func TestRevokeInventoryAccessProcessesItsOwnOutboxEventWhenBacklogExists(t *testing.T) {
	outbox := &fakeOutbox{}
	for index := 0; index < 30; index++ {
		outbox.events = append(outbox.events, ports.AuthorizationOutboxEvent{
			ID:          "older-event-" + strconv.Itoa(index),
			Kind:        ports.AuthorizationOutboxGrantTenantOwner,
			PrincipalID: identity.PrincipalID("other"),
			TenantID:    tenant.ID("tenant-one"),
		})
	}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		accessGrants: []ports.InventoryAccessGrant{
			{
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				PrincipalID:  identity.PrincipalID("viewer"),
				Relationship: ports.InventoryAccessViewer,
			},
		},
		outbox: outbox,
	}
	application := New(Dependencies{
		Observer:                      &fakeObserver{},
		Authorizer:                    &fakeAuthorizer{},
		Tenants:                       &fakeTenantRepository{exists: true},
		Inventories:                   repository,
		Audit:                         &fakeAuditRepository{},
		Outbox:                        outbox,
		IDs:                           &fakeIDGenerator{ids: []string{"audit-revoke", "event-revoke", "claim-revoke"}},
		AuthorizationOutboxDrainLimit: 1,
	})

	removed, err := application.RevokeInventoryAccess(context.Background(), RevokeInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("revoke with backlog: %v", err)
	}
	if !removed {
		t.Fatalf("expected direct grant removal")
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "event-revoke" {
		t.Fatalf("expected targeted revoke event to be processed, got processed=%+v remaining=%+v", outbox.processed, outbox.events)
	}
	if len(outbox.events) != 30 {
		t.Fatalf("expected unrelated backlog to remain untouched, got %+v", outbox.events)
	}
}
