package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

// The caller holds the asset row lock, also acquired by containment and checkout writers.
func validateAssetKindInTx(tx *gorm.DB, item asset.Asset) error {
	if _, ok := asset.NewKind(item.Kind.String()); !ok {
		return ports.ErrForbidden
	}
	var count int64
	if item.Kind == asset.KindItem {
		if err := tx.Model(&assetModel{}).Where(&assetModel{TenantID: item.TenantID.String(), InventoryID: item.InventoryID.String(), ParentAssetID: stringPtr(item.ID.String())}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ports.ErrAssetKindHasContents
		}
	}
	if item.Kind == asset.KindLocation {
		if err := tx.Model(&assetCheckoutModel{}).Where(&assetCheckoutModel{TenantID: item.TenantID.String(), InventoryID: item.InventoryID.String(), AssetID: item.ID.String(), State: asset.CheckoutStateOpen.String()}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ports.ErrAssetKindHasCheckout
		}
	}
	return nil
}
