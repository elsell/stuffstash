package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) realtimeReadTools() agentapp.RealtimeReadTools {
	return agentapp.NewRealtimeReadTools(agentapp.RealtimeReadToolDependencies{
		Queries: realtimeToolQueries{a}, Assets: a.assets, Checkouts: a.checkouts, AssetTags: a.assetTags,
		Preferences: a.notificationService, Clock: a.clock, ToolCallTimeout: a.realtimeVoiceToolCallTimeout,
	})
}
func realtimeReadScope(session RealtimeVoiceSession) agentapp.RealtimeReadToolScope {
	return agentapp.RealtimeReadToolScope{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID}
}

// Embed existing scoped query entrypoints; only adapt envelopes and private wiring.
type realtimeToolQueries struct{ App }

func (q realtimeToolQueries) SearchAssets(ctx context.Context, input SearchAssetsInput) (agentapp.RealtimeToolSearchResult, error) {
	result, err := q.App.SearchAssets(ctx, input)
	return agentapp.RealtimeToolSearchResult{Items: result.Items, HasMore: result.HasMore}, err
}
func (q realtimeToolQueries) GetAssetDetail(ctx context.Context, input GetAssetInput) (assetapp.GetAssetResult, error) {
	result, err := q.App.GetAssetDetail(ctx, input)
	return result.GetAssetResult, err
}
func (q realtimeToolQueries) EnsureInventoryAccessItem(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (inventory.Inventory, error) {
	return q.App.ensureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
}
func (q realtimeToolQueries) EnsureRealtimeVoiceAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return q.App.ensureRealtimeVoiceAccess(ctx, principal, tenantID, inventoryID)
}
func (q realtimeToolQueries) DescribeAssetExpiration(ctx context.Context, scope notificationapp.ScopeInput, item asset.Asset) (*expirationapp.Description, error) {
	return q.App.describeAssetExpiration(ctx, scope, item)
}

var _ agentapp.RealtimeReadToolQueries = realtimeToolQueries{}
