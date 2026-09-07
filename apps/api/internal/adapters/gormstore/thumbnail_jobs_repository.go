package gormstore

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func enqueueThumbnailJob(tx *gorm.DB, attachment media.Attachment, job *media.ThumbnailJob) error {
	if err := media.ValidatePlannedThumbnailJob(attachment, job); err != nil {
		return err
	}
	if job == nil {
		return nil
	}
	model := thumbnailJobModelFromDomain(*job)
	return tx.Omit(clause.Associations).Create(&model).Error
}

func (s Store) ClaimThumbnailJobs(ctx context.Context, claimID string, limit int, now, leaseUntil time.Time) ([]ports.ClaimedThumbnailJob, error) {
	now = now.UTC().Truncate(time.Microsecond)
	leaseUntil = leaseUntil.UTC().Truncate(time.Microsecond)
	if claimID == "" || limit <= 0 || now.IsZero() || !leaseUntil.After(now) {
		return nil, errors.New("invalid thumbnail claim settings")
	}
	claimed := []ports.ClaimedThumbnailJob{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Separate priority queries preserve explicit domain priority without raw SQL.
		for _, priority := range []media.ThumbnailJobPriority{media.ThumbnailJobNewImage, media.ThumbnailJobBackfill} {
			if len(claimed) == limit {
				break
			}
			var models []thumbnailJobModel
			query := tx.Clauses(skipLockedForUpdate()).Where(clause.And(
				clause.Eq{Column: "status", Value: string(ports.ThumbnailJobPending)},
				clause.Eq{Column: "priority", Value: string(priority)},
				clause.Lte{Column: "next_attempt_at", Value: now},
				clause.Or(clause.Eq{Column: "claim_id", Value: ""}, clause.Lte{Column: "claimed_until", Value: now}),
			)).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "attachment_id"}}).Limit(limit - len(claimed))
			if err := query.Find(&models).Error; err != nil {
				return err
			}
			for _, model := range models {
				attempts := model.Attempts + 1
				if err := tx.Model(&model).Updates(map[string]any{"claim_id": claimID, "claimed_until": leaseUntil, "updated_at": now, "attempts": attempts}).Error; err != nil {
					return err
				}
				model.Attempts = attempts
				model.ClaimID = claimID
				model.ClaimedUntil = &leaseUntil
				claimed = append(claimed, model.claim())
			}
		}
		return nil
	})
	return claimed, err
}

func (s Store) ResolveThumbnailJob(ctx context.Context, claim ports.ClaimedThumbnailJob, resolution ports.ThumbnailJobResolution) error {
	if err := resolution.Validate(); err != nil {
		return err
	}
	if claim.ClaimID == "" {
		return ports.ErrOutboxClaimLost
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model thumbnailJobModel
		job := claim.Job
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&thumbnailJobModel{
			AttachmentID: job.AttachmentID.String(), Revision: int(job.Revision),
			TenantID: job.TenantID.String(), InventoryID: job.InventoryID.String(), AssetID: job.AssetID.String(),
			ClaimID: claim.ClaimID,
		}).First(&model).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrOutboxClaimLost
		}
		if err != nil {
			return err
		}
		if !validThumbnailClaim(model, claim, resolution.At) {
			return ports.ErrOutboxClaimLost
		}
		updates := map[string]any{
			"status": string(resolution.Status), "claim_id": "", "claimed_until": nil,
			"failure":         string(resolution.Failure),
			"next_attempt_at": resolution.NextAttemptAt, "updated_at": resolution.At,
		}
		if resolution.Yielded {
			if model.Attempts < 1 {
				return ports.ErrOutboxClaimLost
			}
			updates["attempts"] = model.Attempts - 1
		}
		return tx.Model(&model).Updates(updates).Error
	})
}

func validThumbnailClaim(model thumbnailJobModel, claim ports.ClaimedThumbnailJob, now time.Time) bool {
	job := claim.Job
	return claim.ClaimID != "" && model.ClaimID == claim.ClaimID &&
		model.Status == string(ports.ThumbnailJobPending) && model.Revision == int(job.Revision) &&
		model.ClaimedUntil != nil && model.ClaimedUntil.Equal(claim.ClaimedUntil) && model.ClaimedUntil.After(now) &&
		model.claim().Job.Matches(media.Attachment{ID: job.AttachmentID, TenantID: job.TenantID, InventoryID: job.InventoryID, AssetID: job.AssetID, StorageKey: job.StorageKey, SHA256: job.SHA256})
}
