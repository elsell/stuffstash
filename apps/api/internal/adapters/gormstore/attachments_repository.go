package gormstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveAttachment(ctx context.Context, attachment media.Attachment, auditRecord audit.Record) error {
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
		if err := tx.Create(&attachmentModel{
			ID:          attachment.ID.String(),
			TenantID:    attachment.TenantID.String(),
			InventoryID: attachment.InventoryID.String(),
			AssetID:     attachment.AssetID.String(),
			StorageKey:  attachment.StorageKey.String(),
			FileName:    attachment.FileName.String(),
			ContentType: attachment.ContentType.String(),
			SizeBytes:   attachment.SizeBytes,
			SHA256:      attachment.SHA256.String(),
			CreatedAt:   attachment.CreatedAt,
		}).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) AttachmentByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, bool, error) {
	var model attachmentModel
	err := s.db.WithContext(ctx).Where(&attachmentModel{
		ID:          attachmentID.String(),
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		AssetID:     assetID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return media.Attachment{}, false, nil
	}
	if err != nil {
		return media.Attachment{}, false, err
	}
	attachment, ok := model.toDomain()
	if !ok {
		return media.Attachment{}, false, fmt.Errorf("invalid attachment row %q", model.ID)
	}
	return attachment, true, nil
}

func (s Store) ListAttachmentsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, page ports.AttachmentListPageRequest) ([]media.Attachment, error) {
	var models []attachmentModel
	query := s.db.WithContext(ctx).Where(&attachmentModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		AssetID:     assetID.String(),
	})
	if page.AfterAttachmentID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterAttachmentID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]media.Attachment, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid attachment row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}
