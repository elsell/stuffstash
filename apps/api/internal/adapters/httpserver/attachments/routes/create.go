package routes

import (
	"context"
	"encoding/base64"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCreate(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", func(ctx context.Context, input *dto.CreateAssetAttachmentInput) (*dto.CreateAssetAttachmentOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		content, err := base64.StdEncoding.DecodeString(input.Body.ContentBase64)
		if err != nil {
			return nil, shared.ToHumaError(app.ErrInvalidInput)
		}
		attachment, err := application.CreateAttachment(ctx, app.CreateAttachmentInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			FileName:    input.Body.FileName,
			ContentType: input.Body.ContentType,
			Content:     content,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.CreateAssetAttachmentOutput{
			Body: shared.SuccessEnvelope[dto.AttachmentResponse]{
				Data: mapper.AttachmentToResponse(attachment),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("attachments"), shared.CreatedOperation, shared.SecuredOperation, func(operation *huma.Operation) {
		operation.MaxBodyBytes = application.MaxAttachmentJSONBodyBytes()
	})
}
