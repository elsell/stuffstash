package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) LabelInstance(ctx context.Context) (printing.InstanceID, bool, error) {
	var model labelInstanceModel
	err := s.db.WithContext(ctx).Where(&labelInstanceModel{Singleton: 1}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	return printing.InstanceID(model.InstanceID), err == nil, err
}
func (s Store) BootstrapLabelInstance(ctx context.Context, id printing.InstanceID) (printing.InstanceID, error) {
	if !printing.ValidOpaqueID(string(id)) {
		return "", ports.ErrConflict
	}
	model := labelInstanceModel{Singleton: 1, InstanceID: string(id)}
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&model).Error; err != nil {
		return "", err
	}
	value, found, err := s.LabelInstance(ctx)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ports.ErrConflict
	}
	return value, nil
}
func (s Store) LabelForAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (printing.Label, bool, error) {
	if tenantID == "" || inventoryID == "" || assetID == "" {
		return printing.Label{}, false, ports.ErrForbidden
	}
	var model labelModel
	err := s.db.WithContext(ctx).Where(&labelModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return printing.Label{}, false, nil
	}
	return model.value(), err == nil, err
}
func (s Store) LookupLabel(ctx context.Context, instance printing.InstanceID, id printing.LabelID) (printing.Label, bool, error) {
	if instance == "" || id == "" {
		return printing.Label{}, false, nil
	}
	var model labelModel
	err := s.db.WithContext(ctx).Where(&labelModel{ID: string(id), InstanceID: string(instance)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return printing.Label{}, false, nil
	}
	return model.value(), err == nil, err
}
func (s Store) ProvisionLabel(ctx context.Context, value printing.Label, record audit.Record) (printing.Label, bool, error) {
	if !printing.ValidOpaqueID(string(value.ID)) || record.TenantID.String() != value.TenantID || record.InventoryID.String() != value.InventoryID || record.TargetID != value.AssetID {
		return printing.Label{}, false, ports.ErrForbidden
	}
	var result printing.Label
	created := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockLabelAsset(tx, value.TenantID, value.InventoryID, value.AssetID); err != nil {
			return err
		}
		var instance labelInstanceModel
		if err := tx.Where(&labelInstanceModel{Singleton: 1, InstanceID: string(value.InstanceID)}).First(&instance).Error; err != nil {
			return err
		}
		model := labelModel{ID: string(value.ID), InstanceID: string(value.InstanceID), TenantID: value.TenantID, InventoryID: value.InventoryID, AssetID: value.AssetID, CreatedAt: value.CreatedAt}
		saved := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
		if saved.Error != nil {
			return saved.Error
		}
		if saved.RowsAffected == 1 {
			created = true
			result = model.value()
			return createAuditRecord(tx, record)
		}
		var existing labelModel
		err := tx.Where(&labelModel{TenantID: value.TenantID, InventoryID: value.InventoryID, AssetID: value.AssetID}).First(&existing).Error
		if err != nil {
			return err
		}
		result = existing.value()
		return nil
	})
	return result, created, err
}
func lockLabelAsset(tx *gorm.DB, tenantID, inventoryID, assetID string) error {
	if tenantID == "" || inventoryID == "" || assetID == "" {
		return ports.ErrForbidden
	}
	var model assetModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&assetModel{ID: assetID, TenantID: tenantID, InventoryID: inventoryID}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ErrForbidden
	}
	return err
}
func tombstoneAssetLabels(tx *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) error {
	if err := tx.Model(&labelModel{}).Where(&labelModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()}).Update("tombstoned", true).Error; err != nil {
		return err
	}
	return tx.Where(&labelRenderModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String()}).Delete(&labelRenderModel{}).Error
}
