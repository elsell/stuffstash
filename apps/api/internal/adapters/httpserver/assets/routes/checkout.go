package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCheckout(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/checkout", func(ctx context.Context, input *dto.CheckoutAssetInput) (*dto.CheckoutAssetOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.CheckoutAssetWithOperation(ctx, app.CheckoutAssetInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			Details:     input.Body.Details,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		response := mapper.CheckoutToResponse(result.Checkout)
		response.UndoableOperationID = result.UndoableOperationID
		return &dto.CheckoutAssetOutput{Body: shared.SuccessEnvelope[dto.AssetCheckoutResponse]{
			Data: response,
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("assets"), shared.CreatedOperation, shared.SecuredOperation)
}

func RegisterReturn(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/return", func(ctx context.Context, input *dto.ReturnAssetInput) (*dto.ReturnAssetOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ReturnAssetWithOperation(ctx, app.ReturnAssetInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			Details:     input.Body.Details,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		response := mapper.CheckoutToResponse(result.Checkout)
		response.UndoableOperationID = result.UndoableOperationID
		return &dto.ReturnAssetOutput{Body: shared.SuccessEnvelope[dto.AssetCheckoutResponse]{
			Data: response,
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}

func RegisterUpdateReturnedCheckoutDetails(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/checkouts/{checkoutId}/return-details", func(ctx context.Context, input *dto.UpdateReturnedCheckoutDetailsInput) (*dto.UpdateReturnedCheckoutDetailsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		checkout, err := application.UpdateReturnedCheckoutDetails(ctx, app.UpdateReturnedCheckoutDetailsInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			CheckoutID:  asset.CheckoutID(input.CheckoutID),
			Details:     input.Body.Details,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateReturnedCheckoutDetailsOutput{Body: shared.SuccessEnvelope[dto.AssetCheckoutResponse]{
			Data: mapper.CheckoutToResponse(checkout),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}

func RegisterCheckoutHistory(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/checkouts", func(ctx context.Context, input *dto.ListAssetCheckoutHistoryInput) (*dto.ListAssetCheckoutHistoryOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ListAssetCheckoutHistory(ctx, app.ListAssetCheckoutHistoryInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ListAssetCheckoutHistoryOutput{Body: shared.SuccessEnvelope[[]dto.AssetCheckoutResponse]{
			Data: mapper.CheckoutsToResponse(result.Items),
			Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
		}}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}

func RegisterCheckedOutAssets(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/checked-out-assets", func(ctx context.Context, input *dto.ListCheckedOutAssetsInput) (*dto.ListCheckedOutAssetsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ListCheckedOutAssets(ctx, app.ListCheckedOutAssetsInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ListCheckedOutAssetsOutput{Body: shared.SuccessEnvelope[[]dto.CheckedOutAssetResponse]{
			Data: mapper.CheckedOutAssetsToResponse(result.Items, result.PrimaryPhotos, resolveCheckedOutAssetPrincipals(ctx, application, result.Items)),
			Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
		}}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}
