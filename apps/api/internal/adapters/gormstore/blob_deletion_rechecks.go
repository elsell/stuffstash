package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) ClaimBlobDeletionRechecks(ctx context.Context, claimID string, limit int, now, leaseUntil time.Time, interval time.Duration) ([]ports.BlobDeletionEvent, error) {
	now = now.UTC().Truncate(time.Microsecond)
	leaseUntil = leaseUntil.UTC().Truncate(time.Microsecond)
	if claimID == "" || limit < 1 || limit > 1000 || now.IsZero() || !leaseUntil.After(now) || interval <= 0 {
		return nil, errors.New("invalid deletion recheck claim")
	}
	claimed := []ports.BlobDeletionEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, neverRechecked := range []bool{true, false} {
			if len(claimed) == limit {
				break
			}
			var models []blobDeletionEventModel
			query := tx.Clauses(skipLockedForUpdate()).Where(clause.And(
				clause.Lte{Column: "processed_at", Value: now.Add(-interval)},
				clause.Or(clause.Eq{Column: "claim_id", Value: ""}, clause.Lte{Column: "claimed_until", Value: now}),
			))
			if neverRechecked {
				query = query.Where(clause.Eq{Column: "rechecked_at", Value: nil}).Order(clause.OrderByColumn{Column: clause.Column{Name: "processed_at"}})
			} else {
				query = query.Where(clause.Lte{Column: "rechecked_at", Value: now.Add(-interval)}).Order(clause.OrderByColumn{Column: clause.Column{Name: "rechecked_at"}})
			}
			if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit - len(claimed)).Find(&models).Error; err != nil {
				return err
			}
			for _, model := range models {
				if err := tx.Model(&model).Updates(map[string]any{"claim_id": claimID, "claimed_until": leaseUntil}).Error; err != nil {
					return err
				}
				model.ClaimID = claimID
				model.ClaimedUntil = &leaseUntil
				claimed = append(claimed, model.toPort())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

func (s Store) ResolveBlobDeletionRecheck(ctx context.Context, event ports.BlobDeletionEvent, now time.Time, failed bool) error {
	if event.ID == "" || event.StorageKey == "" || event.ClaimID == "" || now.IsZero() {
		return ports.ErrOutboxClaimLost
	}
	now = now.UTC().Truncate(time.Microsecond)
	result := s.db.WithContext(ctx).Model(&blobDeletionEventModel{}).Where(&blobDeletionEventModel{ID: event.ID, ClaimID: event.ClaimID, StorageKey: event.StorageKey.String()}).Where(clause.And(
		clause.Eq{Column: "claimed_until", Value: event.ClaimedUntil}, clause.Gt{Column: "claimed_until", Value: now}, clause.Neq{Column: "processed_at", Value: nil},
	)).Updates(map[string]any{"rechecked_at": now, "recheck_failed": failed, "claim_id": "", "claimed_until": nil})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ports.ErrOutboxClaimLost
	}
	return nil
}
