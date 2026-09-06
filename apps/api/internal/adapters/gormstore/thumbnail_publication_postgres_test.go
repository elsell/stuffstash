package gormstore

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestPostgresThumbnailCancellationRetainsPublicationLock(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db)
	ctx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	tenantID := tenant.ID("thumbnail-publication-test")
	inventoryID := inventory.InventoryID("thumbnail-publication-inv")
	cleanupSearchTestRows(t, ctx, store, tenantID)
	t.Cleanup(func() { cleanupSearchTestRows(t, context.Background(), store, tenantID) })
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Photos")
	item := assetItem("thumbnail-publication-item", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	attachment := testAttachment(t, "thumbnail-publication-img", item, "photo.jpg", media.ContentTypeJPEG, now)
	if err := store.SaveAttachment(ctx, attachment, auditRecord(t, "thumbnail-publish-create", tenantID, inventoryID, audit.ActionAttachmentCreated), plannedThumbnailJob(t, attachment)); err != nil {
		t.Fatal(err)
	}
	guard, err := NewThumbnailPublicationGuard(store, &publicationClock{now: now})
	if err != nil {
		t.Fatal(err)
	}
	caller, cancel := context.WithCancel(ctx)
	defer cancel()
	entered, cancelled, finish := make(chan struct{}), make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- guard.Publish(caller, attachment, nil, func(writeCtx context.Context) error {
			close(entered)
			<-writeCtx.Done()
			close(cancelled)
			select {
			case <-finish:
			case <-ctx.Done():
			}
			return writeCtx.Err()
		})
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("publisher did not enter")
	}
	cancel()
	select {
	case <-cancelled:
	case <-ctx.Done():
		t.Fatal("publisher did not receive cancellation")
	}
	// A separate connection must remain unable to delete while the cancelled
	// publisher is deliberately still unwinding its in-flight write.
	deleting, stopDelete := context.WithTimeout(ctx, 200*time.Millisecond)
	_, _, deleteErr := store.DeleteAttachmentAndEnqueueBlobDeletion(deleting, "thumbnail-publish-delete", tenantID, inventoryID, item.ID, attachment.ID, auditRecord(t, "thumbnail-publish-audit", tenantID, inventoryID, audit.ActionAttachmentDeleted))
	deadlineReached := errors.Is(deleting.Err(), context.DeadlineExceeded)
	stopDelete()
	close(finish)
	if deleteErr == nil || !deadlineReached {
		t.Fatal("deletion overtook a cancelled publisher still writing")
	}
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("publication cancellation was hidden")
		}
	case <-ctx.Done():
		t.Fatal("publisher failed to unwind")
	}
	_, found, err := store.DeleteAttachmentAndEnqueueBlobDeletion(ctx, "thumbnail-publish-delete", tenantID, inventoryID, item.ID, attachment.ID, auditRecord(t, "thumbnail-publish-audit", tenantID, inventoryID, audit.ActionAttachmentDeleted))
	if err != nil || !found {
		t.Fatalf("deletion did not proceed after publication returned: %v", err)
	}
}
