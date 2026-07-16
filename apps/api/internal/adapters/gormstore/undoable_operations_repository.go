package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func createUndoableOperation(tx *gorm.DB, operation *ports.UndoableOperation) error {
	if operation == nil {
		return nil
	}
	model, err := newUndoableOperationModel(*operation)
	if err != nil {
		return err
	}
	return tx.Create(&model).Error
}

func (s Store) UndoableOperationByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, operationID string) (ports.UndoableOperation, bool, error) {
	var model undoableOperationModel
	err := s.db.WithContext(ctx).Where(&undoableOperationModel{
		ID:          operationID,
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.UndoableOperation{}, false, nil
	}
	if err != nil {
		return ports.UndoableOperation{}, false, err
	}
	operation, ok := model.toPort()
	if !ok {
		return ports.UndoableOperation{}, false, fmt.Errorf("invalid undoable operation row %q", model.ID)
	}
	return operation, true, nil
}

func (s Store) ApplyAssetUndoableOperation(ctx context.Context, operationID string, direction ports.UndoableOperationDirection, expectedCurrent asset.Asset, resulting asset.Asset, auditRecord audit.Record) (ports.UndoableOperation, asset.Asset, error) {
	var saved ports.UndoableOperation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var operationModel undoableOperationModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&undoableOperationModel{ID: operationID}).First(&operationModel).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		operation, ok := operationModel.toPort()
		if !ok {
			return fmt.Errorf("invalid undoable operation row %q", operationModel.ID)
		}
		if operation.TenantID.String() != expectedCurrent.TenantID.String() || operation.InventoryID.String() != expectedCurrent.InventoryID.String() || operation.TargetID != expectedCurrent.ID.String() {
			return ports.ErrForbidden
		}
		var currentModel assetModel
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&assetModel{
			ID:          expectedCurrent.ID.String(),
			TenantID:    expectedCurrent.TenantID.String(),
			InventoryID: expectedCurrent.InventoryID.String(),
		}).First(&currentModel).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrConflict
		}
		if err != nil {
			return err
		}
		current, ok := currentModel.toDomain()
		if !ok {
			return fmt.Errorf("invalid asset row %q", currentModel.ID)
		}
		if !gormAssetsEqual(current, expectedCurrent) {
			return ports.ErrConflict
		}
		if operation.ReplacesTags {
			expectedTagIDs := operation.AfterTagIDs
			if direction == ports.UndoableOperationDirectionRedo {
				expectedTagIDs = operation.BeforeTagIDs
			}
			matches, err := assetTagAssignmentsMatch(tx, operation.TenantID, operation.InventoryID, current.ID, expectedTagIDs)
			if err != nil {
				return err
			}
			if !matches {
				return ports.ErrConflict
			}
		}
		if !gormAssetsSameIdentity(current, resulting) {
			return ports.ErrForbidden
		}
		if err := validateUndoableAssetResult(tx, resulting); err != nil {
			return err
		}
		if err := updateAssetModelForUndoableOperation(tx, currentModel, resulting); err != nil {
			return err
		}
		if operation.ReplacesTags {
			tagIDs := operation.AfterTagIDs
			if direction == ports.UndoableOperationDirectionUndo {
				tagIDs = operation.BeforeTagIDs
			}
			if err := replaceAssetTagsInTx(tx, operation.TenantID, operation.InventoryID, resulting.ID, tagIDs); err != nil {
				return err
			}
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		now := time.Now().UTC()
		switch direction {
		case ports.UndoableOperationDirectionUndo:
			if operation.Status != ports.UndoableOperationAvailable && operation.Status != ports.UndoableOperationRedone {
				return ports.ErrConflict
			}
			operationModel.Status = string(ports.UndoableOperationUndone)
			operationModel.UndoAuditRecordID = stringPtr(auditRecord.ID.String())
		case ports.UndoableOperationDirectionRedo:
			if operation.Status != ports.UndoableOperationUndone {
				return ports.ErrConflict
			}
			operationModel.Status = string(ports.UndoableOperationRedone)
			operationModel.RedoAuditRecordID = stringPtr(auditRecord.ID.String())
		default:
			return ports.ErrConflict
		}
		operationModel.LastAppliedAt = &now
		if err := tx.Save(&operationModel).Error; err != nil {
			return err
		}
		saved, ok = operationModel.toPort()
		if !ok {
			return fmt.Errorf("invalid undoable operation row %q", operationModel.ID)
		}
		return nil
	})
	if err != nil {
		return ports.UndoableOperation{}, asset.Asset{}, err
	}
	return saved, resulting, nil
}

func assetTagAssignmentsMatch(tx *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, expected []assettag.ID) (bool, error) {
	var assignments []assetTagAssignmentModel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&assetTagAssignmentModel{
		TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: assetID.String(),
	}).Find(&assignments).Error; err != nil {
		return false, err
	}
	actual := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		actual[assignment.TagID] = struct{}{}
	}
	wanted := make(map[string]struct{}, len(expected))
	for _, tagID := range expected {
		wanted[tagID.String()] = struct{}{}
	}
	if len(actual) != len(wanted) {
		return false, nil
	}
	for tagID := range wanted {
		if _, ok := actual[tagID]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func (s Store) ApplyAssetCheckoutUndoableOperation(ctx context.Context, operationID string, direction ports.UndoableOperationDirection, expectedCurrent asset.Checkout, resulting asset.Checkout, auditRecord audit.Record) (ports.UndoableOperation, asset.Checkout, error) {
	var saved ports.UndoableOperation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var operationModel undoableOperationModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&undoableOperationModel{ID: operationID}).First(&operationModel).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		operation, ok := operationModel.toPort()
		if !ok {
			return fmt.Errorf("invalid undoable operation row %q", operationModel.ID)
		}
		if operation.TenantID.String() != expectedCurrent.TenantID.String() || operation.InventoryID.String() != expectedCurrent.InventoryID.String() || operation.TargetID != expectedCurrent.AssetID.String() {
			return ports.ErrForbidden
		}
		var currentModel assetCheckoutModel
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&assetCheckoutModel{
			ID:          expectedCurrent.ID.String(),
			TenantID:    expectedCurrent.TenantID.String(),
			InventoryID: expectedCurrent.InventoryID.String(),
		}).First(&currentModel).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrConflict
		}
		if err != nil {
			return err
		}
		current, ok := currentModel.toDomain()
		if !ok {
			return fmt.Errorf("invalid asset checkout row %q", currentModel.ID)
		}
		if !asset.CheckoutsEquivalentForStaleCheck(current, expectedCurrent) {
			return ports.ErrConflict
		}
		if current.ID != resulting.ID || current.TenantID != resulting.TenantID || current.InventoryID != resulting.InventoryID || current.AssetID != resulting.AssetID {
			return ports.ErrForbidden
		}
		if err := tx.Model(&currentModel).Updates(assetCheckoutUpdateMap(resulting)).Error; err != nil {
			return err
		}
		if err := createAuditRecord(tx, auditRecord); err != nil {
			return err
		}
		now := time.Now().UTC()
		switch direction {
		case ports.UndoableOperationDirectionUndo:
			if operation.Status != ports.UndoableOperationAvailable && operation.Status != ports.UndoableOperationRedone {
				return ports.ErrConflict
			}
			operationModel.Status = string(ports.UndoableOperationUndone)
			operationModel.UndoAuditRecordID = stringPtr(auditRecord.ID.String())
		case ports.UndoableOperationDirectionRedo:
			if operation.Status != ports.UndoableOperationUndone {
				return ports.ErrConflict
			}
			operationModel.Status = string(ports.UndoableOperationRedone)
			operationModel.RedoAuditRecordID = stringPtr(auditRecord.ID.String())
		default:
			return ports.ErrConflict
		}
		operationModel.LastAppliedAt = &now
		if err := tx.Save(&operationModel).Error; err != nil {
			return err
		}
		saved, ok = operationModel.toPort()
		if !ok {
			return fmt.Errorf("invalid undoable operation row %q", operationModel.ID)
		}
		return nil
	})
	if err != nil {
		return ports.UndoableOperation{}, asset.Checkout{}, err
	}
	return saved, resulting, nil
}

func validateUndoableAssetResult(tx *gorm.DB, item asset.Asset) error {
	if item.ParentAssetID.String() != "" {
		var parent assetModel
		err := tx.Where(&assetModel{
			ID:          item.ParentAssetID.String(),
			TenantID:    item.TenantID.String(),
			InventoryID: item.InventoryID.String(),
		}).First(&parent).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrForbidden
		}
		if err != nil {
			return err
		}
		parentKind, ok := asset.NewKind(parent.Kind)
		if !ok || !parentKind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive.String() || parent.ID == item.ID.String() {
			return ports.ErrForbidden
		}
		if err := rejectAssetContainmentCycle(tx, item.ID, parent); err != nil {
			return err
		}
	}
	if item.LifecycleState == asset.LifecycleStateArchived {
		hasActiveChildren, err := assetHasActiveChildren(tx, item.TenantID, item.InventoryID, item.ID)
		if err != nil {
			return err
		}
		if hasActiveChildren {
			return ports.ErrForbidden
		}
	}
	return nil
}

func updateAssetModelForUndoableOperation(tx *gorm.DB, model assetModel, item asset.Asset) error {
	customFields, err := json.Marshal(item.CustomFields.Values())
	if err != nil {
		return err
	}
	return tx.Model(&model).Updates(map[string]any{
		"parent_asset_id":      stringPtrFromAssetID(item.ParentAssetID),
		"custom_asset_type_id": stringPtrFromCustomAssetTypeID(item.CustomAssetTypeID),
		"title":                item.Title.String(),
		"description":          item.Description.String(),
		"custom_fields":        string(customFields),
		"lifecycle_state":      item.LifecycleState.String(),
	}).Error
}

func gormAssetsSameIdentity(left asset.Asset, right asset.Asset) bool {
	return left.ID == right.ID && left.TenantID == right.TenantID && left.InventoryID == right.InventoryID && left.Kind == right.Kind && left.CustomAssetTypeID == right.CustomAssetTypeID
}

func gormAssetsEqual(left asset.Asset, right asset.Asset) bool {
	return gormAssetsSameIdentity(left, right) &&
		left.ParentAssetID == right.ParentAssetID &&
		left.Title == right.Title &&
		left.Description == right.Description &&
		left.CustomFields.Equal(right.CustomFields) &&
		left.LifecycleState == right.LifecycleState
}
