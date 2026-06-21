package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

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

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
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

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 1 || events[0].ID != "01ARZ3NDEKTSV4RRFFQ69G5FAX" {
		t.Fatalf("expected one outbox event from first grant, got %+v", events)
	}
}

func TestStoreDeletesInventoryAccessGrantAndEnqueuesRevoke(t *testing.T) {
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
	if _, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "grant-claim", 10, time.Now(), time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("claim initial grant event: %v", err)
	}

	event, removed, err := store.DeleteInventoryAccessGrantAndClaimRevoke(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY", "revoke-claim", time.Now().Add(time.Minute), grant, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryID, audit.ActionInventoryAccessRevoked))
	if err != nil {
		t.Fatalf("delete grant: %v", err)
	}
	if !removed {
		t.Fatalf("expected grant removal")
	}
	if event.Kind != ports.AuthorizationOutboxRevokeInventoryViewer || event.ClaimID != "revoke-claim" || event.ClaimedUntil.IsZero() {
		t.Fatalf("expected claimed revoke viewer outbox event, got %+v", event)
	}
	grants, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryID, ports.InventoryAccessGrantPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list grants after delete: %v", err)
	}
	if len(grants) != 0 {
		t.Fatalf("expected no grants after revoke, got %+v", grants)
	}

	if err := store.MarkAuthorizationOutboxEventProcessed(ctx, event.ID, event.ClaimID); err != nil {
		t.Fatalf("mark revoke processed: %v", err)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "revoke-claim-after-process", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim pending after revoke processed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected claimed revoke event not to be generally claimable after processing, got %+v", events)
	}

	event, removed, err = store.DeleteInventoryAccessGrantAndClaimRevoke(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FB0", "missing-revoke-claim", time.Now().Add(time.Minute), grant, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID, inventoryID, audit.ActionInventoryAccessRevoked))
	if err != nil {
		t.Fatalf("idempotent delete: %v", err)
	}
	if removed {
		t.Fatalf("expected missing grant delete to be idempotent")
	}
	if event.Kind != ports.AuthorizationOutboxRevokeInventoryViewer || event.ClaimID != "missing-revoke-claim" || event.PrincipalID != grant.PrincipalID {
		t.Fatalf("expected claimed idempotent revoke to enqueue stale-relationship cleanup, got %+v", event)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenantID, inventoryID, ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	revocationRecords := 0
	for _, record := range records {
		if record.Action == audit.ActionInventoryAccessRevoked {
			revocationRecords++
		}
	}
	if revocationRecords != 1 {
		t.Fatalf("expected only existing direct grant revoke to write audit, got %d records in %+v", revocationRecords, records)
	}
}

func TestStoreAcceptsInventoryAccessInvitationAtomically(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	email, _ := identity.NewEmail("viewer@example.com")
	expiresAt := time.Now().Add(time.Hour)
	invitation, err := store.SaveInventoryAccessInvitation(ctx, ports.InventoryAccessInvitation{
		ID:                 "01ARZ3NDEKTSV4RRFFQ69G5FAX",
		TenantID:           tenantID,
		InventoryID:        inventoryID,
		Email:              email,
		TokenHash:          "correct-token-hash",
		Relationship:       ports.InventoryAccessViewer,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: identity.PrincipalID("owner"),
		ExpiresAt:          expiresAt,
	}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, inventoryID, audit.ActionInventoryInvitationCreated))
	if err != nil {
		t.Fatalf("save invitation: %v", err)
	}
	if invitation.Status != ports.InventoryAccessInvitationPending || invitation.Email != email {
		t.Fatalf("unexpected invitation: %+v", invitation)
	}
	if invitation.ExpiresAt.IsZero() {
		t.Fatalf("expected invitation expiry")
	}

	_, err = store.SaveInventoryAccessInvitation(ctx, ports.InventoryAccessInvitation{
		ID:                 "01ARZ3NDEKTSV4RRFFQ69G5FB3",
		TenantID:           tenantID,
		InventoryID:        inventoryID,
		Email:              email,
		TokenHash:          "replacement-token-hash",
		Relationship:       ports.InventoryAccessViewer,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: identity.PrincipalID("owner"),
		ExpiresAt:          time.Now().Add(time.Hour),
	}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB4", tenantID, inventoryID, audit.ActionInventoryInvitationCreated))
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate pending invitation conflict, got %v", err)
	}

	_, _, err = store.AcceptInventoryAccessInvitationAndEnqueue(ctx, tenantID, inventoryID, invitation.ID, "wrong-token-hash", identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: email}, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", time.Now(), auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID, inventoryID, audit.ActionInventoryInvitationAccepted))
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected wrong token rejection, got %v", err)
	}

	accepted, grant, err := store.AcceptInventoryAccessInvitationAndEnqueue(ctx, tenantID, inventoryID, invitation.ID, "correct-token-hash", identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: email}, "01ARZ3NDEKTSV4RRFFQ69G5FB1", time.Now(), auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantID, inventoryID, audit.ActionInventoryInvitationAccepted))
	if err != nil {
		t.Fatalf("accept invitation: %v", err)
	}
	if accepted.Status != ports.InventoryAccessInvitationAccepted || accepted.AcceptedPrincipalID != identity.PrincipalID("viewer-user") {
		t.Fatalf("expected accepted invitation, got %+v", accepted)
	}
	if grant.PrincipalID != identity.PrincipalID("viewer-user") || grant.Relationship != ports.InventoryAccessViewer {
		t.Fatalf("unexpected grant: %+v", grant)
	}

	grants, err := store.ListInventoryAccessGrants(ctx, tenantID, inventoryID, ports.InventoryAccessGrantPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list grants: %v", err)
	}
	if len(grants) != 1 || grants[0] != grant {
		t.Fatalf("expected accepted grant persisted, got %+v", grants)
	}
	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-accept", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox: %v", err)
	}
	if len(events) != 1 || events[0].Kind != ports.AuthorizationOutboxGrantInventoryViewer || events[0].PrincipalID != identity.PrincipalID("viewer-user") {
		t.Fatalf("expected grant viewer outbox event, got %+v", events)
	}
}

func TestStoreRejectsRevokedInventoryAccessInvitationAcceptance(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	email, _ := identity.NewEmail("viewer@example.com")
	invitation, err := store.SaveInventoryAccessInvitation(ctx, ports.InventoryAccessInvitation{
		ID:                 "01ARZ3NDEKTSV4RRFFQ69G5FAX",
		TenantID:           tenantID,
		InventoryID:        inventoryID,
		Email:              email,
		TokenHash:          "correct-token-hash",
		Relationship:       ports.InventoryAccessViewer,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: identity.PrincipalID("owner"),
		ExpiresAt:          time.Now().Add(time.Hour),
	}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, inventoryID, audit.ActionInventoryInvitationCreated))
	if err != nil {
		t.Fatalf("save invitation: %v", err)
	}
	revoked, err := store.RevokeInventoryAccessInvitation(ctx, tenantID, inventoryID, invitation.ID, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryID, audit.ActionInventoryInvitationRevoked))
	if err != nil {
		t.Fatalf("revoke invitation: %v", err)
	}
	if !revoked {
		t.Fatalf("expected invitation revoked")
	}
	_, _, err = store.AcceptInventoryAccessInvitationAndEnqueue(ctx, tenantID, inventoryID, invitation.ID, "correct-token-hash", identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: email}, "01ARZ3NDEKTSV4RRFFQ69G5FB0", time.Now(), auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID, inventoryID, audit.ActionInventoryInvitationAccepted))
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected revoked invitation forbidden, got %v", err)
	}
}

func TestStoreRejectsExpiredInventoryAccessInvitationAcceptance(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	email, _ := identity.NewEmail("viewer@example.com")
	invitation, err := store.SaveInventoryAccessInvitation(ctx, ports.InventoryAccessInvitation{
		ID:                 "01ARZ3NDEKTSV4RRFFQ69G5FAX",
		TenantID:           tenantID,
		InventoryID:        inventoryID,
		Email:              email,
		TokenHash:          "correct-token-hash",
		Relationship:       ports.InventoryAccessViewer,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: identity.PrincipalID("owner"),
		ExpiresAt:          time.Now().Add(-time.Hour),
	}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, inventoryID, audit.ActionInventoryInvitationCreated))
	if err != nil {
		t.Fatalf("save invitation: %v", err)
	}

	_, _, err = store.AcceptInventoryAccessInvitationAndEnqueue(ctx, tenantID, inventoryID, invitation.ID, "correct-token-hash", identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: email}, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", time.Now(), auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantID, inventoryID, audit.ActionInventoryInvitationAccepted))
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected expired invitation forbidden, got %v", err)
	}
}

func TestStoreListsInventoryAccessInvitationsWithStatusFilters(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	now := time.Now()
	pendingEmail, _ := identity.NewEmail("pending@example.com")
	expiredEmail, _ := identity.NewEmail("expired@example.com")
	for _, item := range []struct {
		id        string
		auditID   string
		email     identity.Email
		expiresAt time.Time
	}{
		{id: "01ARZ3NDEKTSV4RRFFQ69G5FAX", auditID: "01ARZ3NDEKTSV4RRFFQ69G5FAY", email: pendingEmail, expiresAt: now.Add(time.Hour)},
		{id: "01ARZ3NDEKTSV4RRFFQ69G5FAZ", auditID: "01ARZ3NDEKTSV4RRFFQ69G5FB0", email: expiredEmail, expiresAt: now.Add(-time.Hour)},
	} {
		if _, err := store.SaveInventoryAccessInvitation(ctx, ports.InventoryAccessInvitation{
			ID:                 item.id,
			TenantID:           tenantID,
			InventoryID:        inventoryID,
			Email:              item.email,
			TokenHash:          item.id + "-token-hash",
			Relationship:       ports.InventoryAccessViewer,
			Status:             ports.InventoryAccessInvitationPending,
			InviterPrincipalID: identity.PrincipalID("owner"),
			ExpiresAt:          item.expiresAt,
		}, auditRecord(t, item.auditID, tenantID, inventoryID, audit.ActionInventoryInvitationCreated)); err != nil {
			t.Fatalf("save invitation %s: %v", item.id, err)
		}
	}

	all, err := store.ListInventoryAccessInvitations(ctx, tenantID, inventoryID, ports.InventoryAccessInvitationPageRequest{
		Limit:        10,
		StatusFilter: ports.InventoryAccessInvitationStatusFilterAll,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("list all invitations: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected all invitations, got %+v", all)
	}

	pending, err := store.ListInventoryAccessInvitations(ctx, tenantID, inventoryID, ports.InventoryAccessInvitationPageRequest{
		Limit:        10,
		StatusFilter: ports.InventoryAccessInvitationStatusFilterPending,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("list pending invitations: %v", err)
	}
	if len(pending) != 1 || pending[0].Email != pendingEmail {
		t.Fatalf("expected only unexpired pending invitation, got %+v", pending)
	}

	expired, err := store.ListInventoryAccessInvitations(ctx, tenantID, inventoryID, ports.InventoryAccessInvitationPageRequest{
		Limit:        10,
		StatusFilter: ports.InventoryAccessInvitationStatusFilterExpired,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("list expired invitations: %v", err)
	}
	if len(expired) != 1 || expired[0].Email != expiredEmail {
		t.Fatalf("expected only expired pending invitation, got %+v", expired)
	}
}

func TestStoreUpdatesPendingInventoryAccessInvitationExpirationAndAudits(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	email, _ := identity.NewEmail("viewer@example.com")
	invitation, err := store.SaveInventoryAccessInvitation(ctx, ports.InventoryAccessInvitation{
		ID:                 "01ARZ3NDEKTSV4RRFFQ69G5FAX",
		TenantID:           tenantID,
		InventoryID:        inventoryID,
		Email:              email,
		TokenHash:          "correct-token-hash",
		Relationship:       ports.InventoryAccessViewer,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: identity.PrincipalID("owner"),
		ExpiresAt:          time.Now().Add(time.Hour),
	}, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID, inventoryID, audit.ActionInventoryInvitationCreated))
	if err != nil {
		t.Fatalf("save invitation: %v", err)
	}

	newExpiration := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	updated, ok, err := store.UpdateInventoryAccessInvitationExpiration(ctx, tenantID, inventoryID, invitation.ID, newExpiration, auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FAZ", tenantID, inventoryID, audit.ActionInventoryInvitationExpirationUpdated))
	if err != nil {
		t.Fatalf("update invitation expiration: %v", err)
	}
	if !ok || !updated.ExpiresAt.Equal(newExpiration) || !updated.IsExpired(time.Now()) {
		t.Fatalf("expected expired invitation after update, got ok=%v invitation=%+v", ok, updated)
	}

	records, err := store.ListInventoryAuditRecords(ctx, tenantID, inventoryID, ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	found := false
	for _, record := range records {
		if record.Action == audit.ActionInventoryInvitationExpirationUpdated {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected expiration update audit record, got %+v", records)
	}

	_, _, err = store.AcceptInventoryAccessInvitationAndEnqueue(ctx, tenantID, inventoryID, invitation.ID, "correct-token-hash", identity.Principal{ID: identity.PrincipalID("viewer-user"), Email: email}, "01ARZ3NDEKTSV4RRFFQ69G5FB0", time.Now(), auditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantID, inventoryID, audit.ActionInventoryInvitationAccepted))
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected manually expired invitation acceptance forbidden, got %v", err)
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

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no outbox event for rejected grant, got %+v", events)
	}
}
