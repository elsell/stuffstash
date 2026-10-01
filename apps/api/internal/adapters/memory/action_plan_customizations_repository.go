package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
)

func (s *Store) ExecuteCreateCustomAssetTypeActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, item customfield.AssetType, record audit.Record) (ports.ActionPlanRecord, bool, error) {
	if item.Scope != customfield.ScopeInventory || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() || record.TenantID.String() != tenantID.String() || record.InventoryID.String() != inventoryID.String() {
		return ports.ActionPlanRecord{}, false, ports.ErrForbidden
	}
	return s.executeCustomizationActionPlan(tenantID, inventoryID, planID, transition, func() error { return s.saveCustomAssetTypeLocked(item, record) })
}
func (s *Store) ExecuteCreateCustomFieldActionPlan(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, item customfield.Definition, record audit.Record) (ports.ActionPlanRecord, bool, error) {
	if item.Scope != customfield.ScopeInventory || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() || record.TenantID.String() != tenantID.String() || record.InventoryID.String() != inventoryID.String() {
		return ports.ActionPlanRecord{}, false, ports.ErrForbidden
	}
	return s.executeCustomizationActionPlan(tenantID, inventoryID, planID, transition, func() error { return s.saveCustomFieldDefinitionLocked(item, record) })
}
func (s *Store) executeCustomizationActionPlan(tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, save func() error) (ports.ActionPlanRecord, bool, error) {
	if tenantID == "" || inventoryID == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, found := s.actionPlans[planID]
	if !found || record.TenantID.String() != tenantID.String() || record.InventoryID.String() != inventoryID.String() {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From || transition.At.Before(record.CreatedAt) {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	if err := save(); err != nil {
		return ports.ActionPlanRecord{}, true, err
	}
	record.State = transition.To
	record.UpdatedAt = transition.At
	record.ExecutedAt = transition.At
	s.actionPlans[planID] = record
	return cloneActionPlanRecord(record), true, nil
}
