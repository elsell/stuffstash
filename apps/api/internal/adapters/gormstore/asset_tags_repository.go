package gormstore

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) CreateAssetTag(ctx context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&inventoryModel{}).
			Where(&inventoryModel{ID: tag.InventoryID.String(), TenantID: tag.TenantID.String()}).
			Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return ports.ErrForbidden
		}
		model := assetTagModelFromDomain(tag)
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) SetAssetTags(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, tagIDs []assettag.ID, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&assetModel{}).Where(&assetModel{ID: assetID.String(), TenantID: tenantID.String(), InventoryID: inventoryID.String()}).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return ports.ErrForbidden
		}
		unique := map[string]struct{}{}
		for _, tagID := range tagIDs {
			if tagID.String() == "" {
				return ports.ErrForbidden
			}
			unique[tagID.String()] = struct{}{}
		}
		keys := make([]string, 0, len(unique))
		for tagID := range unique {
			keys = append(keys, tagID)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			var tagCount int64
			if err := tx.Model(&assetTagModel{}).
				Where(&assetTagModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), LifecycleState: assettag.LifecycleStateActive.String()}).
				Where(clause.IN{Column: clause.Column{Name: "id"}, Values: stringValues(keys)}).
				Count(&tagCount).Error; err != nil {
				return err
			}
			if tagCount != int64(len(keys)) {
				return ports.ErrForbidden
			}
		}
		if err := tx.Where(&assetTagAssignmentModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()}).Delete(&assetTagAssignmentModel{}).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, tagID := range keys {
			if err := tx.Create(&assetTagAssignmentModel{
				TenantID:    tenantID.String(),
				InventoryID: inventoryID.String(),
				AssetID:     assetID.String(),
				TagID:       tagID,
				CreatedAt:   now,
			}).Error; err != nil {
				return err
			}
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateAssetTag(ctx context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&assetTagModel{}).
			Where(&assetTagModel{ID: tag.ID.String(), TenantID: tag.TenantID.String(), InventoryID: tag.InventoryID.String(), Key: tag.Key.String(), LifecycleState: assettag.LifecycleStateActive.String()}).
			Updates(map[string]any{
				"display_name": tag.DisplayName.String(),
				"color":        tag.Color.String(),
				"updated_at":   tag.UpdatedAt,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ports.ErrForbidden
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateAssetTagLifecycle(ctx context.Context, tag assettag.Tag, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&assetTagModel{}).
			Where(&assetTagModel{ID: tag.ID.String(), TenantID: tag.TenantID.String(), InventoryID: tag.InventoryID.String(), Key: tag.Key.String(), LifecycleState: assettag.LifecycleStateActive.String()}).
			Updates(map[string]any{
				"lifecycle_state": tag.LifecycleState.String(),
				"updated_at":      tag.UpdatedAt,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ports.ErrForbidden
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) AssetTagByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, tagID assettag.ID) (assettag.Tag, bool, error) {
	var model assetTagModel
	err := s.db.WithContext(ctx).Where(&assetTagModel{
		ID:          tagID.String(),
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return assettag.Tag{}, false, nil
	}
	if err != nil {
		return assettag.Tag{}, false, err
	}
	tag, ok := model.toDomain()
	if !ok {
		return assettag.Tag{}, false, ports.ErrInvalidProviderInput
	}
	return tag, true, nil
}

func (s Store) AssetTagByKey(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, key assettag.Key) (assettag.Tag, bool, error) {
	var model assetTagModel
	err := s.db.WithContext(ctx).Where(&assetTagModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		Key:         key.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return assettag.Tag{}, false, nil
	}
	if err != nil {
		return assettag.Tag{}, false, err
	}
	tag, ok := model.toDomain()
	if !ok {
		return assettag.Tag{}, false, ports.ErrInvalidProviderInput
	}
	return tag, true, nil
}

func (s Store) ListAssetTags(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetTagPageRequest) ([]assettag.Tag, error) {
	query := s.db.WithContext(ctx).
		Where(&assetTagModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), LifecycleState: assettag.LifecycleStateActive.String()}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	if page.AfterTagID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterTagID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	var models []assetTagModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	tags := make([]assettag.Tag, 0, len(models))
	for _, model := range models {
		tag, ok := model.toDomain()
		if !ok {
			return nil, ports.ErrInvalidProviderInput
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (s Store) AssetTagsByAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) ([]assettag.Tag, error) {
	assigned, err := s.AssetTagsByAssets(ctx, tenantID, inventoryID, []asset.ID{assetID})
	if err != nil {
		return nil, err
	}
	return assigned[assetID], nil
}

func (s Store) AssetTagsByAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetIDs []asset.ID) (map[asset.ID][]assettag.Tag, error) {
	return s.assetTagsByAssetsInInventories(ctx, tenantID, []string{inventoryID.String()}, assetIDs)
}

func (s Store) assetTagsByAssetsInInventories(ctx context.Context, tenantID tenant.ID, inventoryIDValues []string, assetIDs []asset.ID) (map[asset.ID][]assettag.Tag, error) {
	out := map[asset.ID][]assettag.Tag{}
	if len(inventoryIDValues) == 0 || len(assetIDs) == 0 {
		return out, nil
	}
	ids := make([]string, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		if assetID.String() == "" {
			continue
		}
		ids = append(ids, assetID.String())
		out[assetID] = []assettag.Tag{}
	}
	if len(ids) == 0 {
		return out, nil
	}
	var assignments []assetTagAssignmentModel
	if err := s.db.WithContext(ctx).
		Where(&assetTagAssignmentModel{TenantID: tenantID.String()}).
		Where(clause.IN{Column: clause.Column{Name: "inventory_id"}, Values: stringValues(inventoryIDValues)}).
		Where(clause.IN{Column: clause.Column{Name: "asset_id"}, Values: stringValues(ids)}).
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	tagIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		tagIDs = append(tagIDs, assignment.TagID)
	}
	if len(tagIDs) == 0 {
		return out, nil
	}
	var models []assetTagModel
	if err := s.db.WithContext(ctx).
		Where(&assetTagModel{TenantID: tenantID.String(), LifecycleState: assettag.LifecycleStateActive.String()}).
		Where(clause.IN{Column: clause.Column{Name: "inventory_id"}, Values: stringValues(inventoryIDValues)}).
		Where(clause.IN{Column: clause.Column{Name: "id"}, Values: stringValues(tagIDs)}).
		Find(&models).Error; err != nil {
		return nil, err
	}
	tagsByID := make(map[string]assettag.Tag, len(models))
	for _, model := range models {
		tag, ok := model.toDomain()
		if !ok {
			return nil, ports.ErrInvalidProviderInput
		}
		tagsByID[model.ID] = tag
	}
	for _, assignment := range assignments {
		tag, ok := tagsByID[assignment.TagID]
		if !ok {
			continue
		}
		out[asset.ID(assignment.AssetID)] = append(out[asset.ID(assignment.AssetID)], tag)
	}
	for assetID := range out {
		sort.Slice(out[assetID], func(i, j int) bool {
			return out[assetID][i].Key.String() < out[assetID][j].Key.String()
		})
	}
	return out, nil
}
