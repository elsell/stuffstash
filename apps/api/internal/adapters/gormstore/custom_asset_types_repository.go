package gormstore

import (
	"context"
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

func (s Store) SaveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	model := customAssetTypeModel{
		ID:             assetType.ID.String(),
		TenantID:       assetType.TenantID.String(),
		Scope:          assetType.Scope.String(),
		TypeKey:        assetType.Key.String(),
		DisplayName:    assetType.DisplayName.String(),
		Description:    assetType.Description.String(),
		LifecycleState: assetType.LifecycleState.String(),
	}
	if assetType.InventoryID.String() != "" {
		inventoryID := assetType.InventoryID.String()
		model.InventoryID = &inventoryID
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customAssetTypeModel
		query := tx.Where(&customAssetTypeModel{
			TenantID: assetType.TenantID.String(),
			TypeKey:  assetType.Key.String(),
		})
		if assetType.Scope == customfield.ScopeInventory {
			query = query.Where(clause.Or(
				clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
				clause.Eq{Column: "inventory_id", Value: assetType.InventoryID.String()},
			))
		}
		err := query.First(&existing).Error
		if err == nil {
			return ports.ErrConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Save(&model).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customAssetTypeModel
		err := tx.Where(&customAssetTypeModel{
			ID:             assetType.ID.String(),
			TenantID:       assetType.TenantID.String(),
			LifecycleState: customfield.AssetTypeLifecycleActive.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Scope != assetType.Scope.String() || existing.TypeKey != assetType.Key.String() || stringFromPtr(existing.InventoryID) != assetType.InventoryID.String() {
			return ports.ErrForbidden
		}

		updates := map[string]any{
			"display_name": assetType.DisplayName.String(),
			"description":  assetType.Description.String(),
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ArchiveCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	if assetType.LifecycleState != customfield.AssetTypeLifecycleArchived {
		return ports.ErrForbidden
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customAssetTypeModel
		err := tx.Where(&customAssetTypeModel{
			ID:             assetType.ID.String(),
			TenantID:       assetType.TenantID.String(),
			LifecycleState: customfield.AssetTypeLifecycleActive.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Scope != assetType.Scope.String() || existing.TypeKey != assetType.Key.String() || stringFromPtr(existing.InventoryID) != assetType.InventoryID.String() {
			return ports.ErrForbidden
		}

		if err := tx.Model(&existing).Update("lifecycle_state", assetType.LifecycleState.String()).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) RestoreCustomAssetType(ctx context.Context, assetType customfield.AssetType, auditRecord audit.Record) error {
	if assetType.LifecycleState != customfield.AssetTypeLifecycleActive {
		return ports.ErrForbidden
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing customAssetTypeModel
		err := tx.Where(&customAssetTypeModel{
			ID:             assetType.ID.String(),
			TenantID:       assetType.TenantID.String(),
			LifecycleState: customfield.AssetTypeLifecycleArchived.String(),
		}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Scope != assetType.Scope.String() || existing.TypeKey != assetType.Key.String() || stringFromPtr(existing.InventoryID) != assetType.InventoryID.String() {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Update("lifecycle_state", assetType.LifecycleState.String()).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) DeleteCustomAssetType(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		refs, err := customAssetTypeHasActiveReferences(tx, tenantID, inventoryID, assetTypeID)
		if err != nil {
			return err
		}
		if refs {
			return ports.ErrForbidden
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		result := scopedCustomAssetTypeQuery(tx, tenantID, inventoryID, assetTypeID).Delete(&customAssetTypeModel{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ports.ErrForbidden
		}
		return nil
	})
}

func (s Store) CustomAssetTypeHasActiveReferences(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (bool, error) {
	return customAssetTypeHasActiveReferences(s.db.WithContext(ctx), tenantID, inventoryID, assetTypeID)
}

func customAssetTypeHasActiveReferences(db *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (bool, error) {
	var assetCount int64
	rawAssetTypeID := assetTypeID.String()
	assetQuery := db.Model(&assetModel{}).Where(&assetModel{
		TenantID:          tenantID.String(),
		LifecycleState:    asset.LifecycleStateActive.String(),
		CustomAssetTypeID: &rawAssetTypeID,
	})
	if inventoryID.String() != "" {
		assetQuery = assetQuery.Where(&assetModel{InventoryID: inventoryID.String()})
	}
	if err := assetQuery.Count(&assetCount).Error; err != nil {
		return false, err
	}
	if assetCount > 0 {
		return true, nil
	}
	var targetCount int64
	if err := db.Model(&customFieldDefinitionAssetTypeModel{}).Where(&customFieldDefinitionAssetTypeModel{TenantID: tenantID.String(), CustomAssetTypeID: assetTypeID.String()}).Count(&targetCount).Error; err != nil {
		return false, err
	}
	return targetCount > 0, nil
}

func (s Store) CustomAssetTypeByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) (customfield.AssetType, bool, error) {
	query := scopedCustomAssetTypeQuery(s.db.WithContext(ctx), tenantID, inventoryID, assetTypeID)
	var model customAssetTypeModel
	err := query.First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return customfield.AssetType{}, false, nil
	}
	if err != nil {
		return customfield.AssetType{}, false, err
	}
	assetType, ok := model.toDomain()
	if !ok {
		return customfield.AssetType{}, false, fmt.Errorf("invalid custom asset type row %q", model.ID)
	}
	return assetType, true, nil
}

func scopedCustomAssetTypeQuery(db *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID, assetTypeID customfield.AssetTypeID) *gorm.DB {
	query := db.Where(&customAssetTypeModel{ID: assetTypeID.String(), TenantID: tenantID.String()})
	if inventoryID.String() == "" {
		return query.Where(&customAssetTypeModel{Scope: customfield.ScopeTenant.String()})
	}
	return query.Where(clause.Or(
		clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
		clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
	))
}

func (s Store) ListTenantCustomAssetTypes(ctx context.Context, tenantID tenant.ID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	query := s.db.WithContext(ctx).Where(&customAssetTypeModel{
		TenantID:       tenantID.String(),
		Scope:          customfield.ScopeTenant.String(),
		LifecycleState: customfield.AssetTypeLifecycleActive.String(),
	})
	return s.listCustomAssetTypes(query, page)
}

func (s Store) ListInventoryCustomAssetTypes(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	query := s.db.WithContext(ctx).
		Where(&customAssetTypeModel{TenantID: tenantID.String(), LifecycleState: customfield.AssetTypeLifecycleActive.String()}).
		Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	return s.listCustomAssetTypes(query, page)
}

func (s Store) CustomAssetTypesByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, ids []customfield.AssetTypeID) ([]customfield.AssetType, error) {
	rawIDs := customAssetTypeIDsToStrings(ids)
	if len(rawIDs) == 0 {
		return nil, nil
	}
	query := s.db.WithContext(ctx).
		Where(&customAssetTypeModel{TenantID: tenantID.String(), LifecycleState: customfield.AssetTypeLifecycleActive.String()}).
		Where(clause.IN{Column: clause.Column{Name: "id"}, Values: stringValues(rawIDs)}).
		Where(clause.Or(
			clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()},
			clause.Eq{Column: "inventory_id", Value: inventoryID.String()},
		))
	return s.listCustomAssetTypes(query, ports.CustomAssetTypePageRequest{})
}

func (s Store) listCustomAssetTypes(query *gorm.DB, page ports.CustomAssetTypePageRequest) ([]customfield.AssetType, error) {
	var models []customAssetTypeModel
	if page.AfterAssetTypeKey != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "cursor_key"}, Value: page.AfterAssetTypeKey})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "cursor_key"}}).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]customfield.AssetType, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid custom asset type row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func customAssetTypeIDsToStrings(ids []customfield.AssetTypeID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}
	return values
}
