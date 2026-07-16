package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/customassettypes/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterListTenant(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/custom-asset-types", func(ctx context.Context, input *dto.ListTenantAssetTypesInput) (*dto.ListAssetTypesOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListTenantCustomAssetTypes(ctx, app.ListCustomAssetTypesInput{
			Principal:      principal,
			Source:         audit.SourceAPI,
			RequestID:      input.RequestID,
			TenantID:       tenant.ID(input.TenantID),
			Limit:          input.Limit,
			Cursor:         input.Cursor,
			LifecycleState: input.LifecycleState,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListAssetTypesOutput{
			Body: shared.SuccessEnvelope[[]dto.AssetTypeResponse]{
				Data: mapper.AssetTypesToResponse(result.Items),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("custom asset types"), shared.SecuredOperation)
}
