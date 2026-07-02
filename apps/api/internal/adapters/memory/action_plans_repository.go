package memory

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) SaveActionPlan(_ context.Context, record ports.ActionPlanRecord) error {
	if err := validateActionPlanRecord(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.actionPlans[record.ID] = cloneActionPlanRecord(record)
	return nil
}

func (s *Store) ActionPlanByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, found := s.actionPlans[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	return cloneActionPlanRecord(record), true, nil
}

func (s *Store) UpdateActionPlanState(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	record, found := s.actionPlans[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From || record.State.Terminal() || transition.At.Before(record.CreatedAt) {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	record.State = transition.To
	record.UpdatedAt = transition.At
	switch transition.To {
	case actionplan.StateApproved:
		record.ApprovedAt = transition.At
	case actionplan.StateCancelled:
		record.CancelledAt = transition.At
	case actionplan.StateExecuted:
		record.ExecutedAt = transition.At
	case actionplan.StateFailed:
		record.FailedAt = transition.At
	}
	s.actionPlans[planID] = record
	return cloneActionPlanRecord(record), true, nil
}

func (s *Store) ExecuteCreateAssetActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	record, found := s.actionPlans[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From || record.State.Terminal() || transition.At.Before(record.CreatedAt) {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	updated := record
	updated.State = transition.To
	updated.UpdatedAt = transition.At
	updated.ExecutedAt = transition.At
	if err := s.createAssetLocked(item, auditRecord, undoableOperation); err != nil {
		return ports.ActionPlanRecord{}, true, err
	}
	s.actionPlans[planID] = updated
	return cloneActionPlanRecord(updated), true, nil
}

func (s *Store) ExecuteCreateAssetsActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || len(creates) == 0 || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	return s.executeCreateAndUpdateAssetsActionPlan(tenantID, inventoryID, planID, transition, creates, nil)
}

func (s *Store) ExecuteCreateAndUpdateAssetsActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation, updates []ports.ActionPlanUpdateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || (len(creates) == 0 && len(updates) == 0) || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	return s.executeCreateAndUpdateAssetsActionPlan(tenantID, inventoryID, planID, transition, creates, updates)
}

func (s *Store) executeCreateAndUpdateAssetsActionPlan(tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation, updates []ports.ActionPlanUpdateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, found := s.actionPlans[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From || record.State.Terminal() || transition.At.Before(record.CreatedAt) {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	assetsSnapshot := cloneAssetMap(s.assets)
	auditSnapshot := cloneAuditMap(s.auditRecords)
	undoableSnapshot := cloneUndoableMap(s.undoables)
	for _, create := range creates {
		operation := create.UndoableOperation
		var err error
		if create.PromotedParent != nil && create.ParentPromotionRecord != nil {
			err = s.createAssetWithParentPromotionLocked(*create.PromotedParent, *create.ParentPromotionRecord, create.Item, create.AuditRecord, &operation)
		} else {
			err = s.createAssetLocked(create.Item, create.AuditRecord, &operation)
		}
		if err != nil {
			s.assets = assetsSnapshot
			s.auditRecords = auditSnapshot
			s.undoables = undoableSnapshot
			return ports.ActionPlanRecord{}, true, err
		}
	}
	for _, update := range updates {
		if err := s.updateAssetLocked(update.ExpectedCurrent, update.Item, update.AuditRecords, update.UndoableOperation); err != nil {
			s.assets = assetsSnapshot
			s.auditRecords = auditSnapshot
			s.undoables = undoableSnapshot
			return ports.ActionPlanRecord{}, true, err
		}
	}
	updated := record
	updated.State = transition.To
	updated.UpdatedAt = transition.At
	updated.ExecutedAt = transition.At
	s.actionPlans[planID] = updated
	return cloneActionPlanRecord(updated), true, nil
}

func (s *Store) ExecuteUpdateAssetActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, expectedCurrent asset.Asset, item asset.Asset, auditRecords []audit.Record, undoableOperation *ports.UndoableOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	record, found := s.actionPlans[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From || record.State.Terminal() || transition.At.Before(record.CreatedAt) {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	updated := record
	updated.State = transition.To
	updated.UpdatedAt = transition.At
	updated.ExecutedAt = transition.At
	if err := s.updateAssetLocked(expectedCurrent, item, auditRecords, undoableOperation); err != nil {
		return ports.ActionPlanRecord{}, true, err
	}
	s.actionPlans[planID] = updated
	return cloneActionPlanRecord(updated), true, nil
}

func (s *Store) ExecuteUpdateAssetLifecycleActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, expectedCurrent asset.Asset, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation) (ports.ActionPlanRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	record, found := s.actionPlans[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From || record.State.Terminal() || transition.At.Before(record.CreatedAt) {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	updated := record
	updated.State = transition.To
	updated.UpdatedAt = transition.At
	updated.ExecutedAt = transition.At
	if err := s.updateAssetLifecycleLocked(expectedCurrent, item, auditRecord, undoableOperation); err != nil {
		return ports.ActionPlanRecord{}, true, err
	}
	s.actionPlans[planID] = updated
	return cloneActionPlanRecord(updated), true, nil
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
	for _, command := range record.Commands {
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

func cloneActionPlanRecord(record ports.ActionPlanRecord) ports.ActionPlanRecord {
	record.Commands = append([]ports.ActionPlanCommandRecord{}, record.Commands...)
	for i := range record.Commands {
		record.Commands[i].ArgumentsJSON = append([]byte{}, record.Commands[i].ArgumentsJSON...)
	}
	record.Risks = append([]string{}, record.Risks...)
	return record
}

func cloneAssetMap(values map[asset.ID]asset.Asset) map[asset.ID]asset.Asset {
	cloned := make(map[asset.ID]asset.Asset, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneAuditMap(values map[audit.ID]audit.Record) map[audit.ID]audit.Record {
	cloned := make(map[audit.ID]audit.Record, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneUndoableMap(values map[string]ports.UndoableOperation) map[string]ports.UndoableOperation {
	cloned := make(map[string]ports.UndoableOperation, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
