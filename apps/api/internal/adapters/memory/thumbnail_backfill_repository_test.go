package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"testing"
	"time"
)

func TestThumbnailBackfillResumesWithoutOverwritingUploads(t *testing.T) {
	store, attachment, job := memoryPublicationFixture(t)
	older := attachment
	older.ID = "older"
	older.StorageKey = "older-original"
	store.attachments[older.ID] = older
	first, err := store.BackfillThumbnailJobs(context.Background(), 1, job.CreatedAt)
	if err != nil || first.Scanned != 1 || first.Enqueued != 1 || first.Complete {
		t.Fatal("bad initial batch", first, err)
	}
	second, err := store.BackfillThumbnailJobs(context.Background(), 2, job.CreatedAt)
	if err != nil || second.Scanned != 1 || second.Enqueued != 0 || !second.Complete {
		t.Fatal("bad resumed batch", second, err)
	}
	claims, err := store.ClaimThumbnailJobs(context.Background(), "worker", 2, job.CreatedAt, job.CreatedAt.Add(time.Minute))
	if err != nil || len(claims) != 2 || claims[0].Job.AttachmentID != attachment.ID || claims[1].Job.Priority != media.ThumbnailJobBackfill {
		t.Fatal("priority lost", claims, err)
	}
}
