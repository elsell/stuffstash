package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterListAssetActivity(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/activity", func(ctx context.Context, input *dto.ListAssetActivityInput) (*dto.ListAssetActivityOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ListAssetActivity(ctx, app.ListAssetActivityInput{
			Principal: principal, TenantID: tenant.ID(input.TenantID), InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID: asset.ID(input.AssetID), View: app.AssetActivityView(input.View), Limit: input.Limit, Cursor: input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ListAssetActivityOutput{Body: shared.SuccessEnvelope[[]dto.AssetActivityResponse]{
			Data: mapper.AssetActivitiesToResponse(result.Items, result.ResolvedPrincipals),
			Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
		}}, nil
	}, huma.OperationTags("audit records"), shared.SecuredOperation)
}
