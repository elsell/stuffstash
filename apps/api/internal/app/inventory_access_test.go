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
	"time"
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
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"event-one"}},
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
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"event-two"}},
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
		Observer:                  observer,
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"event-one", "event-two"}},
		MaxPageLimit:              1,
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
		Observer:                  observer,
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    outbox,
		IDs:                       &fakeIDGenerator{ids: []string{"audit-revoke", "event-revoke", "claim-revoke", "audit-missing", "event-missing", "claim-missing"}},
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
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"audit-unauthorized", "event-unauthorized"}},
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
		TenantUnitOfWork:              &fakeTenantRepository{exists: true},
		Inventories:                   repository,
		InventoryUnitOfWork:           repository,
		InventoryAccess:               repository,
		InventoryAccessUnitOfWork:     repository,
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

func TestInventoryAccessInvitationCreateAcceptAndRevoke(t *testing.T) {
	outbox := &fakeOutbox{}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		outbox: outbox,
	}
	observer := &fakeObserver{}
	now := time.Date(2030, 6, 1, 12, 30, 0, 0, time.UTC)
	application := New(Dependencies{
		Observer:                  observer,
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    outbox,
		IDs:                       &fakeIDGenerator{ids: []string{"invite-one", "audit-invite", "audit-accept", "grant-event", "grant-claim", "invite-two", "audit-invite-two", "audit-revoke"}},
		Clock:                     fakeClock{now: now},
	})

	inviteResult, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Email:        "Viewer@Example.COM",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	invitation := inviteResult.Invitation
	if invitation.Email.String() != "viewer@example.com" || invitation.Status != ports.InventoryAccessInvitationPending {
		t.Fatalf("unexpected invitation: %+v", invitation)
	}
	if inviteResult.AcceptanceToken == "" || invitation.TokenHash != hashInventoryInvitationToken(inviteResult.AcceptanceToken) {
		t.Fatalf("expected invitation response to include a matching one-time acceptance token")
	}
	if !invitation.ExpiresAt.Equal(now.Add(7 * 24 * time.Hour)) {
		t.Fatalf("expected pending invitation expiration from injected clock, got %s", invitation.ExpiresAt)
	}
	if len(repository.auditRecords) != 1 || !repository.auditRecords[0].OccurredAt.Equal(now) {
		t.Fatalf("expected invitation audit timestamp from injected clock, got %+v", repository.auditRecords)
	}

	accepted, grant, err := application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: invitation.Email},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: invitation.ID,
		Token:        inviteResult.AcceptanceToken,
	})
	if err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	if accepted.Status != ports.InventoryAccessInvitationAccepted || accepted.AcceptedPrincipalID != identity.PrincipalID("viewer-user") {
		t.Fatalf("expected accepted invitation, got %+v", accepted)
	}
	if grant.PrincipalID != identity.PrincipalID("viewer-user") || grant.Relationship != ports.InventoryAccessViewer {
		t.Fatalf("unexpected grant: %+v", grant)
	}
	if len(repository.auditRecords) != 2 {
		t.Fatalf("expected invite create/accept audits, got %+v", repository.auditRecords)
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "grant-event" {
		t.Fatalf("expected accepted invitation grant outbox processed, got %+v", outbox.processed)
	}
	if !observer.hasEvent(ports.EventInventoryInvitationCreated) || !observer.hasEvent(ports.EventInventoryInvitationAccepted) {
		t.Fatalf("expected invitation created and accepted observability events, got %+v", observer.events)
	}

	revokedInviteResult, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Email:        "editor@example.com",
		Relationship: "editor",
	})
	if err != nil {
		t.Fatalf("create invitation to revoke: %v", err)
	}
	revokedInvite := revokedInviteResult.Invitation
	revoked, err := application.RevokeInventoryAccessInvitation(context.Background(), RevokeInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: revokedInvite.ID,
	})
	if err != nil {
		t.Fatalf("revoke invitation: %v", err)
	}
	if !revoked {
		t.Fatalf("expected pending invitation revoked")
	}
	if !observer.hasEvent(ports.EventInventoryInvitationRevoked) {
		t.Fatalf("expected invitation revoked observability event, got %+v", observer.events)
	}
	_, _, err = application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("editor-user"), Email: identity.Email("editor@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: revokedInvite.ID,
		Token:        revokedInviteResult.AcceptanceToken,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected revoked invitation rejection, got %v", err)
	}
}

func TestCreateInventoryAccessInvitationRejectsDuplicatePendingInvitation(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		invitations: []ports.InventoryAccessInvitation{
			{
				ID:           "invite-one",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("viewer@example.com"),
				TokenHash:    hashInventoryInvitationToken("old-token"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    time.Now().Add(time.Hour),
			},
		},
	}
	application := New(Dependencies{
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    &fakeOutbox{},
		IDs:                       &fakeIDGenerator{ids: []string{"invite-two", "audit-invite-two"}},
	})

	_, err := application.CreateInventoryAccessInvitation(context.Background(), CreateInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Email:        "viewer@example.com",
		Relationship: "viewer",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected duplicate pending invitation rejection, got %v", err)
	}
}

func TestAcceptInventoryAccessInvitationRejectsMissingOrWrongEmail(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
		invitations: []ports.InventoryAccessInvitation{
			{
				ID:           "invite-one",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("viewer@example.com"),
				TokenHash:    hashInventoryInvitationToken("correct-token"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			{
				ID:           "expired-invite",
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				Email:        identity.Email("expired@example.com"),
				TokenHash:    hashInventoryInvitationToken("expired-token"),
				Relationship: ports.InventoryAccessViewer,
				Status:       ports.InventoryAccessInvitationPending,
				ExpiresAt:    time.Now().Add(-time.Hour),
			},
		},
		outbox: &fakeOutbox{},
	}
	application := New(Dependencies{
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{checkInventoryErr: ports.ErrForbidden},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               repository,
		InventoryUnitOfWork:       repository,
		InventoryAccess:           repository,
		InventoryAccessUnitOfWork: repository,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    repository.outbox,
		IDs:                       &fakeIDGenerator{ids: []string{"audit-missing", "event-missing", "audit-wrong", "event-wrong"}},
	})

	for _, item := range []struct {
		name      string
		principal identity.Principal
		token     string
	}{
		{name: "missing email", principal: identity.Principal{ID: identity.PrincipalID("viewer-user")}, token: "correct-token"},
		{name: "wrong email", principal: identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("wrong@example.com")}, token: "correct-token"},
		{name: "missing token", principal: identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")}},
		{name: "wrong token", principal: identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: identity.Email("viewer@example.com")}, token: "wrong-token"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, _, err := application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
				Principal:    item.principal,
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				InvitationID: "invite-one",
				Token:        item.token,
			})
			if !errors.Is(err, ErrUnauthorized) {
				t.Fatalf("expected unauthorized, got %v", err)
			}
		})
	}

	_, _, err := application.AcceptInventoryAccessInvitation(context.Background(), AcceptInventoryAccessInvitationInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("expired-user"), Email: identity.Email("expired@example.com")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		InvitationID: "expired-invite",
		Token:        "expired-token",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected expired invitation rejection, got %v", err)
	}
}
