package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	maxActionPlanCommands        = 10
	maxActionPlanCommandIDLength = 80
	maxActionPlanSummaryLength   = 500
	maxActionPlanArgumentBytes   = 4096
	maxActionPlanRiskCount       = 10
	maxActionPlanRiskTextLength  = 300
)

type CreateActionPlanInput struct {
	Principal                  identity.Principal
	TenantID                   tenant.ID
	InventoryID                inventory.InventoryID
	Source                     string
	RealtimeSessionID          string
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
}

type ActionPlanCommandInput struct {
	ID        string
	Kind      actionplan.CommandKind
	Summary   string
	Arguments map[string]any
}

type ActionPlanDecisionInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	PlanID      string
}

type ActionPlanExecutionResult struct {
	Record         ports.ActionPlanRecord
	CommandResults []ActionPlanCommandExecutionResult
}

type ActionPlanCommandExecutionResult struct {
	CommandID string
	AssetID   string
	Operation string
	AssetKind string
}

func (a App) CreateActionPlan(ctx context.Context, input CreateActionPlanInput) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	planID := strings.TrimSpace(a.ids.NewID())
	commands, err := a.actionPlanCommands(input.Commands)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	risks, err := boundedActionPlanStrings(input.Risks, maxActionPlanRiskCount, maxActionPlanRiskTextLength)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	now := a.clock.Now()
	record := ports.ActionPlanRecord{
		ID:                         planID,
		TenantID:                   input.TenantID,
		InventoryID:                input.InventoryID,
		PrincipalID:                input.Principal.ID,
		Source:                     strings.TrimSpace(input.Source),
		RealtimeSessionID:          strings.TrimSpace(input.RealtimeSessionID),
		State:                      actionplan.StateProposed,
		IntentSummary:              strings.TrimSpace(input.IntentSummary),
		ModelInterpretationSummary: strings.TrimSpace(input.ModelInterpretationSummary),
		ConfirmationSummary:        strings.TrimSpace(input.ConfirmationSummary),
		Commands:                   commands,
		Risks:                      risks,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
	if err := validateActionPlanApplicationRecord(record); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.actionPlans.SaveActionPlan(ctx, record); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	return record, nil
}

func (a App) ApproveActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.transitionActionPlan(ctx, input, actionplan.StateApproved)
}

func (a App) CancelActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.transitionActionPlan(ctx, input, actionplan.StateCancelled)
}

func (a App) ExecuteActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	result, err := a.ExecuteActionPlanDetailed(ctx, input)
	return result.Record, err
}

func (a App) ExecuteActionPlanDetailed(ctx context.Context, input ActionPlanDecisionInput) (ActionPlanExecutionResult, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ActionPlanExecutionResult{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ActionPlanExecutionResult{}, err
	}
	record, found, err := a.actionPlans.ActionPlanByID(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID))
	if err != nil {
		return ActionPlanExecutionResult{}, err
	}
	if !found {
		return ActionPlanExecutionResult{}, ErrNotFound
	}
	if record.PrincipalID != input.Principal.ID || record.State != actionplan.StateApproved {
		return ActionPlanExecutionResult{}, ErrConflict
	}

	executed, err := a.executeApprovedActionPlanCommands(ctx, input, record)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return ActionPlanExecutionResult{}, err
		}
		failed, failErr := a.transitionActionPlan(ctx, input, actionplan.StateFailed)
		if failErr != nil {
			return ActionPlanExecutionResult{}, failErr
		}
		return ActionPlanExecutionResult{Record: failed}, err
	}
	return executed, nil
}

func (a App) transitionActionPlan(ctx context.Context, input ActionPlanDecisionInput, to actionplan.State) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	permission := ports.InventoryPermissionEditAsset
	if to == actionplan.StateCancelled {
		permission = ports.InventoryPermissionView
	} else if to == actionplan.StateExecuted || to == actionplan.StateFailed {
		permission = ports.InventoryPermissionEditAsset
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, permission); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	record, found, err := a.actionPlans.UpdateActionPlanState(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionPlanTransitionFromState(to),
		To:          to,
		At:          a.clock.Now(),
	})
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ports.ActionPlanRecord{}, ErrConflict
		}
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, ErrNotFound
	}
	return record, nil
}

func actionPlanTransitionFromState(to actionplan.State) actionplan.State {
	switch to {
	case actionplan.StateExecuted, actionplan.StateFailed:
		return actionplan.StateApproved
	default:
		return actionplan.StateProposed
	}
}

func (a App) executeApprovedActionPlanCommands(ctx context.Context, input ActionPlanDecisionInput, record ports.ActionPlanRecord) (ActionPlanExecutionResult, error) {
	if len(record.Commands) == 0 {
		return ActionPlanExecutionResult{}, ErrValidation
	}
	if len(record.Commands) > 1 {
		return a.executeApprovedCreateActionPlanCommands(ctx, input, record)
	}
	command := record.Commands[0]
	switch command.Kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		assetInput, err := actionPlanCreateAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.assetService.PrepareCreateAsset(ctx, assetInput)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.actionPlans.ExecuteCreateAssetsActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, []ports.ActionPlanCreateAssetOperation{{
			Item:                  prepared.Asset,
			AuditRecord:           prepared.AuditRecord,
			PromotedParent:        prepared.PromotedParent,
			ParentPromotionRecord: prepared.ParentPromotionRecord,
			UndoableOperation:     prepared.UndoableOperation,
		}})
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, ErrNotFound
		}
		a.assetService.RecordAssetCreated(ctx, prepared.Asset, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandAssetResult(command, prepared.Asset, "create")},
		}, nil
	case actionplan.CommandKindMoveAsset:
		moveInput, err := actionPlanMoveAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.assetService.PrepareUpdateAsset(ctx, moveInput)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.actionPlans.ExecuteUpdateAssetActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecords, prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, ErrNotFound
		}
		a.assetService.RecordAssetUpdated(ctx, prepared.Asset, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandAssetResult(command, prepared.Asset, "move")},
		}, nil
	case actionplan.CommandKindArchiveAsset:
		archiveInput, err := actionPlanLifecycleAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.assetService.PrepareArchiveAsset(ctx, archiveInput)
		if err != nil {
			if errors.Is(err, ErrValidation) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.actionPlans.ExecuteUpdateAssetLifecycleActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, ErrNotFound
		}
		a.assetService.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
		return ActionPlanExecutionResult{Record: executed}, nil
	case actionplan.CommandKindRestoreAsset:
		restoreInput, err := actionPlanLifecycleAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.assetService.PrepareRestoreAsset(ctx, restoreInput)
		if err != nil {
			if errors.Is(err, ErrValidation) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		executed, found, err := a.actionPlans.ExecuteUpdateAssetLifecycleActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, ErrNotFound
		}
		a.assetService.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
		return ActionPlanExecutionResult{Record: executed}, nil
	case actionplan.CommandKindCheckoutAsset:
		checkoutInput, err := actionPlanCheckoutAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.assetService.PrepareCheckoutAsset(ctx, checkoutInput)
		if err != nil {
			if errors.Is(err, ErrValidation) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		operation := prepared.UndoableOperation
		executed, found, err := a.actionPlans.ExecuteAssetCheckoutActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, ports.ActionPlanCheckoutOperation{
			Checkout:          prepared.Checkout,
			AuditRecord:       prepared.AuditRecord,
			UndoableOperation: &operation,
		})
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, ErrNotFound
		}
		a.assetService.RecordAssetCheckedOut(ctx, prepared.Checkout, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandCheckoutResult(command, prepared.Checkout, "checkout")},
		}, nil
	case actionplan.CommandKindReturnAsset:
		returnInput, err := actionPlanCheckoutAssetInput(input, command)
		if err != nil {
			return ActionPlanExecutionResult{}, err
		}
		prepared, err := a.assetService.PrepareReturnAsset(ctx, ReturnAssetInput(returnInput))
		if err != nil {
			if errors.Is(err, ErrValidation) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		operation := prepared.UndoableOperation
		executed, found, err := a.actionPlans.ExecuteAssetCheckoutActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, ports.ActionPlanCheckoutOperation{
			ExpectedCurrent:   prepared.ExpectedCurrent,
			Checkout:          prepared.Checkout,
			AuditRecord:       prepared.AuditRecord,
			UndoableOperation: &operation,
		})
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ActionPlanExecutionResult{}, ErrConflict
			}
			return ActionPlanExecutionResult{}, err
		}
		if !found {
			return ActionPlanExecutionResult{}, ErrNotFound
		}
		a.assetService.RecordAssetReturned(ctx, prepared.Checkout, input.Principal.ID)
		return ActionPlanExecutionResult{
			Record:         executed,
			CommandResults: []ActionPlanCommandExecutionResult{actionPlanCommandCheckoutResult(command, prepared.Checkout, "return")},
		}, nil
	default:
		return ActionPlanExecutionResult{}, ErrValidation
	}
}

func actionPlanCommandAssetResult(command ports.ActionPlanCommandRecord, item asset.Asset, operation string) ActionPlanCommandExecutionResult {
	return ActionPlanCommandExecutionResult{
		CommandID: command.ID,
		AssetID:   item.ID.String(),
		Operation: operation,
		AssetKind: item.Kind.String(),
	}
}

func actionPlanCommandCheckoutResult(command ports.ActionPlanCommandRecord, checkout asset.Checkout, operation string) ActionPlanCommandExecutionResult {
	return ActionPlanCommandExecutionResult{
		CommandID: command.ID,
		AssetID:   checkout.AssetID.String(),
		Operation: operation,
		AssetKind: asset.KindItem.String(),
	}
}

func actionPlanCreateAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (CreateAssetInput, error) {
	args, err := parseActionPlanCreateArguments(command)
	if err != nil {
		return CreateAssetInput{}, err
	}
	kind := args.Kind
	if command.Kind == actionplan.CommandKindCreateLocation {
		kind = "location"
	}
	if strings.TrimSpace(kind) == "" {
		kind = "item"
	}
	return CreateAssetInput{
		Principal:     input.Principal,
		Source:        audit.SourceConversation,
		RequestID:     command.ID,
		TenantID:      input.TenantID,
		InventoryID:   input.InventoryID,
		Kind:          kind,
		Title:         args.Title,
		Description:   args.Description,
		ParentAssetID: args.ParentAssetID,
		CustomFields:  map[string]any{},
	}, nil
}

type actionPlanCreateArguments struct {
	Title           string
	Kind            string
	Description     string
	ParentAssetID   string
	ParentCommandID string
}

func parseActionPlanCreateArguments(command ports.ActionPlanCommandRecord) (actionPlanCreateArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return actionPlanCreateArguments{}, ErrValidation
	}
	args := actionPlanCreateArguments{}
	for key, value := range raw {
		switch key {
		case "title", "name":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			if strings.TrimSpace(args.Title) == "" {
				args.Title = text
			}
		case "kind":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.Kind = text
		case "description":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.Description = text
		case "parentAssetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.ParentAssetID = text
		case "parentCommandId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.ParentCommandID = text
		default:
			return actionPlanCreateArguments{}, ErrValidation
		}
	}
	if strings.TrimSpace(args.Title) == "" {
		return actionPlanCreateArguments{}, ErrValidation
	}
	if strings.TrimSpace(args.ParentAssetID) != "" && strings.TrimSpace(args.ParentCommandID) != "" {
		return actionPlanCreateArguments{}, ErrValidation
	}
	switch strings.TrimSpace(args.Kind) {
	case "", "item", "container", "location":
		if command.Kind == actionplan.CommandKindCreateLocation && strings.TrimSpace(args.Kind) != "" && strings.TrimSpace(args.Kind) != "location" {
			return actionPlanCreateArguments{}, ErrValidation
		}
		return args, nil
	default:
		return actionPlanCreateArguments{}, ErrValidation
	}
}

func actionPlanStringArgument(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", ErrValidation
	}
	return strings.TrimSpace(value), nil
}

func actionPlanMoveAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (UpdateAssetInput, error) {
	args, err := parseActionPlanMoveArguments(command)
	if err != nil {
		return UpdateAssetInput{}, err
	}
	parent := AssetParentUpdate{Present: true, Value: args.ParentAssetID}
	if args.ParentIsRoot {
		parent.Null = true
		parent.Value = ""
	}
	return UpdateAssetInput{
		Principal:     input.Principal,
		Source:        audit.SourceConversation,
		RequestID:     command.ID,
		TenantID:      input.TenantID,
		InventoryID:   input.InventoryID,
		AssetID:       args.AssetID,
		ParentAssetID: parent,
	}, nil
}

type actionPlanMoveArguments struct {
	AssetID         asset.ID
	ParentAssetID   string
	ParentCommandID string
	ParentIsRoot    bool
}

func parseActionPlanMoveArguments(command ports.ActionPlanCommandRecord) (actionPlanMoveArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return actionPlanMoveArguments{}, ErrValidation
	}
	args := actionPlanMoveArguments{ParentIsRoot: true}
	for key, value := range raw {
		switch key {
		case "assetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanMoveArguments{}, err
			}
			assetID, ok := asset.NewID(text)
			if !ok {
				return actionPlanMoveArguments{}, ErrValidation
			}
			args.AssetID = assetID
		case "parentAssetId":
			if string(value) == "null" {
				args.ParentIsRoot = true
				args.ParentAssetID = ""
				continue
			}
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanMoveArguments{}, err
			}
			args.ParentAssetID = text
			args.ParentIsRoot = strings.TrimSpace(text) == ""
		case "parentCommandId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanMoveArguments{}, err
			}
			args.ParentCommandID = text
			args.ParentIsRoot = false
		default:
			return actionPlanMoveArguments{}, ErrValidation
		}
	}
	if args.AssetID.String() == "" {
		return actionPlanMoveArguments{}, ErrValidation
	}
	if strings.TrimSpace(args.ParentAssetID) != "" && strings.TrimSpace(args.ParentCommandID) != "" {
		return actionPlanMoveArguments{}, ErrValidation
	}
	if strings.TrimSpace(args.ParentCommandID) != "" && !validActionPlanCommandID(args.ParentCommandID) {
		return actionPlanMoveArguments{}, ErrValidation
	}
	if !args.ParentIsRoot {
		if strings.TrimSpace(args.ParentCommandID) == "" {
			if _, ok := asset.NewID(args.ParentAssetID); !ok {
				return actionPlanMoveArguments{}, ErrValidation
			}
		}
		if args.AssetID.String() == strings.TrimSpace(args.ParentAssetID) {
			return actionPlanMoveArguments{}, ErrValidation
		}
	}
	return args, nil
}

func actionPlanLifecycleAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (UpdateAssetLifecycleInput, error) {
	assetID, err := parseActionPlanAssetIDOnlyArguments(command)
	if err != nil {
		return UpdateAssetLifecycleInput{}, err
	}
	return UpdateAssetLifecycleInput{
		Principal:   input.Principal,
		Source:      audit.SourceConversation,
		RequestID:   command.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		AssetID:     assetID,
	}, nil
}

func parseActionPlanAssetIDOnlyArguments(command ports.ActionPlanCommandRecord) (asset.ID, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return "", ErrValidation
	}
	var parsed asset.ID
	for key, value := range raw {
		switch key {
		case "assetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return "", err
			}
			assetID, ok := asset.NewID(text)
			if !ok {
				return "", ErrValidation
			}
			parsed = assetID
		default:
			return "", ErrValidation
		}
	}
	if parsed.String() == "" {
		return "", ErrValidation
	}
	return parsed, nil
}

func actionPlanCheckoutAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (CheckoutAssetInput, error) {
	args, err := parseActionPlanCheckoutArguments(command)
	if err != nil {
		return CheckoutAssetInput{}, err
	}
	return CheckoutAssetInput{
		Principal:   input.Principal,
		Source:      audit.SourceConversation,
		RequestID:   command.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		AssetID:     args.AssetID,
		Details:     args.Details,
	}, nil
}

type actionPlanCheckoutArguments struct {
	AssetID asset.ID
	Details string
}

func parseActionPlanCheckoutArguments(command ports.ActionPlanCommandRecord) (actionPlanCheckoutArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return actionPlanCheckoutArguments{}, ErrValidation
	}
	args := actionPlanCheckoutArguments{}
	for key, value := range raw {
		switch key {
		case "assetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCheckoutArguments{}, err
			}
			assetID, ok := asset.NewID(text)
			if !ok {
				return actionPlanCheckoutArguments{}, ErrValidation
			}
			args.AssetID = assetID
		case "details", "checkoutDetails", "returnDetails":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCheckoutArguments{}, err
			}
			args.Details = text
		default:
			return actionPlanCheckoutArguments{}, ErrValidation
		}
	}
	if args.AssetID.String() == "" {
		return actionPlanCheckoutArguments{}, ErrValidation
	}
	return args, nil
}

func (a App) actionPlanCommands(inputs []ActionPlanCommandInput) ([]ports.ActionPlanCommandRecord, error) {
	if len(inputs) == 0 || len(inputs) > maxActionPlanCommands {
		return nil, ErrValidation
	}
	commands := make([]ports.ActionPlanCommandRecord, 0, len(inputs))
	for _, input := range inputs {
		if !input.Kind.Valid() {
			return nil, ErrValidation
		}
		summary := strings.TrimSpace(input.Summary)
		if summary == "" || len(summary) > maxActionPlanSummaryLength {
			return nil, ErrValidation
		}
		if err := validateSafeActionPlanArguments(input.Arguments); err != nil {
			return nil, err
		}
		arguments, err := json.Marshal(input.Arguments)
		if err != nil || len(arguments) > maxActionPlanArgumentBytes {
			return nil, ErrValidation
		}
		if string(arguments) == "null" {
			arguments = []byte("{}")
		}
		if err := validateExecutableActionPlanArguments(input.Kind, arguments); err != nil {
			return nil, err
		}
		commandID := strings.TrimSpace(input.ID)
		if commandID == "" {
			commandID = strings.TrimSpace(a.ids.NewID())
		}
		if !validActionPlanCommandID(commandID) {
			return nil, ErrValidation
		}
		commands = append(commands, ports.ActionPlanCommandRecord{
			ID:            commandID,
			Kind:          input.Kind,
			Summary:       summary,
			ArgumentsJSON: arguments,
		})
	}
	if err := validateActionPlanCommandDependencies(commands); err != nil {
		return nil, err
	}
	return commands, nil
}

func (a App) ensureActionPlanDependencies() error {
	if a.actionPlans == nil || a.tenants == nil || a.inventories == nil || a.authorizer == nil || a.ids == nil || a.clock == nil {
		return ErrInvalidInput
	}
	return nil
}
