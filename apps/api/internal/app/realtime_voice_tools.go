package app

import (
	"context"
	"encoding/json"
	"math"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoiceToolMaxResults = 20

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	switch call.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets:
		result, err := a.executeRealtimeVoiceSearchTool(ctx, session, call)
		return result, nil, err
	case RealtimeVoiceToolListAuthorizedAssets:
		result, err := a.executeRealtimeVoiceListTool(ctx, session, call)
		return result, nil, err
	case RealtimeVoiceToolProposeActionPlan:
		return a.executeRealtimeVoiceProposeActionPlanTool(ctx, session, call, visibleAssetIDs)
	default:
		return ports.AgentToolResult{}, nil, ports.ErrInvalidProviderInput
	}
}

func (a App) executeRealtimeVoiceProposeActionPlanTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	args, err := parseRealtimeVoiceActionPlanArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceActionPlanVisibleIDs(args.Commands, visibleAssetIDs); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	record, err := a.CreateActionPlan(ctx, CreateActionPlanInput{
		Principal:                  session.Principal,
		TenantID:                   session.TenantID,
		InventoryID:                session.InventoryID,
		Source:                     session.Source,
		RealtimeSessionID:          session.ID,
		IntentSummary:              args.IntentSummary,
		ModelInterpretationSummary: args.ModelInterpretationSummary,
		ConfirmationSummary:        args.ConfirmationSummary,
		Commands:                   args.Commands,
		Risks:                      args.Risks,
	})
	if err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	proposal, err := a.realtimeVoiceActionPlanProposal(ctx, session, record)
	if err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	payload, err := json.Marshal(struct {
		Tool       string                          `json:"tool"`
		ActionPlan RealtimeVoiceActionPlanProposal `json:"actionPlan"`
	}{
		Tool:       call.Name,
		ActionPlan: proposal,
	})
	if err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	return ports.AgentToolResult{
		CallID:  call.ID,
		Name:    call.Name,
		Call:    call,
		Content: string(payload),
	}, &proposal, nil
}

func (a App) executeRealtimeVoiceSearchTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	args, err := parseRealtimeVoiceSearchArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	results, err := a.SearchAssets(ctx, SearchAssetsInput{
		Principal:      session.Principal,
		TenantID:       session.TenantID,
		InventoryIDs:   []inventory.InventoryID{session.InventoryID},
		Query:          args.Query,
		Mode:           "fuzzy",
		LifecycleState: "active",
		Limit:          args.Limit,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}

	items := make([]realtimeVoiceAssetToolItem, 0, len(results.Items))
	for _, result := range results.Items {
		item, err := a.realtimeVoiceAssetToolItem(ctx, session, result.Asset, result.Inventory.Name.String(), realtimeVoiceMatchFields(result.Matches), true)
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		items = append(items, item)
	}
	return realtimeVoiceToolResult(call, realtimeVoiceAssetToolOutput{
		Tool:    call.Name,
		Query:   args.Query,
		Count:   len(items),
		HasMore: results.HasMore,
		Items:   items,
	})
}

func (a App) executeRealtimeVoiceListTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	args, err := parseRealtimeVoiceListArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	inventoryItem, err := a.GetInventory(ctx, GetInventoryInput{
		Principal:   session.Principal,
		Source:      audit.SourceAPI,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}

	items := []realtimeVoiceAssetToolItem{}
	hasMore := false
	cursor := ""
	for page := 0; page < 50 && len(items) < args.Limit; page++ {
		result, err := a.ListAssets(ctx, ListAssetsInput{
			Principal:      session.Principal,
			Source:         audit.SourceAPI,
			TenantID:       session.TenantID,
			InventoryID:    session.InventoryID,
			Limit:          100,
			Cursor:         cursor,
			LifecycleState: args.LifecycleState,
			Sort:           string(ports.AssetListSortIDAsc),
		})
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		for _, visibleAsset := range result.Items {
			toolItem, err := a.realtimeVoiceAssetToolItem(ctx, session, visibleAsset, inventoryItem.Name.String(), nil, true)
			if err != nil {
				return ports.AgentToolResult{}, err
			}
			if args.Kind != "" && toolItem.Kind != args.Kind.String() {
				continue
			}
			if args.ParentTitle != "" && !strings.EqualFold(toolItem.ParentTitle, args.ParentTitle) {
				continue
			}
			if args.LocationTitle != "" && !strings.EqualFold(toolItem.LocationTitle, args.LocationTitle) {
				continue
			}
			items = append(items, toolItem)
			if len(items) >= args.Limit {
				break
			}
		}
		hasMore = result.HasMore
		if !result.HasMore || result.NextCursor == nil {
			break
		}
		cursor = *result.NextCursor
	}
	return realtimeVoiceToolResult(call, realtimeVoiceAssetToolOutput{
		Tool:    call.Name,
		Count:   len(items),
		HasMore: hasMore,
		Filters: map[string]string{
			"kind":           args.Kind.String(),
			"lifecycleState": args.LifecycleState,
			"parentTitle":    args.ParentTitle,
			"locationTitle":  args.LocationTitle,
		},
		Items: items,
	})
}

func (a App) realtimeVoiceAssetToolItem(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (realtimeVoiceAssetToolItem, error) {
	ancestors, err := a.realtimeVoiceAncestors(ctx, session, item)
	if err != nil {
		return realtimeVoiceAssetToolItem{}, err
	}
	path := make([]string, 0, len(ancestors)+1)
	locationTitle := ""
	for _, ancestor := range ancestors {
		path = append(path, ancestor.Title.String())
		if ancestor.Kind == asset.KindLocation {
			locationTitle = ancestor.Title.String()
		}
	}
	path = append(path, item.Title.String())
	parentTitle := ""
	parentKind := ""
	if len(ancestors) > 0 {
		parent := ancestors[len(ancestors)-1]
		parentTitle = parent.Title.String()
		parentKind = parent.Kind.String()
	}
	if item.Kind == asset.KindLocation {
		locationTitle = item.Title.String()
	}

	toolItem := realtimeVoiceAssetToolItem{
		Title:           item.Title.String(),
		Kind:            item.Kind.String(),
		Description:     item.Description.String(),
		InventoryName:   inventoryName,
		LifecycleState:  item.LifecycleState.String(),
		ParentTitle:     parentTitle,
		ParentKind:      parentKind,
		LocationTitle:   locationTitle,
		ContainmentPath: path,
		MatchFields:     matchFields,
	}
	if includeAssetID {
		toolItem.AssetID = item.ID.String()
	}
	return toolItem, nil
}

func (a App) realtimeVoiceAncestors(ctx context.Context, session RealtimeVoiceSession, item asset.Asset) ([]asset.Asset, error) {
	ancestors := []asset.Asset{}
	seen := map[asset.ID]struct{}{item.ID: {}}
	for parentID := item.ParentAssetID; parentID.String() != ""; {
		if _, duplicate := seen[parentID]; duplicate {
			return nil, ports.ErrInvalidProviderInput
		}
		seen[parentID] = struct{}{}
		parent, err := a.GetAsset(ctx, GetAssetInput{
			Principal:   session.Principal,
			Source:      audit.SourceAPI,
			TenantID:    session.TenantID,
			InventoryID: session.InventoryID,
			AssetID:     parentID,
		})
		if err != nil {
			return nil, err
		}
		ancestors = append([]asset.Asset{parent}, ancestors...)
		parentID = parent.ParentAssetID
	}
	return ancestors, nil
}

func realtimeVoiceToolResult(call ports.AgentToolCall, output realtimeVoiceAssetToolOutput) (ports.AgentToolResult, error) {
	if output.Count == 0 {
		output.Note = "No visible matching assets were returned. Do not claim the inventory is empty unless this was a list query broad enough to inspect the relevant asset kind."
	}
	payload, err := json.Marshal(output)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	return ports.AgentToolResult{
		CallID:  call.ID,
		Name:    call.Name,
		Call:    call,
		Content: string(payload),
	}, nil
}

func realtimeVoiceToolErrorResult(call ports.AgentToolCall, code string, message string, retryable bool) (ports.AgentToolResult, error) {
	payload, err := json.Marshal(struct {
		Tool      string `json:"tool"`
		Status    string `json:"status"`
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
	}{
		Tool:      call.Name,
		Status:    "error",
		Code:      code,
		Message:   message,
		Retryable: retryable,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	return ports.AgentToolResult{
		CallID:  call.ID,
		Name:    call.Name,
		Call:    ports.AgentToolCall{ID: call.ID, Name: call.Name, Arguments: map[string]any{}},
		Content: string(payload),
	}, nil
}

func collectRealtimeVoiceVisibleAssetIDs(result ports.AgentToolResult, visibleAssetIDs map[string]struct{}) error {
	if visibleAssetIDs == nil || strings.TrimSpace(result.Content) == "" {
		return nil
	}
	var output realtimeVoiceAssetToolOutput
	if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
		return ports.ErrInvalidProviderInput
	}
	for _, item := range output.Items {
		id := strings.TrimSpace(item.AssetID)
		if id == "" {
			continue
		}
		if _, ok := asset.NewID(id); !ok {
			return ports.ErrInvalidProviderInput
		}
		visibleAssetIDs[id] = struct{}{}
	}
	return nil
}

func realtimeVoiceToolLimit(raw any) (int, error) {
	if raw == nil {
		return 10, nil
	}
	switch value := raw.(type) {
	case float64:
		if math.IsNaN(value) || value != math.Trunc(value) || value < 1 {
			return 0, ports.ErrInvalidProviderInput
		}
		if value > realtimeVoiceToolMaxResults {
			return realtimeVoiceToolMaxResults, nil
		}
		return int(value), nil
	case int:
		if value < 1 {
			return 0, ports.ErrInvalidProviderInput
		}
		if value > realtimeVoiceToolMaxResults {
			return realtimeVoiceToolMaxResults, nil
		}
		return value, nil
	default:
		return 0, ports.ErrInvalidProviderInput
	}
}

func realtimeVoiceOptionalAssetKind(raw any) (asset.Kind, error) {
	if raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	kind, ok := asset.NewKind(value)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	return kind, nil
}

func realtimeVoiceOptionalLifecycleState(raw any) (string, error) {
	if raw == nil {
		return "active", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "active", nil
	}
	switch value {
	case "active", "archived", "all":
		return value, nil
	default:
		return "", ports.ErrInvalidProviderInput
	}
}

func realtimeVoiceMatchFields(matches []search.Match) []string {
	fields := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		field := match.Field.String()
		if field == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	return fields
}

func stringArg(raw any) string {
	value, _ := raw.(string)
	return value
}

func parseRealtimeVoiceSearchArgs(args map[string]any) (realtimeVoiceSearchArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "query", "limit"); err != nil {
		return realtimeVoiceSearchArgs{}, err
	}
	query := strings.TrimSpace(stringArg(args["query"]))
	if query == "" || len(query) > 120 {
		return realtimeVoiceSearchArgs{}, ports.ErrInvalidProviderInput
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceSearchArgs{}, err
	}
	return realtimeVoiceSearchArgs{Query: query, Limit: limit}, nil
}

func parseRealtimeVoiceListArgs(args map[string]any) (realtimeVoiceListArgs, error) {
	if err := rejectUnknownRealtimeVoiceArgs(args, "kind", "lifecycleState", "parentTitle", "locationTitle", "limit"); err != nil {
		return realtimeVoiceListArgs{}, err
	}
	kind, err := realtimeVoiceOptionalAssetKind(args["kind"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	parentTitle, err := optionalRealtimeVoiceTitle(args["parentTitle"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	locationTitle, err := optionalRealtimeVoiceTitle(args["locationTitle"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	limit, err := realtimeVoiceToolLimit(args["limit"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	lifecycleState, err := realtimeVoiceOptionalLifecycleState(args["lifecycleState"])
	if err != nil {
		return realtimeVoiceListArgs{}, err
	}
	return realtimeVoiceListArgs{Kind: kind, LifecycleState: lifecycleState, ParentTitle: parentTitle, LocationTitle: locationTitle, Limit: limit}, nil
}

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
		if !args.ParentIsRoot {
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

func rejectUnknownRealtimeVoiceArgs(args map[string]any, allowed ...string) error {
	allowedSet := map[string]struct{}{}
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range args {
		if _, ok := allowedSet[key]; !ok {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func optionalRealtimeVoiceTitle(raw any) (string, error) {
	if raw == nil {
		return "", nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", ports.ErrInvalidProviderInput
	}
	value = strings.TrimSpace(value)
	if len(value) > 160 {
		return "", ports.ErrInvalidProviderInput
	}
	return value, nil
}

type realtimeVoiceSearchArgs struct {
	Query string
	Limit int
}

type realtimeVoiceListArgs struct {
	Kind           asset.Kind
	LifecycleState string
	ParentTitle    string
	LocationTitle  string
	Limit          int
}

type realtimeVoiceActionPlanArgs struct {
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
}

type realtimeVoiceAssetToolOutput struct {
	Tool    string                       `json:"tool"`
	Query   string                       `json:"query,omitempty"`
	Filters map[string]string            `json:"filters,omitempty"`
	Count   int                          `json:"count"`
	HasMore bool                         `json:"hasMore,omitempty"`
	Note    string                       `json:"note,omitempty"`
	Items   []realtimeVoiceAssetToolItem `json:"items"`
}

type realtimeVoiceAssetToolItem struct {
	AssetID         string   `json:"assetId,omitempty"`
	Title           string   `json:"title"`
	Kind            string   `json:"kind"`
	Description     string   `json:"description,omitempty"`
	InventoryName   string   `json:"inventoryName"`
	LifecycleState  string   `json:"lifecycleState"`
	ParentTitle     string   `json:"parentTitle,omitempty"`
	ParentKind      string   `json:"parentKind,omitempty"`
	LocationTitle   string   `json:"locationTitle,omitempty"`
	ContainmentPath []string `json:"containmentPath,omitempty"`
	MatchFields     []string `json:"matchFields,omitempty"`
}
