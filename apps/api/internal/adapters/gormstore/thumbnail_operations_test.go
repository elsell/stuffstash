package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestThumbnailOperationsRetryOnlyFailedJobs(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	now := job.CreatedAt.Add(time.Minute)
	status, err := store.ThumbnailQueueStatus(ctx, now)
	if err != nil || status.Pending != 1 || status.Leased != 0 || !status.OldestPendingAt.Equal(job.CreatedAt) {
		t.Fatal("incorrect pending status", status, err)
	}
	claims, err := store.ClaimThumbnailJobs(ctx, "worker", 1, now, now.Add(time.Minute))
	if err != nil || len(claims) != 1 {
		t.Fatal(err)
	}
	if count, err := store.RetryFailedThumbnailJobs(ctx, 10, now); err != nil || count != 0 {
		t.Fatal("retried leased job", err)
	}
	status, err = store.ThumbnailQueueStatus(ctx, now)
	if err != nil || status.Leased != 1 {
		t.Fatal("leased count missing", err)
	}
	if err := store.ResolveThumbnailJob(ctx, claims[0], ports.ThumbnailJobResolution{Status: ports.ThumbnailJobFailed, At: now, Failure: ports.ThumbnailFailureProcessing}); err != nil {
		t.Fatal(err)
	}
	if count, err := store.RetryFailedThumbnailJobs(ctx, 10, now); err != nil || count != 1 {
		t.Fatal("failed retry missing", err)
	}
	retry, err := store.ClaimThumbnailJobs(ctx, "retry", 1, now, now.Add(time.Minute))
	if err != nil || len(retry) != 1 || retry[0].Attempts != 1 {
		t.Fatal("retry did not reset attempt budget", retry, err)
	}
	if err := store.ResolveThumbnailJob(ctx, retry[0], ports.ThumbnailJobResolution{Status: ports.ThumbnailJobCompleted, At: now}); err != nil {
		t.Fatal(err)
	}
	if count, err := store.RetryFailedThumbnailJobs(ctx, 10, now); err != nil || count != 0 {
		t.Fatal("completed job retried", err)
	}
}
