package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/identity/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

func Register(api huma.API, application app.App) {
	huma.Get(api, "/me", func(ctx context.Context, input *dto.MeInput) (*dto.MeOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		return &dto.MeOutput{
			Body: shared.SuccessEnvelope[dto.PrincipalResponse]{
				Data: mapper.PrincipalToResponse(principal),
				Meta: shared.Meta{},
			},
		}, nil
	}, huma.OperationTags("identity"), shared.SecuredOperation)

	huma.Get(api, "/me/tenants", func(ctx context.Context, input *dto.ListMyTenantsInput) (*dto.ListMyTenantsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListMyTenants(ctx, app.ListMyTenantsInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			Limit:     input.Limit,
			Cursor:    input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListMyTenantsOutput{
			Body: shared.SuccessEnvelope[[]dto.MyTenantResponse]{
				Data: myTenantsToResponse(result.Items),
				Meta: shared.PaginatedMeta("", result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("identity"), shared.SecuredOperation)
}

func myTenantsToResponse(items []app.MyTenantAccess) []dto.MyTenantResponse {
	data := make([]dto.MyTenantResponse, 0, len(items))
	for _, item := range items {
		data = append(data, dto.MyTenantResponse{
			ID:             item.Tenant.ID.String(),
			Name:           item.Tenant.Name.String(),
			LifecycleState: item.Tenant.LifecycleState.String(),
			Access:         shared.AccessToResponse(item.Access),
		})
	}
	return data
}
