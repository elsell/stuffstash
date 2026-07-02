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
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateAttachmentLifecycle(ctx context.Context, attachment media.Attachment, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing attachmentModel
		err := tx.Where(&attachmentModel{
			ID:          attachment.ID.String(),
			TenantID:    attachment.TenantID.String(),
			InventoryID: attachment.InventoryID.String(),
			AssetID:     attachment.AssetID.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.LifecycleState == attachment.LifecycleState.String() {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Update("lifecycle_state", attachment.LifecycleState.String()).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) DeleteAttachmentAndEnqueueBlobDeletion(ctx context.Context, eventID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID, auditRecord audit.Record) (media.Attachment, bool, error) {
	var deleted media.Attachment
	found := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing attachmentModel
		err := tx.Where(&attachmentModel{ID: attachmentID.String(), TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		item, ok := existing.toDomain()
		if !ok {
			return fmt.Errorf("invalid attachment row %q", existing.ID)
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		if err := tx.Create(&blobDeletionEventModel{
			ID:         eventID,
			StorageKey: existing.StorageKey,
		}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&existing).Error; err != nil {
			return err
		}
		deleted = item
		found = true
		return nil
	})
	return deleted, found, err
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
		TenantID:       tenantID.String(),
		InventoryID:    inventoryID.String(),
		AssetID:        assetID.String(),
		LifecycleState: media.LifecycleStateActive.String(),
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

func (s Store) FirstImageAttachmentsByAssets(ctx context.Context, tenantID tenant.ID, assets []ports.AttachmentAssetReference) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	if len(assets) == 0 {
		return map[ports.AttachmentAssetReference]media.Attachment{}, nil
	}
	ids := make([]string, 0, len(assets))
	inventoryIDs := make([]string, 0, len(assets))
	allowed := map[string]struct{}{}
	for _, item := range assets {
		if item.InventoryID.String() != "" && item.AssetID.String() != "" {
			ids = append(ids, item.AssetID.String())
			inventoryIDs = append(inventoryIDs, item.InventoryID.String())
			allowed[item.InventoryID.String()+":"+item.AssetID.String()] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return map[ports.AttachmentAssetReference]media.Attachment{}, nil
	}

	idValues := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		idValues = append(idValues, id)
	}
	inventoryIDValues := make([]interface{}, 0, len(inventoryIDs))
	for _, id := range inventoryIDs {
		inventoryIDValues = append(inventoryIDValues, id)
	}
	contentTypeValues := []interface{}{
		media.ContentTypeJPEG.String(),
		media.ContentTypePNG.String(),
		media.ContentTypeWEBP.String(),
	}

	var firstRows []struct {
		ID          string
		InventoryID string
		AssetID     string
	}
	err := s.db.WithContext(ctx).
		Model(&attachmentModel{}).
		Select("MIN(id) AS id, inventory_id, asset_id").
		Where(&attachmentModel{
			TenantID:       tenantID.String(),
			LifecycleState: media.LifecycleStateActive.String(),
		}).
		Where(clause.IN{Column: clause.Column{Name: "inventory_id"}, Values: inventoryIDValues}).
		Where(clause.IN{Column: clause.Column{Name: "asset_id"}, Values: idValues}).
		Where(clause.IN{Column: clause.Column{Name: "content_type"}, Values: contentTypeValues}).
		Group("inventory_id").
		Group("asset_id").
		Find(&firstRows).Error
	if err != nil {
		return nil, err
	}
	if len(firstRows) == 0 {
		return map[ports.AttachmentAssetReference]media.Attachment{}, nil
	}
	attachmentIDs := make([]interface{}, 0, len(firstRows))
	firstRefs := map[string]ports.AttachmentAssetReference{}
	for _, row := range firstRows {
		key := row.InventoryID + ":" + row.AssetID
		if _, ok := allowed[key]; !ok || row.ID == "" {
			continue
		}
		attachmentIDs = append(attachmentIDs, row.ID)
		firstRefs[row.ID] = ports.AttachmentAssetReference{
			InventoryID: inventory.InventoryID(row.InventoryID),
			AssetID:     asset.ID(row.AssetID),
		}
	}
	if len(attachmentIDs) == 0 {
		return map[ports.AttachmentAssetReference]media.Attachment{}, nil
	}

	var models []attachmentModel
	err = s.db.WithContext(ctx).
		Where(&attachmentModel{TenantID: tenantID.String()}).
		Where(clause.IN{Column: clause.Column{Name: "id"}, Values: attachmentIDs}).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	items := make(map[ports.AttachmentAssetReference]media.Attachment)
	for _, model := range models {
		ref, ok := firstRefs[model.ID]
		if !ok {
			continue
		}
		if _, exists := items[ref]; exists {
			continue
		}
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid attachment row %q", model.ID)
		}
		items[ref] = item
	}
	return items, nil
}
