package app

import (
	"context"
	"strconv"
	"strings"

	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type SearchAssetsInput struct {
	Principal         identity.Principal
	TenantID          tenant.ID
	Query             string
	Mode              string
	CustomAssetTypeID string
	LifecycleState    string
	Limit             int
	Cursor            string
}

type SearchAssetsResult struct {
	Items      []ports.AssetSearchResult
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (a App) SearchAssets(ctx context.Context, input SearchAssetsInput) (SearchAssetsResult, error) {
	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if !exists {
		return SearchAssetsResult{}, ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return SearchAssetsResult{}, err
	}

	query, ok := search.NewQuery(input.Query)
	if !ok {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	mode, ok := search.NewMode(input.Mode)
	if !ok {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	lifecycleFilter, err := assetapp.LifecycleFilter(input.LifecycleState)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	customAssetTypeID, err := parseSearchCustomAssetTypeID(input.CustomAssetTypeID)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}
	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	cursorScope := searchCursorScope(input.TenantID, query, mode, customAssetTypeID, lifecycleFilter)
	afterResultKey, err := decodePageCursor("search.assets", cursorScope, input.Cursor)
	if err != nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}

	candidateInventoryIDs, err := a.inventoryIDsForTenant(ctx, input.TenantID)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	inventoryIDs, err := a.authorizer.ListViewableInventoryIDs(ctx, input.Principal, input.TenantID, candidateInventoryIDs)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if len(inventoryIDs) == 0 {
		return SearchAssetsResult{Items: []ports.AssetSearchResult{}, Limit: limit}, nil
	}
	if a.search == nil {
		return SearchAssetsResult{}, ErrInvalidInput
	}

	items, err := a.search.SearchAssets(ctx, input.TenantID, inventoryIDs, ports.AssetSearchPageRequest{
		Query:             query,
		Mode:              mode,
		CustomAssetTypeID: customAssetTypeID,
		AfterResultKey:    afterResultKey,
		Limit:             limit + 1,
		LifecycleFilter:   lifecycleFilter,
	})
	if err != nil {
		return SearchAssetsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodePageCursor("search.assets", cursorScope, items[len(items)-1].CursorKey())
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetsSearched,
		Message: "assets searched",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"principal_id":  input.Principal.ID.String(),
			"limit":         strconv.Itoa(limit),
			"mode":          mode.String(),
			"inventory_ids": strconv.Itoa(len(inventoryIDs)),
		},
	})

	return SearchAssetsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (a App) inventoryIDsForTenant(ctx context.Context, tenantID tenant.ID) ([]inventory.InventoryID, error) {
	items, err := a.inventories.ListInventoriesByTenant(ctx, inventory.TenantID(tenantID.String()), ports.InventoryListPageRequest{})
	if err != nil {
		return nil, err
	}

	ids := make([]inventory.InventoryID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func parseSearchCustomAssetTypeID(raw string) (asset.CustomAssetTypeID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	id, ok := asset.NewCustomAssetTypeID(raw)
	if !ok {
		return "", ErrInvalidInput
	}
	return id, nil
}

func searchCursorScope(tenantID tenant.ID, query search.Query, mode search.Mode, customAssetTypeID asset.CustomAssetTypeID, lifecycleFilter ports.AssetLifecycleFilter) string {
	return strings.Join([]string{
		tenantID.String(),
		query.String(),
		mode.String(),
		customAssetTypeID.String(),
		string(lifecycleFilter),
	}, ":")
}
