package media

import (
	"context"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a AttachmentService) ListAttachments(ctx context.Context, input ListAttachmentsInput) (ListAttachmentsResult, error) {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAttachmentsResult{}, err
	}
	if err := a.ensureReadableAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return ListAttachmentsResult{}, err
	}
	limit := appsupport.PageLimit(a.deps.DefaultPageLimit, a.deps.MaxPageLimit, input.Limit)
	afterAttachmentID, err := decodeAttachmentCursor(input.TenantID, input.InventoryID, input.AssetID, input.Cursor)
	if err != nil {
		return ListAttachmentsResult{}, apperrors.ErrInvalidInput
	}
	items, err := a.deps.Attachments.ListAttachmentsByAsset(ctx, input.TenantID, input.InventoryID, input.AssetID, ports.AttachmentListPageRequest{
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
	a.deps.Observer.Record(ctx, ports.Event{
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
		Principal:   input.Principal,
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

func (a AttachmentService) GetAttachment(ctx context.Context, input GetAttachmentInput) (media.Attachment, error) {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return media.Attachment{}, err
	}
	if err := a.ensureReadableAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
	}
	attachment, found, err := a.deps.Attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return media.Attachment{}, err
	}
	if !found {
		return media.Attachment{}, apperrors.ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   input.Principal,
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
	a.deps.Observer.Record(ctx, ports.Event{
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

func (a AttachmentService) DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (AttachmentContentResult, error) {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return AttachmentContentResult{}, err
	}
	if err := a.ensureReadableAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return AttachmentContentResult{}, err
	}
	attachment, found, err := a.deps.Attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return AttachmentContentResult{}, err
	}
	if !found {
		return AttachmentContentResult{}, apperrors.ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   input.Principal,
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
	content, err := a.deps.Blobs.GetBlob(ctx, attachment.StorageKey)
	if err != nil {
		a.deps.Observer.Record(ctx, ports.Event{Name: ports.EventBlobStorageFailed, Message: "blob storage failed"})
		return AttachmentContentResult{}, err
	}
	a.deps.Observer.Record(ctx, ports.Event{
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
