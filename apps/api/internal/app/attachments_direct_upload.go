package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type InitiateAttachmentDirectUploadInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	FileName    string
	ContentType string
	SizeBytes   int64
}

type CompleteAttachmentDirectUploadInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	UploadID    string
}

type verifiedAttachmentInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	ID          media.ID
	StorageKey  media.StorageKey
	FileName    media.FileName
	ContentType media.ContentType
	SizeBytes   int64
	SHA256      media.SHA256
}

func (a App) InitiateAttachmentDirectUpload(ctx context.Context, input InitiateAttachmentDirectUploadInput) (ports.DirectAttachmentUpload, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.DirectAttachmentUpload{}, err
	}
	if a.directUploads == nil {
		return ports.DirectAttachmentUpload{}, ErrInvalidInput
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return ports.DirectAttachmentUpload{}, err
	}
	fileName, ok := media.NewFileName(input.FileName)
	if !ok {
		return ports.DirectAttachmentUpload{}, ErrAttachmentFileNameInvalid
	}
	contentType, ok := media.NewContentType(input.ContentType)
	if !ok {
		return ports.DirectAttachmentUpload{}, ErrAttachmentContentTypeUnsupported
	}
	if input.SizeBytes <= 0 {
		return ports.DirectAttachmentUpload{}, ErrInvalidInput
	}
	if input.SizeBytes > int64(a.maxAttachmentBytes) {
		return ports.DirectAttachmentUpload{}, ErrAttachmentTooLarge
	}
	uploadID := a.ids.NewID()
	attachmentID, ok := media.NewID(a.ids.NewID())
	if !ok || strings.TrimSpace(uploadID) == "" {
		return ports.DirectAttachmentUpload{}, ErrInvalidInput
	}
	storageKey, ok := media.NewStorageKey(input.TenantID.String() + "/" + input.InventoryID.String() + "/" + input.AssetID.String() + "/" + attachmentID.String())
	if !ok {
		return ports.DirectAttachmentUpload{}, ErrInvalidInput
	}
	expiresAt := a.clock.Now().UTC().Add(a.directUploadTTL)
	upload, err := a.directUploads.CreateDirectAttachmentUpload(ctx, ports.DirectAttachmentUploadRequest{
		UploadID:     uploadID,
		AttachmentID: attachmentID,
		TenantID:     input.TenantID,
		InventoryID:  input.InventoryID,
		AssetID:      input.AssetID,
		StorageKey:   storageKey,
		FileName:     fileName,
		ContentType:  contentType,
		SizeBytes:    input.SizeBytes,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "direct upload creation failed"})
		return ports.DirectAttachmentUpload{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentDirectUploadCreated,
		Message: "attachment direct upload created",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachmentID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return upload, nil
}

func (a App) CompleteAttachmentDirectUpload(ctx context.Context, input CompleteAttachmentDirectUploadInput) (media.Attachment, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return media.Attachment{}, err
	}
	if a.directUploads == nil || a.blobs == nil {
		return media.Attachment{}, ErrInvalidInput
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
	}
	completed, err := a.directUploads.CompleteDirectAttachmentUpload(ctx, input.UploadID)
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "direct upload completion failed"})
		switch {
		case errors.Is(err, ports.ErrDirectUploadIncomplete):
			return media.Attachment{}, ErrNotFound
		case errors.Is(err, ports.ErrDirectUploadInvalid), errors.Is(err, ports.ErrDirectUploadExpired), errors.Is(err, ports.ErrDirectUploadMismatch):
			return media.Attachment{}, ErrInvalidInput
		}
		return media.Attachment{}, err
	}
	if completed.UploadID != input.UploadID ||
		completed.TenantID != input.TenantID ||
		completed.InventoryID != input.InventoryID ||
		completed.AssetID != input.AssetID ||
		completed.SizeBytes <= 0 ||
		completed.AttachmentID.String() == "" ||
		completed.StorageKey.String() == "" ||
		completed.FileName.String() == "" ||
		completed.ContentType.String() == "" ||
		completed.SHA256.String() == "" ||
		completed.ExpiresAt.IsZero() ||
		!completed.ExpiresAt.After(a.clock.Now().UTC()) {
		return media.Attachment{}, ErrInvalidInput
	}
	if completed.SizeBytes > int64(a.maxAttachmentBytes) {
		return media.Attachment{}, ErrAttachmentTooLarge
	}
	content, err := a.blobs.GetBlob(ctx, completed.StorageKey)
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "direct upload content validation failed"})
		return media.Attachment{}, err
	}
	hashBytes := sha256.Sum256(content)
	if int64(len(content)) != completed.SizeBytes ||
		completed.SHA256.String() != hex.EncodeToString(hashBytes[:]) {
		return media.Attachment{}, ErrInvalidInput
	}
	if !contentMatchesType(completed.ContentType, content) {
		return media.Attachment{}, ErrAttachmentContentMismatch
	}
	attachment, err := a.persistVerifiedAttachment(ctx, verifiedAttachmentInput{
		Principal:   input.Principal,
		Source:      input.Source,
		RequestID:   input.RequestID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		AssetID:     input.AssetID,
		ID:          completed.AttachmentID,
		StorageKey:  completed.StorageKey,
		FileName:    completed.FileName,
		ContentType: completed.ContentType,
		SizeBytes:   completed.SizeBytes,
		SHA256:      completed.SHA256,
	})
	if err != nil {
		return media.Attachment{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentDirectUploadCompleted,
		Message: "attachment direct upload completed",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return attachment, nil
}

func (a App) persistVerifiedAttachment(ctx context.Context, input verifiedAttachmentInput) (media.Attachment, error) {
	attachment, ok := media.NewAttachment(
		input.ID,
		media.TenantID(input.TenantID.String()),
		media.InventoryID(input.InventoryID.String()),
		media.AssetID(input.AssetID.String()),
		input.StorageKey,
		input.FileName,
		input.ContentType,
		input.SizeBytes,
		input.SHA256,
		a.clock.Now().UTC(),
	)
	if !ok {
		return media.Attachment{}, ErrInvalidInput
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
		return media.Attachment{}, err
	}
	if err := a.attachmentUnitOfWork.SaveAttachment(ctx, attachment, auditRecord); err != nil {
		return media.Attachment{}, err
	}
	a.observer.Record(ctx, ports.Event{
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
	return attachment, nil
}
