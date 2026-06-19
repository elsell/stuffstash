package routes

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterDownload(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content", func(ctx context.Context, input *dto.DownloadAssetAttachmentInput) (*dto.DownloadAssetAttachmentOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		attachmentID, ok := media.NewID(input.AttachmentID)
		if !ok {
			return nil, shared.ToHumaError(app.ErrInvalidInput)
		}

		result, err := application.DownloadAttachment(ctx, app.DownloadAttachmentInput{
			Principal:    principal,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			AssetID:      asset.ID(input.AssetID),
			AttachmentID: attachmentID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.DownloadAssetAttachmentOutput{
			ContentType:        result.Attachment.ContentType.String(),
			ContentDisposition: fmt.Sprintf("attachment; filename=%q", result.Attachment.FileName.String()),
			Body:               result.Content,
		}, nil
	}, huma.OperationTags("attachments"), shared.SecuredOperation)
}
