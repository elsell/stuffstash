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

func RegisterArchiveInventory(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive", func(ctx context.Context, input *dto.ArchiveInventoryAssetTypeInput) (*dto.ArchiveAssetTypeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		assetType, err := application.ArchiveInventoryCustomAssetType(ctx, app.ArchiveCustomAssetTypeInput{
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

		return &dto.ArchiveAssetTypeOutput{
			Body: shared.SuccessEnvelope[dto.AssetTypeResponse]{
				Data: mapper.AssetTypeToResponse(assetType),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}
