package media

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/sync/singleflight"
)

const DefaultPrimarySmallThumbnailWarmLimit = 12
const DefaultPrimarySmallThumbnailWarmConcurrency = 4
const DefaultPrimarySmallThumbnailWarmTimeout = 10 * time.Second

type AppNoopObserver struct{}

func (AppNoopObserver) Record(context.Context, ports.Event) {}

type PrimaryThumbnailWarmState struct {
	sem chan struct{}
}

func NewPrimaryThumbnailWarmState(concurrency int) *PrimaryThumbnailWarmState {
	return &PrimaryThumbnailWarmState{
		sem: make(chan struct{}, NormalizePrimaryThumbnailWarmConcurrency(concurrency)),
	}
}

func (s *PrimaryThumbnailWarmState) tryAcquire() bool {
	if s == nil {
		return false
	}
	select {
	case s.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *PrimaryThumbnailWarmState) release() {
	if s == nil {
		return
	}
	<-s.sem
}

type ThumbnailGenerationState struct {
	group singleflight.Group
}

func NewThumbnailGenerationState() *ThumbnailGenerationState {
	return &ThumbnailGenerationState{}
}

type thumbnailGenerationResult struct {
	contentType media.ContentType
	content     []byte
	source      string
}

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

func (a AttachmentService) DownloadAttachmentThumbnail(ctx context.Context, input DownloadAttachmentThumbnailInput) (AttachmentThumbnailResult, error) {
	if a.deps.ImageProcessor == nil && a.deps.ThumbnailReader == nil {
		return AttachmentThumbnailResult{}, apperrors.ErrInvalidInput
	}
	variant, ok := media.NewThumbnailVariant(input.Variant)
	if !ok {
		return AttachmentThumbnailResult{}, apperrors.ErrInvalidInput
	}
	attachment, err := a.authorizeImageAttachmentRead(ctx, input.Principal, input.Source, input.RequestID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return AttachmentThumbnailResult{}, err
	}
	thumbnail, err := a.getOrGenerateThumbnail(ctx, attachment, variant)
	if err != nil {
		return AttachmentThumbnailResult{}, err
	}
	a.recordAttachmentThumbnailServed(ctx, input, attachment, variant, thumbnail.source)
	return AttachmentThumbnailResult{Attachment: attachment, ContentType: thumbnail.contentType, Content: thumbnail.content}, nil
}

func (a AttachmentService) PrepareAttachmentForModelUse(ctx context.Context, input PrepareAttachmentForModelUseInput) (ports.ModelImage, error) {
	attachment, content, err := a.readAttachmentBlobForImageWork(ctx, input.Principal, input.Source, input.RequestID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return ports.ModelImage{}, err
	}
	if a.deps.ImageProcessor == nil {
		return ports.ModelImage{}, apperrors.ErrInvalidInput
	}
	prepared, err := a.deps.ImageProcessor.PrepareImageForModelUse(ctx, ports.ModelImageRequest{
		Attachment:  attachment,
		ContentType: attachment.ContentType,
		Content:     content,
	})
	if err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "model image preparation failed"})
		return ports.ModelImage{}, err
	}
	if !prepared.ContentType.IsImage() || len(prepared.Content) == 0 || prepared.SizeBytes <= 0 || prepared.SHA256.String() == "" {
		return ports.ModelImage{}, apperrors.ErrInvalidInput
	}
	a.deps.Observer.Record(ctx, ports.Event{
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

func (a AttachmentService) readAttachmentBlobForImageWork(ctx context.Context, principal identity.Principal, source audit.Source, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, []byte, error) {
	attachment, err := a.authorizeImageAttachmentRead(ctx, principal, source, requestID, tenantID, inventoryID, assetID, attachmentID)
	if err != nil {
		return media.Attachment{}, nil, err
	}
	content, err := a.deps.Blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, nil, err
	}
	return attachment, content, nil
}

func (a AttachmentService) authorizeImageAttachmentRead(ctx context.Context, principal identity.Principal, source audit.Source, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, attachmentID media.ID) (media.Attachment, error) {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionView); err != nil {
		return media.Attachment{}, err
	}
	if err := a.ensureReadableAssetForAttachment(ctx, tenantID, inventoryID, assetID); err != nil {
		return media.Attachment{}, err
	}
	attachment, found, err := a.deps.Attachments.AttachmentByID(ctx, tenantID, inventoryID, assetID, attachmentID)
	if err != nil {
		return media.Attachment{}, err
	}
	if !found {
		return media.Attachment{}, apperrors.ErrNotFound
	}
	if !attachment.ContentType.IsImage() {
		return media.Attachment{}, apperrors.ErrInvalidInput
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   principal,
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
		return media.Attachment{}, err
	}
	return attachment, nil
}

func (a AttachmentService) recordAttachmentThumbnailServed(ctx context.Context, input DownloadAttachmentThumbnailInput, attachment media.Attachment, variant media.ThumbnailVariant, source string) {
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentThumbnailGenerated,
		Message: "attachment thumbnail generated",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"variant":       variant.String(),
			"principal_id":  input.Principal.ID.String(),
			"source":        source,
		},
	})
}
