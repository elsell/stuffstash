package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestStorePersistsAndScopesAuditRecords(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)

	tenantOne := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	tenantTwo := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	inventoryOne := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAX")
	inventoryTwo := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAY")
	inventoryThree := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAZ")
	saveTenant(t, ctx, store, tenantOne, "Home")
	saveTenant(t, ctx, store, tenantTwo, "Cabin")
	saveInventory(t, ctx, store, inventoryOne.String(), tenantOne, "Tools")
	saveInventory(t, ctx, store, inventoryTwo.String(), tenantOne, "Supplies")
	saveInventory(t, ctx, store, inventoryThree.String(), tenantTwo, "Cabin Tools")
	occurredAt := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)

	for _, record := range []audit.Record{
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB1", tenantOne, inventoryOne, audit.ActionAssetUpdated, occurredAt),
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB0", tenantOne, inventoryOne, audit.ActionAssetCreated, occurredAt),
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB2", tenantOne, inventoryTwo, audit.ActionAssetMoved, occurredAt.Add(time.Second)),
		auditRecordAt(t, "01ARZ3NDEKTSV4RRFFQ69G5FB3", tenantTwo, inventoryThree, audit.ActionAssetCreated, occurredAt.Add(2*time.Second)),
	} {
		if err := store.SaveAuditRecord(ctx, record); err != nil {
			t.Fatalf("save audit record %s: %v", record.ID, err)
		}
	}

	firstPage, err := store.ListInventoryAuditRecords(ctx, tenantOne, inventoryOne, ports.AuditRecordPageRequest{Limit: 1})
	if err != nil {
		t.Fatalf("list inventory audit records: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].ID != audit.ID("01ARZ3NDEKTSV4RRFFQ69G5FB0") || firstPage[0].Metadata["note"] != "safe" {
		t.Fatalf("unexpected first audit page: %+v", firstPage)
	}
	secondPage, err := store.ListInventoryAuditRecords(ctx, tenantOne, inventoryOne, ports.AuditRecordPageRequest{
		AfterOccurredAt: firstPage[0].OccurredAt,
		AfterRecordID:   firstPage[0].ID,
		Limit:           10,
	})
	if err != nil {
		t.Fatalf("list second audit page: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].ID != audit.ID("01ARZ3NDEKTSV4RRFFQ69G5FB1") {
		t.Fatalf("unexpected second audit page: %+v", secondPage)
	}

	tenantPage, err := store.ListTenantAuditRecords(ctx, tenantOne, ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list tenant audit records: %v", err)
	}
	if len(tenantPage) != 3 {
		t.Fatalf("expected tenant page to include only tenant one records, got %+v", tenantPage)
	}
}

func TestStoreRollsBackTenantAndOutboxWhenAuditInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	existingTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	newTenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	auditID := "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	saveTenant(t, ctx, store, existingTenantID, "Existing")
	if err := store.SaveAuditRecord(ctx, auditRecord(t, auditID, existingTenantID, "", audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save existing audit record: %v", err)
	}

	tenantName, ok := tenant.NewName("Rollback Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenant.Tenant{
		ID:   newTenantID,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, auditID, newTenantID, "", audit.ActionTenantCreated))
	if err == nil {
		t.Fatalf("expected duplicate audit ID to fail")
	}

	exists, err := store.TenantExists(ctx, newTenantID)
	if err != nil {
		t.Fatalf("check tenant exists: %v", err)
	}
	if exists {
		t.Fatalf("expected tenant write to roll back after audit failure")
	}
	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-one", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected outbox write to roll back after audit failure, got %+v", events)
	}
}

func TestStoreRollsBackAssetWhenAuditInsertFails(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	auditID := "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")
	if err := store.SaveAuditRecord(ctx, auditRecord(t, auditID, tenantID, inventoryID, audit.ActionAssetCreated)); err != nil {
		t.Fatalf("save existing audit record: %v", err)
	}

	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	err := store.CreateAsset(ctx, item, auditRecord(t, auditID, tenantID, inventoryID, audit.ActionAssetCreated), nil)
	if err == nil {
		t.Fatalf("expected duplicate audit ID to fail")
	}
	_, found, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}
	if found {
		t.Fatalf("expected asset write to roll back after audit failure")
	}
}
