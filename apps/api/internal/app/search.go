package app

import (
	"context"

	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	searchapp "github.com/stuffstash/stuff-stash/internal/app/search"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type SearchAssetsInput = searchapp.SearchAssetsInput

type SearchAssetsResult struct {
	ExpirationContexts map[ports.AttachmentAssetReference]expirationapp.Description
	Items              []ports.AssetSearchResult
	PrimaryPhotos      map[ports.AttachmentAssetReference]media.Attachment
	Limit              int
	NextCursor         *string
	HasMore            bool
}

func (a App) searchService() searchapp.Service {
	return searchapp.New(searchapp.Dependencies{
		Tenants:          a.tenants,
		Authorizer:       a.authorizer,
		Inventories:      a.inventories,
		Search:           a.search,
		Assets:           a.assets,
		Attachments:      a.attachments,
		Audit:            a.audit,
		IDs:              a.ids,
		Clock:            a.clock,
		Observer:         a.observer,
		DefaultPageLimit: a.defaultPageLimit,
		MaxPageLimit:     a.maxPageLimit,
	})
}

func (a App) SearchAssets(ctx context.Context, input SearchAssetsInput) (SearchAssetsResult, error) {
	result, err := a.searchService().SearchAssets(ctx, input)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	output := SearchAssetsResult{Items: result.Items, PrimaryPhotos: result.PrimaryPhotos, Limit: result.Limit, NextCursor: result.NextCursor, HasMore: result.HasMore}
	if len(result.AuthorizedInventoryIDs) == 0 {
		return output, nil
	}
	a.warmPrimarySmallThumbnails(ctx, searchapp.PrimaryPhotosForSearchResults(result.Items, result.PrimaryPhotos))
	items := make([]asset.Asset, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, item.Asset)
	}
	output.ExpirationContexts, err = a.describeBrowseExpiration(ctx, notificationapp.ScopeInput{Principal: input.Principal, TenantID: input.TenantID, Source: input.Source}, items)
	if err != nil {
		return SearchAssetsResult{}, err
	}
	return output, nil
}
func (a App) primaryImageAttachmentsForSearchResults(ctx context.Context, tenantID tenant.ID, items []ports.AssetSearchResult) (map[ports.AttachmentAssetReference]media.Attachment, error) {
	return searchapp.PrimaryImageAttachmentsForSearchResults(ctx, a.attachments, tenantID, items)
}
