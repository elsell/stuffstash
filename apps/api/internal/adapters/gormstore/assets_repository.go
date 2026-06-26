package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) CreateAsset(ctx context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return createAssetInTx(tx, item, auditRecord, undoableOperation)
	})
}

func createAssetInTx(tx *gorm.DB, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	var containingInventory inventoryModel
	err := tx.Where(&inventoryModel{
		ID:       item.InventoryID.String(),
		TenantID: item.TenantID.String(),
	}).First(&containingInventory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ErrForbidden
	}
	if err != nil {
		return err
	}

	if item.ParentAssetID.String() != "" {
		var parent assetModel
		err = tx.Where(&assetModel{
			ID:          item.ParentAssetID.String(),
			TenantID:    item.TenantID.String(),
			InventoryID: item.InventoryID.String(),
		}).First(&parent).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		parentKind, ok := asset.NewKind(parent.Kind)
		if !ok || !parentKind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive.String() || parent.ID == item.ID.String() {
			return ports.ErrForbidden
		}
		if err := rejectAssetContainmentCycle(tx, item.ID, parent); err != nil {
			return err
		}
	}
	if item.CustomAssetTypeID.String() != "" {
		var assetType customAssetTypeModel
		err = tx.Where(&customAssetTypeModel{
			ID:             item.CustomAssetTypeID.String(),
			TenantID:       item.TenantID.String(),
			LifecycleState: customfield.AssetTypeLifecycleActive.String(),
		}).Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: item.InventoryID.String()},
		)).First(&assetType).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
	}

	parentAssetID := stringPtrFromAssetID(item.ParentAssetID)
	customAssetTypeID := stringPtrFromCustomAssetTypeID(item.CustomAssetTypeID)
	customFields, err := json.Marshal(item.CustomFields.Values())
	if err != nil {
		return err
	}
	if err := tx.Create(&assetModel{
		ID:                item.ID.String(),
		TenantID:          item.TenantID.String(),
		InventoryID:       item.InventoryID.String(),
		ParentAssetID:     parentAssetID,
		CustomAssetTypeID: customAssetTypeID,
		Kind:              item.Kind.String(),
		Title:             item.Title.String(),
		Description:       item.Description.String(),
		CustomFields:      string(customFields),
		LifecycleState:    item.LifecycleState.String(),
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}).Error; err != nil {
		return err
	}

	if err := createAuditRecord(tx, auditRecord); err != nil {
		return err
	}
	return createUndoableOperation(tx, undoableOperation)
}

func (s Store) UpdateAsset(ctx context.Context, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return updateAssetInTx(tx, asset.Asset{}, item, auditRecords, undoableOperation)
	})
}

func updateAssetInTx(tx *gorm.DB, expectedCurrent asset.Asset, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) error {
	var existing assetModel
	err := tx.Where(&assetModel{
		ID:          item.ID.String(),
		TenantID:    item.TenantID.String(),
		InventoryID: item.InventoryID.String(),
	}).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ErrForbidden
	}
	if err != nil {
		return err
	}
	if expectedCurrent.ID.String() != "" {
		matches, err := assetModelMatchesExpected(existing, expectedCurrent)
		if err != nil {
			return err
		}
		if !matches {
			return ports.ErrConflict
		}
	}
	if existing.Kind != item.Kind.String() || existing.LifecycleState != item.LifecycleState.String() || existing.LifecycleState != asset.LifecycleStateActive.String() {
		return ports.ErrForbidden
	}
	if stringFromPtr(existing.CustomAssetTypeID) != item.CustomAssetTypeID.String() {
		return ports.ErrForbidden
	}

	if item.ParentAssetID.String() != "" {
		var parent assetModel
		err = tx.Where(&assetModel{
			ID:          item.ParentAssetID.String(),
			TenantID:    item.TenantID.String(),
			InventoryID: item.InventoryID.String(),
		}).First(&parent).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		parentKind, ok := asset.NewKind(parent.Kind)
		if !ok || !parentKind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive.String() || parent.ID == item.ID.String() {
			return ports.ErrForbidden
		}
		if err := rejectAssetContainmentCycle(tx, item.ID, parent); err != nil {
			return err
		}
	}

	customFields, err := json.Marshal(item.CustomFields.Values())
	if err != nil {
		return err
	}
	updates := map[string]any{
		"parent_asset_id":      stringPtrFromAssetID(item.ParentAssetID),
		"custom_asset_type_id": stringPtrFromCustomAssetTypeID(item.CustomAssetTypeID),
		"title":                item.Title.String(),
		"description":          item.Description.String(),
		"custom_fields":        string(customFields),
		"updated_at":           item.UpdatedAt,
	}
	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return err
	}
	for _, auditRecord := range auditRecords {
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
	}
	return createUndoableOperation(tx, undoableOperation)
}

func assetModelMatchesExpected(model assetModel, expected asset.Asset) (bool, error) {
	existing, ok := model.toDomain()
	if !ok {
		return false, ports.ErrInvalidProviderInput
	}
	return assetsEquivalentForStaleCheck(existing, expected), nil
}

func assetsEquivalentForStaleCheck(left asset.Asset, right asset.Asset) bool {
	return left.ID == right.ID &&
		left.TenantID == right.TenantID &&
		left.InventoryID == right.InventoryID &&
		left.ParentAssetID == right.ParentAssetID &&
		left.CustomAssetTypeID == right.CustomAssetTypeID &&
		left.Kind == right.Kind &&
		left.Title == right.Title &&
		left.Description == right.Description &&
		left.LifecycleState == right.LifecycleState &&
		left.CreatedAt.Equal(right.CreatedAt) &&
		left.UpdatedAt.Equal(right.UpdatedAt) &&
		left.CustomFields.Equal(right.CustomFields)
}

func (s Store) UpdateAssetLifecycle(ctx context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing assetModel
		err := tx.Where(&assetModel{
			ID:          item.ID.String(),
			TenantID:    item.TenantID.String(),
			InventoryID: item.InventoryID.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Kind != item.Kind.String() || existing.Title != item.Title.String() || existing.Description != item.Description.String() || stringFromPtr(existing.ParentAssetID) != item.ParentAssetID.String() || stringFromPtr(existing.CustomAssetTypeID) != item.CustomAssetTypeID.String() {
			return ports.ErrForbidden
		}
		var existingCustomFields map[string]any
		if err := json.Unmarshal([]byte(existing.CustomFields), &existingCustomFields); err != nil {
			return err
		}
		existingFields, ok := asset.NewCustomFields(existingCustomFields)
		if !ok || !existingFields.Equal(item.CustomFields) {
			return ports.ErrForbidden
		}
		if existing.LifecycleState == asset.LifecycleStateActive.String() && item.LifecycleState == asset.LifecycleStateArchived {
			hasActiveChildren, err := assetHasActiveChildren(tx, item.TenantID, item.InventoryID, item.ID)
			if err != nil {
				return err
			}
			if hasActiveChildren {
				return ports.ErrForbidden
			}
		} else if existing.LifecycleState == asset.LifecycleStateArchived.String() && item.LifecycleState == asset.LifecycleStateActive {
			if item.ParentAssetID.String() != "" {
				var parent assetModel
				err := tx.Where(&assetModel{
					ID:          item.ParentAssetID.String(),
					TenantID:    item.TenantID.String(),
					InventoryID: item.InventoryID.String(),
				}).First(&parent).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ports.ErrForbidden
				}
				if err != nil {
					return err
				}
				if parent.LifecycleState != asset.LifecycleStateActive.String() {
					return ports.ErrForbidden
				}
			}
		} else {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Updates(map[string]any{
			"lifecycle_state": item.LifecycleState.String(),
			"updated_at":      item.UpdatedAt,
		}).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		return createUndoableOperation(tx, undoableOperation)
	})
}

func (s Store) DeleteAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hasActiveChildren, err := assetHasActiveChildren(tx, asset.TenantID(tenantID.String()), asset.InventoryID(inventoryID.String()), assetID)
		if err != nil {
			return err
		}
		if hasActiveChildren {
			return ports.ErrForbidden
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		result := tx.Where(&assetModel{ID: assetID.String(), TenantID: tenantID.String(), InventoryID: inventoryID.String()}).Delete(&assetModel{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ports.ErrForbidden
		}
		return nil
	})
}

func (s Store) AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	var model assetModel
	err := s.db.WithContext(ctx).Where(&assetModel{
		ID:          assetID.String(),
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return asset.Asset{}, false, nil
	}
	if err != nil {
		return asset.Asset{}, false, err
	}
	item, ok := model.toDomain()
	if !ok {
		return asset.Asset{}, false, fmt.Errorf("invalid asset row %q", model.ID)
	}
	return item, true, nil
}

func (s Store) AssetHasActiveChildren(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (bool, error) {
	return assetHasActiveChildren(s.db.WithContext(ctx), asset.TenantID(tenantID.String()), asset.InventoryID(inventoryID.String()), assetID)
}

func (s Store) ListAssetsByInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	var models []assetModel
	query := s.db.WithContext(ctx).Where(&assetModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	})
	switch page.LifecycleFilter {
	case "", ports.AssetLifecycleFilterActive:
		query = query.Where(&assetModel{LifecycleState: asset.LifecycleStateActive.String()})
	case ports.AssetLifecycleFilterArchived:
		query = query.Where(&assetModel{LifecycleState: asset.LifecycleStateArchived.String()})
	case ports.AssetLifecycleFilterAll:
	default:
		return nil, ports.ErrForbidden
	}
	sort := page.Sort
	if sort == "" {
		sort = ports.AssetListSortIDAsc
	}
	if page.AfterAssetID.String() != "" && sort == ports.AssetListSortIDAsc {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterAssetID.String()})
	}
	if page.AfterAssetID.String() != "" && sort == ports.AssetListSortUpdatedDesc {
		query = query.Where(clause.Or(
			clause.Lt{Column: clause.Column{Name: "updated_at"}, Value: page.AfterUpdatedAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "updated_at"}, Value: page.AfterUpdatedAt},
				clause.Lt{Column: clause.Column{Name: "id"}, Value: page.AfterAssetID.String()},
			),
		))
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	switch sort {
	case ports.AssetListSortIDAsc:
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	case ports.AssetListSortUpdatedDesc:
		query = query.Order(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{Column: clause.Column{Name: "updated_at"}, Desc: true},
				{Column: clause.Column{Name: "id"}, Desc: true},
			},
		})
	default:
		return nil, ports.ErrForbidden
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]asset.Asset, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid asset row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func assetHasActiveChildren(db *gorm.DB, tenantID asset.TenantID, inventoryID asset.InventoryID, assetID asset.ID) (bool, error) {
	var count int64
	err := db.Model(&assetModel{}).Where(&assetModel{
		TenantID:       tenantID.String(),
		InventoryID:    inventoryID.String(),
		ParentAssetID:  stringPtr(assetID.String()),
		LifecycleState: asset.LifecycleStateActive.String(),
	}).Count(&count).Error
	return count > 0, err
}

func rejectAssetContainmentCycle(tx *gorm.DB, assetID asset.ID, parent assetModel) error {
	for current := parent; ; {
		if current.ID == assetID.String() {
			return ports.ErrForbidden
		}
		if current.ParentAssetID == nil {
			return nil
		}

		nextID := *current.ParentAssetID
		var next assetModel
		err := tx.Where(&assetModel{
			ID:          nextID,
			TenantID:    current.TenantID,
			InventoryID: current.InventoryID,
		}).First(&next).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		current = next
	}
}

func stringPtrFromAssetID(id asset.ID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}

func stringPtrFromCustomAssetTypeID(id asset.CustomAssetTypeID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}
