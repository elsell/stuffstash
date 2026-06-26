package memory

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
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
