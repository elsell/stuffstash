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
	now        time.Time
	event      ports.BlobDeletionEvent
	afterClaim func()
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
	if q.afterClaim != nil {
		q.afterClaim()
	}
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
	deleted         []domain.StorageKey
	failOriginal    bool
	onDelete        func()
	stallOriginal   bool
	liveDerivatives int
}

func (b *cleanupBlobs) DeleteBlob(ctx context.Context, key domain.StorageKey) error {
	if b.onDelete != nil {
		b.onDelete()
	}
	b.deleted = append(b.deleted, key)
	if key == "original" && b.stallOriginal {
		<-ctx.Done()
		return ctx.Err()
	}
	if key != "original" && ctx.Err() == nil {
		b.liveDerivatives++
	}
	if b.failOriginal && key == "original" {
		return errors.New("controlled deletion failure")
	}
	return b.BlobStorage.DeleteBlob(ctx, key)
}

func TestCleanupCancellationLeavesRecoverableLease(t *testing.T) {
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	queue := &cleanupQueue{now: now, event: ports.BlobDeletionEvent{ID: "deleted", StorageKey: "original", ProcessedAt: now.Add(-2 * time.Hour)}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blobs := &cleanupBlobs{BlobStorage: memory.NewStore(), onDelete: cancel}
	worker, err := NewCleanupWorker(queue, blobs, queue, workerIDs{}, workerObserver{}, time.Hour, time.Minute, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worker.Drain(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("cleanup ignored cancellation", err)
	}
	if !queue.event.RecheckedAt.IsZero() || queue.event.ClaimID == "" {
		t.Fatal("cancelled cleanup acknowledged its lease")
	}
}

func TestCleanupStalledOriginalLeavesTimeForDerivatives(t *testing.T) {
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	queue := &cleanupQueue{now: now, event: ports.BlobDeletionEvent{ID: "deleted", StorageKey: "original", ProcessedAt: now.Add(-2 * time.Hour)}}
	blobs := &cleanupBlobs{BlobStorage: memory.NewStore(), stallOriginal: true}
	key, _ := ThumbnailCacheKey("original", domain.ThumbnailVariantSmall)
	if err := blobs.PutBlob(context.Background(), key, domain.ContentTypeJPEG, []byte("late")); err != nil {
		t.Fatal(err)
	}
	worker, err := NewCleanupWorker(queue, blobs, queue, workerIDs{}, workerObserver{}, time.Hour, time.Second, 80*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worker.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if blobs.liveDerivatives != 6 {
		t.Fatal("stalled original consumed derivative deadlines", blobs.liveDerivatives)
	}
	if _, err := blobs.GetBlob(context.Background(), key); !errors.Is(err, ports.ErrBlobNotFound) {
		t.Fatal("late derivative survived stalled original", err)
	}
}

func TestCleanupDoesNotProcessExpiredClaim(t *testing.T) {
	now := time.Date(2026, 9, 6, 22, 0, 0, 0, time.UTC)
	queue := &cleanupQueue{now: now, event: ports.BlobDeletionEvent{ID: "deleted", StorageKey: "original", ProcessedAt: now.Add(-2 * time.Hour)}}
	queue.afterClaim = func() { queue.now = now.Add(2 * time.Minute) }
	blobs := &cleanupBlobs{BlobStorage: memory.NewStore()}
	worker, err := NewCleanupWorker(queue, blobs, queue, workerIDs{}, workerObserver{}, time.Hour, time.Minute, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worker.Drain(context.Background()); !errors.Is(err, ports.ErrOutboxClaimLost) {
		t.Fatal("expired claim processed", err)
	}
	if len(blobs.deleted) != 0 {
		t.Fatal("expired claim touched blob storage")
	}
}
