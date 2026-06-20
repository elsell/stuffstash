package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"time"
)

type attachmentModel struct {
	ID             string         `gorm:"primaryKey;size:26"`
	TenantID       string         `gorm:"not null;size:26;index"`
	Tenant         tenantModel    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	InventoryID    string         `gorm:"not null;size:26;index"`
	Inventory      inventoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:InventoryID;references:ID"`
	AssetID        string         `gorm:"not null;size:26;index"`
	Asset          assetModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:AssetID;references:ID"`
	StorageKey     string         `gorm:"not null;size:512;uniqueIndex"`
	FileName       string         `gorm:"not null;size:255"`
	ContentType    string         `gorm:"not null;size:128"`
	SizeBytes      int64          `gorm:"not null"`
	SHA256         string         `gorm:"not null;size:64"`
	LifecycleState string         `gorm:"not null;size:32;default:'active'"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (attachmentModel) TableName() string {
	return "attachments"
}

func (m attachmentModel) toDomain() (media.Attachment, bool) {
	id, ok := media.NewID(m.ID)
	if !ok {
		return media.Attachment{}, false
	}
	storageKey, ok := media.NewStorageKey(m.StorageKey)
	if !ok {
		return media.Attachment{}, false
	}
	fileName, ok := media.NewFileName(m.FileName)
	if !ok {
		return media.Attachment{}, false
	}
	contentType, ok := media.NewContentType(m.ContentType)
	if !ok {
		return media.Attachment{}, false
	}
	hash, ok := media.NewSHA256(m.SHA256)
	if !ok {
		return media.Attachment{}, false
	}
	lifecycleState, ok := media.NewLifecycleState(m.LifecycleState)
	if !ok {
		return media.Attachment{}, false
	}
	return media.NewAttachmentWithLifecycle(
		id,
		media.TenantID(m.TenantID),
		media.InventoryID(m.InventoryID),
		media.AssetID(m.AssetID),
		storageKey,
		fileName,
		contentType,
		m.SizeBytes,
		hash,
		m.CreatedAt,
		lifecycleState,
	)
}
