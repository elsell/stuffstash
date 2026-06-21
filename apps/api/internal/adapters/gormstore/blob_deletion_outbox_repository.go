package gormstore

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) ClaimPendingBlobDeletionEvents(ctx context.Context, claimID string, limit int, now time.Time, leaseUntil time.Time) ([]ports.BlobDeletionEvent, error) {
	if limit <= 0 {
		limit = 25
	}

	events := []ports.BlobDeletionEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []blobDeletionEventModel
		if err := tx.
			Clauses(skipLockedForUpdate()).
			Where(claimableBlobDeletionEvent(now)).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}
		for _, model := range models {
			model.ClaimID = claimID
			model.ClaimedUntil = &leaseUntil
			if err := tx.Model(&blobDeletionEventModel{}).
				Where(&blobDeletionEventModel{ID: model.ID}).
				Updates(map[string]any{
					"claim_id":      model.ClaimID,
					"claimed_until": model.ClaimedUntil,
				}).Error; err != nil {
				return err
			}
			events = append(events, model.toPort())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

func claimableBlobDeletionEvent(now time.Time) clause.Expression {
	return clause.And(
		clause.Eq{Column: clause.Column{Name: "processed_at"}, Value: nil},
		clause.Eq{Column: clause.Column{Name: "dead_lettered_at"}, Value: nil},
		clause.Or(
			clause.Eq{Column: clause.Column{Name: "claim_id"}, Value: ""},
			clause.Lte{Column: clause.Column{Name: "claimed_until"}, Value: now},
		),
	)
}

func (s Store) MarkBlobDeletionEventProcessed(ctx context.Context, eventID string, claimID string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&blobDeletionEventModel{}).
		Where(&blobDeletionEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"processed_at":  now,
			"last_error":    "",
			"claim_id":      "",
			"claimed_until": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrOutboxClaimLost
	}
	return nil
}

func (s Store) MarkBlobDeletionEventFailed(ctx context.Context, eventID string, claimID string, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model blobDeletionEventModel
		if err := tx.Where(&blobDeletionEventModel{ID: eventID, ClaimID: claimID}).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrOutboxClaimLost
			}
			return err
		}
		model.Attempts++
		model.LastError = reason
		model.ClaimID = ""
		model.ClaimedUntil = nil
		return tx.Save(&model).Error
	})
}

func (s Store) MarkBlobDeletionEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&blobDeletionEventModel{}).
		Where(&blobDeletionEventModel{ID: eventID, ClaimID: claimID}).
		Updates(map[string]any{
			"dead_lettered_at":   now,
			"dead_letter_reason": reason,
			"claim_id":           "",
			"claimed_until":      nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrOutboxClaimLost
	}
	return nil
}
