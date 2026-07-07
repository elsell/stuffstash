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

func RegisterDetail(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", func(ctx context.Context, input *dto.GetAssetInput) (*dto.GetAssetOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.GetAssetDetail(ctx, app.GetAssetInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		var checkouts []asset.Checkout
		if result.CurrentCheckout != nil {
			checkouts = []asset.Checkout{*result.CurrentCheckout}
		}
		return &dto.GetAssetOutput{Body: shared.SuccessEnvelope[dto.AssetResponse]{
			Data: mapper.AssetToResponseWithTags(result.Item, result.Tags, result.PrimaryPhoto, result.CurrentCheckout, resolveCheckoutPrincipals(ctx, application, checkouts)),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}
