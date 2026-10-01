package media

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a AttachmentService) ArchiveAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) (media.Attachment, error) {
	return a.updateAttachmentLifecycle(ctx, input, media.LifecycleStateActive, media.LifecycleStateArchived, audit.ActionAttachmentArchived, ports.EventAttachmentArchived, "attachment archived")
}

func (a AttachmentService) RestoreAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) (media.Attachment, error) {
	return a.updateAttachmentLifecycle(ctx, input, media.LifecycleStateArchived, media.LifecycleStateActive, audit.ActionAttachmentRestored, ports.EventAttachmentRestored, "attachment restored")
}

func (a AttachmentService) updateAttachmentLifecycle(ctx context.Context, input UpdateAttachmentLifecycleInput, from media.LifecycleState, to media.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (media.Attachment, error) {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return media.Attachment{}, err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return media.Attachment{}, err
	}
	attachment, found, err := a.deps.Attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return media.Attachment{}, err
	}
	if !found {
		return media.Attachment{}, apperrors.ErrNotFound
	}
	if attachment.LifecycleState != from {
		return media.Attachment{}, apperrors.ErrInvalidInput
	}
	updated := attachment
	updated.LifecycleState = to
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
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
	if err := a.deps.AttachmentUnitOfWork.UpdateAttachmentLifecycle(ctx, updated, auditRecord); err != nil {
		return media.Attachment{}, err
	}
	a.deps.Observer.Record(ctx, ports.Event{
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

func (a AttachmentService) DeleteAttachment(ctx context.Context, input UpdateAttachmentLifecycleInput) error {
	if err := a.deps.Access.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return err
	}
	if err := a.ensureActiveAssetForAttachment(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return err
	}
	attachment, found, err := a.deps.Attachments.AttachmentByID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID)
	if err != nil {
		return err
	}
	if !found {
		return apperrors.ErrNotFound
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
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
	deletionEventID := a.deps.IDs.NewID()
	_, removed, err := a.deps.AttachmentUnitOfWork.DeleteAttachmentAndEnqueueBlobDeletion(ctx, deletionEventID, input.TenantID, input.InventoryID, input.AssetID, input.AttachmentID, auditRecord)
	if err != nil {
		return err
	}
	if !removed {
		return apperrors.ErrNotFound
	}
	a.drainBlobDeletionOutboxBestEffort(ctx, 1)
	a.deps.Observer.Record(ctx, ports.Event{
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
