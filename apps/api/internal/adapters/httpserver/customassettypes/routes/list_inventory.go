package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterListInventory(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", func(ctx context.Context, input *dto.ListInventoryAssetTypesInput) (*dto.ListAssetTypesOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventoryCustomAssetTypes(ctx, app.ListCustomAssetTypesInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListAssetTypesOutput{
			Body: shared.SuccessEnvelope[[]dto.AssetTypeResponse]{
				Data: mapper.AssetTypesToResponse(result.Items),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}
