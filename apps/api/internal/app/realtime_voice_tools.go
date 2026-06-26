package app

import (
	"context"
	"encoding/json"
	"math"
	"strings"

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
			Description: "List visible assets in the selected inventory. Use this for broad inventory questions like what items do I have or what is in a place. Arguments: optional kind item|container|location, optional parentTitle string, optional locationTitle string, optional limit number. Results are JSON with asset metadata and containment paths.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Properties: map[string]ports.AgentToolParameter{
					"kind": {
						Type:        ports.AgentToolParameterTypeString,
						Description: "Optional asset kind filter.",
						Enum:        []string{"item", "container", "location"},
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
	}
}

func realtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolListAuthorizedAssets:
		return realtimeVoiceListAuthorizedAssetsPublicName
	default:
		return realtimeVoiceSearchAuthorizedAssetsPublicName
	}
}

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	switch call.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets:
		return a.executeRealtimeVoiceSearchTool(ctx, session, call)
	case RealtimeVoiceToolListAuthorizedAssets:
		return a.executeRealtimeVoiceListTool(ctx, session, call)
	default:
		return ports.AgentToolResult{}, ports.ErrForbidden
	}
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
		item, err := a.realtimeVoiceAssetToolItem(ctx, session, result.Asset, result.Inventory.Name.String(), realtimeVoiceMatchFields(result.Matches))
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
			LifecycleState: "active",
			Sort:           string(ports.AssetListSortIDAsc),
		})
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		for _, visibleAsset := range result.Items {
			toolItem, err := a.realtimeVoiceAssetToolItem(ctx, session, visibleAsset, inventoryItem.Name.String(), nil)
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
			"kind":          args.Kind.String(),
			"parentTitle":   args.ParentTitle,
			"locationTitle": args.LocationTitle,
		},
		Items: items,
	})
}

func (a App) realtimeVoiceAssetToolItem(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string) (realtimeVoiceAssetToolItem, error) {
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

	return realtimeVoiceAssetToolItem{
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
	}, nil
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
	if err := rejectUnknownRealtimeVoiceArgs(args, "kind", "parentTitle", "locationTitle", "limit"); err != nil {
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
	return realtimeVoiceListArgs{Kind: kind, ParentTitle: parentTitle, LocationTitle: locationTitle, Limit: limit}, nil
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
	Kind          asset.Kind
	ParentTitle   string
	LocationTitle string
	Limit         int
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
