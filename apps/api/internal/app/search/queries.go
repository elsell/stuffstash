package search

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) SearchAssets(ctx context.Context, input SearchAssetsInput) (SearchAssetsResult, error) {
	exists, err := a.deps.Tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if !exists {
		return SearchAssetsResult{}, apperrors.ErrNotFound
	}

	if err := a.deps.Authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventAuthorizationDenied, Message: "authorization denied", Fields: map[string]string{"tenant_id": input.TenantID.String(), "principal_id": input.Principal.ID.String()}})
		return SearchAssetsResult{}, err
	}

	query, ok := search.NewQuery(input.Query)
	if !ok {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}
	tagIDs := normalizeSearchTagIDs(input.TagIDs)
	if query.String() == "" && len(tagIDs) == 0 {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}
	mode, ok := search.NewMode(input.Mode)
	if !ok {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}
	lifecycleFilter, err := appsupport.LifecycleFilter(input.LifecycleState)
	if err != nil {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}
	checkoutFilter, err := searchCheckoutStateFilter(input.CheckoutState)
	if err != nil {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}
	customAssetTypeID, err := parseSearchCustomAssetTypeID(input.CustomAssetTypeID)
	if err != nil {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}
	limit := appsupport.PageLimit(a.deps.DefaultPageLimit, a.deps.MaxPageLimit, input.Limit)
	cursorScope := searchCursorScope(input.TenantID, input.InventoryIDs, query, mode, tagIDs, customAssetTypeID, lifecycleFilter, checkoutFilter)
	afterResultKey, err := appsupport.DecodePageCursor("search.assets", cursorScope, input.Cursor)
	if err != nil {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}

	candidateInventoryIDs, err := a.inventoryIDsForTenant(ctx, input.TenantID)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if len(input.InventoryIDs) > 0 {
		candidateInventoryIDs = intersectInventoryCandidates(candidateInventoryIDs, input.InventoryIDs)
	}
	inventoryIDs, err := a.deps.Authorizer.ListViewableInventoryIDs(ctx, input.Principal, input.TenantID, candidateInventoryIDs)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	if len(inventoryIDs) == 0 {
		if err := a.saveSearchAssetsReadAudit(ctx, input, inventoryIDs, limit, mode.String(), string(lifecycleFilter), string(checkoutFilter), customAssetTypeID.String(), 0); err != nil {
			return SearchAssetsResult{}, err
		}
		return SearchAssetsResult{Items: []ports.AssetSearchResult{}, Limit: limit}, nil
	}
	if a.deps.Search == nil {
		return SearchAssetsResult{}, apperrors.ErrInvalidInput
	}

	items, err := a.deps.Search.SearchAssets(ctx, input.TenantID, inventoryIDs, ports.AssetSearchPageRequest{
		Query:             query,
		Mode:              mode,
		TagIDs:            tagIDs,
		CustomAssetTypeID: customAssetTypeID,
		AfterResultKey:    afterResultKey,
		Limit:             limit + 1,
		LifecycleFilter:   lifecycleFilter,
		CheckoutFilter:    checkoutFilter,
	})
	if err != nil {
		return SearchAssetsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = appsupport.EncodePageCursor("search.assets", cursorScope, items[len(items)-1].CursorKey())
	}
	primaryPhotos, err := PrimaryImageAttachmentsForSearchResults(ctx, a.deps.Attachments, input.TenantID, items)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	items, err = WithAncestorPaths(ctx, a.deps.Assets, items)
	if err != nil {
		return SearchAssetsResult{}, err
	}

	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetsSearched,
		Message: "assets searched",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"principal_id":  input.Principal.ID.String(),
			"limit":         strconv.Itoa(limit),
			"mode":          mode.String(),
			"tag_filters":   strconv.Itoa(len(tagIDs)),
			"inventory_ids": strconv.Itoa(len(inventoryIDs)),
		},
	})
	if err := a.saveSearchAssetsReadAudit(ctx, input, inventoryIDs, limit, mode.String(), string(lifecycleFilter), string(checkoutFilter), customAssetTypeID.String(), len(items)); err != nil {
		return SearchAssetsResult{}, err
	}
	return SearchAssetsResult{
		AuthorizedInventoryIDs: inventoryIDs,
		Items:                  items,
		PrimaryPhotos:          primaryPhotos,
		Limit:                  limit,
		NextCursor:             nextCursor,
		HasMore:                hasMore,
	}, nil
}
