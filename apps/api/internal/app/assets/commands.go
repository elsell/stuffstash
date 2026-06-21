package assets

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) CreateAsset(ctx context.Context, input CreateAssetInput) (asset.Asset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionCreateAsset); err != nil {
		return asset.Asset{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return asset.Asset{}, err
	}

	kind, ok := asset.NewKind(input.Kind)
	if !ok {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	title, ok := asset.NewTitle(input.Title)
	if !ok {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	customAssetTypeID, err := s.validatedAssetCustomAssetTypeID(ctx, input.TenantID, input.InventoryID, input.CustomAssetTypeID)
	if err != nil {
		return asset.Asset{}, err
	}
	customFields, err := s.validatedCustomFields(ctx, input.TenantID, input.InventoryID, customAssetTypeID, input.CustomFields)
	if err != nil {
		return asset.Asset{}, err
	}

	id, ok := asset.NewID(s.newID())
	if !ok {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}

	parentAssetID := asset.ID("")
	if strings.TrimSpace(input.ParentAssetID) != "" {
		parsedParentID, ok := asset.NewID(input.ParentAssetID)
		if !ok {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
		parent, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, parsedParentID)
		if err != nil {
			return asset.Asset{}, err
		}
		if !found {
			return asset.Asset{}, apperrors.ErrNotFound
		}
		if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
		parentAssetID = parsedParentID
	}

	item := asset.Asset{
		ID:                id,
		TenantID:          asset.TenantID(input.TenantID.String()),
		InventoryID:       asset.InventoryID(input.InventoryID.String()),
		ParentAssetID:     parentAssetID,
		CustomAssetTypeID: customAssetTypeID,
		Kind:              kind,
		Title:             title,
		Description:       asset.NewDescription(input.Description),
		CustomFields:      customFields,
		LifecycleState:    asset.LifecycleStateActive,
	}

	undoableOperation, err := s.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, audit.ActionAssetCreated, nil, item)
	if err != nil {
		return asset.Asset{}, err
	}

	auditRecord, err := s.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetCreated,
		TargetType:  audit.TargetAsset,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"asset_kind":   item.Kind.String(),
			"operation_id": undoableOperation.ID,
			"title":        item.Title.String(),
		},
	})
	if err != nil {
		return asset.Asset{}, err
	}
	if item.CustomAssetTypeID.String() != "" {
		auditRecord.Metadata["custom_asset_type_id"] = item.CustomAssetTypeID.String()
	}

	if s.assetUnitOfWork == nil {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.CreateAsset(ctx, item, auditRecord, &undoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetCreated,
		Message: "asset created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"asset_kind":   item.Kind.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return item, nil
}

func (s Service) UpdateAsset(ctx context.Context, input UpdateAssetInput) (asset.Asset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}

	current, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, apperrors.ErrNotFound
	}
	if current.LifecycleState != asset.LifecycleStateActive {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	updated := current
	parentChanged := false
	fieldsChanged := false

	if input.Title != nil {
		title, ok := asset.NewTitle(*input.Title)
		if !ok {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
		updated.Title = title
		if updated.Title != current.Title {
			fieldsChanged = true
		}
	}
	if input.Description != nil {
		updated.Description = asset.NewDescription(*input.Description)
		if updated.Description != current.Description {
			fieldsChanged = true
		}
	}
	if input.CustomFields != nil {
		customFields, err := s.validatedCustomFields(ctx, input.TenantID, input.InventoryID, current.CustomAssetTypeID, input.CustomFields)
		if err != nil {
			return asset.Asset{}, err
		}
		updated.CustomFields = customFields
		fieldsChanged = true
	}
	if input.ParentAssetID.Present {
		parentAssetID := asset.ID("")
		if !input.ParentAssetID.Null {
			parentAsset, err := s.validatedParentAssetID(ctx, input.TenantID, input.InventoryID, input.AssetID, input.ParentAssetID.Value)
			if err != nil {
				return asset.Asset{}, err
			}
			parentAssetID = parentAsset
		}
		updated.ParentAssetID = parentAssetID
		if updated.ParentAssetID != current.ParentAssetID {
			parentChanged = true
		}
	}

	var undoableOperation *ports.UndoableOperation
	if fieldsChanged || parentChanged {
		operationAction := audit.ActionAssetMoved
		if fieldsChanged {
			operationAction = audit.ActionAssetUpdated
		}
		operation, err := s.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, operationAction, &current, updated)
		if err != nil {
			return asset.Asset{}, err
		}
		undoableOperation = &operation
	}

	auditRecords := []audit.Record{}
	if fieldsChanged {
		metadata := map[string]string{"asset_kind": updated.Kind.String()}
		if undoableOperation != nil {
			metadata["operation_id"] = undoableOperation.ID
		}
		auditRecord, err := s.newAuditRecord(auditRecordInput{
			PrincipalID: input.Principal.ID,
			TenantID:    input.TenantID,
			InventoryID: input.InventoryID,
			Source:      input.Source,
			RequestID:   input.RequestID,
			Action:      audit.ActionAssetUpdated,
			TargetType:  audit.TargetAsset,
			TargetID:    updated.ID.String(),
			Metadata:    metadata,
		})
		if err != nil {
			return asset.Asset{}, err
		}
		auditRecords = append(auditRecords, auditRecord)
	}
	if parentChanged {
		metadata := map[string]string{
			"asset_kind":      updated.Kind.String(),
			"previous_parent": current.ParentAssetID.String(),
			"new_parent":      updated.ParentAssetID.String(),
		}
		if undoableOperation != nil {
			metadata["operation_id"] = undoableOperation.ID
		}
		auditRecord, err := s.newAuditRecord(auditRecordInput{
			PrincipalID: input.Principal.ID,
			TenantID:    input.TenantID,
			InventoryID: input.InventoryID,
			Source:      input.Source,
			RequestID:   input.RequestID,
			Action:      audit.ActionAssetMoved,
			TargetType:  audit.TargetAsset,
			TargetID:    updated.ID.String(),
			Metadata:    metadata,
		})
		if err != nil {
			return asset.Asset{}, err
		}
		auditRecords = append(auditRecords, auditRecord)
	}

	if s.assetUnitOfWork == nil {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.UpdateAsset(ctx, updated, auditRecords, undoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetUpdated,
		Message: "asset updated",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     updated.ID.String(),
			"asset_kind":   updated.Kind.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})

	return updated, nil
}

func (s Service) ArchiveAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	return s.updateAssetLifecycle(ctx, input, asset.LifecycleStateActive, asset.LifecycleStateArchived, audit.ActionAssetArchived, ports.EventAssetArchived, "asset archived")
}

func (s Service) RestoreAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	return s.updateAssetLifecycle(ctx, input, asset.LifecycleStateArchived, asset.LifecycleStateActive, audit.ActionAssetRestored, ports.EventAssetRestored, "asset restored")
}

func (s Service) DeleteAsset(ctx context.Context, input UpdateAssetLifecycleInput) error {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return err
	}
	if input.AssetID.String() == "" {
		return apperrors.ErrInvalidInput
	}
	item, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return err
	}
	if !found {
		return apperrors.ErrNotFound
	}
	auditRecord, err := s.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetDeleted,
		TargetType:  audit.TargetAsset,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"asset_kind":      item.Kind.String(),
			"lifecycle_state": item.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if s.assetUnitOfWork == nil {
		return apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.DeleteAsset(ctx, input.TenantID, input.InventoryID, input.AssetID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetDeleted,
		Message: "asset deleted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"asset_id":     input.AssetID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return nil
}

func (s Service) updateAssetLifecycle(ctx context.Context, input UpdateAssetLifecycleInput, from asset.LifecycleState, to asset.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (asset.Asset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return asset.Asset{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}

	current, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, apperrors.ErrNotFound
	}
	if current.LifecycleState != from {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	if to == asset.LifecycleStateArchived {
		hasActiveChildren, err := s.assets.AssetHasActiveChildren(ctx, input.TenantID, input.InventoryID, input.AssetID)
		if err != nil {
			return asset.Asset{}, err
		}
		if hasActiveChildren {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
	}
	if to == asset.LifecycleStateActive && current.ParentAssetID.String() != "" {
		parent, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, current.ParentAssetID)
		if err != nil {
			return asset.Asset{}, err
		}
		if !found || parent.LifecycleState != asset.LifecycleStateActive {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
	}

	updated := current
	updated.LifecycleState = to
	undoableOperation, err := s.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, action, &current, updated)
	if err != nil {
		return asset.Asset{}, err
	}
	auditRecord, err := s.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      action,
		TargetType:  audit.TargetAsset,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"asset_kind":      updated.Kind.String(),
			"operation_id":    undoableOperation.ID,
			"previous_state":  current.LifecycleState.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	if err != nil {
		return asset.Asset{}, err
	}

	if s.assetUnitOfWork == nil {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.UpdateAssetLifecycle(ctx, updated, auditRecord, &undoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, apperrors.ErrInvalidInput
		}
		return asset.Asset{}, err
	}

	s.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       input.TenantID.String(),
			"inventory_id":    input.InventoryID.String(),
			"asset_id":        updated.ID.String(),
			"asset_kind":      updated.Kind.String(),
			"principal_id":    input.Principal.ID.String(),
			"lifecycle_state": updated.LifecycleState.String(),
			"previous_state":  current.LifecycleState.String(),
		},
	})

	return updated, nil
}
