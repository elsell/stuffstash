package memory

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) UndoableOperationByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, operationID string) (ports.UndoableOperation, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	operation, ok := s.undoables[operationID]
	if !ok || operation.TenantID != tenantID || operation.InventoryID != inventoryID {
		return ports.UndoableOperation{}, false, nil
	}
	return operation, true, nil
}

func (s *Store) ApplyAssetUndoableOperation(_ context.Context, operationID string, direction ports.UndoableOperationDirection, expectedCurrent asset.Asset, resulting asset.Asset, auditRecord audit.Record) (ports.UndoableOperation, asset.Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	operation, ok := s.undoables[operationID]
	if !ok {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
	}
	if operation.TenantID != tenant.ID(expectedCurrent.TenantID.String()) || operation.InventoryID != inventory.InventoryID(expectedCurrent.InventoryID.String()) || operation.TargetID != expectedCurrent.ID.String() {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
	}
	current, ok := s.assets[expectedCurrent.ID]
	if !ok || !memoryAssetsEqual(current, expectedCurrent) {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
	}
	if current.TenantID != resulting.TenantID || current.InventoryID != resulting.InventoryID || current.ID != resulting.ID {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
	}
	if resulting.LifecycleState == asset.LifecycleStateArchived {
		for _, child := range s.assets {
			if child.TenantID == resulting.TenantID && child.InventoryID == resulting.InventoryID && child.ParentAssetID == resulting.ID && child.LifecycleState == asset.LifecycleStateActive {
				return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
			}
		}
	}
	if resulting.LifecycleState == asset.LifecycleStateActive && resulting.ParentAssetID.String() != "" {
		parent, ok := s.assets[resulting.ParentAssetID]
		if !ok || parent.TenantID != resulting.TenantID || parent.InventoryID != resulting.InventoryID || parent.LifecycleState != asset.LifecycleStateActive || !parent.Kind.CanContainChildren() {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
		}
	}
	if resulting.ParentAssetID.String() != "" {
		parent, ok := s.assets[resulting.ParentAssetID]
		if !ok || parent.TenantID != resulting.TenantID || parent.InventoryID != resulting.InventoryID || !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
		}
		if parent.ID == resulting.ID {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
		}
		for currentParent := parent; currentParent.ParentAssetID.String() != ""; {
			next, ok := s.assets[currentParent.ParentAssetID]
			if !ok || next.TenantID != resulting.TenantID || next.InventoryID != resulting.InventoryID || next.ID == resulting.ID {
				return ports.UndoableOperation{}, asset.Asset{}, ports.ErrForbidden
			}
			currentParent = next
		}
	}

	switch direction {
	case ports.UndoableOperationDirectionUndo:
		if operation.Status != ports.UndoableOperationAvailable && operation.Status != ports.UndoableOperationRedone {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationUndone
		operation.UndoAuditRecordID = auditRecord.ID
	case ports.UndoableOperationDirectionRedo:
		if operation.Status != ports.UndoableOperationUndone {
			return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationRedone
		operation.RedoAuditRecordID = auditRecord.ID
	default:
		return ports.UndoableOperation{}, asset.Asset{}, ports.ErrConflict
	}
	operation.LastAppliedAt = time.Now().UTC()
	s.assets[resulting.ID] = resulting
	s.auditRecords[auditRecord.ID] = auditRecord
	s.undoables[operation.ID] = operation
	return operation, resulting, nil
}

func (s *Store) ApplyAssetCheckoutUndoableOperation(_ context.Context, operationID string, direction ports.UndoableOperationDirection, expectedCurrent asset.Checkout, resulting asset.Checkout, auditRecord audit.Record) (ports.UndoableOperation, asset.Checkout, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	operation, ok := s.undoables[operationID]
	if !ok {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrForbidden
	}
	if operation.TenantID != tenant.ID(expectedCurrent.TenantID.String()) || operation.InventoryID != inventory.InventoryID(expectedCurrent.InventoryID.String()) || operation.TargetID != expectedCurrent.AssetID.String() {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrForbidden
	}
	current, ok := s.checkouts[expectedCurrent.ID]
	if !ok || !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
	}
	if current.TenantID != resulting.TenantID || current.InventoryID != resulting.InventoryID || current.AssetID != resulting.AssetID || current.ID != resulting.ID {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrForbidden
	}
	if _, exists := s.auditRecords[auditRecord.ID]; exists {
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
	}
	for _, candidate := range s.checkouts {
		if candidate.ID == current.ID || candidate.TenantID != current.TenantID || candidate.InventoryID != current.InventoryID || candidate.AssetID != current.AssetID {
			continue
		}
		if candidate.State == asset.CheckoutStateOpen {
			return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
		}
	}

	switch direction {
	case ports.UndoableOperationDirectionUndo:
		if operation.Status != ports.UndoableOperationAvailable && operation.Status != ports.UndoableOperationRedone {
			return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationUndone
		operation.UndoAuditRecordID = auditRecord.ID
	case ports.UndoableOperationDirectionRedo:
		if operation.Status != ports.UndoableOperationUndone {
			return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
		}
		operation.Status = ports.UndoableOperationRedone
		operation.RedoAuditRecordID = auditRecord.ID
	default:
		return ports.UndoableOperation{}, asset.Checkout{}, ports.ErrConflict
	}
	operation.LastAppliedAt = time.Now().UTC()
	s.checkouts[resulting.ID] = resulting
	s.auditRecords[auditRecord.ID] = auditRecord
	s.undoables[operation.ID] = operation
	return operation, resulting, nil
}

func memoryAssetsEqual(left asset.Asset, right asset.Asset) bool {
	return left.ID == right.ID &&
		left.TenantID == right.TenantID &&
		left.InventoryID == right.InventoryID &&
		left.ParentAssetID == right.ParentAssetID &&
		left.CustomAssetTypeID == right.CustomAssetTypeID &&
		left.Kind == right.Kind &&
		left.Title == right.Title &&
		left.Description == right.Description &&
		left.CustomFields.Equal(right.CustomFields) &&
		left.LifecycleState == right.LifecycleState
}
