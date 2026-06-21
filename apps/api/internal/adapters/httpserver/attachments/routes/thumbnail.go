package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterThumbnail(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/thumbnail", func(ctx context.Context, input *dto.DownloadAssetAttachmentThumbnailInput) (*dto.DownloadAssetAttachmentThumbnailOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		attachmentID, ok := media.NewID(input.AttachmentID)
		if !ok {
			return nil, shared.ToHumaError(app.ErrInvalidInput)
		}
		result, err := application.DownloadAttachmentThumbnail(ctx, app.DownloadAttachmentThumbnailInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			AssetID:      asset.ID(input.AssetID),
			AttachmentID: attachmentID,
			Variant:      input.Variant,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DownloadAssetAttachmentThumbnailOutput{
			ContentType: result.ContentType.String(),
			Body:        result.Content,
		}, nil
	}, huma.OperationTags("attachments"), shared.SecuredOperation)
}
