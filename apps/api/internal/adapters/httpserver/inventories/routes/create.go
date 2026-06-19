package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCreate(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories", func(ctx context.Context, input *dto.CreateInventoryInput) (*dto.CreateInventoryOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.CreateInventory(ctx, app.CreateInventoryInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
			Name:      input.Body.Name,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.CreateInventoryOutput{
			Body: shared.SuccessEnvelope[dto.InventoryResponse]{
				Data: mapper.InventoryToResponse(item),
				Meta: shared.Meta{TenantID: item.TenantID.String()},
			},
		}, nil
	}, huma.OperationTags("inventories"), shared.CreatedOperation, shared.SecuredOperation)
}
