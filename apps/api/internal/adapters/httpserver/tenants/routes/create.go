package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/mapper"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

func Register(api huma.API, application app.App) {
	huma.Post(api, "/tenants", func(ctx context.Context, input *dto.CreateTenantInput) (*dto.CreateTenantOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.CreateTenant(ctx, app.CreateTenantInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			Name:      input.Body.Name,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.CreateTenantOutput{
			Body: shared.SuccessEnvelope[dto.TenantResponse]{
				Data: mapper.TenantToResponse(item),
				Meta: shared.Meta{TenantID: item.ID.String()},
			},
		}, nil
	}, huma.OperationTags("tenants"), shared.CreatedOperation, shared.SecuredOperation)
}
