package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type AssetSearchRepository interface {
	SearchAssets(ctx context.Context, tenantID tenant.ID, inventoryIDs []inventory.InventoryID, page AssetSearchPageRequest) ([]AssetSearchResult, error)
}

type AssetSearchPageRequest struct {
	Query             search.Query
	Mode              search.Mode
	CustomAssetTypeID asset.CustomAssetTypeID
	AfterResultKey    string
	Limit             int
	LifecycleFilter   AssetLifecycleFilter
}

type AssetSearchResult struct {
	Type      search.ResultType
	TenantID  tenant.ID
	Inventory inventory.Inventory
	Asset     asset.Asset
	Matches   []search.Match
}

func (r AssetSearchResult) CursorKey() string {
	return r.Inventory.ID.String() + ":" + r.Asset.ID.String()
}
