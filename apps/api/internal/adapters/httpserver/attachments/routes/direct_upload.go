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
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterDirectUpload(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads", func(ctx context.Context, input *dto.InitiateAssetAttachmentDirectUploadInput) (*dto.InitiateAssetAttachmentDirectUploadOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		upload, err := application.InitiateAttachmentDirectUpload(ctx, app.InitiateAttachmentDirectUploadInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			FileName:    input.Body.FileName,
			ContentType: input.Body.ContentType,
			SizeBytes:   input.Body.SizeBytes,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.InitiateAssetAttachmentDirectUploadOutput{
			Body: shared.SuccessEnvelope[dto.DirectUploadResponse]{
				Data: mapper.DirectUploadToResponse(upload),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("attachments"), shared.CreatedOperation, shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete", func(ctx context.Context, input *dto.CompleteAssetAttachmentDirectUploadInput) (*dto.CompleteAssetAttachmentDirectUploadOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		attachment, err := application.CompleteAttachmentDirectUpload(ctx, app.CompleteAttachmentDirectUploadInput{
			Principal:   principal,
			Source:      audit.SourceAPI,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			AssetID:     asset.ID(input.AssetID),
			UploadID:    input.UploadID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.CompleteAssetAttachmentDirectUploadOutput{
			Body: shared.SuccessEnvelope[dto.AttachmentResponse]{
				Data: mapper.AttachmentToResponse(attachment),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("attachments"), shared.CreatedOperation, shared.SecuredOperation)
}
