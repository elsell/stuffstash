package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/search/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/search/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterSearchAssets(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/search/assets", func(ctx context.Context, input *dto.SearchAssetsInput) (*dto.SearchAssetsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.SearchAssets(ctx, app.SearchAssetsInput{
			Principal:         principal,
			TenantID:          tenant.ID(input.TenantID),
			Query:             input.Query,
			Mode:              input.Mode,
			CustomAssetTypeID: input.CustomAssetTypeID,
			LifecycleState:    input.LifecycleState,
			Limit:             input.Limit,
			Cursor:            input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.SearchAssetsOutput{
			Body: shared.SuccessEnvelope[[]dto.AssetSearchResultResponse]{
				Data: mapper.AssetSearchResultsToResponse(result.Items, result.PrimaryPhotos),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("search"), shared.SecuredOperation)
}

func Register(api huma.API, application app.App) {
	RegisterSearchAssets(api, application)
}
