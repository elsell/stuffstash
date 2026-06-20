package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type blobDeletionEventModel struct {
	ID               string `gorm:"primaryKey;size:26"`
	StorageKey       string `gorm:"not null;size:512;index"`
	Attempts         int    `gorm:"not null;default:0"`
	LastError        string `gorm:"not null;default:''"`
	ClaimID          string `gorm:"not null;default:'';size:26;index"`
	ClaimedUntil     *time.Time
	ProcessedAt      *time.Time
	DeadLetteredAt   *time.Time `gorm:"index"`
	DeadLetterReason string     `gorm:"not null;default:''"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (blobDeletionEventModel) TableName() string {
	return "blob_deletion_events"
}

func (m blobDeletionEventModel) toPort() ports.BlobDeletionEvent {
	event := ports.BlobDeletionEvent{
		ID:               m.ID,
		StorageKey:       media.StorageKey(m.StorageKey),
		Attempts:         m.Attempts,
		LastError:        m.LastError,
		ClaimID:          m.ClaimID,
		DeadLetterReason: m.DeadLetterReason,
		CreatedAt:        m.CreatedAt,
	}
	if m.ClaimedUntil != nil {
		event.ClaimedUntil = *m.ClaimedUntil
	}
	if m.ProcessedAt != nil {
		event.ProcessedAt = *m.ProcessedAt
	}
	if m.DeadLetteredAt != nil {
		event.DeadLetteredAt = *m.DeadLetteredAt
	}
	return event
}
