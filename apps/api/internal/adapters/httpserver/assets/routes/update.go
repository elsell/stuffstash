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

func RegisterUpdate(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", func(ctx context.Context, input *dto.UpdateAssetInput) (*dto.UpdateAssetOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.UpdateAsset(ctx, app.UpdateAssetInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			Title:       input.Body.Title,
			Description: input.Body.Description,
			ParentAssetID: app.AssetParentUpdate{
				Present: input.Body.ParentAssetID.Present(),
				Null:    input.Body.ParentAssetID.Null(),
				Value:   input.Body.ParentAssetID.Value(),
			},
			CustomFields: input.Body.CustomFields,
			TagIDs:       input.Body.TagIDs,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		tags := mapper.TagsToResponse(nil)
		if input.Body.TagIDs != nil {
			result, err := application.GetAssetDetail(ctx, app.GetAssetInput{
				Principal:   principal,
				Source:      audit.SourceAPI,
				RequestID:   input.RequestID,
				TenantID:    tenant.ID(input.TenantID),
				InventoryID: inventory.InventoryID(input.InventoryID),
				AssetID:     item.ID,
			})
			if err != nil {
				return nil, shared.ToHumaError(err)
			}
			tags = mapper.TagsToResponse(result.Tags)
		}

		return &dto.UpdateAssetOutput{
			Body: shared.SuccessEnvelope[dto.AssetResponse]{
				Data: func() dto.AssetResponse {
					response := mapper.AssetToResponse(item, nil, nil, nil)
					response.Tags = tags
					return response
				}(),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}
