package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) ClaimPendingAuthorizationOutboxEvents(ctx context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}

	events := []ports.AuthorizationOutboxEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []authorizationOutboxEventModel
		now := time.Now()
		if err := tx.
			Clauses(skipLockedForUpdate()).
			Where(claimableAuthorizationOutboxEvent(now)).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
			Limit(limit).
			Find(&models).Error; err != nil {
			return err
		}

		claimed := make([]authorizationOutboxEventModel, 0, len(models))
		for _, model := range models {
			model.ClaimID = claimID
			model.ClaimedUntil = &leaseUntil
			claimed = append(claimed, model)
		}

		for _, model := range claimed {
			if err := tx.
				Model(&authorizationOutboxEventModel{}).
				Where(&authorizationOutboxEventModel{ID: model.ID}).
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

func skipLockedForUpdate() clause.Locking {
	return clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}
}

func claimableAuthorizationOutboxEvent(now time.Time) clause.Expression {
	return clause.And(
		clause.Eq{Column: clause.Column{Name: "processed_at"}, Value: nil},
		clause.Eq{Column: clause.Column{Name: "dead_lettered_at"}, Value: nil},
		clause.Or(
			clause.Eq{Column: clause.Column{Name: "claim_id"}, Value: ""},
			clause.Lte{Column: clause.Column{Name: "claimed_until"}, Value: now},
		),
	)
}

func (s Store) MarkAuthorizationOutboxEventProcessed(ctx context.Context, eventID string, claimID string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
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
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}

func (s Store) MarkAuthorizationOutboxEventFailed(ctx context.Context, eventID string, claimID string, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model authorizationOutboxEventModel
		if err := tx.Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ports.ErrAuthorizationOutboxClaimLost
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

func (s Store) MarkAuthorizationOutboxEventDeadLettered(ctx context.Context, eventID string, claimID string, reason string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&authorizationOutboxEventModel{}).
		Where(&authorizationOutboxEventModel{ID: eventID, ClaimID: claimID}).
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
		return ports.ErrAuthorizationOutboxClaimLost
	}
	return nil
}
