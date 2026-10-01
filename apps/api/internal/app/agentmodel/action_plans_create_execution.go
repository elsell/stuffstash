package agentmodel

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ActionPlanService) executeApprovedCreateActionPlanCommands(ctx context.Context, input ActionPlanDecisionInput, record ports.ActionPlanRecord) (ActionPlanExecutionResult, error) {
	if err := validateActionPlanCommandDependencies(record.Commands); err != nil {
		return ActionPlanExecutionResult{}, err
	}
	preparedCreates := make([]ports.ActionPlanCreateAssetOperation, 0, len(record.Commands))
	preparedUpdates := make([]ports.ActionPlanUpdateAssetOperation, 0, len(record.Commands))
	commandResults := make([]ActionPlanCommandExecutionResult, 0, len(record.Commands))
	createdByCommand := map[string]asset.Asset{}
	pendingParentKinds := map[asset.ID]asset.Kind{}
	pendingParentIDs := map[asset.ID]asset.ID{}
	for _, command := range record.Commands {
		switch command.Kind {
		case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
			assetInput, err := actionPlanCreateAssetInput(input, command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			args, err := ParseActionPlanCreateArguments(command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			if args.ParentCommandID != "" {
				parent, ok := createdByCommand[args.ParentCommandID]
				if !ok || !parent.Kind.CanContainChildren() {
					return ActionPlanExecutionResult{}, apperrors.ErrValidation
				}
				assetInput.ParentAssetID = parent.ID.String()
			}
			prepared, err := a.deps.AssetPreparation.PrepareCreateAssetWithPendingParents(ctx, assetInput, pendingParentKinds)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			if prepared.PromotedParent != nil {
				pendingParentKinds[prepared.PromotedParent.ID] = prepared.PromotedParent.Kind
			}
			createdByCommand[command.ID] = prepared.Asset
			pendingParentKinds[prepared.Asset.ID] = prepared.Asset.Kind
			pendingParentIDs[prepared.Asset.ID] = prepared.Asset.ParentAssetID
			commandResults = append(commandResults, actionPlanCommandAssetResult(command, prepared.Asset, "create"))
			preparedCreates = append(preparedCreates, ports.ActionPlanCreateAssetOperation{
				Item:                  prepared.Asset,
				AuditRecord:           prepared.AuditRecord,
				PromotedParent:        prepared.PromotedParent,
				ParentPromotionRecord: prepared.ParentPromotionRecord,
				UndoableOperation:     prepared.UndoableOperation,
			})
		case actionplan.CommandKindMoveAsset:
			moveInput, err := actionPlanMoveAssetInput(input, command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			args, err := ParseActionPlanMoveArguments(command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			if args.ParentCommandID != "" {
				parent, ok := createdByCommand[args.ParentCommandID]
				if !ok || !parent.Kind.CanContainChildren() {
					return ActionPlanExecutionResult{}, apperrors.ErrValidation
				}
				if err := a.validatePendingMoveParentDoesNotCreateCycle(ctx, input, args.AssetID, parent.ID, pendingParentIDs); err != nil {
					return ActionPlanExecutionResult{}, err
				}
				moveInput.ParentAssetID = ports.AssetParentUpdate{Present: true, Value: parent.ID.String()}
			}
			prepared, err := a.deps.AssetPreparation.PrepareUpdateAssetWithPendingParents(ctx, moveInput, pendingParentKinds)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			commandResults = append(commandResults, actionPlanCommandAssetResult(command, prepared.Asset, "move"))
			preparedUpdates = append(preparedUpdates, ports.ActionPlanUpdateAssetOperation{
				ExpectedCurrent:   prepared.PreviousAsset,
				Item:              prepared.Asset,
				AuditRecords:      prepared.AuditRecords,
				UndoableOperation: prepared.UndoableOperation,
			})
		default:
			return ActionPlanExecutionResult{}, apperrors.ErrValidation
		}
	}
	executed, found, err := a.deps.ActionPlans.ExecuteCreateAndUpdateAssetsActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          a.deps.Clock.Now(),
	}, preparedCreates, preparedUpdates)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ActionPlanExecutionResult{}, apperrors.ErrConflict
		}
		return ActionPlanExecutionResult{}, err
	}
	if !found {
		return ActionPlanExecutionResult{}, apperrors.ErrNotFound
	}
	for _, create := range preparedCreates {
		a.deps.AssetPreparation.RecordAssetCreated(ctx, create.Item, input.Principal.ID)
	}
	for _, update := range preparedUpdates {
		a.deps.AssetPreparation.RecordAssetUpdated(ctx, update.Item, input.Principal.ID)
	}
	return ActionPlanExecutionResult{Record: executed, CommandResults: commandResults}, nil
}

func (a ActionPlanService) validatePendingMoveParentDoesNotCreateCycle(ctx context.Context, input ActionPlanDecisionInput, movedAssetID asset.ID, parentAssetID asset.ID, pendingParentIDs map[asset.ID]asset.ID) error {
	for currentParentID := parentAssetID; currentParentID.String() != ""; {
		if currentParentID == movedAssetID {
			return apperrors.ErrValidation
		}
		if pendingParentID, ok := pendingParentIDs[currentParentID]; ok {
			currentParentID = pendingParentID
			continue
		}
		parent, found, err := a.deps.Assets.AssetByID(ctx, input.TenantID, input.InventoryID, currentParentID)
		if err != nil {
			return err
		}
		if !found {
			return apperrors.ErrValidation
		}
		currentParentID = parent.ParentAssetID
	}
	return nil
}

func validateActionPlanCommandDependencies(commands []ports.ActionPlanCommandRecord) error {
	seenCreateKinds := map[string]asset.Kind{}
	seenIDs := map[string]struct{}{}
	for _, command := range commands {
		if strings.TrimSpace(command.ID) == "" {
			return apperrors.ErrValidation
		}
		if _, exists := seenIDs[command.ID]; exists {
			return apperrors.ErrValidation
		}
		seenIDs[command.ID] = struct{}{}
		if len(commands) > 1 && command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation && command.Kind != actionplan.CommandKindMoveAsset {
			return apperrors.ErrValidation
		}
		if command.Kind == actionplan.CommandKindCreateAsset || command.Kind == actionplan.CommandKindCreateLocation {
			args, err := ParseActionPlanCreateArguments(command)
			if err != nil {
				return err
			}
			if args.ParentCommandID != "" {
				parentKind, ok := seenCreateKinds[args.ParentCommandID]
				if !ok || !parentKind.CanContainChildren() {
					return apperrors.ErrValidation
				}
			}
			kind := strings.TrimSpace(args.Kind)
			if command.Kind == actionplan.CommandKindCreateLocation {
				kind = asset.KindLocation.String()
			}
			if kind == "" {
				kind = asset.KindItem.String()
			}
			parsedKind, ok := asset.NewKind(kind)
			if !ok {
				return apperrors.ErrValidation
			}
			seenCreateKinds[command.ID] = parsedKind
			continue
		}
		if command.Kind == actionplan.CommandKindMoveAsset {
			args, err := ParseActionPlanMoveArguments(command)
			if err != nil {
				return err
			}
			if args.ParentCommandID != "" {
				parentKind, ok := seenCreateKinds[args.ParentCommandID]
				if !ok || !parentKind.CanContainChildren() {
					return apperrors.ErrValidation
				}
			}
		}
	}
	return nil
}
