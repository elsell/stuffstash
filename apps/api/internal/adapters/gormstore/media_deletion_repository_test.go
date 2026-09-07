package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestDeletedAttachmentStorageKeyCannotBeReused(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	tenantID, inventoryID, assetID := tenant.ID(attachment.TenantID), inventory.InventoryID(attachment.InventoryID), asset.ID(attachment.AssetID)
	_, found, err := store.DeleteAttachmentAndEnqueueBlobDeletion(ctx, "delete-original", tenantID, inventoryID, assetID, attachment.ID, auditRecord(t, "deleted", tenantID, inventoryID, audit.ActionAttachmentDeleted))
	if err != nil || !found {
		t.Fatal("delete failed", err)
	}
	attachment.ID = "replacement"
	if err := store.SaveAttachment(ctx, attachment, auditRecord(t, "replacement", tenantID, inventoryID, audit.ActionAttachmentCreated), plannedThumbnailJob(t, attachment)); !errors.Is(err, ports.ErrConflict) {
		t.Fatal("retired key reused", err)
	}
}

func TestAssetDeletionEnqueuesAllMediaAndRemovesJobs(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	tenantID, inventoryID, assetID := tenant.ID(attachment.TenantID), inventory.InventoryID(attachment.InventoryID), asset.ID(attachment.AssetID)
	if err := store.DeleteAsset(ctx, tenantID, inventoryID, assetID, auditRecord(t, "asset-deleted", tenantID, inventoryID, audit.ActionAssetDeleted)); err != nil {
		t.Fatal(err)
	}
	_, found, err := store.AttachmentByID(ctx, tenantID, inventoryID, assetID, attachment.ID)
	if err != nil || found {
		t.Fatal("asset deletion retained attachment", err)
	}
	now := time.Now()
	events, err := store.ClaimPendingBlobDeletionEvents(ctx, "cleanup", 10, now, now.Add(time.Minute))
	if err != nil || len(events) != 1 || events[0].StorageKey != attachment.StorageKey {
		t.Fatal("asset deletion lost blob cleanup", events, err)
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker", 10, now, now.Add(time.Minute))
	if err != nil || len(claims) != 0 {
		t.Fatal("deleted image remained queued", claims, err)
	}
}
