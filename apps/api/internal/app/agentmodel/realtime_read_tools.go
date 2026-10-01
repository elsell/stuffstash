package agentmodel

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
	searchapp "github.com/stuffstash/stuff-stash/internal/app/search"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
	"time"
)

const realtimeVoiceToolMaxResults = tools.RealtimeVoiceToolMaxResults

func (a RealtimeReadTools) ExecuteRealtimeVoiceTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	toolCtx, cancel := context.WithTimeout(ctx, a.realtimeVoiceToolCallTimeout)
	defer cancel()
	switch call.Name {
	case RealtimeVoiceToolGetExpirationCalendar:
		result, err := a.ExecuteRealtimeVoiceExpirationCalendar(toolCtx, session, call)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolQueryExpiringAssets:
		result, err := a.ExecuteRealtimeVoiceExpirationQuery(toolCtx, session, call)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolGetInventoryVocabulary:
		result, err := a.ExecuteRealtimeVoiceVocabularyTool(toolCtx, session, call)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolSearchAuthorizedAssets:
		result, err := a.ExecuteRealtimeVoiceSearchTool(toolCtx, session, call)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolGetAssetDetail:
		result, err := a.ExecuteRealtimeVoiceAssetDetailTool(toolCtx, session, call, visibleAssetIDs)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListAuthorizedAssets:
		result, err := a.ExecuteRealtimeVoiceListTool(toolCtx, session, call, visibleAssetIDs)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListAssetAuditHistory:
		result, err := a.ExecuteRealtimeVoiceAssetAuditHistoryTool(toolCtx, session, call, visibleAssetIDs)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListCheckedOutAssets:
		result, err := a.ExecuteRealtimeVoiceCheckedOutAssetsTool(toolCtx, session, call)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListAssetCheckoutHistory:
		result, err := a.ExecuteRealtimeVoiceAssetCheckoutHistoryTool(toolCtx, session, call, visibleAssetIDs)
		return result, RealtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	default:
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
}

func RealtimeVoiceToolDeadlineError(parentCtx context.Context, toolCtx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) && errors.Is(toolCtx.Err(), context.DeadlineExceeded) && parentCtx.Err() == nil {
		return ErrRealtimeVoiceToolCallTimedOut
	}
	return err
}

func (a RealtimeReadTools) ExecuteRealtimeVoiceSearchTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	args, err := tools.ParseRealtimeVoiceSearchArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	input := searchapp.SearchAssetsInput{
		Principal: session.Principal, TenantID: session.TenantID, InventoryIDs: []inventory.InventoryID{session.InventoryID}, Source: audit.SourceConversation, Query: args.Query, LifecycleState: args.LifecycleState, Limit: args.Limit,
	}
	var results RealtimeToolSearchResult
	for _, mode := range realtimeReadToolSearchModes(session) {
		input.Mode = mode.String()
		results, err = a.SearchAssets(ctx, input)
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		if len(results.Items) > 0 {
			break
		}
	}

	items := make([]RealtimeVoiceAssetToolItem, 0, len(results.Items))
	for _, result := range results.Items {
		item, err := a.RealtimeVoiceAssetToolItem(ctx, session, result.Asset, result.Inventory.Name.String(), RealtimeVoiceMatchFields(result.Matches), true)
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		names := make([]string, 0, len(result.AssignedTags))
		for _, tag := range result.AssignedTags {
			names = append(names, tag.DisplayName.String())
		}
		item.TagNames = agentmodel.BoundedObservationTagNames(names)
		items = append(items, item)
	}
	return RealtimeVoiceToolResult(call, RealtimeVoiceAssetToolOutput{
		Tool:    call.Name,
		Query:   args.Query,
		Count:   len(items),
		HasMore: results.HasMore,
		Items:   items,
	})
}

func (a RealtimeReadTools) ExecuteRealtimeVoiceListTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := tools.ParseRealtimeVoiceListArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if args.ParentAssetID != "" {
		if _, visible := visibleAssetIDs[args.ParentAssetID]; !visible {
			return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
		}
	}
	inventoryItem, err := a.GetInventory(ctx, inventoryapp.GetInventoryInput{
		Principal:   session.Principal,
		Source:      audit.SourceAPI,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}

	items := []RealtimeVoiceAssetToolItem{}
	hasMore := false
	cursor := ""
	for page := 0; page < 50 && len(items) < args.Limit; page++ {
		result, err := a.ListAssets(ctx, assetapp.ListAssetsInput{
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
			toolItem, err := a.RealtimeVoiceAssetToolItem(ctx, session, visibleAsset, inventoryItem.Name.String(), nil, true)
			if err != nil {
				return ports.AgentToolResult{}, err
			}
			if args.Kind != "" && toolItem.Kind != args.Kind.String() {
				continue
			}
			if args.ParentAssetID != "" && toolItem.ParentAssetID != args.ParentAssetID {
				continue
			}
			if args.ParentTitle != "" && !strings.EqualFold(toolItem.ParentTitle, args.ParentTitle) {
				continue
			}
			if args.LocationTitle != "" && !strings.EqualFold(toolItem.LocationTitle, args.LocationTitle) {
				continue
			}
			if args.ParentScope == tools.RealtimeVoiceParentScopeRoot && toolItem.ParentTitle != "" {
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
	return RealtimeVoiceToolResult(call, RealtimeVoiceAssetToolOutput{
		Tool:    call.Name,
		Count:   len(items),
		HasMore: hasMore,
		Filters: map[string]string{
			"kind":           args.Kind.String(),
			"lifecycleState": args.LifecycleState,
			"parentAssetId":  args.ParentAssetID,
			"parentTitle":    args.ParentTitle,
			"locationTitle":  args.LocationTitle,
			"parentScope":    args.ParentScope,
		},
		Items: items,
	})
}

func (a RealtimeReadTools) RealtimeVoiceAssetToolItem(ctx context.Context, session RealtimeReadToolScope, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (RealtimeVoiceAssetToolItem, error) {
	return a.RealtimeVoiceAssetToolItemWithCheckout(ctx, session, item, inventoryName, matchFields, includeAssetID, true)
}

func (a RealtimeReadTools) RealtimeVoiceAssetToolItemWithoutCheckoutLookup(ctx context.Context, session RealtimeReadToolScope, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (RealtimeVoiceAssetToolItem, error) {
	return a.RealtimeVoiceAssetToolItemWithCheckout(ctx, session, item, inventoryName, matchFields, includeAssetID, false)
}

func (a RealtimeReadTools) RealtimeVoiceAssetToolItemWithCheckout(ctx context.Context, session RealtimeReadToolScope, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool, includeCheckout bool) (RealtimeVoiceAssetToolItem, error) {
	ancestors, err := a.RealtimeVoiceAncestors(ctx, session, item)
	if err != nil {
		return RealtimeVoiceAssetToolItem{}, err
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

	toolItem := RealtimeVoiceAssetToolItem{
		Title:           item.Title.String(),
		Kind:            item.Kind.String(),
		Description:     item.Description.String(),
		InventoryName:   inventoryName,
		LifecycleState:  item.LifecycleState.String(),
		ParentAssetID:   item.ParentAssetID.String(),
		ParentTitle:     parentTitle,
		ParentKind:      parentKind,
		LocationTitle:   locationTitle,
		ContainmentPath: path,
		MatchFields:     matchFields,
	}
	toolItem.Expiration, err = a.RealtimeVoiceExpiration(ctx, session, item)
	if err != nil {
		return RealtimeVoiceAssetToolItem{}, err
	}
	if includeAssetID {
		toolItem.AssetID = item.ID.String()
	}
	if includeCheckout && a.checkouts != nil {
		checkout, found, err := a.checkouts.CurrentAssetCheckout(ctx, session.TenantID, session.InventoryID, item.ID)
		if err != nil {
			return RealtimeVoiceAssetToolItem{}, err
		}
		if found {
			toolItem.CurrentCheckout = &RealtimeVoiceCurrentCheckoutEntry{
				ID:                      checkout.ID.String(),
				CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
				CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
			}
		}
	}
	return toolItem, nil
}

func (a RealtimeReadTools) RealtimeVoiceAncestors(ctx context.Context, session RealtimeReadToolScope, item asset.Asset) ([]asset.Asset, error) {
	ancestors := []asset.Asset{}
	seen := map[asset.ID]struct{}{item.ID: {}}
	for parentID := item.ParentAssetID; parentID.String() != ""; {
		if _, duplicate := seen[parentID]; duplicate {
			return nil, ports.ErrInvalidProviderInput
		}
		seen[parentID] = struct{}{}
		parent, err := a.GetAsset(ctx, assetapp.GetAssetInput{
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

func RealtimeVoiceToolResult(call ports.AgentToolCall, output RealtimeVoiceAssetToolOutput) (ports.AgentToolResult, error) {
	if output.Count == 0 {
		output.Note = "No visible assets were returned within this request's scope; hasMore indicates incomplete coverage."
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

func RealtimeVoiceMatchFields(matches []search.Match) []string {
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

func StringArg(raw any) string {
	value, _ := raw.(string)
	return value
}

func RealtimeVoiceToolLimit(raw any) (int, error) { return tools.RealtimeVoiceToolLimit(raw) }

func RealtimeVoiceOptionalAssetKind(raw any) (asset.Kind, error) {
	return tools.RealtimeVoiceOptionalAssetKind(raw)
}

func RealtimeVoiceOptionalLifecycleState(raw any) (string, error) {
	return tools.RealtimeVoiceOptionalLifecycleState(raw)
}
