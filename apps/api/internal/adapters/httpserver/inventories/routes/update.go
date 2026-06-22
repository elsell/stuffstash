package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/inventories/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterUpdate(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}", func(ctx context.Context, input *dto.UpdateInventoryInput) (*dto.UpdateInventoryOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		item, err := application.UpdateInventory(ctx, app.UpdateInventoryInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Name:        input.Body.Name,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		response, err := inventoryResponse(ctx, application, principal, item)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateInventoryOutput{Body: shared.SuccessEnvelope[dto.InventoryResponse]{
			Data: response,
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("inventories"), shared.SecuredOperation)
}
