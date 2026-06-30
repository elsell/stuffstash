package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func parseRealtimeVoiceActionPlanArgs(args map[string]any, transcript string) (realtimeVoiceActionPlanArgs, error) {
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
		Commands:                   canonicalRealtimeVoiceActionPlanCommandDependencies(canonicalRealtimeVoiceTranscriptCreateHierarchy(commands, transcript)),
		Risks:                      risks,
	}
	if parsed.IntentSummary == "" || parsed.ModelInterpretationSummary == "" || parsed.ConfirmationSummary == "" || len(parsed.Commands) == 0 {
		return realtimeVoiceActionPlanArgs{}, ports.ErrInvalidProviderInput
	}
	if err := validateRealtimeVoiceActionPlanMoveDependencies(parsed); err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	if err := validateRealtimeVoiceRootMoves(parsed.Commands, transcript); err != nil {
		return realtimeVoiceActionPlanArgs{}, err
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
			summary := strings.TrimSpace(firstStringArg(command["summary"], command["commandSummary"]))
			if summary == "" {
				return nil, ports.ErrInvalidProviderInput
			}
			arguments, err := realtimeVoiceActionPlanArguments(command)
			if err != nil {
				return nil, err
			}
			kind, arguments, err := canonicalRealtimeVoiceCommandKind(firstStringArg(command["kind"], command["commandKind"]), arguments)
			if err != nil {
				return nil, err
			}
			commandID := strings.TrimSpace(firstStringArg(command["id"], command["commandId"]))
			arguments = canonicalRealtimeVoiceDependentParentReference(arguments, previousCommandIDs)
			arguments = canonicalRealtimeVoiceConflictingParentReference(arguments)
			kind = canonicalRealtimeVoiceCreateLocationKind(kind, arguments)
			kind, arguments = canonicalRealtimeVoiceCreateLocationStorageKind(kind, arguments)
			arguments = canonicalRealtimeVoiceCreateAssetKind(kind, arguments)
			arguments = canonicalRealtimeVoiceCreateParentAssetID(kind, arguments)
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
		commands = canonicalRealtimeVoiceCreatedItemMoveCommands(commands)
		commands = canonicalRealtimeVoiceCreatedItemContainerDestination(commands)
		commands = canonicalRealtimeVoiceActionPlanCommandDependencies(commands)
		return commands, nil
	}

	arguments, err := realtimeVoiceActionPlanArguments(args)
	if err != nil {
		return nil, err
	}
	commandKind, arguments, err := canonicalRealtimeVoiceCommandKind(stringArg(args["commandKind"]), arguments)
	if err != nil {
		return nil, err
	}
	commandKind = canonicalRealtimeVoiceCreateLocationKind(commandKind, arguments)
	arguments = canonicalRealtimeVoiceConflictingParentReference(arguments)
	commandKind, arguments = canonicalRealtimeVoiceCreateLocationStorageKind(commandKind, arguments)
	arguments = canonicalRealtimeVoiceCreateAssetKind(commandKind, arguments)
	arguments = canonicalRealtimeVoiceCreateParentAssetID(commandKind, arguments)
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

func canonicalRealtimeVoiceCommandKind(raw string, arguments map[string]any) (actionplan.CommandKind, map[string]any, error) {
	kind := actionplan.CommandKind(strings.TrimSpace(raw))
	switch kind {
	case "create_item":
		return actionplan.CommandKindCreateAsset, realtimeVoiceArgumentsWithDefaultKind(arguments, asset.KindItem.String()), nil
	case "create_container":
		return actionplan.CommandKindCreateAsset, realtimeVoiceArgumentsWithDefaultKind(arguments, asset.KindContainer.String()), nil
	default:
		if !kind.Valid() {
			return "", nil, ports.ErrInvalidProviderInput
		}
		return kind, arguments, nil
	}
}

func realtimeVoiceArgumentsWithDefaultKind(arguments map[string]any, defaultKind string) map[string]any {
	if arguments == nil {
		arguments = map[string]any{}
	}
	if strings.TrimSpace(stringArg(arguments["kind"])) != "" {
		return arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		canonical[key] = value
	}
	canonical["kind"] = defaultKind
	return canonical
}

func validateRealtimeVoiceActionPlanMoveDependencies(plan realtimeVoiceActionPlanArgs) error {
	createdTitles := []string{}
	planText := strings.ToLower(plan.IntentSummary + " " + plan.ModelInterpretationSummary + " " + plan.ConfirmationSummary)
	if planText == "" {
		return nil
	}
	for _, command := range plan.Commands {
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			continue
		}
		title := strings.ToLower(strings.TrimSpace(firstStringArg(command.Arguments["title"], command.Arguments["name"])))
		if title != "" {
			createdTitles = append(createdTitles, title)
		}
	}
	if len(createdTitles) == 0 {
		return nil
	}
	for _, command := range plan.Commands {
		if command.Kind != actionplan.CommandKindMoveAsset {
			continue
		}
		if strings.TrimSpace(stringArg(command.Arguments["parentAssetId"])) != "" || strings.TrimSpace(stringArg(command.Arguments["parentCommandId"])) != "" {
			continue
		}
		for _, title := range createdTitles {
			if realtimeVoicePlanTextTargetsCreatedTitle(planText, title) {
				return ports.ErrInvalidProviderInput
			}
		}
	}
	return nil
}

func realtimeVoicePlanTextTargetsCreatedTitle(planText string, title string) bool {
	title = strings.TrimSpace(title)
	if title == "" {
		return false
	}
	for _, prefix := range []string{" to ", " into ", " inside ", " in "} {
		if strings.Contains(planText, prefix+title) ||
			strings.Contains(planText, prefix+"the "+title) ||
			strings.Contains(planText, prefix+"a "+title) ||
			strings.Contains(planText, prefix+"an "+title) {
			return true
		}
	}
	return false
}

func validateRealtimeVoiceRootMoves(commands []ActionPlanCommandInput, transcript string) error {
	if !realtimeVoiceTranscriptNamesDestination(transcript) || realtimeVoiceTranscriptAllowsRootDestination(transcript) {
		return nil
	}
	for _, command := range commands {
		if command.Kind != actionplan.CommandKindMoveAsset {
			continue
		}
		parentAssetID := strings.TrimSpace(stringArg(command.Arguments["parentAssetId"]))
		parentCommandID := strings.TrimSpace(stringArg(command.Arguments["parentCommandId"]))
		if parentAssetID == "" && parentCommandID == "" {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func realtimeVoiceTranscriptNamesDestination(transcript string) bool {
	normalized := " " + strings.ToLower(strings.TrimSpace(transcript)) + " "
	for _, token := range []string{" to ", " into ", " inside ", " in "} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceTranscriptAllowsRootDestination(transcript string) bool {
	normalized := " " + strings.ToLower(strings.TrimSpace(transcript)) + " "
	for _, replacer := range []string{".", ",", "!", "?", ";", ":", "\"", "'", "(", ")", "[", "]", "{", "}"} {
		normalized = strings.ReplaceAll(normalized, replacer, " ")
	}
	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = " " + normalized + " "
	for _, phrase := range []string{" root ", " top level ", " top-level ", " inventory root ", " no parent ", " out of ", " remove from ", " take out of "} {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
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

func canonicalRealtimeVoiceConflictingParentReference(arguments map[string]any) map[string]any {
	if strings.TrimSpace(stringArg(arguments["parentAssetId"])) == "" || strings.TrimSpace(stringArg(arguments["parentCommandId"])) == "" {
		return arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		if key == "parentCommandId" {
			continue
		}
		canonical[key] = value
	}
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

func canonicalRealtimeVoiceCreateLocationStorageKind(kind actionplan.CommandKind, arguments map[string]any) (actionplan.CommandKind, map[string]any) {
	if kind != actionplan.CommandKindCreateLocation || !realtimeVoiceTitleLooksLikeContainer(firstStringArg(arguments["title"], arguments["name"])) {
		return kind, arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		if key == "kind" {
			continue
		}
		canonical[key] = value
	}
	canonical["kind"] = asset.KindContainer.String()
	return actionplan.CommandKindCreateAsset, canonical
}

func canonicalRealtimeVoiceCreateAssetKind(kind actionplan.CommandKind, arguments map[string]any) map[string]any {
	if kind != actionplan.CommandKindCreateAsset || strings.TrimSpace(stringArg(arguments["kind"])) != asset.KindContainer.String() {
		return arguments
	}
	title := strings.TrimSpace(firstStringArg(arguments["title"], arguments["name"]))
	if title == "" || realtimeVoiceTitleLooksLikeContainer(title) {
		return arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		canonical[key] = value
	}
	canonical["kind"] = asset.KindItem.String()
	return canonical
}

func realtimeVoiceTitleLooksLikeContainer(title string) bool {
	for _, word := range realtimeVoiceMeaningfulWords(title) {
		if realtimeVoiceContainerSegmentWords[word] {
			return true
		}
	}
	return false
}

func canonicalRealtimeVoiceCreateParentAssetID(kind actionplan.CommandKind, arguments map[string]any) map[string]any {
	if kind != actionplan.CommandKindCreateAsset && kind != actionplan.CommandKindCreateLocation {
		return arguments
	}
	if strings.TrimSpace(stringArg(arguments["assetId"])) == "" {
		return arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		if key == "assetId" {
			continue
		}
		canonical[key] = value
	}
	if strings.TrimSpace(stringArg(arguments["parentAssetId"])) != "" ||
		strings.TrimSpace(stringArg(arguments["parentCommandId"])) != "" ||
		strings.TrimSpace(firstStringArg(arguments["title"], arguments["name"])) == "" {
		return canonical
	}
	canonical["parentAssetId"] = arguments["assetId"]
	return canonical
}

func canonicalRealtimeVoiceCreatedItemMoveCommands(commands []ActionPlanCommandInput) []ActionPlanCommandInput {
	if len(commands) < 2 {
		return commands
	}
	createItemIndexByID := map[string]int{}
	createItemIndexByGuessedID := map[string]int{}
	for index, command := range commands {
		if command.ID == "" || command.Kind != actionplan.CommandKindCreateAsset {
			continue
		}
		kind := strings.TrimSpace(stringArg(command.Arguments["kind"]))
		if kind == "" {
			kind = asset.KindItem.String()
		}
		if kind == asset.KindItem.String() {
			createItemIndexByID[command.ID] = index
			for _, guessedID := range realtimeVoiceGuessedAssetIDsFromTitle(firstStringArg(command.Arguments["title"], command.Arguments["name"])) {
				createItemIndexByGuessedID[guessedID] = index
			}
		}
	}
	if len(createItemIndexByID) == 0 {
		return commands
	}
	dropMoveIndexes := map[int]struct{}{}
	normalized := append([]ActionPlanCommandInput{}, commands...)
	for index, command := range commands {
		if command.Kind != actionplan.CommandKindMoveAsset {
			continue
		}
		assetID := strings.TrimSpace(stringArg(command.Arguments["assetId"]))
		parentAssetID := strings.TrimSpace(stringArg(command.Arguments["parentAssetId"]))
		parentCommandID := strings.TrimSpace(stringArg(command.Arguments["parentCommandId"]))
		if createIndex, ok := createItemIndexByID[assetID]; ok && (parentAssetID != "" || parentCommandID != "") {
			normalized[createIndex].Arguments = realtimeVoiceArgumentsWithParentReference(normalized[createIndex].Arguments, parentAssetID, parentCommandID)
			dropMoveIndexes[index] = struct{}{}
			continue
		}
		if createIndex, ok := createItemIndexByGuessedID[assetID]; ok && (parentAssetID != "" || parentCommandID != "") {
			normalized[createIndex].Arguments = realtimeVoiceArgumentsWithParentReference(normalized[createIndex].Arguments, parentAssetID, parentCommandID)
			dropMoveIndexes[index] = struct{}{}
			continue
		}
		if createIndex, ok := createItemIndexByID[parentCommandID]; ok && assetID != "" && parentAssetID == "" {
			normalized[createIndex].Arguments = realtimeVoiceArgumentsWithParentReference(normalized[createIndex].Arguments, assetID, "")
			dropMoveIndexes[index] = struct{}{}
		}
	}
	if len(dropMoveIndexes) == 0 {
		return commands
	}
	compacted := make([]ActionPlanCommandInput, 0, len(normalized)-len(dropMoveIndexes))
	for index, command := range normalized {
		if _, drop := dropMoveIndexes[index]; drop {
			continue
		}
		compacted = append(compacted, command)
	}
	return compacted
}

func canonicalRealtimeVoiceCreatedItemContainerDestination(commands []ActionPlanCommandInput) []ActionPlanCommandInput {
	if len(commands) < 2 {
		return commands
	}
	itemIndexes := []int{}
	containerIndexes := []int{}
	for index, command := range commands {
		if command.Kind != actionplan.CommandKindCreateAsset {
			continue
		}
		kind := strings.TrimSpace(stringArg(command.Arguments["kind"]))
		if kind == "" {
			kind = asset.KindItem.String()
		}
		switch kind {
		case asset.KindItem.String():
			itemIndexes = append(itemIndexes, index)
		case asset.KindContainer.String():
			containerIndexes = append(containerIndexes, index)
		}
	}
	if len(itemIndexes) != 1 || len(containerIndexes) != 1 {
		return commands
	}
	itemIndex := itemIndexes[0]
	container := commands[containerIndexes[0]]
	if container.ID == "" {
		return commands
	}
	item := commands[itemIndex]
	itemParentCommandID := strings.TrimSpace(stringArg(item.Arguments["parentCommandId"]))
	containerParentAssetID := strings.TrimSpace(stringArg(container.Arguments["parentAssetId"]))
	itemParentAssetID := strings.TrimSpace(stringArg(item.Arguments["parentAssetId"]))
	if itemParentCommandID != "" || (itemParentAssetID != "" && itemParentAssetID != containerParentAssetID) {
		return commands
	}
	normalized := append([]ActionPlanCommandInput{}, commands...)
	canonical := map[string]any{}
	for key, value := range item.Arguments {
		if key == "parentAssetId" {
			continue
		}
		canonical[key] = value
	}
	canonical["parentCommandId"] = container.ID
	normalized[itemIndex].Arguments = canonical
	return normalized
}

func realtimeVoiceArgumentsWithParentReference(arguments map[string]any, parentAssetID string, parentCommandID string) map[string]any {
	if strings.TrimSpace(stringArg(arguments["parentCommandId"])) != "" {
		return arguments
	}
	canonical := map[string]any{}
	for key, value := range arguments {
		if parentCommandID != "" && key == "parentAssetId" {
			continue
		}
		canonical[key] = value
	}
	if parentCommandID != "" {
		canonical["parentCommandId"] = parentCommandID
		return canonical
	}
	if strings.TrimSpace(stringArg(arguments["parentAssetId"])) != "" {
		return arguments
	}
	if parentAssetID != "" {
		canonical["parentAssetId"] = parentAssetID
	}
	return canonical
}

func realtimeVoiceGuessedAssetIDsFromTitle(title string) []string {
	words := realtimeVoiceMeaningfulWords(title)
	if len(words) == 0 {
		return nil
	}
	values := []string{strings.Join(words, "-") + "-1"}
	withoutModifiers := make([]string, 0, len(words))
	for _, word := range words {
		switch word {
		case "spare", "pack":
			continue
		default:
			withoutModifiers = append(withoutModifiers, word)
		}
	}
	if len(withoutModifiers) > 0 && len(withoutModifiers) != len(words) {
		values = append(values, strings.Join(withoutModifiers, "-")+"-1")
	}
	return values
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

type realtimeVoiceActionPlanArgs struct {
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
}
