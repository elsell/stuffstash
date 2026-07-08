package app

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoiceToolMaxResults = 20

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, transcript string, priorResults []ports.AgentToolResult, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	toolCtx, cancel := context.WithTimeout(ctx, a.realtimeVoiceToolCallTimeout)
	defer cancel()
	switch call.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets:
		result, err := a.executeRealtimeVoiceSearchTool(toolCtx, session, call)
		return result, nil, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolGetAssetDetail:
		result, err := a.executeRealtimeVoiceAssetDetailTool(toolCtx, session, call, visibleAssetIDs)
		return result, nil, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListAuthorizedAssets:
		result, err := a.executeRealtimeVoiceListTool(toolCtx, session, call)
		return result, nil, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListAssetAuditHistory:
		result, err := a.executeRealtimeVoiceAssetAuditHistoryTool(toolCtx, session, call, visibleAssetIDs)
		return result, nil, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListCheckedOutAssets:
		result, err := a.executeRealtimeVoiceCheckedOutAssetsTool(toolCtx, session, call)
		return result, nil, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolListAssetCheckoutHistory:
		result, err := a.executeRealtimeVoiceAssetCheckoutHistoryTool(toolCtx, session, call, visibleAssetIDs)
		return result, nil, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	case RealtimeVoiceToolProposeActionPlan:
		result, proposal, err := a.executeRealtimeVoiceProposeActionPlanTool(toolCtx, session, transcript, priorResults, call, visibleAssetIDs)
		return result, proposal, realtimeVoiceToolDeadlineError(ctx, toolCtx, err)
	default:
		return ports.AgentToolResult{}, nil, ports.ErrInvalidProviderInput
	}
}

func realtimeVoiceToolDeadlineError(parentCtx context.Context, toolCtx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) && errors.Is(toolCtx.Err(), context.DeadlineExceeded) && parentCtx.Err() == nil {
		return errRealtimeVoiceToolCallTimedOut
	}
	return err
}

func (a App) executeRealtimeVoiceProposeActionPlanTool(ctx context.Context, session RealtimeVoiceSession, transcript string, priorResults []ports.AgentToolResult, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	args, err := parseRealtimeVoiceActionPlanArgs(call.Arguments, transcript)
	if err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceActionPlanVisibleIDs(args.Commands, visibleAssetIDs); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := a.validateRealtimeVoiceActionPlanTranscriptAlignment(ctx, session, args.Commands, transcript); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceMoveRequestUsesVisibleSource(args.Commands, transcript, priorResults); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceMoveRequestDoesNotCreateMissingSource(args.Commands, transcript, priorResults); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceRootCreatesUseVisibleParents(args.Commands, transcript, priorResults); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceMissingDestinationSegmentsAccountedFor(args.Commands, transcript, priorResults); err != nil {
		return ports.AgentToolResult{}, nil, err
	}
	if err := validateRealtimeVoiceMissingDestinationHierarchy(args.Commands, transcript, priorResults); err != nil {
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

func realtimeVoiceActionPlanCommandRecord(command ActionPlanCommandInput) (ports.ActionPlanCommandRecord, error) {
	payload, err := json.Marshal(command.Arguments)
	if err != nil {
		return ports.ActionPlanCommandRecord{}, ports.ErrInvalidProviderInput
	}
	return ports.ActionPlanCommandRecord{Kind: command.Kind, ArgumentsJSON: payload}, nil
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
		Source:         audit.SourceConversation,
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
			if args.ParentScope == realtimeVoiceParentScopeRoot && toolItem.ParentTitle != "" {
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
			"parentScope":    args.ParentScope,
		},
		Items: items,
	})
}

func (a App) realtimeVoiceAssetToolItem(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (realtimeVoiceAssetToolItem, error) {
	return a.realtimeVoiceAssetToolItemWithCheckout(ctx, session, item, inventoryName, matchFields, includeAssetID, true)
}

func (a App) realtimeVoiceAssetToolItemWithoutCheckoutLookup(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (realtimeVoiceAssetToolItem, error) {
	return a.realtimeVoiceAssetToolItemWithCheckout(ctx, session, item, inventoryName, matchFields, includeAssetID, false)
}

func (a App) realtimeVoiceAssetToolItemWithCheckout(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool, includeCheckout bool) (realtimeVoiceAssetToolItem, error) {
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
	if includeCheckout && a.checkouts != nil {
		checkout, found, err := a.checkouts.CurrentAssetCheckout(ctx, session.TenantID, session.InventoryID, item.ID)
		if err != nil {
			return realtimeVoiceAssetToolItem{}, err
		}
		if found {
			toolItem.CurrentCheckout = &realtimeVoiceCurrentCheckoutEntry{
				ID:                      checkout.ID.String(),
				CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
				CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
			}
		}
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
