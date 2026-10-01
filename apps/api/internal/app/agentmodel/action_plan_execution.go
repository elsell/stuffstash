package agentmodel

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ActionPlanService) executeApprovedActionPlanCommands(ctx context.Context, input ActionPlanDecisionInput, record ports.ActionPlanRecord) (ActionPlanExecutionResult, error) {
	if len(record.Commands) == 0 {
		return ActionPlanExecutionResult{}, apperrors.ErrValidation
	}
	if len(record.Commands) > 1 {
		return a.executeApprovedCreateActionPlanCommands(ctx, input, record)
	}
	command := record.Commands[0]
	switch command.Kind {
	case actionplan.CommandKindCreateCustomAssetType, actionplan.CommandKindCreateCustomFieldDefinition:
		return a.executeApprovedCustomization(ctx, input, command)
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		assetInput, err := actionPlanCreateAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.deps.AssetPreparation.PrepareCreateAsset(ctx, assetInput)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.deps.ActionPlans.ExecuteCreateAssetsActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.deps.Clock.Now(),
		}, []ports.ActionPlanCreateAssetOperation{{
			Item:                  prepared.Asset,
			AuditRecord:           prepared.AuditRecord,
			PromotedParent:        prepared.PromotedParent,
			ParentPromotionRecord: prepared.ParentPromotionRecord,
			UndoableOperation:     prepared.UndoableOperation,
		}})
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, apperrors.ErrNotFound
		}
		a.deps.AssetPreparation.RecordAssetCreated(ctx, prepared.Asset, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandAssetResult(command, prepared.Asset, "create")},
		}, nil
	case actionplan.CommandKindMoveAsset, actionplan.CommandKindUpdateAsset:
		updateInput, operation, err := actionPlanAssetUpdateInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.deps.AssetPreparation.PrepareUpdateAsset(ctx, updateInput)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.deps.ActionPlans.ExecuteUpdateAssetActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.deps.Clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecords, prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, apperrors.ErrNotFound
		}
		a.deps.AssetPreparation.RecordAssetUpdated(ctx, prepared.Asset, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandAssetResult(command, prepared.Asset, operation)},
		}, nil
	case actionplan.CommandKindArchiveAsset:
		archiveInput, err := actionPlanLifecycleAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.deps.AssetPreparation.PrepareArchiveAsset(ctx, archiveInput)
		if err != nil {
			if errors.Is(err, apperrors.ErrValidation) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.deps.ActionPlans.ExecuteUpdateAssetLifecycleActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.deps.Clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, apperrors.ErrNotFound
		}
		a.deps.AssetPreparation.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandAssetResult(command, prepared.Asset, "archive")},
		}, nil
	case actionplan.CommandKindRestoreAsset:
		restoreInput, err := actionPlanLifecycleAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.deps.AssetPreparation.PrepareRestoreAsset(ctx, restoreInput)
		if err != nil {
			if errors.Is(err, apperrors.ErrValidation) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.deps.ActionPlans.ExecuteUpdateAssetLifecycleActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.deps.Clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, apperrors.ErrNotFound
		}
		a.deps.AssetPreparation.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandAssetResult(command, prepared.Asset, "restore")},
		}, nil
	case actionplan.CommandKindCheckoutAsset:
		checkoutInput, err := actionPlanCheckoutAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.deps.AssetPreparation.PrepareCheckoutAsset(ctx, checkoutInput)
		if err != nil {
			if errors.Is(err, apperrors.ErrValidation) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		operation := prepared.UndoableOperation
		executed, found, err := a.deps.ActionPlans.ExecuteAssetCheckoutActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.deps.Clock.Now(),
		}, ports.ActionPlanCheckoutOperation{
			Checkout:          prepared.Checkout,
			AuditRecord:       prepared.AuditRecord,
			UndoableOperation: &operation,
		})
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, apperrors.ErrNotFound
		}
		a.deps.AssetPreparation.RecordAssetCheckedOut(ctx, prepared.Checkout, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandCheckoutResult(command, prepared.Checkout, "checkout")},
		}, nil
	case actionplan.CommandKindReturnAsset:
		returnInput, err := actionPlanCheckoutAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.deps.AssetPreparation.PrepareReturnAsset(ctx, ports.ReturnAssetInput(returnInput))
		if err != nil {
			if errors.Is(err, apperrors.ErrValidation) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		operation := prepared.UndoableOperation
		executed, found, err := a.deps.ActionPlans.ExecuteAssetCheckoutActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.deps.Clock.Now(),
		}, ports.ActionPlanCheckoutOperation{
			ExpectedCurrent:   prepared.ExpectedCurrent,
			Checkout:          prepared.Checkout,
			AuditRecord:       prepared.AuditRecord,
			UndoableOperation: &operation,
		})
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, apperrors.ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, apperrors.ErrNotFound
		}
		a.deps.AssetPreparation.RecordAssetReturned(ctx, prepared.Checkout, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandCheckoutResult(command, prepared.Checkout, "return")},
		}, nil
	default:
		return ActionPlanExecutionResult{}, apperrors.ErrValidation
	}
}
