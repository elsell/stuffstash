package agentmodel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
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
	MaxActionPlanSummaryLength   = 500
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
	Principal    identity.Principal
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	PlanID       string
	CommandEdits []ActionPlanCommandEditInput
}

type ActionPlanCommandEditInput struct {
	CommandID       string
	Title           *string
	ParentSelection *ActionPlanParentSelectionInput
}

type ActionPlanParentSelectionInput struct {
	Kind string
	ID   string
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

func (a ActionPlanService) CreateActionPlan(ctx context.Context, input CreateActionPlanInput) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.deps.InventoryAccess.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	planID := strings.TrimSpace(a.deps.IDs.NewID())
	commands, err := a.ActionPlanCommands(input.Commands)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	for _, command := range commands {
		if IsCustomizationCommand(command.Kind) {
			if len(commands) != 1 {
				return ports.ActionPlanRecord{}, apperrors.ErrValidation
			}
			if err := a.deps.InventoryAccess.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
				return ports.ActionPlanRecord{}, err
			}
		}
	}
	risks, err := boundedActionPlanStrings(input.Risks, maxActionPlanRiskCount, maxActionPlanRiskTextLength)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	now := a.deps.Clock.Now()
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
	if err := a.deps.ActionPlans.SaveActionPlan(ctx, record); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	return record, nil
}

func (a ActionPlanService) ApproveActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	if len(input.CommandEdits) > 0 {
		return a.approveEditedActionPlan(ctx, input)
	}
	return a.transitionActionPlan(ctx, input, actionplan.StateApproved)
}

func (a ActionPlanService) CancelActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.transitionActionPlan(ctx, input, actionplan.StateCancelled)
}

func (a ActionPlanService) ExecuteActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	result, err := a.ExecuteActionPlanDetailed(ctx, input)
	return result.Record, err
}

func (a ActionPlanService) ExecuteActionPlanDetailed(ctx context.Context, input ActionPlanDecisionInput) (ActionPlanExecutionResult, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ActionPlanExecutionResult{}, err
	}
	if err := a.deps.InventoryAccess.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ActionPlanExecutionResult{}, err
	}
	record, found, err := a.deps.ActionPlans.ActionPlanByID(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID))
	if err != nil {
		return ActionPlanExecutionResult{}, err
	}
	if !found {
		return ActionPlanExecutionResult{}, apperrors.ErrNotFound
	}
	if record.PrincipalID != input.Principal.ID || record.State != actionplan.StateApproved {
		return ActionPlanExecutionResult{}, apperrors.ErrConflict
	}

	executed, err := a.executeApprovedActionPlanCommands(ctx, input, record)
	if err != nil {
		failed, failErr := a.transitionActionPlan(ctx, input, actionplan.StateFailed)
		if failErr != nil {
			return ActionPlanExecutionResult{}, failErr
		}
		return ActionPlanExecutionResult{Record: failed}, err
	}
	return executed, nil
}

func (a ActionPlanService) transitionActionPlan(ctx context.Context, input ActionPlanDecisionInput, to actionplan.State) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	permission := ports.InventoryPermissionEditAsset
	if to == actionplan.StateCancelled {
		permission = ports.InventoryPermissionView
	} else if to == actionplan.StateExecuted || to == actionplan.StateFailed {
		permission = ports.InventoryPermissionEditAsset
	}
	if err := a.deps.InventoryAccess.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, permission); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	record, found, err := a.deps.ActionPlans.UpdateActionPlanState(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionPlanTransitionFromState(to),
		To:          to,
		At:          a.deps.Clock.Now(),
	})
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ports.ActionPlanRecord{}, apperrors.ErrConflict
		}
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, apperrors.ErrNotFound
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

func actionPlanStringArgument(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", apperrors.ErrValidation
	}
	return strings.TrimSpace(value), nil
}

func actionPlanMoveAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ports.UpdateAssetInput, error) {
	args, err := ParseActionPlanMoveArguments(command)
	if err != nil {
		return ports.UpdateAssetInput{}, err
	}
	parent := ports.AssetParentUpdate{Present: true, Value: args.ParentAssetID}
	if args.ParentIsRoot {
		parent.Null = true
		parent.Value = ""
	}
	return ports.UpdateAssetInput{
		Principal:     input.Principal,
		Source:        audit.SourceConversation,
		RequestID:     command.ID,
		TenantID:      input.TenantID,
		InventoryID:   input.InventoryID,
		AssetID:       args.AssetID,
		ParentAssetID: parent,
	}, nil
}

type ActionPlanMoveArguments struct {
	AssetID         asset.ID
	ParentAssetID   string
	ParentCommandID string
	ParentIsRoot    bool
}

func ParseActionPlanMoveArguments(command ports.ActionPlanCommandRecord) (ActionPlanMoveArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return ActionPlanMoveArguments{}, apperrors.ErrValidation
	}
	args := ActionPlanMoveArguments{ParentIsRoot: true}
	for key, value := range raw {
		switch key {
		case "assetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanMoveArguments{}, err
			}
			assetID, ok := asset.NewID(text)
			if !ok {
				return ActionPlanMoveArguments{}, apperrors.ErrValidation
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
				return ActionPlanMoveArguments{}, err
			}
			args.ParentAssetID = text
			args.ParentIsRoot = strings.TrimSpace(text) == ""
		case "parentCommandId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanMoveArguments{}, err
			}
			args.ParentCommandID = text
			args.ParentIsRoot = false
		default:
			return ActionPlanMoveArguments{}, apperrors.ErrValidation
		}
	}
	if args.AssetID.String() == "" {
		return ActionPlanMoveArguments{}, apperrors.ErrValidation
	}
	if strings.TrimSpace(args.ParentAssetID) != "" && strings.TrimSpace(args.ParentCommandID) != "" {
		return ActionPlanMoveArguments{}, apperrors.ErrValidation
	}
	if strings.TrimSpace(args.ParentCommandID) != "" && !ValidActionPlanCommandID(args.ParentCommandID) {
		return ActionPlanMoveArguments{}, apperrors.ErrValidation
	}
	if !args.ParentIsRoot {
		if strings.TrimSpace(args.ParentCommandID) == "" {
			if _, ok := asset.NewID(args.ParentAssetID); !ok {
				return ActionPlanMoveArguments{}, apperrors.ErrValidation
			}
		}
		if args.AssetID.String() == strings.TrimSpace(args.ParentAssetID) {
			return ActionPlanMoveArguments{}, apperrors.ErrValidation
		}
	}
	return args, nil
}

func actionPlanLifecycleAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ports.UpdateAssetLifecycleInput, error) {
	assetID, err := ParseActionPlanAssetIDOnlyArguments(command)
	if err != nil {
		return ports.UpdateAssetLifecycleInput{}, err
	}
	return ports.UpdateAssetLifecycleInput{
		Principal:   input.Principal,
		Source:      audit.SourceConversation,
		RequestID:   command.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		AssetID:     assetID,
	}, nil
}

func ParseActionPlanAssetIDOnlyArguments(command ports.ActionPlanCommandRecord) (asset.ID, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return "", apperrors.ErrValidation
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
				return "", apperrors.ErrValidation
			}
			parsed = assetID
		default:
			return "", apperrors.ErrValidation
		}
	}
	if parsed.String() == "" {
		return "", apperrors.ErrValidation
	}
	return parsed, nil
}

func actionPlanCheckoutAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ports.CheckoutAssetInput, error) {
	args, err := ParseActionPlanCheckoutArguments(command)
	if err != nil {
		return ports.CheckoutAssetInput{}, err
	}
	return ports.CheckoutAssetInput{
		Principal:   input.Principal,
		Source:      audit.SourceConversation,
		RequestID:   command.ID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		AssetID:     args.AssetID,
		Details:     args.Details,
	}, nil
}

type ActionPlanCheckoutArguments struct {
	AssetID asset.ID
	Details string
}

func ParseActionPlanCheckoutArguments(command ports.ActionPlanCommandRecord) (ActionPlanCheckoutArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return ActionPlanCheckoutArguments{}, apperrors.ErrValidation
	}
	args := ActionPlanCheckoutArguments{}
	for key, value := range raw {
		switch key {
		case "assetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCheckoutArguments{}, err
			}
			assetID, ok := asset.NewID(text)
			if !ok {
				return ActionPlanCheckoutArguments{}, apperrors.ErrValidation
			}
			args.AssetID = assetID
		case "details", "checkoutDetails", "returnDetails":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCheckoutArguments{}, err
			}
			args.Details = text
		default:
			return ActionPlanCheckoutArguments{}, apperrors.ErrValidation
		}
	}
	if args.AssetID.String() == "" {
		return ActionPlanCheckoutArguments{}, apperrors.ErrValidation
	}
	return args, nil
}

func (a ActionPlanService) ActionPlanCommands(inputs []ActionPlanCommandInput) ([]ports.ActionPlanCommandRecord, error) {
	if len(inputs) == 0 || len(inputs) > maxActionPlanCommands {
		return nil, apperrors.ErrValidation
	}
	commands := make([]ports.ActionPlanCommandRecord, 0, len(inputs))
	for _, input := range inputs {
		if !input.Kind.Valid() {
			return nil, apperrors.ErrValidation
		}
		summary := strings.TrimSpace(input.Summary)
		if summary == "" || len(summary) > MaxActionPlanSummaryLength {
			return nil, apperrors.ErrValidation
		}
		arguments, err := json.Marshal(input.Arguments)
		if err != nil || len(arguments) > maxActionPlanArgumentBytes {
			return nil, apperrors.ErrValidation
		}
		if string(arguments) == "null" {
			arguments = []byte("{}")
		}
		if err := ValidateExecutableActionPlanArguments(input.Kind, arguments); err != nil {
			return nil, err
		}
		commandID := strings.TrimSpace(input.ID)
		if commandID == "" {
			commandID = strings.TrimSpace(a.deps.IDs.NewID())
		}
		if !ValidActionPlanCommandID(commandID) {
			return nil, apperrors.ErrValidation
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

func (a ActionPlanService) ensureActionPlanDependencies() error {
	if a.deps.ActionPlans == nil || a.deps.Tenants == nil || a.deps.Inventories == nil || a.deps.Authorizer == nil || a.deps.IDs == nil || a.deps.Clock == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}
