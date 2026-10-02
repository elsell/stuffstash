package gormstore

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func restoreArchiveAssets(tx *gorm.DB, d ports.InventoryExportDocument, principal string) error {
	for _, a := range d.Assets {
		values, err := json.Marshal(a.CustomFields)
		if err != nil {
			return err
		}
		if a.CustomFields == nil {
			values = []byte("{}")
		}
		m := assetModel{ID: a.ID, TenantID: d.TenantID, InventoryID: d.InventoryID, Kind: a.Kind, Title: a.Title, Description: a.Description, CustomFields: string(values), LifecycleState: a.LifecycleState, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt, ExpirationDate: a.ExpirationDate, ExpirationPrecision: a.ExpirationPrecision}
		if a.CustomAssetTypeID != "" {
			m.CustomAssetTypeID = &a.CustomAssetTypeID
		}
		if err = tx.Create(&m).Error; err != nil {
			return err
		}
	}
	for _, a := range d.Assets {
		if a.ParentAssetID != "" {
			// UpdateColumn avoids replacing the preserved source updated_at timestamp.
			if err := tx.Model(&assetModel{}).Where(clause.Eq{Column: "tenant_id", Value: d.TenantID}).Where(clause.Eq{Column: "inventory_id", Value: d.InventoryID}).Where(clause.Eq{Column: "id", Value: a.ID}).UpdateColumn("parent_asset_id", a.ParentAssetID).Error; err != nil {
				return err
			}
		}
		for _, id := range a.TagIDs {
			if err := tx.Create(&assetTagAssignmentModel{TenantID: d.TenantID, InventoryID: d.InventoryID, AssetID: a.ID, TagID: id}).Error; err != nil {
				return err
			}
		}
		for _, m := range a.Attachments {
			key := d.TenantID + "/" + d.InventoryID + "/" + a.ID + "/" + m.ID.String()
			if m.TenantID.String() != d.TenantID || m.InventoryID.String() != d.InventoryID || m.AssetID.String() != a.ID || m.StorageKey.String() != key {
				return ports.ErrArchiveJobScope
			}
			if err := reserveMediaBlobKey(tx, m.StorageKey); err != nil {
				return err
			}
			row := attachmentModel{ID: m.ID.String(), TenantID: d.TenantID, InventoryID: d.InventoryID, AssetID: a.ID, StorageKey: key, FileName: m.FileName.String(), ContentType: m.ContentType.String(), SHA256: m.SHA256.String(), SizeBytes: m.SizeBytes, CreatedAt: m.CreatedAt, LifecycleState: m.LifecycleState.String()}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		if c := a.CurrentCheckout; c != nil {
			if c.TenantID.String() != d.TenantID || c.InventoryID.String() != d.InventoryID || c.AssetID.String() != a.ID || c.CheckedOutByPrincipal != principal {
				return ports.ErrArchiveJobScope
			}
			row := newAssetCheckoutModel(*c)
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
