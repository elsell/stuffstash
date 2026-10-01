package media

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a AttachmentService) drainBlobDeletionOutboxBestEffort(ctx context.Context, limit int) {
	if err := a.DrainBlobDeletionOutbox(ctx, limit); err != nil {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobDeletionOutboxFailed,
			Message: "blob deletion outbox drain failed",
			Fields:  map[string]string{"error": err.Error()},
		})
	}
}

func (a AttachmentService) DrainBlobDeletionOutbox(ctx context.Context, limit int) error {
	if a.deps.BlobDeletionOutbox == nil || a.deps.Blobs == nil {
		return nil
	}
	if limit <= 0 {
		limit = 1
	}
	claimID := a.deps.IDs.NewID()
	now := a.deps.Clock.Now().UTC()
	events, err := a.deps.BlobDeletionOutbox.ClaimPendingBlobDeletionEvents(ctx, claimID, limit, now, now.Add(a.deps.BlobDeletionClaimLease))
	if err != nil {
		return err
	}
	if len(events) > 0 {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobDeletionOutboxClaimed,
			Message: "blob deletion outbox events claimed",
			Fields: map[string]string{
				"event_count": strconv.Itoa(len(events)),
			},
		})
	}
	for _, event := range events {
		if err := a.deleteBlobAndThumbnailDerivatives(ctx, event.StorageKey); err != nil {
			a.deps.Observer.Record(ctx, ports.Event{
				Name:    ports.EventBlobDeletionOutboxFailed,
				Message: "blob deletion outbox event failed",
				Fields: map[string]string{
					"event_id": event.ID,
					"attempts": strconv.Itoa(event.Attempts + 1),
				},
			})
			if event.Attempts+1 >= a.deps.BlobDeletionMaxAttempts {
				if markErr := a.deps.BlobDeletionOutbox.MarkBlobDeletionEventDeadLettered(ctx, event.ID, claimID, err.Error()); markErr != nil {
					return markErr
				}
				a.deps.Observer.Record(ctx, ports.Event{
					Name:    ports.EventBlobDeletionOutboxDeadLettered,
					Message: "blob deletion outbox event dead-lettered",
					Fields: map[string]string{
						"event_id": event.ID,
						"attempts": strconv.Itoa(event.Attempts + 1),
					},
				})
			} else {
				if markErr := a.deps.BlobDeletionOutbox.MarkBlobDeletionEventFailed(ctx, event.ID, claimID, err.Error()); markErr != nil {
					return markErr
				}
			}
			continue
		}
		if err := a.deps.BlobDeletionOutbox.MarkBlobDeletionEventProcessed(ctx, event.ID, claimID); err != nil {
			return err
		}
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobDeletionOutboxProcessed,
			Message: "blob deletion outbox event processed",
			Fields: map[string]string{
				"event_id": event.ID,
			},
		})
	}
	return nil
}

func (a AttachmentService) deleteBlobAndThumbnailDerivatives(ctx context.Context, storageKey media.StorageKey) error {
	keys := append([]media.StorageKey{storageKey}, ThumbnailStorageKeysForBlob(storageKey)...)
	var deletionErrors []error
	for _, key := range keys {
		if err := a.deps.Blobs.DeleteBlob(ctx, key); err != nil && !errors.Is(err, ports.ErrBlobNotFound) {
			deletionErrors = append(deletionErrors, err)
		}
	}
	return errors.Join(deletionErrors...)
}
