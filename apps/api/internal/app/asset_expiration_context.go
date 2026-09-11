package app

import (
	"context"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
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
