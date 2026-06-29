package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func parseRealtimeVoiceActionPlanArgs(args map[string]any) (realtimeVoiceActionPlanArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "commandKind", "intentSummary", "modelInterpretationSummary", "confirmationSummary", "commandSummary", "arguments", "argumentsJson", "commands", "risks", "riskSummary"); err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	commands, err := realtimeVoiceActionPlanCommands(args)
	if err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	risks, err := realtimeVoiceActionPlanRisks(args)
	if err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	parsed := realtimeVoiceActionPlanArgs{
		IntentSummary:              strings.TrimSpace(stringArg(args["intentSummary"])),
		ModelInterpretationSummary: strings.TrimSpace(stringArg(args["modelInterpretationSummary"])),
		ConfirmationSummary:        strings.TrimSpace(stringArg(args["confirmationSummary"])),
		Commands:                   commands,
		Risks:                      risks,
	}
	if parsed.IntentSummary == "" || parsed.ModelInterpretationSummary == "" || parsed.ConfirmationSummary == "" || len(parsed.Commands) == 0 {
		return realtimeVoiceActionPlanArgs{}, ports.ErrInvalidProviderInput
	}
	return parsed, nil
}

func realtimeVoiceActionPlanCommands(args map[string]any) ([]ActionPlanCommandInput, error) {
	if rawCommands, exists := args["commands"]; exists {
		values, ok := rawCommands.([]any)
		if !ok || len(values) == 0 || len(values) > maxActionPlanCommands {
			return nil, ports.ErrInvalidProviderInput
		}
		commands := make([]ActionPlanCommandInput, 0, len(values))
		for _, value := range values {
			command, ok := value.(map[string]any)
			if !ok {
				return nil, ports.ErrInvalidProviderInput
			}
			if err := rejectUnknownRealtimeVoiceArgs(command, "id", "commandId", "kind", "commandKind", "summary", "commandSummary", "arguments", "argumentsJson"); err != nil {
				return nil, err
			}
			kind := actionplan.CommandKind(strings.TrimSpace(firstStringArg(command["kind"], command["commandKind"])))
			if !kind.Valid() {
				return nil, ports.ErrInvalidProviderInput
			}
			summary := strings.TrimSpace(firstStringArg(command["summary"], command["commandSummary"]))
			if summary == "" {
				return nil, ports.ErrInvalidProviderInput
			}
			arguments, err := realtimeVoiceActionPlanArguments(command)
			if err != nil {
				return nil, err
			}
			commands = append(commands, ActionPlanCommandInput{
				ID:        strings.TrimSpace(firstStringArg(command["id"], command["commandId"])),
				Kind:      kind,
				Summary:   summary,
				Arguments: arguments,
			})
		}
		return commands, nil
	}

	commandKind := actionplan.CommandKind(strings.TrimSpace(stringArg(args["commandKind"])))
	if !commandKind.Valid() {
		return nil, ports.ErrInvalidProviderInput
	}
	arguments, err := realtimeVoiceActionPlanArguments(args)
	if err != nil {
		return nil, err
	}
	summary := strings.TrimSpace(stringArg(args["commandSummary"]))
	if summary == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	return []ActionPlanCommandInput{{
		Kind:      commandKind,
		Summary:   summary,
		Arguments: arguments,
	}}, nil
}

func firstStringArg(values ...any) string {
	for _, value := range values {
		if text := stringArg(value); strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}

func realtimeVoiceActionPlanArguments(args map[string]any) (map[string]any, error) {
	if raw, exists := args["arguments"]; exists {
		arguments, ok := raw.(map[string]any)
		if !ok {
			return nil, ports.ErrInvalidProviderInput
		}
		return arguments, nil
	}
	if raw, exists := args["argumentsJson"]; exists {
		if arguments, ok := raw.(map[string]any); ok {
			return arguments, nil
		}
	}
	rawJSON := strings.TrimSpace(stringArg(args["argumentsJson"]))
	if rawJSON == "" {
		return map[string]any{}, nil
	}
	var arguments map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &arguments); err != nil {
		return nil, ports.ErrInvalidProviderInput
	}
	if arguments == nil {
		return map[string]any{}, nil
	}
	return arguments, nil
}

func realtimeVoiceActionPlanRisks(args map[string]any) ([]string, error) {
	risks := []string{}
	if rawRisks, exists := args["risks"]; exists {
		values, ok := rawRisks.([]any)
		if !ok {
			return nil, ports.ErrInvalidProviderInput
		}
		for _, value := range values {
			risk, ok := value.(string)
			if !ok {
				return nil, ports.ErrInvalidProviderInput
			}
			if risk = strings.TrimSpace(risk); risk != "" {
				risks = append(risks, risk)
			}
		}
	}
	if risk := strings.TrimSpace(stringArg(args["riskSummary"])); risk != "" {
		risks = append(risks, risk)
	}
	return risks, nil
}

func validateRealtimeVoiceActionPlanVisibleIDs(commands []ActionPlanCommandInput, visibleAssetIDs map[string]struct{}) error {
	if len(commands) == 0 {
		return ports.ErrInvalidProviderInput
	}
	for _, command := range commands {
		ids, err := realtimeVoiceActionPlanReferencedAssetIDs(command)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if _, ok := visibleAssetIDs[id]; !ok {
				return ports.ErrInvalidProviderInput
			}
		}
	}
	return nil
}

func realtimeVoiceActionPlanReferencedAssetIDs(command ActionPlanCommandInput) ([]string, error) {
	payload, err := json.Marshal(command.Arguments)
	if err != nil {
		return nil, ports.ErrInvalidProviderInput
	}
	record := ports.ActionPlanCommandRecord{Kind: command.Kind, ArgumentsJSON: payload}
	ids := []string{}
	switch command.Kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		args, err := parseActionPlanCreateArguments(record)
		if err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		if strings.TrimSpace(args.ParentAssetID) != "" {
			ids = append(ids, args.ParentAssetID)
		}
	case actionplan.CommandKindMoveAsset:
		args, err := parseActionPlanMoveArguments(record)
		if err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		ids = append(ids, args.AssetID.String())
		if strings.TrimSpace(args.ParentAssetID) != "" {
			ids = append(ids, args.ParentAssetID)
		}
	case actionplan.CommandKindArchiveAsset, actionplan.CommandKindRestoreAsset:
		id, err := parseActionPlanAssetIDOnlyArguments(record)
		if err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		ids = append(ids, id.String())
	default:
		return nil, ports.ErrInvalidProviderInput
	}
	return ids, nil
}

func (a App) realtimeVoiceActionPlanProposal(ctx context.Context, session RealtimeVoiceSession, record ports.ActionPlanRecord) (RealtimeVoiceActionPlanProposal, error) {
	commands := make([]RealtimeVoiceActionPlanCommand, 0, len(record.Commands))
	for _, command := range record.Commands {
		proposalCommand, err := a.realtimeVoiceActionPlanCommand(ctx, session, command)
		if err != nil {
			return RealtimeVoiceActionPlanProposal{}, err
		}
		commands = append(commands, proposalCommand)
	}
	return RealtimeVoiceActionPlanProposal{
		PlanID:              record.ID,
		ConfirmationSummary: record.ConfirmationSummary,
		Commands:            commands,
		Risks:               append([]string{}, record.Risks...),
	}, nil
}

func (a App) realtimeVoiceActionPlanCommand(ctx context.Context, session RealtimeVoiceSession, command ports.ActionPlanCommandRecord) (RealtimeVoiceActionPlanCommand, error) {
	proposal := RealtimeVoiceActionPlanCommand{
		ID:        command.ID,
		Kind:      string(command.Kind),
		Summary:   command.Summary,
		Operation: actionPlanCommandOperation(command.Kind),
	}
	if command.Kind == actionplan.CommandKindCreateAsset || command.Kind == actionplan.CommandKindCreateLocation {
		args, err := parseActionPlanCreateArguments(command)
		if err == nil {
			proposal.Title = args.Title
			proposal.AssetKind = args.Kind
			if command.Kind == actionplan.CommandKindCreateLocation {
				proposal.AssetKind = asset.KindLocation.String()
			}
			if proposal.AssetKind == "" {
				proposal.AssetKind = asset.KindItem.String()
			}
			proposal.ParentAssetID = args.ParentAssetID
			if args.ParentAssetID != "" {
				parentID, ok := asset.NewID(args.ParentAssetID)
				if !ok {
					return RealtimeVoiceActionPlanCommand{}, ports.ErrInvalidProviderInput
				}
				parent, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, parentID)
				if err != nil {
					return RealtimeVoiceActionPlanCommand{}, err
				}
				if !found {
					return RealtimeVoiceActionPlanCommand{}, ports.ErrInvalidProviderInput
				}
				proposal.ParentTitle = parent.Title.String()
				proposal.ParentKind = parent.Kind.String()
			}
			proposal.ParentCommandID = args.ParentCommandID
		}
	} else if command.Kind == actionplan.CommandKindMoveAsset {
		args, err := parseActionPlanMoveArguments(command)
		if err == nil {
			proposal.ParentAssetID = args.ParentAssetID
			if args.ParentAssetID != "" {
				parentID, ok := asset.NewID(args.ParentAssetID)
				if !ok {
					return RealtimeVoiceActionPlanCommand{}, ports.ErrInvalidProviderInput
				}
				parent, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, parentID)
				if err != nil {
					return RealtimeVoiceActionPlanCommand{}, err
				}
				if !found {
					return RealtimeVoiceActionPlanCommand{}, ports.ErrInvalidProviderInput
				}
				proposal.ParentTitle = parent.Title.String()
				proposal.ParentKind = parent.Kind.String()
			}
			proposal.ParentCommandID = args.ParentCommandID
		}
	}
	return proposal, nil
}

func actionPlanCommandOperation(kind actionplan.CommandKind) string {
	switch kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		return "create"
	case actionplan.CommandKindMoveAsset:
		return "move"
	case actionplan.CommandKindArchiveAsset:
		return "archive"
	case actionplan.CommandKindRestoreAsset:
		return "restore"
	default:
		return "update"
	}
}

type realtimeVoiceActionPlanArgs struct {
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
}
