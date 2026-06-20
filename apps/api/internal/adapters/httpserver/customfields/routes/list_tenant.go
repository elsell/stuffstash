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

func RegisterListTenant(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/custom-field-definitions", func(ctx context.Context, input *dto.ListTenantDefinitionsInput) (*dto.ListDefinitionsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListTenantCustomFieldDefinitions(ctx, app.ListCustomFieldDefinitionsInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
			Limit:     input.Limit,
			Cursor:    input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListDefinitionsOutput{
			Body: shared.SuccessEnvelope[[]dto.DefinitionResponse]{
				Data: mapper.DefinitionsToResponse(result.Items),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("custom field definitions"), shared.SecuredOperation)
}
