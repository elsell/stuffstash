package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterDetail(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", func(ctx context.Context, input *dto.GetAssetAttachmentInput) (*dto.GetAssetAttachmentOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		attachmentID, ok := media.NewID(input.AttachmentID)
		if !ok {
			return nil, shared.ToHumaError(app.ErrInvalidInput)
		}
		attachment, err := application.GetAttachment(ctx, app.GetAttachmentInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			AssetID:      asset.ID(input.AssetID),
			AttachmentID: attachmentID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetAssetAttachmentOutput{Body: shared.SuccessEnvelope[dto.AttachmentResponse]{
			Data: mapper.AttachmentToResponse(attachment),
			Meta: shared.Meta{TenantID: input.TenantID},
		}}, nil
	}, huma.OperationTags("attachments"), shared.SecuredOperation)
}
