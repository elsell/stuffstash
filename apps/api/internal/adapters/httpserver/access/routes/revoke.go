package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/access/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterRevoke(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", func(ctx context.Context, input *dto.RevokeInventoryAccessInput) (*dto.RevokeInventoryAccessOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		if _, err := application.RevokeInventoryAccess(ctx, app.RevokeInventoryAccessInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			TargetUserID: input.PrincipalID,
			Relationship: input.Relationship,
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.RevokeInventoryAccessOutput{}, nil
	}, huma.OperationTags("inventory access"), shared.NoContentOperation, shared.SecuredOperation)
}
