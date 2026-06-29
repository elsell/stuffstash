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
		previousCommandIDs := map[string]struct{}{}
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
			commandID := strings.TrimSpace(firstStringArg(command["id"], command["commandId"]))
			arguments = canonicalRealtimeVoiceDependentParentReference(arguments, previousCommandIDs)
			kind = canonicalRealtimeVoiceCreateLocationKind(kind, arguments)
			commands = append(commands, ActionPlanCommandInput{
				ID:        commandID,
				Kind:      kind,
				Summary:   summary,
				Arguments: arguments,
			})
			if commandID != "" {
				previousCommandIDs[commandID] = struct{}{}
			}
		}
		return canonicalRealtimeVoiceActionPlanCommandDependencies(commands), nil
	}

	commandKind := actionplan.CommandKind(strings.TrimSpace(stringArg(args["commandKind"])))
	if !commandKind.Valid() {
		return nil, ports.ErrInvalidProviderInput
	}
	arguments, err := realtimeVoiceActionPlanArguments(args)
	if err != nil {
		return nil, err
	}
	commandKind = canonicalRealtimeVoiceCreateLocationKind(commandKind, arguments)
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

func canonicalRealtimeVoiceDependentParentReference(arguments map[string]any, previousCommandIDs map[string]struct{}) map[string]any {
	if len(arguments) == 0 || len(previousCommandIDs) == 0 {
		return arguments
	}
	if strings.TrimSpace(stringArg(arguments["parentCommandId"])) != "" {
		return arguments
	}
	parentAssetID := strings.TrimSpace(stringArg(arguments["parentAssetId"]))
	if parentAssetID == "" {
		return arguments
	}
	if _, ok := previousCommandIDs[parentAssetID]; !ok {
		return arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		if key == "parentAssetId" {
			continue
		}
		canonical[key] = value
	}
	canonical["parentCommandId"] = parentAssetID
	return canonical
}

func canonicalRealtimeVoiceCreateLocationKind(kind actionplan.CommandKind, arguments map[string]any) actionplan.CommandKind {
	if kind != actionplan.CommandKindCreateAsset {
		return kind
	}
	if strings.TrimSpace(stringArg(arguments["kind"])) != asset.KindLocation.String() {
		return kind
	}
	return actionplan.CommandKindCreateLocation
}

func canonicalRealtimeVoiceActionPlanCommandDependencies(commands []ActionPlanCommandInput) []ActionPlanCommandInput {
	if len(commands) < 2 {
		return commands
	}
	normalized := make([]ActionPlanCommandInput, 0, len(commands))
	indexByID := map[string]int{}
	for _, command := range commands {
		if command.ID != "" {
			indexByID[command.ID] = len(normalized)
		}
		normalized = append(normalized, command)
	}

	indegree := make([]int, len(normalized))
	dependents := map[int][]int{}
	for index, command := range normalized {
		parentCommandID := strings.TrimSpace(stringArg(command.Arguments["parentCommandId"]))
		parentIndex, ok := indexByID[parentCommandID]
		if !ok {
			continue
		}
		indegree[index]++
		dependents[parentIndex] = append(dependents[parentIndex], index)
	}
	ready := make([]int, 0, len(normalized))
	for index, count := range indegree {
		if count == 0 {
			ready = append(ready, index)
		}
	}
	ordered := make([]ActionPlanCommandInput, 0, len(normalized))
	for len(ready) > 0 {
		index := ready[0]
		ready = ready[1:]
		ordered = append(ordered, normalized[index])
		for _, dependent := range dependents[index] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = appendStableRealtimeVoiceCommandIndex(ready, dependent)
			}
		}
	}
	if len(ordered) != len(normalized) {
		return normalized
	}
	return ordered
}

func appendStableRealtimeVoiceCommandIndex(indexes []int, index int) []int {
	insertAt := len(indexes)
	for readyIndex, existing := range indexes {
		if index < existing {
			insertAt = readyIndex
			break
		}
	}
	indexes = append(indexes, 0)
	copy(indexes[insertAt+1:], indexes[insertAt:])
	indexes[insertAt] = index
	return indexes
}

type realtimeVoiceActionPlanArgs struct {
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
}
