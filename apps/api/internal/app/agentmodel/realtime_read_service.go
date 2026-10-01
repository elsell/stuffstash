package agentmodel

import (
	"context"
	"errors"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/app/audithistory"
	customfieldapp "github.com/stuffstash/stuff-stash/internal/app/customfields"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	searchapp "github.com/stuffstash/stuff-stash/internal/app/search"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

var ErrRealtimeVoiceToolCallTimedOut = errors.New("realtime voice tool call timed out")

type RealtimeReadToolScope struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
}

type RealtimeToolSearchResult struct {
	Items   []ports.AssetSearchResult
	HasMore bool
}

// Implementations retain application authorization, enrichment and read audit behavior.
type RealtimeReadToolQueries interface {
	SearchAssets(context.Context, searchapp.SearchAssetsInput) (RealtimeToolSearchResult, error)
	ListAssets(context.Context, assetapp.ListAssetsInput) (assetapp.ListAssetsResult, error)
	GetAsset(context.Context, assetapp.GetAssetInput) (asset.Asset, error)
	GetAssetDetail(context.Context, assetapp.GetAssetInput) (assetapp.GetAssetResult, error)
	GetInventory(context.Context, inventoryapp.GetInventoryInput) (inventory.Inventory, error)
	ListCheckedOutAssets(context.Context, assetapp.ListCheckedOutAssetsInput) (assetapp.CheckedOutAssetsResult, error)
	ListAssetCheckoutHistory(context.Context, assetapp.ListAssetCheckoutHistoryInput) (assetapp.AssetCheckoutHistoryResult, error)
	ListAssetAuditHistory(context.Context, audithistory.ListAssetAuditHistoryInput) (audithistory.ListAssetAuditHistoryResult, error)
	ListInventoryCustomAssetTypes(context.Context, customfieldapp.ListCustomAssetTypesInput) (customfieldapp.ListCustomAssetTypesResult, error)
	ListInventoryCustomFieldDefinitions(context.Context, customfieldapp.ListCustomFieldDefinitionsInput) (customfieldapp.ListCustomFieldDefinitionsResult, error)
	ListAssetTags(context.Context, assetapp.ListAssetTagsInput) (assetapp.ListAssetTagsResult, error)
	EnsureInventoryAccessItem(context.Context, identity.Principal, tenant.ID, inventory.InventoryID, ports.InventoryPermission) (inventory.Inventory, error)
	EnsureRealtimeVoiceAccess(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error
	DescribeAssetExpiration(context.Context, notificationapp.ScopeInput, asset.Asset) (*expirationapp.Description, error)
}

type RealtimeReadPreferences interface {
	GetPreferences(context.Context, notificationapp.ScopeInput) (ports.NotificationPreferencesRecord, error)
}

type RealtimeReadToolDependencies struct {
	Queries         RealtimeReadToolQueries
	Assets          ports.AssetRepository
	Checkouts       ports.AssetCheckoutRepository
	AssetTags       ports.AssetTagRepository
	Preferences     RealtimeReadPreferences
	Clock           ports.Clock
	ToolCallTimeout time.Duration
}

type RealtimeReadTools struct {
	RealtimeReadToolQueries
	assets                       ports.AssetRepository
	checkouts                    ports.AssetCheckoutRepository
	assetTags                    ports.AssetTagRepository
	notificationService          RealtimeReadPreferences
	clock                        ports.Clock
	realtimeVoiceToolCallTimeout time.Duration
}

func NewRealtimeReadTools(deps RealtimeReadToolDependencies) RealtimeReadTools {
	return RealtimeReadTools{RealtimeReadToolQueries: deps.Queries, assets: deps.Assets,
		checkouts: deps.Checkouts, assetTags: deps.AssetTags, notificationService: deps.Preferences,
		clock: deps.Clock, realtimeVoiceToolCallTimeout: deps.ToolCallTimeout}
}
func realtimeReadToolSearchModes(RealtimeReadToolScope) []search.Mode {
	return []search.Mode{search.ModeFuzzy}
}
