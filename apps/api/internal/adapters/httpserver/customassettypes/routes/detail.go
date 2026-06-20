package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterDetailTenant(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", func(ctx context.Context, input *dto.GetTenantAssetTypeInput) (*dto.GetAssetTypeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		assetType, err := application.GetTenantCustomAssetType(ctx, app.GetCustomAssetTypeInput{
			Principal:         principal,
			Source:            audit.SourceAPI,
			RequestID:         input.RequestID,
			TenantID:          tenant.ID(input.TenantID),
			CustomAssetTypeID: customfield.AssetTypeID(input.CustomAssetTypeID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetAssetTypeOutput{Body: shared.SuccessEnvelope[dto.AssetTypeResponse]{
			Data: mapper.AssetTypeToResponse(assetType),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}

func RegisterDetailInventory(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", func(ctx context.Context, input *dto.GetInventoryAssetTypeInput) (*dto.GetAssetTypeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		assetType, err := application.GetInventoryCustomAssetType(ctx, app.GetCustomAssetTypeInput{
			Principal:         principal,
			Source:            audit.SourceAPI,
			RequestID:         input.RequestID,
			TenantID:          tenant.ID(input.TenantID),
			InventoryID:       inventory.InventoryID(input.InventoryID),
			CustomAssetTypeID: customfield.AssetTypeID(input.CustomAssetTypeID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetAssetTypeOutput{Body: shared.SuccessEnvelope[dto.AssetTypeResponse]{
			Data: mapper.AssetTypeToResponse(assetType),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}
