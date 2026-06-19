package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterRestore(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore", func(ctx context.Context, input *dto.UpdateAssetLifecycleInput) (*dto.UpdateAssetLifecycleOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.RestoreAsset(ctx, app.UpdateAssetLifecycleInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.UpdateAssetLifecycleOutput{
			Body: shared.SuccessEnvelope[dto.AssetResponse]{
				Data: mapper.AssetToResponse(item),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}
