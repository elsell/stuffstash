package media

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestCleanupRemovesLateWritesAndAttemptsEveryKey(t *testing.T) {
	for _, failOriginal := range []bool{false, true} {
		now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
		queue := &cleanupQueue{now: now, event: ports.BlobDeletionEvent{ID: "deleted", StorageKey: "original", ProcessedAt: now.Add(-2 * time.Hour)}}
		blobs := &cleanupBlobs{BlobStorage: memory.NewStore(), failOriginal: failOriginal}
		worker, err := NewCleanupWorker(queue, blobs, queue, workerIDs{}, workerObserver{}, time.Hour, time.Minute, 30*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		// These bytes model a PUT accepted remotely before deletion, completing late.
		key, _ := ThumbnailCacheKey(queue.event.StorageKey, domain.ThumbnailVariantSmall)
		if err := blobs.PutBlob(context.Background(), key, domain.ContentTypeJPEG, []byte("late thumbnail")); err != nil {
			t.Fatal(err)
		}
		if err := blobs.PutBlob(context.Background(), thumbnailMetadataKey(key), domain.ContentType("text/plain"), []byte("image/jpeg")); err != nil {
			t.Fatal(err)
		}
		worked, err := worker.Drain(context.Background())
		if err != nil || !worked || queue.event.RecheckFailed != failOriginal {
			t.Fatal("cleanup did not resolve correctly", err)
		}
		if len(blobs.deleted) != 7 {
			t.Fatal("failed earlier delete skipped a derivative", blobs.deleted)
		}
		if _, err := blobs.GetBlob(context.Background(), key); !errors.Is(err, ports.ErrBlobNotFound) {
			t.Fatal("late thumbnail survived", err)
		}
		if _, err := blobs.GetBlob(context.Background(), thumbnailMetadataKey(key)); !errors.Is(err, ports.ErrBlobNotFound) {
			t.Fatal("late metadata survived", err)
		}
		// A later accepted write is still cleaned because the tombstone remains.
		queue.now = now.Add(2 * time.Hour)
		if err := blobs.PutBlob(context.Background(), key, domain.ContentTypeJPEG, []byte("later thumbnail")); err != nil {
			t.Fatal(err)
		}
		if _, err := worker.Drain(context.Background()); err != nil {
			t.Fatal(err)
		}
		if _, err := blobs.GetBlob(context.Background(), key); !errors.Is(err, ports.ErrBlobNotFound) {
			t.Fatal("tombstone was discarded", err)
		}
	}
}

type cleanupQueue struct {
	now   time.Time
	event ports.BlobDeletionEvent
}

func (q *cleanupQueue) Now() time.Time { return q.now }
func (q *cleanupQueue) ClaimBlobDeletionRechecks(ctx context.Context, id string, _ int, now, until time.Time, interval time.Duration) ([]ports.BlobDeletionEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if q.event.RecheckedAt.After(now.Add(-interval)) || q.event.ProcessedAt.After(now.Add(-interval)) {
		return nil, nil
	}
	q.event.ClaimID = id
	q.event.ClaimedUntil = until
	return []ports.BlobDeletionEvent{q.event}, nil
}
func (q *cleanupQueue) ResolveBlobDeletionRecheck(ctx context.Context, event ports.BlobDeletionEvent, now time.Time, failed bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.ClaimID != q.event.ClaimID || !event.ClaimedUntil.Equal(q.event.ClaimedUntil) {
		return ports.ErrOutboxClaimLost
	}
	q.event.RecheckedAt = now
	q.event.RecheckFailed = failed
	q.event.ClaimID = ""
	return nil
}

type cleanupBlobs struct {
	ports.BlobStorage
	deleted      []domain.StorageKey
	failOriginal bool
}

func (b *cleanupBlobs) DeleteBlob(ctx context.Context, key domain.StorageKey) error {
	b.deleted = append(b.deleted, key)
	if b.failOriginal && key == "original" {
		return errors.New("controlled deletion failure")
	}
	return b.BlobStorage.DeleteBlob(ctx, key)
}
