package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

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
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
	Limit       int
	Cursor      string
}

type DownloadAttachmentInput struct {
	Principal    identity.Principal
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

func (a App) CreateAttachment(ctx context.Context, input CreateAttachmentInput) (media.Attachment, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return media.Attachment{}, err
	}
	if a.attachments == nil || a.blobs == nil {
		return media.Attachment{}, ErrInvalidInput
	}
	if input.AssetID.String() == "" {
		return media.Attachment{}, ErrInvalidInput
	}
	if _, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
	} else if !found {
		return media.Attachment{}, ErrNotFound
	}
	fileName, ok := media.NewFileName(input.FileName)
	if !ok {
		return media.Attachment{}, ErrInvalidInput
	}
	contentType, ok := media.NewContentType(input.ContentType)
	if !ok || len(input.Content) == 0 || len(input.Content) > a.maxAttachmentBytes || !contentMatchesType(contentType, input.Content) {
		return media.Attachment{}, ErrInvalidInput
	}
	attachmentID, ok := media.NewID(a.ids.NewID())
	if !ok {
		return media.Attachment{}, ErrInvalidInput
	}
	storageKey, ok := media.NewStorageKey(input.TenantID.String() + "/" + input.InventoryID.String() + "/" + input.AssetID.String() + "/" + attachmentID.String())
	if !ok {
		return media.Attachment{}, ErrInvalidInput
	}
	hashBytes := sha256.Sum256(input.Content)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		return media.Attachment{}, ErrInvalidInput
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
		time.Now().UTC(),
	)
	if !ok {
		return media.Attachment{}, ErrInvalidInput
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
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
	if err := a.blobs.PutBlob(ctx, storageKey, contentType, input.Content); err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return media.Attachment{}, err
	}
	if err := a.attachments.SaveAttachment(ctx, attachment, auditRecord); err != nil {
		if deleteErr := a.blobs.DeleteBlob(ctx, storageKey); deleteErr != nil {
			a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob cleanup failed"})
		}
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

func (a App) ListAttachments(ctx context.Context, input ListAttachmentsInput) (ListAttachmentsResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAttachmentsResult{}, err
	}
	if _, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return ListAttachmentsResult{}, err
	} else if !found {
		return ListAttachmentsResult{}, ErrNotFound
	}
	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterAttachmentID, err := decodeAttachmentCursor(input.TenantID, input.InventoryID, input.AssetID, input.Cursor)
	if err != nil {
		return ListAttachmentsResult{}, ErrInvalidInput
	}
	items, err := a.attachments.ListAttachmentsByAsset(ctx, input.TenantID, input.InventoryID, input.AssetID, ports.AttachmentListPageRequest{
		AfterAttachmentID: afterAttachmentID,
		Limit:             limit + 1,
	})
	if err != nil {
		return ListAttachmentsResult{}, err
	}
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeAttachmentCursor(input.TenantID, input.InventoryID, input.AssetID, items[len(items)-1].ID)
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentsListed,
		Message: "attachments listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     input.AssetID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strings.TrimSpace(strconv.Itoa(limit)),
		},
	})
	return ListAttachmentsResult{Items: items, Limit: limit, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (a App) DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (AttachmentContentResult, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return AttachmentContentResult{}, err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return AttachmentContentResult{}, err
	}
	if !found {
		return AttachmentContentResult{}, ErrNotFound
	}
	content, err := a.blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return AttachmentContentResult{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentContentDownloaded,
		Message: "attachment content downloaded",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": attachment.ID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return AttachmentContentResult{Attachment: attachment, Content: content}, nil
}

func encodeAttachmentCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, id media.ID) *string {
	return encodePageCursor("attachments", tenantID.String()+":"+inventoryID.String()+":"+assetID.String(), id.String())
}

func decodeAttachmentCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, cursor string) (media.ID, error) {
	decoded, err := decodePageCursor("attachments", tenantID.String()+":"+inventoryID.String()+":"+assetID.String(), cursor)
	if err != nil {
		return "", err
	}
	if decoded == "" {
		return "", nil
	}
	id, ok := media.NewID(decoded)
	if !ok {
		return "", ErrInvalidInput
	}
	return id, nil
}

func contentMatchesType(contentType media.ContentType, content []byte) bool {
	switch contentType {
	case media.ContentTypePNG:
		return len(content) >= 8 &&
			content[0] == 0x89 &&
			content[1] == 'P' &&
			content[2] == 'N' &&
			content[3] == 'G' &&
			content[4] == '\r' &&
			content[5] == '\n' &&
			content[6] == 0x1a &&
			content[7] == '\n'
	case media.ContentTypeJPEG:
		return len(content) >= 3 && content[0] == 0xff && content[1] == 0xd8 && content[2] == 0xff
	case media.ContentTypeWEBP:
		return len(content) >= 12 &&
			string(content[0:4]) == "RIFF" &&
			string(content[8:12]) == "WEBP"
	case media.ContentTypePDF:
		return len(content) >= 5 && string(content[0:5]) == "%PDF-"
	default:
		return false
	}
}
