package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func deleteAssetAttachments(tx *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, now time.Time) error {
	var attachments []attachmentModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&attachmentModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&attachments).Error; err != nil {
		return err
	}
	for _, attachment := range attachments {
		if err := tx.Create(&blobDeletionEventModel{ID: attachment.ID, StorageKey: attachment.StorageKey, CreatedAt: now}).Error; err != nil {
			return err
		}
		if err := deleteAttachmentRows(tx, attachment); err != nil {
			return err
		}
	}
	return nil
}

func deleteAttachmentRows(tx *gorm.DB, attachment attachmentModel) error {
	if err := tx.Where(&thumbnailJobModel{AttachmentID: attachment.ID}).Delete(&thumbnailJobModel{}).Error; err != nil {
		return err
	}
	return tx.Delete(&attachment).Error
}
