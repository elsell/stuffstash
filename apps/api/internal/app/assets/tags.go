package assets

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateAssetTagInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Key         string
	DisplayName string
	Color       string
}

type UpdateAssetTagInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	TagID       assettag.ID
	DisplayName *string
	Color       *string
}

type ListAssetTagsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type AssetTagLifecycleInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	TagID       assettag.ID
}

type GetAssetAssignedTagsInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	AssetID     asset.ID
}

type ListAssetTagsResult struct {
	Items      []assettag.Tag
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (s Service) CreateAssetTag(ctx context.Context, input CreateAssetTagInput) (assettag.Tag, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return assettag.Tag{}, err
	}
	if err := s.ensureAssetTagDependencies(); err != nil {
		return assettag.Tag{}, err
	}
	key, ok := assettag.NewKey(input.Key)
	if !ok {
		if key, ok = assettag.KeyFromDisplayName(input.DisplayName); !ok {
			return assettag.Tag{}, apperrors.ErrInvalidInput
		}
	}
	displayName, ok := assettag.NewDisplayName(input.DisplayName)
	if !ok {
		return assettag.Tag{}, apperrors.ErrInvalidInput
	}
	color, ok := assettag.NewColor(input.Color)
	if !ok {
		return assettag.Tag{}, apperrors.ErrInvalidInput
	}
	if _, found, err := s.assetTags.AssetTagByKey(ctx, input.TenantID, input.InventoryID, key); err != nil {
		return assettag.Tag{}, err
	} else if found {
		return assettag.Tag{}, apperrors.ErrInvalidInput
	}
	id, ok := assettag.NewID(s.newID())
	if !ok {
		return assettag.Tag{}, apperrors.ErrInvalidInput
	}
	now := s.now().UTC()
	tag, ok := assettag.NewTag(id, assettag.TenantID(input.TenantID.String()), assettag.InventoryID(input.InventoryID.String()), key, displayName, color, now)
	if !ok {
		return assettag.Tag{}, apperrors.ErrInvalidInput
	}
	auditRecord, err := s.newAssetTagAuditRecord(input, audit.ActionAssetTagCreated, tag)
	if err != nil {
		return assettag.Tag{}, err
	}
	if err := s.assetTagUnitOfWork.CreateAssetTag(ctx, tag, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return assettag.Tag{}, apperrors.ErrInvalidInput
		}
		return assettag.Tag{}, err
	}
	return tag, nil
}

func (s Service) UpdateAssetTag(ctx context.Context, input UpdateAssetTagInput) (assettag.Tag, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return assettag.Tag{}, err
	}
	if err := s.ensureAssetTagDependencies(); err != nil {
		return assettag.Tag{}, err
	}
	current, found, err := s.assetTags.AssetTagByID(ctx, input.TenantID, input.InventoryID, input.TagID)
	if err != nil {
		return assettag.Tag{}, err
	}
	if !found || current.LifecycleState != assettag.LifecycleStateActive {
		return assettag.Tag{}, apperrors.ErrNotFound
	}
	updated := current
	if input.DisplayName != nil {
		displayName, ok := assettag.NewDisplayName(*input.DisplayName)
		if !ok {
			return assettag.Tag{}, apperrors.ErrInvalidInput
		}
		updated.DisplayName = displayName
	}
	if input.Color != nil {
		color, ok := assettag.NewColor(*input.Color)
		if !ok {
			return assettag.Tag{}, apperrors.ErrInvalidInput
		}
		updated.Color = color
	}
	updated.UpdatedAt = s.now().UTC()
	auditRecord, err := s.newAssetTagAuditRecord(CreateAssetTagInput{
		Principal:   input.Principal,
		Source:      input.Source,
		RequestID:   input.RequestID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
	}, audit.ActionAssetTagUpdated, updated)
	if err != nil {
		return assettag.Tag{}, err
	}
	if err := s.assetTagUnitOfWork.UpdateAssetTag(ctx, updated, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return assettag.Tag{}, apperrors.ErrInvalidInput
		}
		return assettag.Tag{}, err
	}
	return updated, nil
}

func (s Service) ArchiveAssetTag(ctx context.Context, input AssetTagLifecycleInput) (assettag.Tag, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return assettag.Tag{}, err
	}
	if err := s.ensureAssetTagDependencies(); err != nil {
		return assettag.Tag{}, err
	}
	current, found, err := s.assetTags.AssetTagByID(ctx, input.TenantID, input.InventoryID, input.TagID)
	if err != nil {
		return assettag.Tag{}, err
	}
	if !found || current.LifecycleState != assettag.LifecycleStateActive {
		return assettag.Tag{}, apperrors.ErrNotFound
	}
	updated := current
	updated.LifecycleState = assettag.LifecycleStateArchived
	updated.UpdatedAt = s.now().UTC()
	auditRecord, err := s.newAssetTagAuditRecord(CreateAssetTagInput{
		Principal:   input.Principal,
		Source:      input.Source,
		RequestID:   input.RequestID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
	}, audit.ActionAssetTagArchived, updated)
	if err != nil {
		return assettag.Tag{}, err
	}
	if err := s.assetTagUnitOfWork.UpdateAssetTagLifecycle(ctx, updated, auditRecord); err != nil {
		return assettag.Tag{}, err
	}
	return updated, nil
}

func (s Service) ListAssetTags(ctx context.Context, input ListAssetTagsInput) (ListAssetTagsResult, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ListAssetTagsResult{}, err
	}
	if err := s.ensureAssetTagDependencies(); err != nil {
		return ListAssetTagsResult{}, err
	}
	limit := pageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	after := assettag.ID(strings.TrimSpace(input.Cursor))
	items, err := s.assetTags.ListAssetTags(ctx, input.TenantID, input.InventoryID, ports.AssetTagPageRequest{AfterTagID: after, Limit: limit + 1})
	if err != nil {
		return ListAssetTagsResult{}, err
	}
	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		cursor := items[len(items)-1].ID.String()
		nextCursor = &cursor
	}
	if err := s.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetTagListed,
		TargetType:  audit.TargetInventory,
		TargetID:    input.InventoryID.String(),
	}); err != nil {
		return ListAssetTagsResult{}, err
	}
	return ListAssetTagsResult{Items: items, Limit: limit, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s Service) GetAssetAssignedTags(ctx context.Context, input GetAssetAssignedTagsInput) ([]assettag.Tag, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return nil, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return nil, err
	}
	if s.assetTags == nil {
		return nil, nil
	}
	if input.AssetID.String() == "" {
		return nil, apperrors.ErrInvalidInput
	}
	item, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, apperrors.ErrNotFound
	}
	return s.tagsForAsset(ctx, item)
}

func (s Service) ensureAssetTagDependencies() error {
	if s.assetTags == nil || s.assetTagUnitOfWork == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func (s Service) newAssetTagAuditRecord(input CreateAssetTagInput, action audit.Action, tag assettag.Tag) (audit.Record, error) {
	return s.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      action,
		TargetType:  audit.TargetAssetTag,
		TargetID:    tag.ID.String(),
		Metadata: map[string]string{
			"tag_id":  tag.ID.String(),
			"tag_key": tag.Key.String(),
		},
	})
}

func (s Service) setAssetTagAssignments(ctx context.Context, principal identity.Principal, source audit.Source, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, rawTagIDs []string) error {
	if s.assetTagUnitOfWork == nil {
		return apperrors.ErrInvalidInput
	}
	tagIDs, err := s.validateAssignableAssetTagIDs(ctx, tenantID, inventoryID, rawTagIDs)
	if err != nil {
		return err
	}
	auditRecord, err := s.newAuditRecord(auditRecordInput{
		Principal:   principal,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		Source:      source,
		RequestID:   requestID,
		Action:      audit.ActionAssetUpdated,
		TargetType:  audit.TargetAsset,
		TargetID:    assetID.String(),
		Metadata: map[string]string{
			"asset_id":  assetID.String(),
			"tag_count": strconv.Itoa(len(tagIDs)),
		},
	})
	if err != nil {
		return err
	}
	if err := s.assetTagUnitOfWork.SetAssetTags(ctx, tenantID, inventoryID, assetID, tagIDs, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	return nil
}

func (s Service) SetAssetTagAssignmentsForImport(ctx context.Context, principal identity.Principal, requestID string, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, rawTagIDs []string) error {
	return s.setAssetTagAssignments(ctx, principal, audit.SourceImport, requestID, tenantID, inventoryID, assetID, rawTagIDs)
}

func (s Service) validateAssignableAssetTagIDs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, rawTagIDs []string) ([]assettag.ID, error) {
	if len(rawTagIDs) == 0 {
		return nil, nil
	}
	if s.assetTags == nil {
		return nil, apperrors.ErrInvalidInput
	}
	tagIDs := make([]assettag.ID, 0, len(rawTagIDs))
	seen := map[assettag.ID]struct{}{}
	for _, raw := range rawTagIDs {
		tagID, ok := assettag.NewID(raw)
		if !ok {
			return nil, apperrors.ErrInvalidInput
		}
		if _, exists := seen[tagID]; exists {
			continue
		}
		tag, found, err := s.assetTags.AssetTagByID(ctx, tenantID, inventoryID, tagID)
		if err != nil {
			return nil, err
		}
		if !found || tag.LifecycleState != assettag.LifecycleStateActive {
			return nil, apperrors.ErrInvalidInput
		}
		seen[tagID] = struct{}{}
		tagIDs = append(tagIDs, tagID)
	}
	return tagIDs, nil
}
