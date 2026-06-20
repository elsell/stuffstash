package gormstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveInventory(ctx context.Context, item inventory.Inventory) error {
	model := inventoryModel{
		ID:             item.ID.String(),
		TenantID:       item.TenantID.String(),
		Name:           item.Name.String(),
		LifecycleState: lifecycleStateOrActive(item.LifecycleState.String()),
	}

	return s.db.WithContext(ctx).Save(&model).Error
}

func (s Store) InventoryByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	var model inventoryModel
	err := s.db.WithContext(ctx).Where(&inventoryModel{
		ID:       inventoryID.String(),
		TenantID: tenantID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return inventory.Inventory{}, false, nil
	}
	if err != nil {
		return inventory.Inventory{}, false, err
	}
	item, ok := model.toDomain()
	return item, ok, nil
}

func (s Store) UpdateInventory(ctx context.Context, item inventory.Inventory, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing inventoryModel
		err := tx.Where(&inventoryModel{ID: item.ID.String(), TenantID: item.TenantID.String()}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.LifecycleState != inventory.LifecycleStateActive.String() || item.LifecycleState != inventory.LifecycleStateActive {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Update("name", item.Name.String()).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateInventoryLifecycle(ctx context.Context, item inventory.Inventory, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing inventoryModel
		err := tx.Where(&inventoryModel{ID: item.ID.String(), TenantID: item.TenantID.String()}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		if existing.Name != item.Name.String() || existing.LifecycleState == item.LifecycleState.String() {
			return ports.ErrForbidden
		}
		if err := tx.Model(&existing).Update("lifecycle_state", item.LifecycleState.String()).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) DeleteInventory(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hasAssets, err := inventoryHasAssets(tx, tenantID, inventoryID)
		if err != nil {
			return err
		}
		if hasAssets {
			return ports.ErrForbidden
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		result := tx.Where(&inventoryModel{ID: inventoryID.String(), TenantID: tenantID.String()}).Delete(&inventoryModel{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ports.ErrForbidden
		}
		return nil
	})
}

func (s Store) InventoryHasActiveAssets(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (bool, error) {
	return inventoryHasActiveAssets(s.db.WithContext(ctx), tenantID, inventoryID)
}

func inventoryHasActiveAssets(db *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID) (bool, error) {
	var count int64
	if err := db.Model(&assetModel{}).Where(&assetModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), LifecycleState: asset.LifecycleStateActive.String()}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func inventoryHasAssets(db *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID) (bool, error) {
	var count int64
	if err := db.Model(&assetModel{}).Where(&assetModel{TenantID: tenantID.String(), InventoryID: inventoryID.String()}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s Store) SaveInventoryAndEnqueueOwnerGrant(ctx context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&inventoryModel{
			ID:             item.ID.String(),
			TenantID:       item.TenantID.String(),
			Name:           item.Name.String(),
			LifecycleState: lifecycleStateOrActive(item.LifecycleState.String()),
		}).Error; err != nil {
			return err
		}

		inventoryID := item.ID.String()
		if err := tx.Create(&authorizationOutboxEventModel{
			ID:          eventID,
			Kind:        string(ports.AuthorizationOutboxGrantInventoryOwner),
			PrincipalID: principal.ID.String(),
			TenantID:    tenantID.String(),
			InventoryID: &inventoryID,
		}).Error; err != nil {
			return err
		}

		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ListInventoriesByTenant(ctx context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	var models []inventoryModel
	query := s.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String(), LifecycleState: inventory.LifecycleStateActive.String()})
	if page.AfterInventoryID.String() != "" {
		query = query.Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterInventoryID.String()})
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]inventory.Inventory, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid inventory row %q", model.ID)
		}
		items = append(items, item)
	}

	return items, nil
}
