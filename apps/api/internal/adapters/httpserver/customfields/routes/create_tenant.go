package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customfields/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCreateTenant(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/custom-field-definitions", func(ctx context.Context, input *dto.CreateTenantDefinitionInput) (*dto.CreateDefinitionOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		definition, err := application.CreateTenantCustomFieldDefinition(ctx, app.CreateCustomFieldDefinitionInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			Key:         input.Body.Key,
			DisplayName: input.Body.DisplayName,
			Type:        input.Body.Type,
			EnumOptions: input.Body.EnumOptions,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.CreateDefinitionOutput{
			Body: shared.SuccessEnvelope[dto.DefinitionResponse]{
				Data: mapper.DefinitionToResponse(definition),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("custom field definitions"), shared.CreatedOperation, shared.SecuredOperation)
}
