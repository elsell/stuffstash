package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterGrant(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", func(ctx context.Context, input *dto.GrantInventoryAccessInput) (*dto.GrantInventoryAccessOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		grant, err := application.GrantInventoryAccess(ctx, app.GrantInventoryAccessInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			TargetUserID: input.Body.PrincipalID,
			Relationship: input.Body.Relationship,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.GrantInventoryAccessOutput{
			Body: shared.SuccessEnvelope[dto.GrantResponse]{
				Data: mapper.GrantToResponse(grant),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("inventory access"), shared.CreatedOperation, shared.SecuredOperation)
}
