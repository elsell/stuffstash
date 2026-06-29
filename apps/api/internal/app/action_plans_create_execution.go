package app

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeApprovedCreateActionPlanCommands(ctx context.Context, input ActionPlanDecisionInput, record ports.ActionPlanRecord) (ports.ActionPlanRecord, error) {
	if err := validateActionPlanCommandDependencies(record.Commands); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	preparedCreates := make([]ports.ActionPlanCreateAssetOperation, 0, len(record.Commands))
	createdByCommand := map[string]asset.Asset{}
	pendingParentKinds := map[asset.ID]asset.Kind{}
	for _, command := range record.Commands {
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			return ports.ActionPlanRecord{}, ErrValidation
		}
		assetInput, err := actionPlanCreateAssetInput(input, command)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		args, err := parseActionPlanCreateArguments(command)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		if args.ParentCommandID != "" {
			parent, ok := createdByCommand[args.ParentCommandID]
			if !ok || !parent.Kind.CanContainChildren() {
				return ports.ActionPlanRecord{}, ErrValidation
			}
			assetInput.ParentAssetID = parent.ID.String()
		}
		prepared, err := a.assetService.PrepareCreateAssetWithPendingParents(ctx, assetInput, pendingParentKinds)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		createdByCommand[command.ID] = prepared.Asset
		pendingParentKinds[prepared.Asset.ID] = prepared.Asset.Kind
		preparedCreates = append(preparedCreates, ports.ActionPlanCreateAssetOperation{
			Item:              prepared.Asset,
			AuditRecord:       prepared.AuditRecord,
			UndoableOperation: prepared.UndoableOperation,
		})
	}
	executed, found, err := a.actionPlans.ExecuteCreateAssetsActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionplan.StateApproved,
		To:          actionplan.StateExecuted,
		At:          a.clock.Now(),
	}, preparedCreates)
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ports.ActionPlanRecord{}, ErrConflict
		}
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, ErrNotFound
	}
	for _, create := range preparedCreates {
		a.assetService.RecordAssetCreated(ctx, create.Item, input.Principal.ID)
	}
	return executed, nil
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
		if len(commands) > 1 && command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
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
		}
	}
	return nil
}
