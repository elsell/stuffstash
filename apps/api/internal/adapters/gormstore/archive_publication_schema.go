package gormstore

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func restoreArchiveSchema(tx *gorm.DB, d ports.InventoryExportDocument) error {
	for _, t := range d.CustomAssetTypes {
		if t.TenantID.String() != d.TenantID || t.InventoryID.String() != d.InventoryID || t.Scope != customfield.ScopeInventory {
			return ports.ErrArchiveJobScope
		}
		var count int64
		if err := tx.Model(&customAssetTypeModel{}).Where(clause.Eq{Column: "tenant_id", Value: d.TenantID}).Where(clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()}).Where(clause.Eq{Column: "type_key", Value: t.Key.String()}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return ports.ErrConflict
		}
		m := customAssetTypeModel{ID: t.ID.String(), TenantID: d.TenantID, InventoryID: &d.InventoryID, Scope: customfield.ScopeInventory.String(), TypeKey: t.Key.String(), DisplayName: t.DisplayName.String(), Description: t.Description.String(), LifecycleState: customfield.AssetTypeLifecycleActive.String(), ExpirationEnabled: t.ExpirationEnabled}
		if err := tx.Create(&m).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
	}
	for _, f := range d.CustomFieldDefinitions {
		if f.TenantID.String() != d.TenantID || f.InventoryID.String() != d.InventoryID || f.Scope != customfield.ScopeInventory {
			return ports.ErrArchiveJobScope
		}
		var count int64
		if err := tx.Model(&customFieldDefinitionModel{}).Where(clause.Eq{Column: "tenant_id", Value: d.TenantID}).Where(clause.Eq{Column: "scope", Value: customfield.ScopeTenant.String()}).Where(clause.Eq{Column: "field_key", Value: f.Key.String()}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return ports.ErrConflict
		}
		options, err := json.Marshal(customFieldKeysToStrings(f.EnumOptions))
		if err != nil {
			return err
		}
		m := customFieldDefinitionModel{ID: f.ID.String(), TenantID: d.TenantID, InventoryID: &d.InventoryID, Scope: customfield.ScopeInventory.String(), FieldKey: f.Key.String(), DisplayName: f.DisplayName.String(), FieldType: f.Type.String(), EnumOptions: string(options), Applicability: f.Applicability.String(), LifecycleState: f.LifecycleState.String()}
		if err = tx.Create(&m).Error; err != nil {
			return customFieldDefinitionWriteError(err)
		}
		for _, id := range f.CustomAssetTypeIDs {
			if err = tx.Create(&customFieldDefinitionAssetTypeModel{CustomFieldDefinitionID: f.ID.String(), CustomAssetTypeID: id.String(), TenantID: d.TenantID, InventoryID: &d.InventoryID}).Error; err != nil {
				return err
			}
		}
	}
	for _, t := range d.Tags {
		if t.TenantID.String() != d.TenantID || t.InventoryID.String() != d.InventoryID {
			return ports.ErrArchiveJobScope
		}
		m := assetTagModelFromDomain(t)
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
	}
	return nil
}

func restoreArchiveTypeLifecycles(tx *gorm.DB, d ports.InventoryExportDocument) error {
	for _, t := range d.CustomAssetTypes {
		if t.LifecycleState != customfield.AssetTypeLifecycleArchived {
			continue
		}
		if err := tx.Model(&customAssetTypeModel{}).Where(clause.Eq{Column: "tenant_id", Value: d.TenantID}).Where(clause.Eq{Column: "inventory_id", Value: d.InventoryID}).Where(clause.Eq{Column: "id", Value: t.ID.String()}).UpdateColumn("lifecycle_state", t.LifecycleState.String()).Error; err != nil {
			return err
		}
	}
	return nil
}
