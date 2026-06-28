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

func realtimeVoiceToolDescriptors() []ports.AgentToolDescriptor {
	return []ports.AgentToolDescriptor{
		{
			Name:        RealtimeVoiceToolSearchAuthorizedAssets,
			Label:       realtimeVoiceSearchAuthorizedAssetsPublicName,
			Description: "Search visible assets in the selected inventory by natural-language keywords. Use this for where-is, do-I-have, or specific-item questions. Arguments: query string, optional limit number. Results are JSON with asset metadata and containment paths.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"query"},
				Properties: map[string]ports.AgentToolParameter{
					"query": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Short natural-language keywords for the visible asset, container, or location the user asked about.",
					},
					"limit": {
						Type:        ports.AgentToolParameterTypeInteger,
						Description: "Maximum number of visible matching assets to return. Defaults to 10 and is capped at 20.",
					},
				},
			},
		},
		{
			Name:        RealtimeVoiceToolListAuthorizedAssets,
			Label:       realtimeVoiceListAuthorizedAssetsPublicName,
			Description: "List visible assets in the selected inventory. Use this for broad inventory questions like what items do I have, what is in a place, or what archived item should be restored. Arguments: optional kind item|container|location, optional lifecycleState active|archived|all, optional parentTitle string, optional locationTitle string, optional limit number. Results are JSON with asset metadata, internal asset IDs for action-plan arguments, and containment paths.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Properties: map[string]ports.AgentToolParameter{
					"kind": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional asset kind filter.",
						Enum:        []string{"item", "container", "location"},
					},
					"lifecycleState": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional lifecycle filter. Defaults to active. Use archived when the user asks to restore an archived asset.",
						Enum:        []string{"active", "archived", "all"},
					},
					"parentTitle": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional direct parent title filter for questions about what is inside a specific container or location.",
					},
					"locationTitle": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional containing location title filter for questions about what is in a place.",
					},
					"limit": {
						Type:        ports.AgentToolParameterTypeInteger,
						Description: "Maximum number of visible assets to return. Defaults to 10 and is capped at 20.",
					},
				},
			},
		},
		{
			Name:        RealtimeVoiceToolProposeActionPlan,
			Label:       realtimeVoiceProposeActionPlanPublicName,
			Description: "Prepare a user-reviewable action plan for a requested inventory change. This does not execute the change. Use only when the user asks to create, move, update, archive, or restore an inventory item or location. Arguments: commandKind enum, intentSummary, modelInterpretationSummary, confirmationSummary, commandSummary, optional argumentsJson object string, optional riskSummary.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"commandKind", "intentSummary", "modelInterpretationSummary", "confirmationSummary", "commandSummary"},
				Properties: map[string]ports.AgentToolParameter{
					"commandKind": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Action-plan command kind.",
						Enum:        []string{"create_asset", "create_location", "move_asset", "update_asset", "archive_asset", "restore_asset"},
					},
					"intentSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Safe user-facing summary of what the user asked to change.",
					},
					"modelInterpretationSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Safe user-facing summary of how the request was interpreted.",
					},
					"confirmationSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Short confirmation question shown to the user.",
					},
					"commandSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Short safe summary of the single proposed command.",
					},
					"argumentsJson": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional JSON object containing safe command arguments such as title, kind, or parent title. Do not include secrets, prompts, transcripts, provider data, hidden IDs, or approval claims.",
					},
					"riskSummary": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional short safe risk summary.",
					},
				},
			},
		},
	}
}

func realtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolProposeActionPlan:
		return realtimeVoiceProposeActionPlanPublicName
	case RealtimeVoiceToolListAuthorizedAssets:
		return realtimeVoiceListAuthorizedAssetsPublicName
	default:
		return realtimeVoiceSearchAuthorizedAssetsPublicName
	}
}

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	switch call.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets:
		result, err := a.executeRealtimeVoiceSearchTool(ctx, session, call)
		return result, nil, err
	case RealtimeVoiceToolListAuthorizedAssets:
		result, err := a.executeRealtimeVoiceListTool(ctx, session, call)
		return result, nil, err
	case RealtimeVoiceToolProposeActionPlan:
		return a.executeRealtimeVoiceProposeActionPlanTool(ctx, session, call)
	default:
		return ports.AgentToolResult{}, nil, ports.ErrForbidden
	}
}

func (a App) executeRealtimeVoiceProposeActionPlanTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	args, err := parseRealtimeVoiceActionPlanArgs(call.Arguments)
	if err != nil {
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
		Commands: []ActionPlanCommandInput{{
			Kind:      args.CommandKind,
			Summary:   args.CommandSummary,
			Arguments: args.Arguments,
		}},
		Risks: args.Risks,
	})
	if err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	proposal := realtimeVoiceActionPlanProposal(record)
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
		item, err := a.realtimeVoiceAssetToolItem(ctx, session, result.Asset, result.Inventory.Name.String(), realtimeVoiceMatchFields(result.Matches), false)
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
	if err := rejectUnknownRealtimeVoiceArgs(args, "commandKind", "intentSummary", "modelInterpretationSummary", "confirmationSummary", "commandSummary", "arguments", "argumentsJson", "risks", "riskSummary"); err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	commandKind := actionplan.CommandKind(strings.TrimSpace(stringArg(args["commandKind"])))
	if !commandKind.Valid() {
		return realtimeVoiceActionPlanArgs{}, ports.ErrInvalidProviderInput
	}
	arguments, err := realtimeVoiceActionPlanArguments(args)
	if err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	risks, err := realtimeVoiceActionPlanRisks(args)
	if err != nil {
		return realtimeVoiceActionPlanArgs{}, err
	}
	parsed := realtimeVoiceActionPlanArgs{
		CommandKind:                commandKind,
		IntentSummary:              strings.TrimSpace(stringArg(args["intentSummary"])),
		ModelInterpretationSummary: strings.TrimSpace(stringArg(args["modelInterpretationSummary"])),
		ConfirmationSummary:        strings.TrimSpace(stringArg(args["confirmationSummary"])),
		CommandSummary:             strings.TrimSpace(stringArg(args["commandSummary"])),
		Arguments:                  arguments,
		Risks:                      risks,
	}
	if parsed.IntentSummary == "" || parsed.ModelInterpretationSummary == "" || parsed.ConfirmationSummary == "" || parsed.CommandSummary == "" {
		return realtimeVoiceActionPlanArgs{}, ports.ErrInvalidProviderInput
	}
	return parsed, nil
}

func realtimeVoiceActionPlanArguments(args map[string]any) (map[string]any, error) {
	if raw, exists := args["arguments"]; exists {
		arguments, ok := raw.(map[string]any)
		if !ok {
			return nil, ports.ErrInvalidProviderInput
		}
		return arguments, nil
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

func realtimeVoiceActionPlanProposal(record ports.ActionPlanRecord) RealtimeVoiceActionPlanProposal {
	commands := make([]RealtimeVoiceActionPlanCommand, 0, len(record.Commands))
	for _, command := range record.Commands {
		commands = append(commands, RealtimeVoiceActionPlanCommand{
			Kind:    string(command.Kind),
			Summary: command.Summary,
		})
	}
	return RealtimeVoiceActionPlanProposal{
		PlanID:              record.ID,
		ConfirmationSummary: record.ConfirmationSummary,
		Commands:            commands,
		Risks:               append([]string{}, record.Risks...),
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
	CommandKind                actionplan.CommandKind
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	CommandSummary             string
	Arguments                  map[string]any
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
