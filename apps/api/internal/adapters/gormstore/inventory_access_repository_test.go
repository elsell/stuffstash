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
	if _, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "grant-claim", 10, time.Now().Add(time.Minute)); err != nil {
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

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "revoke-claim-after-process", 10, time.Now().Add(time.Minute))
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
