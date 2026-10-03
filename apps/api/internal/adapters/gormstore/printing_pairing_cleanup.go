package gormstore

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) CleanupExpiredPrintPairings(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit < 1 || limit > 1000 {
		return 0, ports.ErrPrintConflict
	}
	count := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var expired []printingPairingModel
		err := tx.Select("id").Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where(clause.Lte{Column: "expires_at", Value: now}).Order(clause.OrderByColumn{Column: clause.Column{Name: "expires_at"}}).Limit(limit).Find(&expired).Error
		if err != nil {
			return err
		}
		if len(expired) == 0 {
			return nil
		}
		ids := make([]any, 0, len(expired))
		for _, pairing := range expired {
			ids = append(ids, pairing.ID)
		}
		result := tx.Where(clause.IN{Column: "id", Values: ids}).Where(clause.Lte{Column: "expires_at", Value: now}).Delete(&printingPairingModel{})
		count = int(result.RowsAffected)
		return result.Error
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
