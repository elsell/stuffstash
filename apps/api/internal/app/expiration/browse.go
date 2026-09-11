package expiration

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

// DescribeItems enriches already authorized inventory assets with one type read per distinct type.
func DescribeItems(ctx context.Context, items []asset.Asset, settings notification.Settings, types ports.CustomAssetTypeRepository, now time.Time) (map[ports.AttachmentAssetReference]Description, error) {
	result := make(map[ports.AttachmentAssetReference]Description)
	enabledTypes := make(map[customfield.AssetTypeID]bool)
	for _, item := range items {
		if item.Expiration.Value() == "" {
			continue
		}
		id := customfield.AssetTypeID(item.CustomAssetTypeID)
		enabled, known := enabledTypes[id]
		if !known {
			kind, found, err := types.CustomAssetTypeByID(ctx, tenant.ID(item.TenantID), inventory.InventoryID(item.InventoryID), id)
			if err != nil {
				return nil, err
			}
			enabled = found && kind.IsActive() && kind.ExpirationEnabled
			enabledTypes[id] = enabled
		}
		value, err := Describe(item.Expiration, notification.AssetTypeID(id), enabled && item.LifecycleState == asset.LifecycleStateActive, settings, now)
		if err != nil {
			return nil, err
		}
		result[ports.AttachmentAssetReference{InventoryID: inventory.InventoryID(item.InventoryID), AssetID: item.ID}] = value
	}
	return result, nil
}
