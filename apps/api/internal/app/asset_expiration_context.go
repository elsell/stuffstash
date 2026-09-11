package app

import (
	"context"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// The caller has already authorized and read this scoped asset.
func (a App) describeAssetExpiration(ctx context.Context, scope notificationapp.ScopeInput, item asset.Asset) (*expirationapp.Description, error) {
	if item.Expiration.Value() == "" {
		return nil, nil
	}
	preferences, err := a.notificationService.GetPreferences(ctx, scope)
	if err != nil {
		return nil, err
	}
	if a.customAssetTypes == nil || a.clock == nil {
		return nil, ports.ErrInvalidProviderInput
	}
	kind, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, scope.TenantID, scope.InventoryID, customfield.AssetTypeID(item.CustomAssetTypeID))
	if err != nil {
		return nil, err
	}
	enabled := found && kind.ExpirationEnabled && kind.LifecycleState == customfield.AssetTypeLifecycleActive && item.LifecycleState == asset.LifecycleStateActive
	value, err := expirationapp.Describe(item.Expiration, notification.AssetTypeID(item.CustomAssetTypeID), enabled, preferences.Settings, a.clock.Now())
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// Inputs are the results of an authorized browse or search operation.
func (a App) describeBrowseExpiration(ctx context.Context, scope notificationapp.ScopeInput, items []asset.Asset) (map[ports.AttachmentAssetReference]expirationapp.Description, error) {
	result := make(map[ports.AttachmentAssetReference]expirationapp.Description)
	groups := make(map[inventory.InventoryID][]asset.Asset)
	for _, item := range items {
		if item.Expiration.Value() != "" {
			groups[inventory.InventoryID(item.InventoryID)] = append(groups[inventory.InventoryID(item.InventoryID)], item)
		}
	}
	for inventoryID, values := range groups {
		scope.InventoryID = inventoryID
		settings, err := a.notificationService.GetPreferences(ctx, scope)
		if err != nil {
			return nil, err
		}
		if a.customAssetTypes == nil || a.clock == nil {
			return nil, ports.ErrInvalidProviderInput
		}
		descriptions, err := expirationapp.DescribeItems(ctx, values, settings.Settings, a.customAssetTypes, a.clock.Now())
		if err != nil {
			return nil, err
		}
		for key, value := range descriptions {
			result[key] = value
		}
	}
	return result, nil
}
