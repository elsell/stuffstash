package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterList(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories", func(ctx context.Context, input *dto.ListInventoriesInput) (*dto.ListInventoriesOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventories(ctx, app.ListInventoriesInput{
			Principal: principal,
			TenantID:  tenant.ID(input.TenantID),
			Limit:     input.Limit,
			Cursor:    input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListInventoriesOutput{
			Body: shared.SuccessEnvelope[[]dto.InventoryResponse]{
				Data: mapper.InventoriesToResponse(result.Items),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("inventories"), shared.SecuredOperation)
}

func Register(api huma.API, application app.App) {
	RegisterCreate(api, application)
	RegisterList(api, application)
}
