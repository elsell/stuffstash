package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// importTargets composes import orchestration with target-domain validation.
// Prepared records still commit through the import unit of work.
type importTargets struct{ app App }

var _ ports.ImportTargets = importTargets{}

func (t importTargets) CreateField(ctx context.Context, command ports.ImportJobCommand, field importplan.FieldDefinition) error {
	_, err := t.app.CreateInventoryCustomFieldDefinition(ctx, CreateCustomFieldDefinitionInput{
		Principal:     command.Principal,
		Source:        audit.SourceImport,
		RequestID:     command.RequestID,
		TenantID:      command.TenantID,
		InventoryID:   command.InventoryID,
		Key:           field.Key,
		DisplayName:   field.DisplayName,
		Type:          field.Type,
		Applicability: customfield.ApplicabilityAllAssets.String(),
	})
	return err
}
func (t importTargets) CreateTag(ctx context.Context, command ports.ImportJobCommand, tag importplan.TagDefinition) (assettag.Tag, error) {
	return t.app.CreateAssetTag(ctx, CreateAssetTagInput{
		Principal:   command.Principal,
		Source:      audit.SourceImport,
		RequestID:   command.RequestID,
		TenantID:    command.TenantID,
		InventoryID: command.InventoryID,
		Key:         tag.Key,
		DisplayName: tag.DisplayName,
		Color:       tag.Color,
	})
}
func (t importTargets) PrepareAsset(ctx context.Context, command ports.ImportJobCommand, planned importplan.Asset, parentAssetID string) (ports.PreparedImportAsset, error) {
	prepared, err := t.app.assetService.PrepareCreateAsset(ctx, CreateAssetInput{
		Principal:     command.Principal,
		Source:        audit.SourceImport,
		RequestID:     command.RequestID,
		TenantID:      command.TenantID,
		InventoryID:   command.InventoryID,
		Kind:          planned.Kind,
		Title:         planned.Title,
		Description:   planned.Description,
		ParentAssetID: parentAssetID,
		CustomFields:  planned.CustomFields,
	})
	return ports.PreparedImportAsset{Asset: prepared.Asset, AuditRecord: prepared.AuditRecord, PromotedParent: prepared.PromotedParent, ParentPromotionRecord: prepared.ParentPromotionRecord, UndoableOperation: prepared.UndoableOperation}, err
}
func (t importTargets) PrepareAttachment(ctx context.Context, command ports.ImportJobCommand, assetID asset.ID, planned importplan.Attachment) (ports.PreparedImportAttachment, error) {
	prepared, err := t.app.prepareAttachment(ctx, CreateAttachmentInput{
		Principal:   command.Principal,
		Source:      audit.SourceImport,
		RequestID:   command.RequestID,
		TenantID:    command.TenantID,
		InventoryID: command.InventoryID,
		AssetID:     assetID,
		FileName:    planned.FileName,
		ContentType: planned.ContentType,
		Content:     planned.Content,
	})
	return ports.PreparedImportAttachment{ThumbnailJob: prepared.ThumbnailJob, Attachment: prepared.Attachment, AuditRecord: prepared.AuditRecord, StorageKey: prepared.StorageKey, ContentType: prepared.ContentType}, err
}
func (t importTargets) RecordAttachmentCreated(ctx context.Context, command ports.ImportJobCommand, assetID asset.ID, attachment media.Attachment) {
	t.app.recordAttachmentCreated(ctx, CreateAttachmentInput{
		Principal:   command.Principal,
		TenantID:    command.TenantID,
		InventoryID: command.InventoryID,
		AssetID:     assetID,
	}, attachment)
}
func (t importTargets) DeleteAttachment(ctx context.Context, command ports.ImportJobCommand, assetID asset.ID, attachmentID media.ID) error {
	return t.app.DeleteAttachment(ctx, UpdateAttachmentLifecycleInput{
		Principal:    command.Principal,
		Source:       audit.SourceImport,
		RequestID:    command.RequestID,
		TenantID:     command.TenantID,
		InventoryID:  command.InventoryID,
		AssetID:      assetID,
		AttachmentID: attachmentID,
	})
}
func (t importTargets) DeleteAsset(ctx context.Context, command ports.ImportJobCommand, assetID asset.ID) error {
	return t.app.DeleteAsset(ctx, UpdateAssetLifecycleInput{
		Principal:   command.Principal,
		Source:      audit.SourceImport,
		RequestID:   command.RequestID,
		TenantID:    command.TenantID,
		InventoryID: command.InventoryID,
		AssetID:     assetID,
	})
}

func (t importTargets) EnsureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	return t.app.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, permission)
}
func (t importTargets) SetAssetTags(ctx context.Context, command ports.ImportJobCommand, assetID asset.ID, ids []string) error {
	return t.app.assetService.SetAssetTagAssignmentsForImport(ctx, command.Principal, command.RequestID, command.TenantID, command.InventoryID, assetID, ids)
}
func (t importTargets) RecordAssetCreated(ctx context.Context, item asset.Asset, principalID identity.PrincipalID) {
	t.app.assetService.RecordAssetCreated(ctx, item, principalID)
}
