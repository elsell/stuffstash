package gormstore

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type thumbnailJobModel struct {
	AttachmentID  string          `gorm:"primaryKey;size:26"`
	Revision      int             `gorm:"primaryKey"`
	Attachment    attachmentModel `gorm:"foreignKey:AttachmentID;references:ID;constraint:OnDelete:CASCADE"`
	TenantID      string          `gorm:"not null;size:26"`
	InventoryID   string          `gorm:"not null;size:26"`
	AssetID       string          `gorm:"not null;size:26"`
	StorageKey    string          `gorm:"not null;size:512"`
	SHA256        string          `gorm:"not null;size:64"`
	Priority      string          `gorm:"not null;size:32"`
	Status        string          `gorm:"not null;size:32;index:idx_thumbnail_jobs_pending,priority:1"`
	Attempts      int             `gorm:"not null;default:0"`
	Failure       string          `gorm:"not null;default:''"`
	ClaimID       string          `gorm:"not null;default:'';size:26"`
	ClaimedUntil  *time.Time
	NextAttemptAt time.Time `gorm:"not null;index:idx_thumbnail_jobs_pending,priority:2"`
	CreatedAt     time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime:false"`
}

func (thumbnailJobModel) TableName() string { return "thumbnail_jobs" }

func thumbnailJobModelFromDomain(job media.ThumbnailJob) thumbnailJobModel {
	return thumbnailJobModel{
		AttachmentID: job.AttachmentID.String(), TenantID: job.TenantID.String(),
		InventoryID: job.InventoryID.String(), AssetID: job.AssetID.String(),
		StorageKey: job.StorageKey.String(), SHA256: job.SHA256.String(),
		Revision: int(job.Revision), Priority: string(job.Priority),
		Status: string(ports.ThumbnailJobPending), NextAttemptAt: job.CreatedAt,
		CreatedAt: job.CreatedAt, UpdatedAt: job.CreatedAt,
	}
}

func (m thumbnailJobModel) claim() ports.ClaimedThumbnailJob {
	result := ports.ClaimedThumbnailJob{Job: media.ThumbnailJob{
		AttachmentID: media.ID(m.AttachmentID), TenantID: media.TenantID(m.TenantID),
		InventoryID: media.InventoryID(m.InventoryID), AssetID: media.AssetID(m.AssetID),
		StorageKey: media.StorageKey(m.StorageKey), SHA256: media.SHA256(m.SHA256),
		Revision: media.ThumbnailRevision(m.Revision), Priority: media.ThumbnailJobPriority(m.Priority), CreatedAt: m.CreatedAt,
	}, ClaimID: m.ClaimID, Attempts: m.Attempts}
	if m.ClaimedUntil != nil {
		result.ClaimedUntil = *m.ClaimedUntil
	}
	return result
}
