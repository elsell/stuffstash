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

func (a App) CreateAttachment(ctx context.Context, input CreateAttachmentInput) (media.Attachment, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return media.Attachment{}, err
	}
	if a.attachments == nil || a.blobs == nil {
		return media.Attachment{}, ErrInvalidInput
	}
	if input.AssetID.String() == "" {
		return media.Attachment{}, ErrInvalidInput
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
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
		a.clock.Now().UTC(),
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
	if err := a.attachmentUnitOfWork.SaveAttachment(ctx, attachment, auditRecord); err != nil {
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
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAttachmentsResult{}, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return ListAttachmentsResult{}, err
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
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAttachmentListed,
		TargetType:  audit.TargetAsset,
		TargetID:    input.AssetID.String(),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListAttachmentsResult{}, err
	}
	return ListAttachmentsResult{Items: items, Limit: limit, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (a App) GetAttachment(ctx context.Context, input GetAttachmentInput) (media.Attachment, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return media.Attachment{}, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return media.Attachment{}, err
	}
	if !found {
		return media.Attachment{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAttachmentViewed,
		TargetType:  audit.TargetAttachment,
		TargetID:    attachment.ID.String(),
		Metadata: map[string]string{
			"asset_id":         input.AssetID.String(),
			"lifecycle_state":  attachment.LifecycleState.String(),
			"attachment_bytes": strconv.FormatInt(attachment.SizeBytes, 10),
		},
	}); err != nil {
		return media.Attachment{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentViewed,
		Message: "attachment viewed",
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

func (a App) DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (AttachmentContentResult, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return AttachmentContentResult{}, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return AttachmentContentResult{}, err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return AttachmentContentResult{}, err
	}
	if !found {
		return AttachmentContentResult{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAttachmentContentDownloaded,
		TargetType:  audit.TargetAttachment,
		TargetID:    attachment.ID.String(),
		Metadata: map[string]string{
			"asset_id": input.AssetID.String(),
		},
	}); err != nil {
		return AttachmentContentResult{}, err
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

func (a App) ArchiveAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) (media.Attachment, error) {
	return a.updateAttachmentLifecycle(ctx, input, media.LifecycleStateActive, media.LifecycleStateArchived, audit.ActionAttachmentArchived, ports.EventAttachmentArchived, "attachment archived")
}

func (a App) RestoreAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) (media.Attachment, error) {
	return a.updateAttachmentLifecycle(ctx, input, media.LifecycleStateArchived, media.LifecycleStateActive, audit.ActionAttachmentRestored, ports.EventAttachmentRestored, "attachment restored")
}

func (a App) updateAttachmentLifecycle(ctx context.Context, input UpdateAttachmentLifecycleInput, from media.LifecycleState, to media.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (media.Attachment, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return media.Attachment{}, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return media.Attachment{}, err
	}
	if !found {
		return media.Attachment{}, ErrNotFound
	}
	if attachment.LifecycleState != from {
		return media.Attachment{}, ErrInvalidInput
	}
	updated := attachment
	updated.LifecycleState = to
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      action,
		TargetType:  audit.TargetAttachment,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"asset_id":         input.AssetID.String(),
			"previous_state":   attachment.LifecycleState.String(),
			"lifecycle_state":  updated.LifecycleState.String(),
			"attachment_bytes": strconv.FormatInt(updated.SizeBytes, 10),
		},
	})
	if err != nil {
		return media.Attachment{}, err
	}
	if err := a.attachmentUnitOfWork.UpdateAttachmentLifecycle(ctx, updated, auditRecord); err != nil {
		return media.Attachment{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       input.TenantID.String(),
			"inventory_id":    input.InventoryID.String(),
			"asset_id":        input.AssetID.String(),
			"attachment_id":   updated.ID.String(),
			"principal_id":    input.Principal.ID.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	return updated, nil
}

func (a App) DeleteAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) error {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return err
	}
	attachment, found, err := a.attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAttachmentDeleted,
		TargetType:  audit.TargetAttachment,
		TargetID:    attachment.ID.String(),
		Metadata: map[string]string{
			"asset_id":         input.AssetID.String(),
			"lifecycle_state":  attachment.LifecycleState.String(),
			"attachment_bytes": strconv.FormatInt(attachment.SizeBytes, 10),
		},
	})
	if err != nil {
		return err
	}
	deletionEventID := a.ids.NewID()
	_, removed, err := a.attachmentUnitOfWork.DeleteAttachmentAndEnqueueBlobDeletion(ctx, deletionEventID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID, auditRecord)
	if err != nil {
		return err
	}
	if !removed {
		return ErrNotFound
	}
	a.drainBlobDeletionOutboxBestEffort(ctx, 1)
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventAttachmentDeleted,
		Message: "attachment deleted",
		Fields: map[string]string{
			"tenant_id":     input.TenantID.String(),
			"inventory_id":  input.InventoryID.String(),
			"asset_id":      input.AssetID.String(),
			"attachment_id": input.AttachmentID.String(),
			"principal_id":  input.Principal.ID.String(),
		},
	})
	return nil
}

func (a App) drainBlobDeletionOutboxBestEffort(ctx context.Context, limit int) {
	if err := a.DrainBlobDeletionOutbox(ctx, limit); err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobDeletionOutboxFailed,
			Message: "blob deletion outbox drain failed",
			Fields:  map[string]string{"error": err.Error()},
		})
	}
}

func (a App) DrainBlobDeletionOutbox(ctx context.Context, limit int) error {
	if a.blobDeletionOutbox == nil || a.blobs == nil {
		return nil
	}
	if limit <= 0 {
		limit = 1
	}
	claimID := a.ids.NewID()
	now := a.clock.Now().UTC()
	events, err := a.blobDeletionOutbox.ClaimPendingBlobDeletionEvents(ctx, claimID, limit, now, now.Add(a.blobDeletionClaimLease))
	if err != nil {
		return err
	}
	if len(events) > 0 {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobDeletionOutboxClaimed,
			Message: "blob deletion outbox events claimed",
			Fields: map[string]string{
				"event_count": strconv.Itoa(len(events)),
			},
		})
	}
	for _, event := range events {
		if err := a.blobs.DeleteBlob(ctx, event.StorageKey); err != nil && !errors.Is(err, ports.ErrBlobNotFound) {
			a.observer.Record(ctx, ports.Event{
				Name:    ports.EventBlobDeletionOutboxFailed,
				Message: "blob deletion outbox event failed",
				Fields: map[string]string{
					"event_id": event.ID,
					"attempts": strconv.Itoa(event.Attempts + 1),
				},
			})
			if event.Attempts+1 >= a.blobDeletionMaxAttempts {
				if markErr := a.blobDeletionOutbox.MarkBlobDeletionEventDeadLettered(ctx, event.ID, claimID, err.Error()); markErr != nil {
					return markErr
				}
				a.observer.Record(ctx, ports.Event{
					Name:    ports.EventBlobDeletionOutboxDeadLettered,
					Message: "blob deletion outbox event dead-lettered",
					Fields: map[string]string{
						"event_id": event.ID,
						"attempts": strconv.Itoa(event.Attempts + 1),
					},
				})
			} else {
				if markErr := a.blobDeletionOutbox.MarkBlobDeletionEventFailed(ctx, event.ID, claimID, err.Error()); markErr != nil {
					return markErr
				}
			}
			continue
		}
		if err := a.blobDeletionOutbox.MarkBlobDeletionEventProcessed(ctx, event.ID, claimID); err != nil {
			return err
		}
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventBlobDeletionOutboxProcessed,
			Message: "blob deletion outbox event processed",
			Fields: map[string]string{
				"event_id": event.ID,
			},
		})
	}
	return nil
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

func (a App) ensureActiveAssetForAttachment(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) error {
	item, found, err := a.assets.AssetByID(ctx, tenantID, inventoryID, assetID)
	if err != nil {
		return err
	}
	if !found || item.LifecycleState != asset.LifecycleStateActive {
		return ErrNotFound
	}
	return nil
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
