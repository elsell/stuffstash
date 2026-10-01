package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"strings"
)

func (s Store) ExecuteCreateCustomAssetTypeActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, item customfield.AssetType, record audit.Record) (ports.ActionPlanRecord, bool, error) {
	if item.Scope != customfield.ScopeInventory || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() || record.TenantID.String() != tenantID.String() || record.InventoryID.String() != inventoryID.String() {
		return ports.ActionPlanRecord{}, false, ports.ErrForbidden
	}
	return s.executeCustomizationActionPlan(ctx, tenantID, inventoryID, planID, transition, func(tx Store) error { return tx.SaveCustomAssetType(ctx, item, record) })
}
func (s Store) ExecuteCreateCustomFieldActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, item customfield.Definition, record audit.Record) (ports.ActionPlanRecord, bool, error) {
	if item.Scope != customfield.ScopeInventory || item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String() || record.TenantID.String() != tenantID.String() || record.InventoryID.String() != inventoryID.String() {
		return ports.ActionPlanRecord{}, false, ports.ErrForbidden
	}
	return s.executeCustomizationActionPlan(ctx, tenantID, inventoryID, planID, transition, func(tx Store) error { return tx.SaveCustomFieldDefinition(ctx, item, record) })
}
func (s Store) executeCustomizationActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, save func(Store) error) (ports.ActionPlanRecord, bool, error) {
	if tenantID == "" || inventoryID == "" || strings.TrimSpace(planID) == "" || validateActionPlanTransition(transition) != nil || transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	found := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		found, err = updateActionPlanStateInDB(tx, tenantID, inventoryID, planID, transition)
		if err != nil || !found {
			return err
		}
		return save(NewStore(tx))
	})
	if err != nil {
		return ports.ActionPlanRecord{}, found, err
	}
	if !found {
		return ports.ActionPlanRecord{}, false, nil
	}
	return s.ActionPlanByID(ctx, tenantID, inventoryID, planID)
}
