package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"gorm.io/gorm"
)

func checkoutsEquivalentForStorage(db *gorm.DB, current, expected asset.Checkout) bool {
	if db.Dialector.Name() == "postgres" {
		// pgx writes PostgreSQL timestamps at microsecond precision. Older JSON
		// undo snapshots retain nanoseconds, so compare at the persisted precision.
		current = checkoutAtPostgresPrecision(current)
		expected = checkoutAtPostgresPrecision(expected)
	}
	return asset.CheckoutsEquivalentForStaleCheck(current, expected)
}

func checkoutAtPostgresPrecision(checkout asset.Checkout) asset.Checkout {
	checkout.CheckedOutAt = checkout.CheckedOutAt.Truncate(time.Microsecond)
	checkout.ReturnedAt = checkout.ReturnedAt.Truncate(time.Microsecond)
	checkout.CreatedAt = checkout.CreatedAt.Truncate(time.Microsecond)
	checkout.UpdatedAt = checkout.UpdatedAt.Truncate(time.Microsecond)
	return checkout
}
