package app

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeApprovedCreateActionPlanCommands(ctx context.Context, input ActionPlanDecisionInput, record ports.ActionPlanRecord) (ActionPlanExecutionResult, error) {
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
			args, err := parseActionPlanCreateArguments(command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			if args.ParentCommandID != "" {
				parent, ok := createdByCommand[args.ParentCommandID]
				if !ok || !parent.Kind.CanContainChildren() {
					return ActionPlanExecutionResult{}, ErrValidation
				}
				assetInput.ParentAssetID = parent.ID.String()
			}
			prepared, err := a.assetService.PrepareCreateAssetWithPendingParents(ctx, assetInput, pendingParentKinds)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			createdByCommand[command.ID] = prepared.Asset
			pendingParentKinds[prepared.Asset.ID] = prepared.Asset.Kind
			pendingParentIDs[prepared.Asset.ID] = prepared.Asset.ParentAssetID
			commandResults = append(commandResults, actionPlanCommandAssetResult(command, prepared.Asset, "create"))
			preparedCreates = append(preparedCreates, ports.ActionPlanCreateAssetOperation{
				Item:              prepared.Asset,
				AuditRecord:       prepared.AuditRecord,
				UndoableOperation: prepared.UndoableOperation,
			})
		case actionplan.CommandKindMoveAsset:
			moveInput, err := actionPlanMoveAssetInput(input, command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			args, err := parseActionPlanMoveArguments(command)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			if args.ParentCommandID != "" {
				parent, ok := createdByCommand[args.ParentCommandID]
				if !ok || !parent.Kind.CanContainChildren() {
					return ActionPlanExecutionResult{}, ErrValidation
				}
				if err := a.validatePendingMoveParentDoesNotCreateCycle(ctx, input, args.AssetID, parent.ID, pendingParentIDs); err != nil {
					return ActionPlanExecutionResult{}, err
				}
				moveInput.ParentAssetID = AssetParentUpdate{Present: true, Value: parent.ID.String()}
			}
			prepared, err := a.assetService.PrepareUpdateAssetWithPendingParents(ctx, moveInput, pendingParentKinds)
			if err != nil {
				return ActionPlanExecutionResult{}, err
			}
			preparedUpdates = append(preparedUpdates, ports.ActionPlanUpdateAssetOperation{
				ExpectedCurrent:   prepared.PreviousAsset,
				Item:              prepared.Asset,
				AuditRecords:      prepared.AuditRecords,
				UndoableOperation: prepared.UndoableOperation,
			})
		default:
			return ActionPlanExecutionResult{}, ErrValidation
		}
	}
	executed, found, err := a.actionPlans.ExecuteCreateAndUpdateAssetsActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          a.clock.Now(),
	}, preparedCreates, preparedUpdates)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ActionPlanExecutionResult{}, ErrConflict
		}
		return ActionPlanExecutionResult{}, err
	}
	if !found {
		return ActionPlanExecutionResult{}, ErrNotFound
	}
	for _, create := range preparedCreates {
		a.assetService.RecordAssetCreated(ctx, create.Item, input.Principal.ID)
	}
	for _, update := range preparedUpdates {
		a.assetService.RecordAssetUpdated(ctx, update.Item, input.Principal.ID)
	}
	return ActionPlanExecutionResult{Record: executed, CommandResults: commandResults}, nil
}

func (a App) validatePendingMoveParentDoesNotCreateCycle(ctx context.Context, input ActionPlanDecisionInput, movedAssetID asset.ID, parentAssetID asset.ID, pendingParentIDs map[asset.ID]asset.ID) error {
	for currentParentID := parentAssetID; currentParentID.String() != ""; {
		if currentParentID == movedAssetID {
			return ErrValidation
		}
		if pendingParentID, ok := pendingParentIDs[currentParentID]; ok {
			currentParentID = pendingParentID
			continue
		}
		parent, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, currentParentID)
		if err != nil {
			return err
		}
		if !found {
			return ErrValidation
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
			return ErrValidation
		}
		if _, exists := seenIDs[command.ID]; exists {
			return ErrValidation
		}
		seenIDs[command.ID] = struct{}{}
		if len(commands) > 1 && command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation && command.Kind != actionplan.CommandKindMoveAsset {
			return ErrValidation
		}
		if command.Kind == actionplan.CommandKindCreateAsset || command.Kind == actionplan.CommandKindCreateLocation {
			args, err := parseActionPlanCreateArguments(command)
			if err != nil {
				return err
			}
			if args.ParentCommandID != "" {
				parentKind, ok := seenCreateKinds[args.ParentCommandID]
				if !ok || !parentKind.CanContainChildren() {
					return ErrValidation
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
				return ErrValidation
			}
			seenCreateKinds[command.ID] = parsedKind
			continue
		}
		if command.Kind == actionplan.CommandKindMoveAsset {
			args, err := parseActionPlanMoveArguments(command)
			if err != nil {
				return err
			}
			if args.ParentCommandID != "" {
				parentKind, ok := seenCreateKinds[args.ParentCommandID]
				if !ok || !parentKind.CanContainChildren() {
					return ErrValidation
				}
			}
		}
	}
	return nil
}
