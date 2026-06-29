package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (f *fakeActionPlanRepository) ExecuteCreateAssetsActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	if transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted || len(creates) == 0 {
		return ports.ActionPlanRecord{}, false, ports.ErrInvalidProviderInput
	}
	record, found := f.records[planID]
	if !found || record.TenantID != tenantID || record.InventoryID != inventoryID {
		return ports.ActionPlanRecord{}, false, nil
	}
	if record.PrincipalID != transition.PrincipalID || record.State != transition.From {
		return ports.ActionPlanRecord{}, true, ports.ErrConflict
	}
	if f.assetUnitOfWork == nil {
		return ports.ActionPlanRecord{}, true, ErrInvalidInput
	}
	assetRepository, ok := f.assetUnitOfWork.(*fakeAssetRepository)
	if !ok {
		return ports.ActionPlanRecord{}, true, ErrInvalidInput
	}
	itemsSnapshot := map[asset.ID]asset.Asset{}
	for key, value := range assetRepository.items {
		itemsSnapshot[key] = value
	}
	undoablesSnapshot := map[string]ports.UndoableOperation{}
	for key, value := range assetRepository.undoables {
		undoablesSnapshot[key] = value
	}
	auditSnapshot := append([]audit.Record{}, assetRepository.auditRecords...)
	for _, create := range creates {
		operation := create.UndoableOperation
		if err := f.assetUnitOfWork.CreateAsset(ctx, create.Item, create.AuditRecord, &operation); err != nil {
			assetRepository.items = itemsSnapshot
			assetRepository.undoables = undoablesSnapshot
			assetRepository.auditRecords = auditSnapshot
			return ports.ActionPlanRecord{}, true, err
		}
	}
	updated := record
	updated.State = transition.To
	updated.UpdatedAt = transition.At
	updated.ExecutedAt = transition.At
	f.records[planID] = updated
	return updated, true, nil
}
