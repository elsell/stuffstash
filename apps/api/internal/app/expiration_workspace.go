package app

import (
	"context"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ExpirationWorkspaceResult struct {
	expirationapp.WorkspaceResult
	PrimaryPhotos map[ports.AttachmentAssetReference]media.Attachment
}

func (a App) ListExpirationAssets(ctx context.Context, input expirationapp.WorkspaceInput) (ExpirationWorkspaceResult, error) {
	preferences, err := a.notificationService.GetPreferences(ctx, notificationapp.ScopeInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: input.Source, RequestID: input.RequestID})
	if err != nil {
		return ExpirationWorkspaceResult{}, err
	}
	service := expirationapp.NewWorkspaceService(expirationapp.WorkspaceDependencies{Checkouts: a.checkouts, Assets: a.assets, Types: a.customAssetTypes, Tags: a.assetTags, Inventories: a.inventories, Authorizer: a.authorizer, Audit: a.audit, IDs: a.ids, Clock: a.clock, Observer: a.observer})
	result, err := service.List(ctx, input, preferences.Settings)
	if err != nil {
		return ExpirationWorkspaceResult{}, err
	}
	photos, err := a.primaryImageAttachmentsForSearchResults(ctx, input.TenantID, result.Records)
	if err != nil {
		return ExpirationWorkspaceResult{}, err
	}
	return ExpirationWorkspaceResult{WorkspaceResult: result, PrimaryPhotos: photos}, nil
}
