package gormstore

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ThumbnailPublicationGuard struct {
	store Store
	clock ports.Clock
}

var _ ports.ThumbnailPublicationGuard = (*ThumbnailPublicationGuard)(nil)

func NewThumbnailPublicationGuard(store Store, clock ports.Clock) (*ThumbnailPublicationGuard, error) {
	if store.db == nil || clock == nil {
		return nil, errors.New("thumbnail publication requires storage and clock")
	}
	return &ThumbnailPublicationGuard{store: store, clock: clock}, nil
}

func (g *ThumbnailPublicationGuard) Publish(ctx context.Context, expected media.Attachment, claim *ports.ClaimedThumbnailJob, publish func(context.Context) error) error {
	if publish == nil || expected.ID == "" || expected.TenantID == "" || expected.InventoryID == "" || expected.AssetID == "" || expected.StorageKey == "" || expected.SHA256 == "" || !expected.ContentType.IsImage() {
		return errors.New("thumbnail publication requires a scoped image and publisher")
	}
	// Connection acquisition and row-lock queries remain cancellable. The transaction
	// itself must outlive cancellation until the publisher has actually returned.
	return g.store.db.WithContext(ctx).Connection(func(connection *gorm.DB) error {
		return connection.WithContext(context.WithoutCancel(ctx)).Transaction(func(tx *gorm.DB) error {
			if tx.Dialector.Name() == "sqlite" {
				// SQLite ignores FOR UPDATE. A no-op identity write acquires its writer
				// lock without changing domain state, including when WAL is enabled.
				err := tx.WithContext(ctx).Model(&attachmentModel{}).Where(&attachmentModel{
					ID: expected.ID.String(), TenantID: expected.TenantID.String(), InventoryID: expected.InventoryID.String(), AssetID: expected.AssetID.String(),
				}).UpdateColumn("id", expected.ID.String()).Error
				if err != nil {
					return err
				}
			}
			var model attachmentModel
			err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(&attachmentModel{
				ID: expected.ID.String(), TenantID: expected.TenantID.String(),
				InventoryID: expected.InventoryID.String(), AssetID: expected.AssetID.String(),
			}).First(&model).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrBlobNotFound
			}
			if err != nil {
				return err
			}
			actual, valid := model.toDomain()
			if !valid || !actual.ContentType.IsImage() || actual.StorageKey != expected.StorageKey || actual.SHA256 != expected.SHA256 {
				return ports.ErrBlobNotFound
			}
			if claim != nil {
				if !claim.Job.Matches(actual) || claim.ClaimID == "" || claim.Job.Revision != media.CurrentThumbnailRevision {
					return ports.ErrOutboxClaimLost
				}
				var job thumbnailJobModel
				err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(&thumbnailJobModel{AttachmentID: actual.ID.String(), Revision: int(claim.Job.Revision)}).First(&job).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ports.ErrOutboxClaimLost
				}
				if err != nil {
					return err
				}
				if !validThumbnailClaim(job, *claim, g.clock.Now()) {
					return ports.ErrOutboxClaimLost
				}
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			return publish(ctx)
		})
	})
}
