package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
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
	TagIDs            []assettag.ID
	CustomAssetTypeID asset.CustomAssetTypeID
	AfterResultKey    string
	Limit             int
	LifecycleFilter   AssetLifecycleFilter
	CheckoutFilter    AssetCheckoutStateFilter
}

type AssetCheckoutStateFilter string

const (
	AssetCheckoutStateFilterAny        AssetCheckoutStateFilter = "any"
	AssetCheckoutStateFilterCheckedOut AssetCheckoutStateFilter = "checked_out"
	AssetCheckoutStateFilterAvailable  AssetCheckoutStateFilter = "available"
)

type AssetSearchResult struct {
	Type            search.ResultType
	TenantID        tenant.ID
	Inventory       inventory.Inventory
	Asset           asset.Asset
	CurrentCheckout *asset.Checkout
	AssignedTags    []assettag.Tag
	Matches         []search.Match
}

func (r AssetSearchResult) CursorKey() string {
	return r.Inventory.ID.String() + ":" + r.Asset.ID.String()
}
