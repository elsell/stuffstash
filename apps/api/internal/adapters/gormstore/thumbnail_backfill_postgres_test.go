package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPostgresThumbnailConcurrentBackfillAndClaims(t *testing.T) {
	store, now := postgresBackfillFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var batches [2]ports.ThumbnailBackfillProgress
	var failures [2]error
	var wait sync.WaitGroup
	for i := range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			batches[i], failures[i] = NewStore(store.db).BackfillThumbnailJobs(ctx, 1, now)
		}()
	}
	wait.Wait()
	for i := range 2 {
		if failures[i] != nil || batches[i].Scanned != 1 || batches[i].Enqueued != 1 {
			t.Fatal("concurrent backfill lost progress", batches, failures)
		}
	}
	if batches[0].Cursor == batches[1].Cursor {
		t.Fatal("concurrent scans repeated cursor")
	}
	var claims [2][]ports.ClaimedThumbnailJob
	for i := range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			claims[i], failures[i] = NewStore(store.db).ClaimThumbnailJobs(ctx, []string{"worker-a", "worker-b"}[i], 1, now, now.Add(time.Minute))
		}()
	}
	wait.Wait()
	for i := range 2 {
		if failures[i] != nil || len(claims[i]) != 1 {
			t.Fatal("concurrent claim failed", failures)
		}
	}
	if claims[0][0].Job.AttachmentID == claims[1][0].Job.AttachmentID {
		t.Fatal("two workers claimed the same image")
	}
}

func TestPostgresThumbnailBackfillFailureRollsBackJobsAndCursor(t *testing.T) {
	store, now := postgresBackfillFixture(t)
	ctx := context.Background()
	if err := store.db.Model(&attachmentModel{ID: "backfill-image-b"}).UpdateColumn("file_name", "").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := store.BackfillThumbnailJobs(ctx, 2, now); err == nil {
		t.Fatal("invalid attachment did not fail batch")
	}
	var jobs int64
	if err := store.db.Model(&thumbnailJobModel{}).Count(&jobs).Error; err != nil {
		t.Fatal(err)
	}
	if jobs != 0 {
		t.Fatal("failed batch committed derivative jobs")
	}
	if err := store.db.Model(&attachmentModel{ID: "backfill-image-b"}).UpdateColumn("file_name", "photo.jpg").Error; err != nil {
		t.Fatal(err)
	}
	progress, err := NewStore(store.db).BackfillThumbnailJobs(ctx, 2, now)
	if err != nil || progress.Scanned != 2 || progress.Enqueued != 2 {
		t.Fatal("rollback advanced cursor", progress, err)
	}
}

func postgresBackfillFixture(t *testing.T) (Store, time.Time) {
	t.Helper()
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
	pool.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = pool.Close() })
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	store := NewStore(db)
	ctx := context.Background()
	tenantID := tenant.ID("backfill-tenant")
	inventoryID := inventory.InventoryID("backfill-inventory")
	cleanupSearchTestRows(t, ctx, store, tenantID)
	clearCursor := func() {
		if err := store.db.Delete(&thumbnailBackfillModel{Revision: int(media.CurrentThumbnailRevision)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	clearCursor()
	t.Cleanup(func() { cleanupSearchTestRows(t, ctx, store, tenantID); clearCursor() })
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Photos")
	item := assetItem("backfill-asset", tenantID.String(), inventoryID.String(), asset.KindItem, "")
	if err := store.CreateAsset(ctx, item, postgresAuditRecord(t, "backfill-create-asset", tenantID, inventoryID, audit.ActionAssetCreated), nil); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	for _, id := range []string{"backfill-image-a", "backfill-image-b"} {
		attachment := testAttachment(t, id, item, "photo.jpg", media.ContentTypeJPEG, now)
		if err := store.SaveAttachment(ctx, attachment, postgresAuditRecord(t, id+"-audit", tenantID, inventoryID, audit.ActionAttachmentCreated), plannedThumbnailJob(t, attachment)); err != nil {
			t.Fatal(err)
		}
		if err := store.db.Delete(&thumbnailJobModel{AttachmentID: id, Revision: int(media.CurrentThumbnailRevision)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return store, now
}
