package memory

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ThumbnailPublicationGuard struct {
	store *Store
	clock ports.Clock
}

var _ ports.ThumbnailPublicationGuard = (*ThumbnailPublicationGuard)(nil)

func NewThumbnailPublicationGuard(store *Store, clock ports.Clock) (*ThumbnailPublicationGuard, error) {
	if store == nil || clock == nil {
		return nil, errors.New("thumbnail publication requires storage and clock")
	}
	return &ThumbnailPublicationGuard{store: store, clock: clock}, nil
}

func (g *ThumbnailPublicationGuard) Publish(ctx context.Context, expected media.Attachment, claim *ports.ClaimedThumbnailJob, publish func(context.Context) error) error {
	if publish == nil || expected.ID == "" || expected.TenantID == "" || expected.InventoryID == "" || expected.AssetID == "" || expected.StorageKey == "" || expected.SHA256 == "" || !expected.ContentType.IsImage() {
		return errors.New("thumbnail publication requires a scoped image and publisher")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Blob storage has its own mutex, so the lifecycle lock remains held until
	// publication returns without deadlocking the blob callback.
	g.store.mu.Lock()
	defer g.store.mu.Unlock()
	actual, exists := g.store.attachments[expected.ID]
	if !exists || actual.TenantID != expected.TenantID || actual.InventoryID != expected.InventoryID || actual.AssetID != expected.AssetID || actual.StorageKey != expected.StorageKey || actual.SHA256 != expected.SHA256 || !actual.ContentType.IsImage() {
		return ports.ErrBlobNotFound
	}
	if claim != nil {
		record, exists := g.store.thumbnailJobs[thumbnailKey(claim.Job)]
		if !exists || record.status != ports.ThumbnailJobPending || claim.Job.Revision != media.CurrentThumbnailRevision || !claim.Job.Matches(actual) || !record.claim.Job.Matches(actual) || claim.ClaimID == "" || record.claim.ClaimID != claim.ClaimID || !record.claim.ClaimedUntil.Equal(claim.ClaimedUntil) || !record.claim.ClaimedUntil.After(g.clock.Now()) {
			return ports.ErrOutboxClaimLost
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return publish(ctx)
}
