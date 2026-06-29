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
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	record, found, err := a.actionPlans.ActionPlanByID(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID))
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, ErrNotFound
	}
	if record.PrincipalID != input.Principal.ID || record.State != actionplan.StateApproved {
		return ports.ActionPlanRecord{}, ErrConflict
	}

	executed, err := a.executeApprovedActionPlanCommands(ctx, input, record)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) {
			return ports.ActionPlanRecord{}, err
		}
		failed, failErr := a.transitionActionPlan(ctx, input, actionplan.StateFailed)
		if failErr != nil {
			return ports.ActionPlanRecord{}, failErr
		}
		return failed, err
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

func (a App) executeApprovedActionPlanCommands(ctx context.Context, input ActionPlanDecisionInput, record ports.ActionPlanRecord) (ports.ActionPlanRecord, error) {
	if len(record.Commands) == 0 {
		return ports.ActionPlanRecord{}, ErrValidation
	}
	if len(record.Commands) > 1 {
		return a.executeApprovedCreateActionPlanCommands(ctx, input, record)
	}
	command := record.Commands[0]
	switch command.Kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		assetInput, err := actionPlanCreateAssetInput(input, command)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		prepared, err := a.assetService.PrepareCreateAsset(ctx, assetInput)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		executed, found, err := a.actionPlans.ExecuteCreateAssetActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ports.ActionPlanRecord{}, ErrConflict
			}
			return ports.ActionPlanRecord{}, err
		}
		if !found {
			return ports.ActionPlanRecord{}, ErrNotFound
		}
		a.assetService.RecordAssetCreated(ctx, prepared.Asset, input.Principal.ID)
		return executed, nil
	case actionplan.CommandKindMoveAsset:
		moveInput, err := actionPlanMoveAssetInput(input, command)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		prepared, err := a.assetService.PrepareUpdateAsset(ctx, moveInput)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		executed, found, err := a.actionPlans.ExecuteUpdateAssetActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecords, prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ports.ActionPlanRecord{}, ErrConflict
			}
			return ports.ActionPlanRecord{}, err
		}
		if !found {
			return ports.ActionPlanRecord{}, ErrNotFound
		}
		a.assetService.RecordAssetUpdated(ctx, prepared.Asset, input.Principal.ID)
		return executed, nil
	case actionplan.CommandKindArchiveAsset:
		archiveInput, err := actionPlanLifecycleAssetInput(input, command)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		prepared, err := a.assetService.PrepareArchiveAsset(ctx, archiveInput)
		if err != nil {
			if errors.Is(err, ErrValidation) {
				return ports.ActionPlanRecord{}, ErrConflict
			}
			return ports.ActionPlanRecord{}, err
		}
		executed, found, err := a.actionPlans.ExecuteUpdateAssetLifecycleActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ports.ActionPlanRecord{}, ErrConflict
			}
			return ports.ActionPlanRecord{}, err
		}
		if !found {
			return ports.ActionPlanRecord{}, ErrNotFound
		}
		a.assetService.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
		return executed, nil
	case actionplan.CommandKindRestoreAsset:
		restoreInput, err := actionPlanLifecycleAssetInput(input, command)
		if err != nil {
			return ports.ActionPlanRecord{}, err
		}
		prepared, err := a.assetService.PrepareRestoreAsset(ctx, restoreInput)
		if err != nil {
			if errors.Is(err, ErrValidation) {
				return ports.ActionPlanRecord{}, ErrConflict
			}
			return ports.ActionPlanRecord{}, err
		}
		executed, found, err := a.actionPlans.ExecuteUpdateAssetLifecycleActionPlan(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
			PrincipalID: input.Principal.ID,
			From:        actionplan.StateApproved,
			To:          actionplan.StateExecuted,
			At:          a.clock.Now(),
		}, prepared.PreviousAsset, prepared.Asset, prepared.AuditRecord, &prepared.UndoableOperation)
		if err != nil {
			if errors.Is(err, ports.ErrConflict) {
				return ports.ActionPlanRecord{}, ErrConflict
			}
			return ports.ActionPlanRecord{}, err
		}
		if !found {
			return ports.ActionPlanRecord{}, ErrNotFound
		}
		a.assetService.RecordAssetLifecycleUpdated(ctx, prepared, input.Principal.ID)
		return executed, nil
	default:
		return ports.ActionPlanRecord{}, ErrValidation
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
	AssetID       asset.ID
	ParentAssetID string
	ParentIsRoot  bool
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
		default:
			return actionPlanMoveArguments{}, ErrValidation
		}
	}
	if args.AssetID.String() == "" {
		return actionPlanMoveArguments{}, ErrValidation
	}
	if !args.ParentIsRoot {
		if _, ok := asset.NewID(args.ParentAssetID); !ok {
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

func validActionPlanCommandID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxActionPlanCommandIDLength {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func validateExecutableActionPlanArguments(kind actionplan.CommandKind, arguments []byte) error {
	command := ports.ActionPlanCommandRecord{Kind: kind, ArgumentsJSON: arguments}
	switch kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		_, err := parseActionPlanCreateArguments(command)
		return err
	case actionplan.CommandKindMoveAsset:
		_, err := parseActionPlanMoveArguments(command)
		return err
	case actionplan.CommandKindArchiveAsset, actionplan.CommandKindRestoreAsset:
		_, err := parseActionPlanAssetIDOnlyArguments(command)
		return err
	default:
		return ErrValidation
	}
}

func validateSafeActionPlanArguments(arguments any) error {
	if arguments == nil {
		return nil
	}
	switch value := arguments.(type) {
	case map[string]any:
		for key, nested := range value {
			if unsafeActionPlanArgumentKey(key) {
				return ErrValidation
			}
			if err := validateSafeActionPlanArguments(nested); err != nil {
				return err
			}
		}
	case []any:
		for _, nested := range value {
			if err := validateSafeActionPlanArguments(nested); err != nil {
				return err
			}
		}
	case string:
		if unsafeActionPlanArgumentString(value) {
			return ErrValidation
		}
	case bool, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return nil
	default:
		return ErrValidation
	}
	return nil
}

func unsafeActionPlanArgumentKey(key string) bool {
	normalized := normalizeActionPlanSafetyText(key)
	unsafeTokens := []string{
		"audio",
		"approval",
		"approved",
		"apikey",
		"bearer",
		"credential",
		"generatedspeech",
		"modelresponse",
		"password",
		"prompt",
		"providerid",
		"providerresponse",
		"providersessionid",
		"secret",
		"sessiontoken",
		"token",
		"transcript",
	}
	for _, token := range unsafeTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func unsafeActionPlanArgumentString(value string) bool {
	normalized := normalizeActionPlanSafetyText(value)
	unsafePhrases := []string{
		"apikey",
		"bearer",
		"beginprivatekey",
		"credential",
		"modelresponse",
		"providerresponse",
		"rawprompt",
		"systemprompt",
	}
	for _, phrase := range unsafePhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func normalizeActionPlanSafetyText(value string) string {
	replacer := strings.NewReplacer("_", "", "-", "", " ", "", ".", "", ":", "")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}

func boundedActionPlanStrings(values []string, maxCount int, maxLength int) ([]string, error) {
	if len(values) > maxCount {
		return nil, ErrValidation
	}
	bounded := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > maxLength {
			return nil, ErrValidation
		}
		bounded = append(bounded, trimmed)
	}
	return bounded, nil
}

func validateActionPlanApplicationRecord(record ports.ActionPlanRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		strings.TrimSpace(record.ConfirmationSummary) == "" ||
		len(record.ConfirmationSummary) > maxActionPlanSummaryLength ||
		record.State != actionplan.StateProposed ||
		record.CreatedAt.IsZero() ||
		record.UpdatedAt.IsZero() ||
		len(record.Commands) == 0 {
		return ErrValidation
	}
	if len(record.IntentSummary) > maxActionPlanSummaryLength || len(record.ModelInterpretationSummary) > maxActionPlanSummaryLength {
		return ErrValidation
	}
	return nil
}

func (a App) ensureActionPlanDependencies() error {
	if a.actionPlans == nil || a.tenants == nil || a.inventories == nil || a.authorizer == nil || a.ids == nil || a.clock == nil {
		return ErrInvalidInput
	}
	return nil
}
