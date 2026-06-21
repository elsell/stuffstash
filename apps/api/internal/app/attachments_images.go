package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type DownloadAttachmentThumbnailInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
	Variant      string
}

type PrepareAttachmentForModelUseInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
}

type AttachmentThumbnailResult struct {
	Attachment  media.Attachment
	ContentType media.ContentType
	Content     []byte
}

func (a App) DownloadAttachmentThumbnail(ctx context.Context, input DownloadAttachmentThumbnailInput) (AttachmentThumbnailResult, error) {
	attachment, content, err := a.readAttachmentBlobForImageWork(ctx, input.Principal, input.Source, input.RequestID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return AttachmentThumbnailResult{}, err
	}
	if a.imageProcessor == nil {
		return AttachmentThumbnailResult{}, ErrInvalidInput
	}
	variant, ok := media.NewThumbnailVariant(input.Variant)
	if !ok {
		return AttachmentThumbnailResult{}, ErrInvalidInput
	}
	thumbnail, err := a.imageProcessor.CreateThumbnail(ctx, ports.ImageDerivativeRequest{
		Attachment:  attachment,
		Variant:     variant,
		ContentType: attachment.ContentType,
		Content:     content,
	})
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "thumbnail generation failed"})
		return AttachmentThumbnailResult{}, err
	}
	if !thumbnail.ContentType.IsImage() || len(thumbnail.Content) == 0 {
		return AttachmentThumbnailResult{}, ErrInvalidInput
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentThumbnailGenerated,
		Message: "attachment thumbnail generated",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"variant":       variant.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return AttachmentThumbnailResult{Attachment: attachment, ContentType: thumbnail.ContentType, Content: thumbnail.Content}, nil
}

func (a App) PrepareAttachmentForModelUse(ctx context.Context, input PrepareAttachmentForModelUseInput) (ports.ModelImage, error) {
	attachment, content, err := a.readAttachmentBlobForImageWork(ctx, input.Principal, input.Source, input.RequestID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return ports.ModelImage{}, err
	}
	if a.imageProcessor == nil {
		return ports.ModelImage{}, ErrInvalidInput
	}
	prepared, err := a.imageProcessor.PrepareImageForModelUse(ctx, ports.ModelImageRequest{
		Attachment:  attachment,
		ContentType: attachment.ContentType,
		Content:     content,
	})
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "model image preparation failed"})
		return ports.ModelImage{}, err
	}
	if !prepared.ContentType.IsImage() || len(prepared.Content) == 0 || prepared.SizeBytes <= 0 || prepared.SHA256.String() == "" {
		return ports.ModelImage{}, ErrInvalidInput
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentModelImagePrepared,
		Message: "attachment model image prepared",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return prepared, nil
}

func (a App) readAttachmentBlobForImageWork(ctx context.Context, principal identity.Principal, source audit.Source, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, []byte, error) {
	if err := a.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionView); err != nil {
		return media.Attachment{}, nil, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, tenantID, inventoryID, assetID); err != nil {
		return media.Attachment{}, nil, err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, tenantID, inventoryID, assetID, attachmentID)
	if err != nil {
		return media.Attachment{}, nil, err
	}
	if !found {
		return media.Attachment{}, nil, ErrNotFound
	}
	if !attachment.ContentType.IsImage() {
		return media.Attachment{}, nil, ErrInvalidInput
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: principal.ID,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		Source:      source,
		RequestID:   requestID,
		Action:      audit.ActionAttachmentContentDownloaded,
		TargetType:  audit.TargetAttachment,
		TargetID:    attachment.ID.String(),
		Metadata: map[string]string{
			"asset_id": assetID.String(),
		},
	}); err != nil {
		return media.Attachment{}, nil, err
	}
	content, err := a.blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, nil, err
	}
	return attachment, content, nil
}
