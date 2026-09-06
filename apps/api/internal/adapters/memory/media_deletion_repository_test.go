package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestAssetDeletionQueuesMediaAndRetiresItsKey(t *testing.T) {
	store, attachment, job := memoryPublicationFixture(t)
	delete(store.attachments, attachment.ID)
	delete(store.thumbnailJobs, thumbnailKey(job))
	item := asset.Asset{ID: asset.ID(attachment.AssetID), TenantID: asset.TenantID(attachment.TenantID), InventoryID: asset.InventoryID(attachment.InventoryID)}
	store.assets[item.ID] = item
	ctx := context.Background()
	if err := store.SaveAttachment(ctx, attachment, audit.Record{ID: "created", OccurredAt: job.CreatedAt}, &job); err != nil {
		t.Fatal(err)
	}
	tenantID, inventoryID := tenant.ID(attachment.TenantID), inventory.InventoryID(attachment.InventoryID)
	if err := store.DeleteAsset(ctx, tenantID, inventoryID, item.ID, audit.Record{ID: "deleted", OccurredAt: job.CreatedAt}); err != nil {
		t.Fatal(err)
	}
	if _, exists := store.attachments[attachment.ID]; exists {
		t.Fatal("deleted asset kept image metadata")
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(claims) != 0 {
		t.Fatal("deleted image remained queued")
	}
	events, err := store.ClaimPendingBlobDeletionEvents(ctx, "cleanup", 1, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(events) != 1 || events[0].StorageKey != attachment.StorageKey {
		t.Fatal("blob cleanup lost")
	}
	store.assets[item.ID] = item
	attachment.ID = "replacement"
	replacement, err := media.PlanThumbnailJob(attachment)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAttachment(ctx, attachment, audit.Record{ID: "replacement"}, replacement); !errors.Is(err, ports.ErrConflict) {
		t.Fatal("retired key reused", err)
	}
}
