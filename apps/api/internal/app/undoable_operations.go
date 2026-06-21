package app

import (
	"context"
	"errors"

	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ApplyUndoableOperationInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	OperationID string
}

func (a App) UndoOperation(ctx context.Context, input ApplyUndoableOperationInput) (asset.Asset, error) {
	return a.applyUndoableOperation(ctx, input, ports.UndoableOperationDirectionUndo)
}

func (a App) RedoOperation(ctx context.Context, input ApplyUndoableOperationInput) (asset.Asset, error) {
	return a.applyUndoableOperation(ctx, input, ports.UndoableOperationDirectionRedo)
}

func (a App) applyUndoableOperation(ctx context.Context, input ApplyUndoableOperationInput, direction ports.UndoableOperationDirection) (asset.Asset, error) {
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if input.OperationID == "" || a.undoables == nil {
		return asset.Asset{}, ErrInvalidInput
	}
	operation, found, err := a.undoables.UndoableOperationByID(ctx, input.TenantID, input.InventoryID, input.OperationID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ErrNotFound
	}
	expectedCurrent, resulting, err := assetUndoableOperationStates(operation, direction)
	if err != nil {
		return asset.Asset{}, err
	}
	if err := a.validateUndoableAssetResult(ctx, input.TenantID, input.InventoryID, resulting); err != nil {
		return asset.Asset{}, err
	}
	auditAction := audit.ActionUndoableOperationUndone
	eventName := ports.EventUndoableOperationUndone
	eventMessage := "undoable operation undone"
	if direction == ports.UndoableOperationDirectionRedo {
		auditAction = audit.ActionUndoableOperationRedone
		eventName = ports.EventUndoableOperationRedone
		eventMessage = "undoable operation redone"
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		PrincipalID: input.Principal.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      auditAction,
		TargetType:  audit.TargetUndoableOperation,
		TargetID:    operation.ID,
		Metadata: map[string]string{
			"operation_id":    operation.ID,
			"original_action": operation.OriginalAction.String(),
			"target_type":     operation.TargetType.String(),
			"target_id":       operation.TargetID,
		},
	})
	if err != nil {
		return asset.Asset{}, err
	}
	applied, item, err := a.undoables.ApplyAssetUndoableOperation(ctx, operation.ID, direction, expectedCurrent, resulting, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return asset.Asset{}, ErrInvalidInput
		}
		return asset.Asset{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       input.TenantID.String(),
			"inventory_id":    input.InventoryID.String(),
			"operation_id":    applied.ID,
			"original_action": applied.OriginalAction.String(),
			"target_type":     applied.TargetType.String(),
			"target_id":       applied.TargetID,
			"principal_id":    input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (a App) validateUndoableAssetResult(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, item asset.Asset) error {
	customAssetTypeID := item.CustomAssetTypeID
	if customAssetTypeID.String() != "" {
		err := a.ensureSnapshotCustomAssetTypeExists(ctx, tenantID, inventoryID, customAssetTypeID)
		if err != nil {
			return err
		}
	}
	if _, err := assetapp.ValidateCustomFields(ctx, a.customFields, tenantID, inventoryID, customAssetTypeID, item.CustomFields.Values()); err != nil {
		return err
	}
	return nil
}

func (a App) ensureSnapshotCustomAssetTypeExists(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, customAssetTypeID asset.CustomAssetTypeID) error {
	parsed, ok := customfield.NewAssetTypeID(customAssetTypeID.String())
	if !ok {
		return ErrInvalidInput
	}
	if a.customAssetTypes == nil {
		return ErrInvalidInput
	}
	_, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, tenantID, inventoryID, parsed)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	return nil
}

func assetUndoableOperationStates(operation ports.UndoableOperation, direction ports.UndoableOperationDirection) (asset.Asset, asset.Asset, error) {
	if operation.TargetType != audit.TargetAsset {
		return asset.Asset{}, asset.Asset{}, ErrInvalidInput
	}
	after := operation.AfterAsset
	switch direction {
	case ports.UndoableOperationDirectionUndo:
		switch operation.OriginalAction {
		case audit.ActionAssetCreated:
			resulting := after
			resulting.LifecycleState = asset.LifecycleStateArchived
			return after, resulting, nil
		case audit.ActionAssetUpdated, audit.ActionAssetMoved, audit.ActionAssetArchived, audit.ActionAssetRestored:
			if operation.BeforeAsset == nil {
				return asset.Asset{}, asset.Asset{}, ErrInvalidInput
			}
			return after, *operation.BeforeAsset, nil
		default:
			return asset.Asset{}, asset.Asset{}, ErrInvalidInput
		}
	case ports.UndoableOperationDirectionRedo:
		switch operation.OriginalAction {
		case audit.ActionAssetCreated:
			expected := after
			expected.LifecycleState = asset.LifecycleStateArchived
			return expected, after, nil
		case audit.ActionAssetUpdated, audit.ActionAssetMoved, audit.ActionAssetArchived, audit.ActionAssetRestored:
			if operation.BeforeAsset == nil {
				return asset.Asset{}, asset.Asset{}, ErrInvalidInput
			}
			return *operation.BeforeAsset, after, nil
		default:
			return asset.Asset{}, asset.Asset{}, ErrInvalidInput
		}
	default:
		return asset.Asset{}, asset.Asset{}, ErrInvalidInput
	}
}
