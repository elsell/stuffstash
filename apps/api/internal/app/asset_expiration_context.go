package app

import (
	"context"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) expirationContext(scope notificationapp.ScopeInput) expirationapp.ContextService {
	return expirationapp.ContextService{
		Types: a.customAssetTypes, Clock: a.clock,
		Preferences: func(ctx context.Context, inventoryID inventory.InventoryID) (ports.NotificationPreferencesRecord, error) {
			scoped := scope
			scoped.InventoryID = inventoryID
			return a.notificationService.GetPreferences(ctx, scoped)
		},
	}
}
func (a App) describeAssetExpiration(ctx context.Context, scope notificationapp.ScopeInput, item asset.Asset) (*expirationapp.Description, error) {
	return a.expirationContext(scope).DescribeAsset(ctx, scope.TenantID, scope.InventoryID, item)
}
func (a App) describeBrowseExpiration(ctx context.Context, scope notificationapp.ScopeInput, items []asset.Asset) (map[ports.AttachmentAssetReference]expirationapp.Description, error) {
	return a.expirationContext(scope).DescribeBrowse(ctx, items)
}
