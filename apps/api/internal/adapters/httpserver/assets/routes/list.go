package routes

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterList(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets", func(ctx context.Context, input *dto.ListAssetsInput) (*dto.ListAssetsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListAssets(ctx, app.ListAssetsInput{
			Principal:      principal,
			Source:         audit.SourceAPI,
			RequestID:      input.RequestID,
			TenantID:       tenant.ID(input.TenantID),
			InventoryID:    inventory.InventoryID(input.InventoryID),
			Limit:          input.Limit,
			Cursor:         input.Cursor,
			LifecycleState: input.LifecycleState,
			Sort:           input.Sort,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		data := mapper.AssetsToResponseWithTags(result.Items, result.Tags, result.PrimaryPhotos, result.Checkouts, resolveCheckoutPrincipalsFromMap(ctx, application, result.Checkouts))
		for index := range data {
			ref := ports.AttachmentAssetReference{InventoryID: inventory.InventoryID(result.Items[index].InventoryID), AssetID: result.Items[index].ID}
			if value, ok := result.ExpirationContexts[ref]; ok {
				data[index].ExpirationContext = mapper.ExpirationContextToResponse(value.State, value.TrackingEnabled, value.AdvanceDays, value.Timezone)
			}
		}

		return &dto.ListAssetsOutput{
			Body: shared.SuccessEnvelope[[]dto.AssetResponse]{
				Data: data,
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("assets"), shared.SecuredOperation)
}

func Register(api huma.API, application app.App) {
	RegisterCreate(api, application)
	RegisterDetail(api, application)
	RegisterUpdate(api, application)
	RegisterArchive(api, application)
	RegisterRestore(api, application)
	RegisterDelete(api, application)
	RegisterCheckout(api, application)
	RegisterReturn(api, application)
	RegisterUpdateReturnedCheckoutDetails(api, application)
	RegisterCheckoutHistory(api, application)
	RegisterCheckedOutAssets(api, application)
	RegisterList(api, application)
	RegisterExpirationWorkspace(api, application)
}
