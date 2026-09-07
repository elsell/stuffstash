package media

import (
	"context"
	"errors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

type CleanupWorker struct {
	queue                    ports.BlobDeletionRechecks
	blobs                    ports.BlobStorage
	clock                    ports.Clock
	ids                      ports.IDGenerator
	observer                 ports.Observer
	interval, lease, timeout time.Duration
}

func NewCleanupWorker(queue ports.BlobDeletionRechecks, blobs ports.BlobStorage, clock ports.Clock, ids ports.IDGenerator, observer ports.Observer, interval, lease, timeout time.Duration) (*CleanupWorker, error) {
	if queue == nil || blobs == nil || clock == nil || ids == nil || observer == nil || interval <= 0 || timeout <= 0 || lease <= timeout {
		return nil, errors.New("invalid cleanup worker dependencies or timing")
	}
	return &CleanupWorker{queue: queue, blobs: blobs, clock: clock, ids: ids, observer: observer, interval: interval, lease: lease, timeout: timeout}, nil
}
func (w *CleanupWorker) Drain(ctx context.Context) (bool, error) {
	now := w.clock.Now()
	events, err := w.queue.ClaimBlobDeletionRechecks(ctx, w.ids.NewID(), 1, now, now.Add(w.lease), w.interval)
	if err != nil || len(events) == 0 {
		return false, err
	}
	if len(events) != 1 || events[0].ID == "" || events[0].StorageKey == "" {
		return true, ports.ErrOutboxClaimLost
	}
	event := events[0]
	budget := event.ClaimedUntil.Sub(w.clock.Now()) - (w.lease - w.timeout)
	if budget <= 0 {
		return true, ports.ErrOutboxClaimLost
	}
	if budget > w.timeout {
		budget = w.timeout
	}
	cleaning, cancel := context.WithTimeout(ctx, budget)
	keys := []domain.StorageKey{event.StorageKey}
	for _, variant := range []domain.ThumbnailVariant{domain.ThumbnailVariantSmall, domain.ThumbnailVariantMedium, domain.ThumbnailVariantLarge} {
		key, err := ThumbnailCacheKey(event.StorageKey, variant)
		if err != nil {
			cancel()
			return true, err
		}
		keys = append(keys, key, thumbnailMetadataKey(key))
	}
	failed := false
	// Equal bounded shares leave one share for scheduling overhead.
	perKeyBudget := budget / time.Duration(len(keys)+1)
	for _, key := range keys {
		deleting, endDelete := context.WithTimeout(cleaning, perKeyBudget)
		err := w.blobs.DeleteBlob(deleting, key)
		if err == nil {
			err = deleting.Err()
		}
		endDelete()
		if err != nil && !errors.Is(err, ports.ErrBlobNotFound) {
			failed = true
		}
	}
	if cleaning.Err() != nil {
		failed = true
	}
	cancel()
	if err := ctx.Err(); err != nil {
		return true, err
	}
	if err := w.queue.ResolveBlobDeletionRecheck(ctx, event, w.clock.Now(), failed); err != nil {
		return true, err
	}
	outcome := "completed"
	if failed {
		outcome = "failed"
	}
	w.observer.Record(ctx, ports.Event{Name: ports.EventBlobDeletionRechecked, Message: "deleted blob keys rechecked", Fields: map[string]string{"outcome": outcome}})
	return true, nil
}
