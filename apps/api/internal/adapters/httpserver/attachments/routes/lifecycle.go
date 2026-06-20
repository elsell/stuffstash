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

func RegisterArchive(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive", func(ctx context.Context, input *dto.UpdateAssetAttachmentLifecycleInput) (*dto.UpdateAssetAttachmentLifecycleOutput, error) {
		return updateAttachmentLifecycle(ctx, application, input, application.ArchiveAttachment)
	}, huma.OperationTags("attachments"), shared.SecuredOperation)
}

func RegisterRestore(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore", func(ctx context.Context, input *dto.UpdateAssetAttachmentLifecycleInput) (*dto.UpdateAssetAttachmentLifecycleOutput, error) {
		return updateAttachmentLifecycle(ctx, application, input, application.RestoreAttachment)
	}, huma.OperationTags("attachments"), shared.SecuredOperation)
}

func updateAttachmentLifecycle(ctx context.Context, application app.App, input *dto.UpdateAssetAttachmentLifecycleInput, operation func(context.Context, app.UpdateAttachmentLifecycleInput) (media.Attachment, error)) (*dto.UpdateAssetAttachmentLifecycleOutput, error) {
	principal, err := shared.Authenticate(ctx, application, input.Authorization)
	if err != nil {
		return nil, err
	}
	attachmentID, ok := media.NewID(input.AttachmentID)
	if !ok {
		return nil, shared.ToHumaError(app.ErrInvalidInput)
	}
	attachment, err := operation(ctx, app.UpdateAttachmentLifecycleInput{
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
	return &dto.UpdateAssetAttachmentLifecycleOutput{Body: shared.SuccessEnvelope[dto.AttachmentResponse]{
		Data: mapper.AttachmentToResponse(attachment),
		Meta: shared.Meta{TenantID: input.TenantID},
	}}, nil
}

func RegisterDelete(api huma.API, application app.App) {
	huma.Delete(api, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", func(ctx context.Context, input *dto.UpdateAssetAttachmentLifecycleInput) (*dto.DeleteAssetAttachmentOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		attachmentID, ok := media.NewID(input.AttachmentID)
		if !ok {
			return nil, shared.ToHumaError(app.ErrInvalidInput)
		}
		if err := application.DeleteAttachment(ctx, app.UpdateAttachmentLifecycleInput{
			Principal:    principal,
			Source:       audit.SourceAPI,
			RequestID:    input.RequestID,
			TenantID:     tenant.ID(input.TenantID),
			InventoryID:  inventory.InventoryID(input.InventoryID),
			AssetID:      asset.ID(input.AssetID),
			AttachmentID: attachmentID,
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteAssetAttachmentOutput{}, nil
	}, huma.OperationTags("attachments"), shared.NoContentOperation, shared.SecuredOperation)
}
