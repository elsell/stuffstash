package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAttachmentInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	FileName    string
	ContentType string
	Content     []byte
}

type ListAttachmentsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	Limit       int
	Cursor      string
}

type GetAttachmentInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
}

type DownloadAttachmentInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
}

type UpdateAttachmentLifecycleInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	AssetID      asset.ID
	AttachmentID media.ID
}

type ListAttachmentsResult struct {
	Items      []media.Attachment
	Limit      int
	NextCursor *string
	HasMore    bool
}

type AttachmentContentResult struct {
	Attachment media.Attachment
	Content    []byte
}

type PreparedAttachment struct {
	ThumbnailJob *media.ThumbnailJob
	Attachment   media.Attachment
	AuditRecord  audit.Record
	StorageKey   media.StorageKey
	ContentType  media.ContentType
}

func (a AttachmentService) CreateAttachment(ctx context.Context, input CreateAttachmentInput) (media.Attachment, error) {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return media.Attachment{}, err
	}
	prepared, err := a.PrepareAttachment(ctx, input)
	if err != nil {
		return media.Attachment{}, err
	}
	if err := a.deps.Blobs.PutBlob(ctx, prepared.StorageKey, prepared.ContentType, input.Content); err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, err
	}
	if err := a.deps.AttachmentUnitOfWork.SaveAttachment(ctx, prepared.Attachment, prepared.AuditRecord, prepared.ThumbnailJob); err != nil {
		if deleteErr := a.deps.Blobs.DeleteBlob(ctx, prepared.StorageKey); deleteErr != nil {
			a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob cleanup failed"})
		}
		return media.Attachment{}, err
	}
	a.RecordAttachmentCreated(ctx, input, prepared.Attachment)
	return prepared.Attachment, nil
}

func (a AttachmentService) PrepareAttachment(ctx context.Context, input CreateAttachmentInput) (PreparedAttachment, error) {
	if a.deps.Attachments == nil || a.deps.Blobs == nil {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	if input.AssetID.String() == "" {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return PreparedAttachment{}, err
	}
	fileName, ok := media.NewFileName(input.FileName)
	if !ok {
		return PreparedAttachment{}, apperrors.ErrAttachmentFileNameInvalid
	}
	contentType, ok := media.NewContentType(input.ContentType)
	if !ok {
		return PreparedAttachment{}, apperrors.ErrAttachmentContentTypeUnsupported
	}
	if len(input.Content) == 0 {
		return PreparedAttachment{}, apperrors.ErrAttachmentContentEmpty
	}
	if len(input.Content) > a.deps.MaxAttachmentBytes {
		return PreparedAttachment{}, apperrors.ErrAttachmentTooLarge
	}
	if !contentMatchesType(contentType, input.Content) {
		return PreparedAttachment{}, apperrors.ErrAttachmentContentMismatch
	}
	attachmentID, ok := media.NewID(a.deps.IDs.NewID())
	if !ok {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	storageKey, ok := media.NewStorageKey(input.TenantID.String() + "/" + input.InventoryID.String() + "/" + input.AssetID.String() + "/" + attachmentID.String())
	if !ok {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	hashBytes := sha256.Sum256(input.Content)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	attachment, ok := media.NewAttachment(
		attachmentID,
		media.TenantID(input.TenantID.String()),
		media.InventoryID(input.InventoryID.String()),
		media.AssetID(input.AssetID.String()),
		storageKey,
		fileName,
		contentType,
		int64(len(input.Content)),
		hash,
		a.deps.Clock.Now().UTC(),
	)
	if !ok {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAttachmentCreated,
		TargetType:  audit.TargetAsset,
		TargetID:    input.AssetID.String(),
		Metadata: map[string]string{
			"attachment_id": attachment.ID.String(),
			"content_type":  attachment.ContentType.String(),
			"size_bytes":    strconv.FormatInt(attachment.SizeBytes, 10),
		},
	})
	if err != nil {
		return PreparedAttachment{}, err
	}
	thumbnailJob, err := media.PlanThumbnailJob(attachment)
	if err != nil {
		return PreparedAttachment{}, apperrors.ErrInvalidInput
	}
	return PreparedAttachment{ThumbnailJob: thumbnailJob, Attachment: attachment, AuditRecord: auditRecord, StorageKey: storageKey, ContentType: contentType}, nil
}

func (a AttachmentService) RecordAttachmentCreated(ctx context.Context, input CreateAttachmentInput, attachment media.Attachment) {
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentCreated,
		Message: "attachment created",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
}
