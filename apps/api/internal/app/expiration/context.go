package expiration

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// The caller has already authorized and read this scoped asset.
func (a ContextService) DescribeAsset(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, item asset.Asset) (*Description, error) {
	if item.Expiration.Value() == "" {
		return nil, nil
	}
	preferences, err := a.Preferences(ctx, inventoryID)
	if err != nil {
		return nil, err
	}
	if a.Types == nil || a.Clock == nil {
		return nil, ports.ErrInvalidProviderInput
	}
	kind, found, err := a.Types.CustomAssetTypeByID(ctx, tenantID, inventoryID, customfield.AssetTypeID(item.CustomAssetTypeID))
	if err != nil {
		return nil, err
	}
	enabled := found && kind.ExpirationEnabled && kind.LifecycleState == customfield.AssetTypeLifecycleActive && item.LifecycleState == asset.LifecycleStateActive
	value, err := Describe(item.Expiration, notification.AssetTypeID(item.CustomAssetTypeID), enabled, preferences.Settings, a.Clock.Now())
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// Inputs are the results of an authorized browse or search operation.
func (a ContextService) DescribeBrowse(ctx context.Context, items []asset.Asset) (map[ports.AttachmentAssetReference]Description, error) {
	result := make(map[ports.AttachmentAssetReference]Description)
	groups := make(map[inventory.InventoryID][]asset.Asset)
	for _, item := range items {
		if item.Expiration.Value() != "" {
			groups[inventory.InventoryID(item.InventoryID)] = append(groups[inventory.InventoryID(item.InventoryID)], item)
		}
	}
	for inventoryID, values := range groups {
		settings, err := a.Preferences(ctx, inventoryID)
		if err != nil {
			return nil, err
		}
		if a.Types == nil || a.Clock == nil {
			return nil, ports.ErrInvalidProviderInput
		}
		descriptions, err := DescribeItems(ctx, values, settings.Settings, a.Types, a.Clock.Now())
		if err != nil {
			return nil, err
		}
		for key, value := range descriptions {
			result[key] = value
		}
	}
	return result, nil
}

// Preferences is bound to the authorized caller by the composition root.
type ContextService struct {
	Preferences func(context.Context, inventory.InventoryID) (ports.NotificationPreferencesRecord, error)
	Types       ports.CustomAssetTypeRepository
	Clock       ports.Clock
}
