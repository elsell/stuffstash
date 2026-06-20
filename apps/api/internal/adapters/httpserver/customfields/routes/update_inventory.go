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

func RegisterUpdateInventory(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", func(ctx context.Context, input *dto.UpdateInventoryDefinitionInput) (*dto.UpdateDefinitionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		definition, err := application.UpdateInventoryCustomFieldDefinition(ctx, app.UpdateCustomFieldDefinitionInput{
			Principal:          principal,
			Source:             audit.SourceAPI,
			RequestID:          input.RequestID,
			TenantID:           tenant.ID(input.TenantID),
			InventoryID:        inventory.InventoryID(input.InventoryID),
			DefinitionID:       customfield.ID(input.DefinitionID),
			DisplayName:        input.Body.DisplayName,
			Key:                input.Body.Key,
			Type:               input.Body.Type,
			EnumOptions:        input.Body.EnumOptions,
			Applicability:      input.Body.Applicability,
			CustomAssetTypeIDs: input.Body.CustomAssetTypeIDs,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.UpdateDefinitionOutput{
			Body: shared.SuccessEnvelope[dto.DefinitionResponse]{
				Data: mapper.DefinitionToResponse(definition),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}
