package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterList(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", func(ctx context.Context, input *dto.ListInventoryAccessInput) (*dto.ListInventoryAccessOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListInventoryAccessGrants(ctx, app.ListInventoryAccessGrantsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Limit:       input.Limit,
			Cursor:      input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListInventoryAccessOutput{
			Body: shared.SuccessEnvelope[[]dto.GrantResponse]{
				Data: mapper.GrantsToResponse(result.Items),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.SecuredOperation)
}
