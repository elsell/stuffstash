package gormstore

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func (s Store) CreateImportedAttachment(ctx context.Context, attachment media.Attachment, auditRecord audit.Record, link ports.ImportSourceLink, record ports.ImportJobResource) error {
	if err := validateImportSourceLink(link); err != nil {
		return err
	}
	if err := validateImportJobResource(record); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item assetModel
		err := tx.Where(&assetModel{
			ID:          attachment.AssetID.String(),
			TenantID:    attachment.TenantID.String(),
			InventoryID: attachment.InventoryID.String(),
		}).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if item.LifecycleState != asset.LifecycleStateActive.String() {
			return ports.ErrForbidden
		}
		if err := tx.Create(&attachmentModel{
			ID:             attachment.ID.String(),
			TenantID:       attachment.TenantID.String(),
			InventoryID:    attachment.InventoryID.String(),
			AssetID:        attachment.AssetID.String(),
			StorageKey:     attachment.StorageKey.String(),
			FileName:       attachment.FileName.String(),
			ContentType:    attachment.ContentType.String(),
			SizeBytes:      attachment.SizeBytes,
			SHA256:         attachment.SHA256.String(),
			LifecycleState: lifecycleStateOrActive(attachment.LifecycleState.String()),
			CreatedAt:      attachment.CreatedAt,
		}).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		linkModel := importSourceLinkModelFromRecord(link)
		if err := tx.Create(&linkModel).Error; err != nil {
			return importLinkPersistenceError(err)
		}
		resourceModel := importJobResourceModelFromRecord(record)
		if err := tx.Create(&resourceModel).Error; err != nil {
			return importLinkPersistenceError(err)
		}
		return nil
	})
}
