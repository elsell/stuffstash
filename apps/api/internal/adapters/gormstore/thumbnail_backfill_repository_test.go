package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"testing"
	"time"
)

func TestThumbnailBackfillResumesAndPreservesNewUploadPriority(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	for _, id := range []media.ID{"backfill-a", "backfill-b"} {
		older := attachment
		older.ID = id
		older.StorageKey = media.StorageKey(id)
		record := auditRecord(t, "audit-"+id.String(), tenant.ID(older.TenantID), inventory.InventoryID(older.InventoryID), audit.ActionAttachmentCreated)
		if err := store.SaveAttachment(ctx, older, record, plannedThumbnailJob(t, older)); err != nil {
			t.Fatal(err)
		}
		// Simulate attachments that predate transactional thumbnail scheduling.
		if err := store.db.Delete(&thumbnailJobModel{AttachmentID: id.String(), Revision: int(media.CurrentThumbnailRevision)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	now := job.CreatedAt.Add(time.Hour)
	first, err := store.BackfillThumbnailJobs(ctx, 1, now)
	if err != nil || first.Scanned != 1 || first.Enqueued != 1 || first.Complete {
		t.Fatal("first batch incorrect", first, err)
	}
	// A new repository instance resumes from the persisted cursor.
	resumed := NewStore(store.db)
	second, err := resumed.BackfillThumbnailJobs(ctx, 1, now)
	if err != nil || second.Scanned != 1 || second.Enqueued != 1 || second.Cursor == first.Cursor {
		t.Fatal("cursor did not advance", second, err)
	}
	third, err := resumed.BackfillThumbnailJobs(ctx, 2, now)
	if err != nil || third.Enqueued != 0 || !third.Complete {
		t.Fatal("existing upload was duplicated", third, err)
	}
	final, err := store.BackfillThumbnailJobs(ctx, 2, now)
	if err != nil || final.Scanned != 0 || !final.Complete {
		t.Fatal("completed backfill restarted", final, err)
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker", 3, now, now.Add(time.Minute))
	if err != nil || len(claims) != 3 {
		t.Fatal("backfill missing", err)
	}
	if claims[0].Job.AttachmentID != attachment.ID || claims[0].Job.Priority != media.ThumbnailJobNewImage {
		t.Fatal("new upload lost priority")
	}
	for _, claim := range claims[1:] {
		if claim.Job.Priority != media.ThumbnailJobBackfill {
			t.Fatal("backfill has wrong priority")
		}
	}
}

func TestThumbnailBackfillRejectsUnboundedBatch(t *testing.T) {
	store := newTestStore(t, context.Background())
	for _, limit := range []int{0, 1001} {
		if _, err := store.BackfillThumbnailJobs(context.Background(), limit, time.Now()); err == nil {
			t.Fatal("invalid batch accepted")
		}
	}
}
