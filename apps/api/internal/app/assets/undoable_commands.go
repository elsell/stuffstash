package assets

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
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

func (s Service) UndoOperation(ctx context.Context, input ApplyUndoableOperationInput) (asset.Asset, error) {
	return s.applyUndoableOperation(ctx, input, ports.UndoableOperationDirectionUndo)
}

func (s Service) RedoOperation(ctx context.Context, input ApplyUndoableOperationInput) (asset.Asset, error) {
	return s.applyUndoableOperation(ctx, input, ports.UndoableOperationDirectionRedo)
}

func (s Service) applyUndoableOperation(ctx context.Context, input ApplyUndoableOperationInput, direction ports.UndoableOperationDirection) (asset.Asset, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Asset{}, err
	}
	if input.OperationID == "" || s.undoables == nil {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	operation, found, err := s.undoables.UndoableOperationByID(ctx, input.TenantID, input.InventoryID, input.OperationID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, apperrors.ErrNotFound
	}
	auditAction := audit.ActionUndoableOperationUndone
	eventName := ports.EventUndoableOperationUndone
	eventMessage := "undoable operation undone"
	if direction == ports.UndoableOperationDirectionRedo {
		auditAction = audit.ActionUndoableOperationRedone
		eventName = ports.EventUndoableOperationRedone
		eventMessage = "undoable operation redone"
	}
	auditRecord, err := s.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      auditAction,
		TargetType:  operation.TargetType,
		TargetID:    operation.TargetID,
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
	if operation.AfterCheckout != nil {
		return s.applyCheckoutUndoableOperation(ctx, input, direction, operation, auditRecord, eventName, eventMessage)
	}
	expectedCurrent, resulting, err := assetUndoableOperationStates(operation, direction)
	if err != nil {
		return asset.Asset{}, err
	}
	if err := s.validateUndoableAssetResult(ctx, input.TenantID, input.InventoryID, resulting, resulting.CustomAssetTypeID != expectedCurrent.CustomAssetTypeID); err != nil {
		return asset.Asset{}, err
	}
	applied, item, err := s.undoables.ApplyAssetUndoableOperation(ctx, operation.ID, direction, expectedCurrent, resulting, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
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
			"operation_id":    applied.ID,
			"original_action": applied.OriginalAction.String(),
			"target_type":     applied.TargetType.String(),
			"target_id":       applied.TargetID,
			"principal_id":    input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (s Service) applyCheckoutUndoableOperation(ctx context.Context, input ApplyUndoableOperationInput, direction ports.UndoableOperationDirection, operation ports.UndoableOperation, auditRecord audit.Record, eventName ports.EventName, eventMessage string) (asset.Asset, error) {
	if s.checkouts == nil || s.assets == nil {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	expectedCurrent, resulting, err := checkoutUndoableOperationStates(ctx, s.checkouts, operation, direction)
	if err != nil {
		return asset.Asset{}, err
	}
	itemID, ok := asset.NewID(operation.TargetID)
	if !ok || itemID != resulting.AssetID {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	item, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, itemID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, apperrors.ErrNotFound
	}
	if resulting.State == asset.CheckoutStateOpen && (item.LifecycleState != asset.LifecycleStateActive || !item.Kind.IsPortable()) {
		return asset.Asset{}, apperrors.ErrInvalidInput
	}
	applied, _, err := s.undoables.ApplyAssetCheckoutUndoableOperation(ctx, operation.ID, direction, expectedCurrent, resulting, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
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
			"operation_id":    applied.ID,
			"original_action": applied.OriginalAction.String(),
			"target_type":     applied.TargetType.String(),
			"target_id":       applied.TargetID,
			"principal_id":    input.Principal.ID.String(),
		},
	})
	return item, nil
}
