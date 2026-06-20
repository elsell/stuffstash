package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterDetailTenant(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/custom-field-definitions/{definitionId}", func(ctx context.Context, input *dto.GetTenantDefinitionInput) (*dto.GetDefinitionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		definition, err := application.GetTenantCustomFieldDefinition(ctx, app.GetCustomFieldDefinitionInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			DefinitionID: customfield.ID(input.DefinitionID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetDefinitionOutput{Body: shared.SuccessEnvelope[dto.DefinitionResponse]{
			Data: mapper.DefinitionToResponse(definition),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}

func RegisterDetailInventory(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", func(ctx context.Context, input *dto.GetInventoryDefinitionInput) (*dto.GetDefinitionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		definition, err := application.GetInventoryCustomFieldDefinition(ctx, app.GetCustomFieldDefinitionInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			DefinitionID: customfield.ID(input.DefinitionID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetDefinitionOutput{Body: shared.SuccessEnvelope[dto.DefinitionResponse]{
			Data: mapper.DefinitionToResponse(definition),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}
