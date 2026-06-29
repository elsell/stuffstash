package assets

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) CreateAsset(ctx context.Context, input CreateAssetInput) (asset.Asset, error) {
	prepared, err := s.PrepareCreateAsset(ctx, input)
	if err != nil {
		return asset.Asset{}, err
	}
	if err := s.persistPreparedCreateAsset(ctx, prepared); err != nil {
		return asset.Asset{}, err
	}
	s.RecordAssetCreated(ctx, prepared.Asset, input.Principal.ID)
	return prepared.Asset, nil
}

type PreparedCreateAsset struct {
	Asset             asset.Asset
	AuditRecord       audit.Record
	UndoableOperation ports.UndoableOperation
}

func (s Service) PrepareCreateAsset(ctx context.Context, input CreateAssetInput) (PreparedCreateAsset, error) {
	return s.prepareCreateAsset(ctx, input, nil)
}

func (s Service) PrepareCreateAssetWithPendingParents(ctx context.Context, input CreateAssetInput, pendingParents map[asset.ID]asset.Kind) (PreparedCreateAsset, error) {
	return s.prepareCreateAsset(ctx, input, pendingParents)
}

func (s Service) prepareCreateAsset(ctx context.Context, input CreateAssetInput, pendingParents map[asset.ID]asset.Kind) (PreparedCreateAsset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionCreateAsset); err != nil {
		return PreparedCreateAsset{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return PreparedCreateAsset{}, err
	}

	kind, ok := asset.NewKind(input.Kind)
	if !ok {
		return PreparedCreateAsset{}, apperrors.ErrInvalidInput
	}
	title, ok := asset.NewTitle(input.Title)
	if !ok {
		return PreparedCreateAsset{}, apperrors.ErrInvalidInput
	}
	customAssetTypeID, err := s.validatedAssetCustomAssetTypeID(ctx, input.TenantID, input.InventoryID, input.CustomAssetTypeID)
	if err != nil {
		return PreparedCreateAsset{}, err
	}
	customFields, err := s.validatedCustomFields(ctx, input.TenantID, input.InventoryID, customAssetTypeID, input.CustomFields)
	if err != nil {
		return PreparedCreateAsset{}, err
	}

	id, ok := asset.NewID(s.newID())
	if !ok {
		return PreparedCreateAsset{}, apperrors.ErrInvalidInput
	}
	now := s.now().UTC()

	parentAssetID := asset.ID("")
	if strings.TrimSpace(input.ParentAssetID) != "" {
		parsedParentID, ok := asset.NewID(input.ParentAssetID)
		if !ok {
			return PreparedCreateAsset{}, apperrors.ErrInvalidInput
		}
		if pendingKind, ok := pendingParents[parsedParentID]; ok {
			if !pendingKind.CanContainChildren() {
				return PreparedCreateAsset{}, apperrors.ErrInvalidInput
			}
		} else {
			parent, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, parsedParentID)
			if err != nil {
				return PreparedCreateAsset{}, err
			}
			if !found {
				return PreparedCreateAsset{}, apperrors.ErrNotFound
			}
			if !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
				return PreparedCreateAsset{}, apperrors.ErrInvalidInput
			}
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
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	undoableOperation, err := s.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, audit.ActionAssetCreated, nil, item)
	if err != nil {
		return PreparedCreateAsset{}, err
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
		return PreparedCreateAsset{}, err
	}
	if item.CustomAssetTypeID.String() != "" {
		auditRecord.Metadata["custom_asset_type_id"] = item.CustomAssetTypeID.String()
	}

	return PreparedCreateAsset{Asset: item, AuditRecord: auditRecord, UndoableOperation: undoableOperation}, nil
}

func (s Service) persistPreparedCreateAsset(ctx context.Context, prepared PreparedCreateAsset) error {
	if s.assetUnitOfWork == nil {
		return apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.CreateAsset(ctx, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	return nil
}

func (s Service) RecordAssetCreated(ctx context.Context, item asset.Asset, principalID identity.PrincipalID) {
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetCreated,
		Message: "asset created",
		Fields: map[string]string{
			"tenant_id":    item.TenantID.String(),
			"inventory_id": item.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"asset_kind":   item.Kind.String(),
			"principal_id": principalID.String(),
		},
	})
}

func (s Service) UpdateAsset(ctx context.Context, input UpdateAssetInput) (asset.Asset, error) {
	prepared, err := s.PrepareUpdateAsset(ctx, input)
	if err != nil {
		return asset.Asset{}, err
	}
	if err := s.persistPreparedUpdateAsset(ctx, prepared); err != nil {
		return asset.Asset{}, err
	}
	s.RecordAssetUpdated(ctx, prepared.Asset, input.Principal.ID)
	return prepared.Asset, nil
}

type PreparedUpdateAsset struct {
	PreviousAsset     asset.Asset
	Asset             asset.Asset
	AuditRecords      []audit.Record
	UndoableOperation *ports.UndoableOperation
}

func (s Service) PrepareUpdateAsset(ctx context.Context, input UpdateAssetInput) (PreparedUpdateAsset, error) {
	return s.prepareUpdateAsset(ctx, input, nil)
}

func (s Service) PrepareUpdateAssetWithPendingParents(ctx context.Context, input UpdateAssetInput, pendingParents map[asset.ID]asset.Kind) (PreparedUpdateAsset, error) {
	return s.prepareUpdateAsset(ctx, input, pendingParents)
}

func (s Service) prepareUpdateAsset(ctx context.Context, input UpdateAssetInput, pendingParents map[asset.ID]asset.Kind) (PreparedUpdateAsset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return PreparedUpdateAsset{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return PreparedUpdateAsset{}, err
	}
	if input.AssetID.String() == "" {
		return PreparedUpdateAsset{}, apperrors.ErrInvalidInput
	}

	current, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return PreparedUpdateAsset{}, err
	}
	if !found {
		return PreparedUpdateAsset{}, apperrors.ErrNotFound
	}
	if current.LifecycleState != asset.LifecycleStateActive {
		return PreparedUpdateAsset{}, apperrors.ErrInvalidInput
	}
	updated := current
	parentChanged := false
	fieldsChanged := false

	if input.Title != nil {
		title, ok := asset.NewTitle(*input.Title)
		if !ok {
			return PreparedUpdateAsset{}, apperrors.ErrInvalidInput
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
			return PreparedUpdateAsset{}, err
		}
		updated.CustomFields = customFields
		fieldsChanged = true
	}
	if input.ParentAssetID.Present {
		parentAssetID := asset.ID("")
		if !input.ParentAssetID.Null {
			parentAsset, err := s.validatedParentAssetIDWithPendingParents(ctx, input.TenantID, input.InventoryID, input.AssetID, input.ParentAssetID.Value, pendingParents)
			if err != nil {
				return PreparedUpdateAsset{}, err
			}
			parentAssetID = parentAsset
		}
		updated.ParentAssetID = parentAssetID
		if updated.ParentAssetID != current.ParentAssetID {
			parentChanged = true
		}
	}
	if fieldsChanged || parentChanged {
		updated.UpdatedAt = s.now().UTC()
	}

	var undoableOperation *ports.UndoableOperation
	if fieldsChanged || parentChanged {
		operationAction := audit.ActionAssetMoved
		if fieldsChanged {
			operationAction = audit.ActionAssetUpdated
		}
		operation, err := s.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, operationAction, &current, updated)
		if err != nil {
			return PreparedUpdateAsset{}, err
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
			return PreparedUpdateAsset{}, err
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
			return PreparedUpdateAsset{}, err
		}
		auditRecords = append(auditRecords, auditRecord)
	}

	return PreparedUpdateAsset{PreviousAsset: current, Asset: updated, AuditRecords: auditRecords, UndoableOperation: undoableOperation}, nil
}

func (s Service) persistPreparedUpdateAsset(ctx context.Context, prepared PreparedUpdateAsset) error {
	if s.assetUnitOfWork == nil {
		return apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.UpdateAsset(ctx, prepared.Asset, prepared.AuditRecords, prepared.UndoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	return nil
}

func (s Service) RecordAssetUpdated(ctx context.Context, item asset.Asset, principalID identity.PrincipalID) {
	s.observer.Record(ctx, ports.Event{
		Name:    ports.EventAssetUpdated,
		Message: "asset updated",
		Fields: map[string]string{
			"tenant_id":    item.TenantID.String(),
			"inventory_id": item.InventoryID.String(),
			"asset_id":     item.ID.String(),
			"asset_kind":   item.Kind.String(),
			"principal_id": principalID.String(),
		},
	})
}

func (s Service) ArchiveAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	prepared, err := s.PrepareArchiveAsset(ctx, input)
	if err != nil {
		return asset.Asset{}, err
	}
	if err := s.persistPreparedUpdateAssetLifecycle(ctx, prepared); err != nil {
		return asset.Asset{}, err
	}
	s.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
	return prepared.Asset, nil
}

func (s Service) RestoreAsset(ctx context.Context, input UpdateAssetLifecycleInput) (asset.Asset, error) {
	prepared, err := s.PrepareRestoreAsset(ctx, input)
	if err != nil {
		return asset.Asset{}, err
	}
	if err := s.persistPreparedUpdateAssetLifecycle(ctx, prepared); err != nil {
		return asset.Asset{}, err
	}
	s.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
	return prepared.Asset, nil
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

type PreparedUpdateAssetLifecycle struct {
	PreviousAsset     asset.Asset
	Asset             asset.Asset
	AuditRecord       audit.Record
	UndoableOperation ports.UndoableOperation
	EventName         ports.EventName
	EventMessage      string
}

func (s Service) PrepareArchiveAsset(ctx context.Context, input UpdateAssetLifecycleInput) (PreparedUpdateAssetLifecycle, error) {
	return s.prepareUpdateAssetLifecycle(ctx, input, asset.LifecycleStateActive, asset.LifecycleStateArchived, audit.ActionAssetArchived, ports.EventAssetArchived, "asset archived")
}

func (s Service) PrepareRestoreAsset(ctx context.Context, input UpdateAssetLifecycleInput) (PreparedUpdateAssetLifecycle, error) {
	return s.prepareUpdateAssetLifecycle(ctx, input, asset.LifecycleStateArchived, asset.LifecycleStateActive, audit.ActionAssetRestored, ports.EventAssetRestored, "asset restored")
}

func (s Service) prepareUpdateAssetLifecycle(ctx context.Context, input UpdateAssetLifecycleInput, from asset.LifecycleState, to asset.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (PreparedUpdateAssetLifecycle, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return PreparedUpdateAssetLifecycle{}, err
	}
	if err := s.ensureAssetRepository(); err != nil {
		return PreparedUpdateAssetLifecycle{}, err
	}
	if input.AssetID.String() == "" {
		return PreparedUpdateAssetLifecycle{}, apperrors.ErrInvalidInput
	}

	current, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return PreparedUpdateAssetLifecycle{}, err
	}
	if !found {
		return PreparedUpdateAssetLifecycle{}, apperrors.ErrNotFound
	}
	if current.LifecycleState != from {
		return PreparedUpdateAssetLifecycle{}, apperrors.ErrInvalidInput
	}
	if to == asset.LifecycleStateArchived {
		hasActiveChildren, err := s.assets.AssetHasActiveChildren(ctx, input.TenantID, input.InventoryID, input.AssetID)
		if err != nil {
			return PreparedUpdateAssetLifecycle{}, err
		}
		if hasActiveChildren {
			return PreparedUpdateAssetLifecycle{}, apperrors.ErrInvalidInput
		}
	}
	if to == asset.LifecycleStateActive && current.ParentAssetID.String() != "" {
		parent, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, current.ParentAssetID)
		if err != nil {
			return PreparedUpdateAssetLifecycle{}, err
		}
		if !found || parent.LifecycleState != asset.LifecycleStateActive {
			return PreparedUpdateAssetLifecycle{}, apperrors.ErrInvalidInput
		}
	}

	updated := current
	updated.LifecycleState = to
	updated.UpdatedAt = s.now().UTC()
	undoableOperation, err := s.newAssetUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, action, &current, updated)
	if err != nil {
		return PreparedUpdateAssetLifecycle{}, err
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
		return PreparedUpdateAssetLifecycle{}, err
	}
	return PreparedUpdateAssetLifecycle{PreviousAsset: current, Asset: updated, AuditRecord: auditRecord, UndoableOperation: undoableOperation, EventName: eventName, EventMessage: eventMessage}, nil
}

func (s Service) persistPreparedUpdateAssetLifecycle(ctx context.Context, prepared PreparedUpdateAssetLifecycle) error {
	if s.assetUnitOfWork == nil {
		return apperrors.ErrInvalidInput
	}
	if err := s.assetUnitOfWork.UpdateAssetLifecycle(ctx, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation); err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	return nil
}

func (s Service) RecordAssetLifecycleUpdated(ctx context.Context, prepared PreparedUpdateAssetLifecycle, principalID identity.PrincipalID) {
	s.observer.Record(ctx, ports.Event{
		Name:    prepared.EventName,
		Message: prepared.EventMessage,
		Fields: map[string]string{
			"tenant_id":       prepared.Asset.TenantID.String(),
			"inventory_id":    prepared.Asset.InventoryID.String(),
			"asset_id":        prepared.Asset.ID.String(),
			"asset_kind":      prepared.Asset.Kind.String(),
			"principal_id":    principalID.String(),
			"lifecycle_state": prepared.Asset.LifecycleState.String(),
			"previous_state":  prepared.PreviousAsset.LifecycleState.String(),
		},
	})
}
