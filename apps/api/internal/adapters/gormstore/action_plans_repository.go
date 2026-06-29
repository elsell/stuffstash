package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveActionPlan(ctx context.Context, record ports.ActionPlanRecord) error {
	if err := validateActionPlanRecord(record); err != nil {
		return err
	}
	model, err := actionPlanModelFromRecord(record)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(&model).Error
}

func (s Store) ActionPlanByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	var model actionPlanModel
	err := s.db.WithContext(ctx).Where(&actionPlanModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), ID: strings.TrimSpace(planID)}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ActionPlanRecord{}, false, nil
	}
	if err != nil {
		return ports.ActionPlanRecord{}, false, err
	}
	return actionPlanRecordFromModel(model)
}

func (s Store) UpdateActionPlanState(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	found, err := updateActionPlanStateInDB(s.db.WithContext(ctx), tenantID, inventoryID, planID, transition)
	if err != nil {
		return ports.ActionPlanRecord{}, found, err
	}
	if !found {
		return ports.ActionPlanRecord{}, false, nil
	}
	return s.ActionPlanByID(ctx, tenantID, inventoryID, planID)
}

func (s Store) ExecuteCreateAssetActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	found := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transitionFound, err := updateActionPlanStateInDB(tx, tenantID, inventoryID, planID, transition)
		if err != nil {
			return err
		}
		found = transitionFound
		if !found {
			return nil
		}
		return createAssetInTx(tx, item, auditRecord, undoableOperation)
	})
	if err != nil {
		return ports.ActionPlanRecord{}, found, err
	}
	if !found {
		return ports.ActionPlanRecord{}, false, nil
	}
	return s.ActionPlanByID(ctx, tenantID, inventoryID, planID)
}

func (s Store) ExecuteCreateAssetsActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || len(creates) == 0 || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	return s.ExecuteCreateAndUpdateAssetsActionPlan(ctx, tenantID, inventoryID, planID, transition, creates, nil)
}

func (s Store) ExecuteCreateAndUpdateAssetsActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation, updates []ports.ActionPlanUpdateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || (len(creates) == 0 && len(updates) == 0) || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	found := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transitionFound, err := updateActionPlanStateInDB(tx, tenantID, inventoryID, planID, transition)
		if err != nil {
			return err
		}
		found = transitionFound
		if !found {
			return nil
		}
		for _, create := range creates {
			if err := createAssetInTx(tx, create.Item, create.AuditRecord, actionPlanUndoableOperationPtr(create.UndoableOperation)); err != nil {
				return err
			}
		}
		for _, update := range updates {
			if err := updateAssetInTx(tx, update.ExpectedCurrent, update.Item, update.AuditRecords, update.UndoableOperation); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ports.ActionPlanRecord{}, found, err
	}
	if !found {
		return ports.ActionPlanRecord{}, false, nil
	}
	return s.ActionPlanByID(ctx, tenantID, inventoryID, planID)
}

func actionPlanUndoableOperationPtr(operation ports.UndoableOperation) *ports.UndoableOperation {
	if strings.TrimSpace(operation.ID) == "" {
		return nil
	}
	return &operation
}

func (s Store) ExecuteUpdateAssetActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, expectedCurrent asset.Asset, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	found := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transitionFound, err := updateActionPlanStateInDB(tx, tenantID, inventoryID, planID, transition)
		if err != nil {
			return err
		}
		found = transitionFound
		if !found {
			return nil
		}
		return updateAssetInTx(tx, expectedCurrent, item, auditRecords, undoableOperation)
	})
	if err != nil {
		return ports.ActionPlanRecord{}, found, err
	}
	if !found {
		return ports.ActionPlanRecord{}, false, nil
	}
	return s.ActionPlanByID(ctx, tenantID, inventoryID, planID)
}

func (s Store) ExecuteUpdateAssetLifecycleActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, expectedCurrent asset.Asset, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	found := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transitionFound, err := updateActionPlanStateInDB(tx, tenantID, inventoryID, planID, transition)
		if err != nil {
			return err
		}
		found = transitionFound
		if !found {
			return nil
		}
		return updateAssetLifecycleInTx(tx, expectedCurrent, item, auditRecord, undoableOperation)
	})
	if err != nil {
		return ports.ActionPlanRecord{}, found, err
	}
	if !found {
		return ports.ActionPlanRecord{}, false, nil
	}
	return s.ActionPlanByID(ctx, tenantID, inventoryID, planID)
}

func updateActionPlanStateInDB(db *gorm.DB, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition) (bool, error) {
	scope := actionPlanModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		ID:          strings.TrimSpace(planID),
		PrincipalID: transition.PrincipalID.String(),
		State:       string(transition.From),
	}
	result := db.Model(&actionPlanModel{}).
		Where(&scope).
		Where(clause.Lte{Column: clause.Column{Name: "created_at"}, Value: transition.At}).
		Updates(actionPlanStateUpdates(transition))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}
	var existing actionPlanModel
	err := db.Where(&actionPlanModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		ID:          strings.TrimSpace(planID),
	}).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, ports.ErrConflict
}

func actionPlanStateUpdates(transition ports.ActionPlanStateTransition) map[string]any {
	updates := map[string]any{
		"state":      string(transition.To),
		"updated_at": transition.At,
	}
	switch transition.To {
	case actionplan.StateApproved:
		updates["approved_at"] = transition.At
	case actionplan.StateCancelled:
		updates["cancelled_at"] = transition.At
	case actionplan.StateExecuted:
		updates["executed_at"] = transition.At
	case actionplan.StateFailed:
		updates["failed_at"] = transition.At
	}
	return updates
}

type persistedActionPlanCommand struct {
	ID            string                 `json:"id"`
	Kind          actionplan.CommandKind `json:"kind"`
	Summary       string                 `json:"summary"`
	ArgumentsJSON json.RawMessage        `json:"arguments"`
}

func actionPlanModelFromRecord(record ports.ActionPlanRecord) (actionPlanModel, error) {
	commands := make([]persistedActionPlanCommand, 0, len(record.Commands))
	for _, command := range record.Commands {
		commands = append(commands, persistedActionPlanCommand{
			ID:            strings.TrimSpace(command.ID),
			Kind:          command.Kind,
			Summary:       strings.TrimSpace(command.Summary),
			ArgumentsJSON: append([]byte{}, command.ArgumentsJSON...),
		})
	}
	commandsJSON, err := json.Marshal(commands)
	if err != nil {
		return actionPlanModel{}, ports.ErrInvalidProviderInput
	}
	risksJSON, err := json.Marshal(record.Risks)
	if err != nil {
		return actionPlanModel{}, ports.ErrInvalidProviderInput
	}
	return actionPlanModel{
		ID:                         record.ID,
		TenantID:                   record.TenantID.String(),
		InventoryID:                record.InventoryID.String(),
		PrincipalID:                record.PrincipalID.String(),
		Source:                     strings.TrimSpace(record.Source),
		RealtimeSessionID:          strings.TrimSpace(record.RealtimeSessionID),
		State:                      string(record.State),
		IntentSummary:              strings.TrimSpace(record.IntentSummary),
		ModelInterpretationSummary: strings.TrimSpace(record.ModelInterpretationSummary),
		ConfirmationSummary:        strings.TrimSpace(record.ConfirmationSummary),
		CommandsJSON:               commandsJSON,
		RisksJSON:                  risksJSON,
		CreatedAt:                  record.CreatedAt,
		UpdatedAt:                  record.UpdatedAt,
	}, nil
}

func actionPlanRecordFromModel(model actionPlanModel) (ports.ActionPlanRecord, bool, error) {
	var commands []persistedActionPlanCommand
	if err := json.Unmarshal(model.CommandsJSON, &commands); err != nil {
		return ports.ActionPlanRecord{}, false, fmt.Errorf("invalid action plan commands row %q: %w", model.ID, err)
	}
	var risks []string
	if err := json.Unmarshal(model.RisksJSON, &risks); err != nil {
		return ports.ActionPlanRecord{}, false, fmt.Errorf("invalid action plan risks row %q: %w", model.ID, err)
	}
	record := ports.ActionPlanRecord{
		ID:                         model.ID,
		TenantID:                   tenant.ID(model.TenantID),
		InventoryID:                inventory.InventoryID(model.InventoryID),
		PrincipalID:                identity.PrincipalID(model.PrincipalID),
		Source:                     model.Source,
		RealtimeSessionID:          model.RealtimeSessionID,
		State:                      actionplan.State(model.State),
		IntentSummary:              model.IntentSummary,
		ModelInterpretationSummary: model.ModelInterpretationSummary,
		ConfirmationSummary:        model.ConfirmationSummary,
		Commands:                   make([]ports.ActionPlanCommandRecord, 0, len(commands)),
		Risks:                      risks,
		CreatedAt:                  model.CreatedAt,
		UpdatedAt:                  model.UpdatedAt,
	}
	for _, command := range commands {
		record.Commands = append(record.Commands, ports.ActionPlanCommandRecord{
			ID:            command.ID,
			Kind:          command.Kind,
			Summary:       command.Summary,
			ArgumentsJSON: append([]byte{}, command.ArgumentsJSON...),
		})
	}
	record.ApprovedAt = timeFromPointer(model.ApprovedAt)
	record.CancelledAt = timeFromPointer(model.CancelledAt)
	record.ExecutedAt = timeFromPointer(model.ExecutedAt)
	record.FailedAt = timeFromPointer(model.FailedAt)
	if err := validateActionPlanReadRecord(record); err != nil {
		return ports.ActionPlanRecord{}, false, fmt.Errorf("invalid action plan row %q: %w", model.ID, err)
	}
	return record, true, nil
}

func timeFromPointer(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func validateActionPlanRecord(record ports.ActionPlanRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		record.State != actionplan.StateProposed ||
		strings.TrimSpace(record.ConfirmationSummary) == "" ||
		record.CreatedAt.IsZero() ||
		record.UpdatedAt.IsZero() ||
		!record.ApprovedAt.IsZero() ||
		!record.CancelledAt.IsZero() ||
		!record.ExecutedAt.IsZero() ||
		!record.FailedAt.IsZero() ||
		len(record.Commands) == 0 {
		return ports.ErrInvalidProviderInput
	}
	return validateActionPlanCommands(record.Commands)
}

func validateActionPlanReadRecord(record ports.ActionPlanRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		!record.State.Valid() ||
		strings.TrimSpace(record.ConfirmationSummary) == "" ||
		record.CreatedAt.IsZero() ||
		record.UpdatedAt.IsZero() ||
		len(record.Commands) == 0 {
		return ports.ErrInvalidProviderInput
	}
	if err := validateActionPlanCommands(record.Commands); err != nil {
		return err
	}
	if record.State == actionplan.StateApproved && record.ApprovedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if record.State == actionplan.StateCancelled && record.CancelledAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if record.State == actionplan.StateExecuted && record.ExecutedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if record.State == actionplan.StateFailed && record.FailedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateActionPlanCommands(commands []ports.ActionPlanCommandRecord) error {
	for _, command := range commands {
		if strings.TrimSpace(command.ID) == "" ||
			!command.Kind.Valid() ||
			strings.TrimSpace(command.Summary) == "" ||
			!json.Valid(command.ArgumentsJSON) {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func validateActionPlanTransition(transition ports.ActionPlanStateTransition) error {
	if transition.PrincipalID.String() == "" || !validActionPlanTransition(transition.From, transition.To) || transition.At.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validActionPlanTransition(from actionplan.State, to actionplan.State) bool {
	switch from {
	case actionplan.StateProposed:
		return to == actionplan.StateApproved || to == actionplan.StateCancelled
	case actionplan.StateApproved:
		return to == actionplan.StateExecuted || to == actionplan.StateFailed
	default:
		return false
	}
}
