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

func (f *fakeActionPlanRepository) ExecuteCreateAndUpdateAssetsActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ports.ActionPlanStateTransition, creates []ports.ActionPlanCreateAssetOperation, updates []ports.ActionPlanUpdateAssetOperation) (ports.ActionPlanRecord, bool, error) {
	if transition.From != actionplan.StateApproved || transition.To != actionplan.StateExecuted {
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
	assetsSnapshot := map[asset.ID]asset.Asset{}
	for id, item := range assetRepository.items {
		assetsSnapshot[id] = item
	}
	undoablesSnapshot := map[string]ports.UndoableOperation{}
	for id, operation := range assetRepository.undoables {
		undoablesSnapshot[id] = operation
	}
	auditSnapshot := append([]audit.Record{}, assetRepository.auditRecords...)
	for _, create := range creates {
		operation := create.UndoableOperation
		var err error
		if create.PromotedParent != nil && create.ParentPromotionRecord != nil {
			err = f.assetUnitOfWork.CreateAssetWithParentPromotion(ctx, *create.PromotedParent, *create.ParentPromotionRecord, create.Item, create.AuditRecord, &operation)
		} else {
			err = f.assetUnitOfWork.CreateAsset(ctx, create.Item, create.AuditRecord, &operation)
		}
		if err != nil {
			assetRepository.items = assetsSnapshot
			assetRepository.undoables = undoablesSnapshot
			assetRepository.auditRecords = auditSnapshot
			return ports.ActionPlanRecord{}, true, err
		}
	}
	for _, update := range updates {
		current, found, err := assetRepository.AssetByID(ctx, tenantID, inventoryID, update.ExpectedCurrent.ID)
		if err != nil {
			assetRepository.items = assetsSnapshot
			assetRepository.undoables = undoablesSnapshot
			assetRepository.auditRecords = auditSnapshot
			return ports.ActionPlanRecord{}, true, err
		}
		if !found || !testAssetsEquivalentForStaleCheck(current, update.ExpectedCurrent) {
			assetRepository.items = assetsSnapshot
			assetRepository.undoables = undoablesSnapshot
			assetRepository.auditRecords = auditSnapshot
			return ports.ActionPlanRecord{}, true, ports.ErrConflict
		}
		if err := f.assetUnitOfWork.UpdateAsset(ctx, update.Item, update.AuditRecords, update.UndoableOperation); err != nil {
			assetRepository.items = assetsSnapshot
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
