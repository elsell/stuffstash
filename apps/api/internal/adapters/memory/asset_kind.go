package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) validateAssetKindLocked(item asset.Asset) error {
	if _, ok := asset.NewKind(item.Kind.String()); !ok {
		return ports.ErrForbidden
	}
	if item.Kind == asset.KindItem {
		for _, child := range s.assets {
			if child.TenantID == item.TenantID && child.InventoryID == item.InventoryID && child.ParentAssetID == item.ID {
				return ports.ErrAssetKindHasContents
			}
		}
	}
	if item.Kind == asset.KindLocation {
		for _, checkout := range s.checkouts {
			if checkout.TenantID == item.TenantID && checkout.InventoryID == item.InventoryID && checkout.AssetID == item.ID && checkout.State == asset.CheckoutStateOpen {
				return ports.ErrAssetKindHasCheckout
			}
		}
	}
	return nil
}
