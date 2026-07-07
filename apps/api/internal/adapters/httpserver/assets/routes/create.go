package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCreate(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets", func(ctx context.Context, input *dto.CreateAssetInput) (*dto.CreateAssetOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.CreateAsset(ctx, app.CreateAssetInput{
			Principal:         principal,
			Source:            audit.SourceAPI,
			RequestID:         input.RequestID,
			TenantID:          tenant.ID(input.TenantID),
			InventoryID:       inventory.InventoryID(input.InventoryID),
			Kind:              input.Body.Kind,
			Title:             input.Body.Title,
			Description:       input.Body.Description,
			ParentAssetID:     input.Body.ParentAssetID,
			CustomAssetTypeID: input.Body.CustomAssetTypeID,
			CustomFields:      input.Body.CustomFields,
			TagIDs:            input.Body.TagIDs,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		tags, err := application.GetAssetAssignedTags(ctx, app.GetAssetAssignedTagsInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     item.ID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.CreateAssetOutput{
			Body: shared.SuccessEnvelope[dto.AssetResponse]{
				Data: mapper.AssetToResponseWithTags(item, tags, nil, nil, nil),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("assets"), shared.CreatedOperation, shared.SecuredOperation)
}
