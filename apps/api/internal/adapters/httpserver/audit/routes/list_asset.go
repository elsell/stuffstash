package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterListAsset(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/audit-records", func(ctx context.Context, input *dto.ListAssetAuditHistoryInput) (*dto.ListAuditRecordsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListAssetAuditHistory(ctx, app.ListAssetAuditHistoryInput{
			Principal:   principal,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     input.AssetID,
			Limit:       input.Limit,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListAuditRecordsOutput{
			Body: shared.SuccessEnvelope[[]dto.RecordResponse]{
				Data: mapper.RecordsToResponse(result.Items, result.ResolvedPrincipals),
				Meta: shared.PaginatedMeta(input.AssetID, result.Limit, nil, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("audit records"), shared.SecuredOperation)
}
