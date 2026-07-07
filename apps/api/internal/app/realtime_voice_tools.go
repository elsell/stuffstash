package app

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoiceToolMaxResults = 20

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, transcript string, priorResults []ports.AgentToolResult, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, *RealtimeVoiceActionPlanProposal, error) {
	switch call.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets:
		result, err := a.executeRealtimeVoiceSearchTool(ctx, session, call)
		return result, nil, err
	case RealtimeVoiceToolListAuthorizedAssets:
		result, err := a.executeRealtimeVoiceListTool(ctx, session, call)
		return result, nil, err
	case RealtimeVoiceToolListAssetAuditHistory:
		result, err := a.executeRealtimeVoiceAssetAuditHistoryTool(ctx, session, call, visibleAssetIDs)
		return result, nil, err
	case RealtimeVoiceToolProposeActionPlan:
		return a.executeRealtimeVoiceProposeActionPlanTool(ctx, session, transcript, priorResults, call, visibleAssetIDs)
	default:
		return ports.AgentToolResult{}, nil, ports.ErrInvalidProviderInput
	}
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

func (a App) validateRealtimeVoiceActionPlanTranscriptAlignment(ctx context.Context, session RealtimeVoiceSession, commands []ActionPlanCommandInput, transcript string) error {
	for _, command := range commands {
		record, err := realtimeVoiceActionPlanCommandRecord(command)
		if err != nil {
			return err
		}
		switch command.Kind {
		case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
			args, err := parseActionPlanCreateArguments(record)
			if err != nil {
				return ports.ErrInvalidProviderInput
			}
			if strings.TrimSpace(args.ParentAssetID) != "" {
				if err := a.validateRealtimeVoiceMentionedAssetID(ctx, session, args.ParentAssetID, transcript); err != nil {
					return err
				}
			}
			if strings.TrimSpace(args.ParentAssetID) == "" && strings.TrimSpace(args.ParentCommandID) == "" {
				if err := validateRealtimeVoiceRootCreate(command.Kind, args.Kind, transcript); err != nil {
					return err
				}
				if err := a.validateRealtimeVoiceNoDuplicateRootCreate(ctx, session, command.Kind, args.Title, args.Kind); err != nil {
					return err
				}
			}
		case actionplan.CommandKindMoveAsset:
			args, err := parseActionPlanMoveArguments(record)
			if err != nil {
				return ports.ErrInvalidProviderInput
			}
			if err := a.validateRealtimeVoiceMentionedAssetID(ctx, session, args.AssetID.String(), transcript); err != nil {
				return err
			}
			if strings.TrimSpace(args.ParentAssetID) != "" {
				if err := a.validateRealtimeVoiceMentionedAssetID(ctx, session, args.ParentAssetID, transcript); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateRealtimeVoiceRootCreate(commandKind actionplan.CommandKind, rawKind string, transcript string) error {
	if !realtimeVoiceTranscriptNamesDestination(transcript) || realtimeVoiceTranscriptAllowsRootDestination(transcript) {
		return nil
	}
	if commandKind != actionplan.CommandKindCreateAsset {
		return nil
	}
	kind := strings.TrimSpace(rawKind)
	if kind == "" {
		kind = asset.KindItem.String()
	}
	if kind == asset.KindItem.String() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func realtimeVoiceNoMatchQueries(results []ports.AgentToolResult) []string {
	queries := []string{}
	for _, result := range results {
		if result.Name != RealtimeVoiceToolSearchAuthorizedAssets {
			continue
		}
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			continue
		}
		if output.Count == 0 && strings.TrimSpace(output.Query) != "" {
			queries = append(queries, output.Query)
		}
	}
	return queries
}

func realtimeVoiceQueryLooksLikeDestinationSegment(query string, transcript string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	transcript = strings.ToLower(strings.TrimSpace(transcript))
	if query == "" || !strings.Contains(transcript, query) {
		return false
	}
	for _, word := range realtimeVoiceMeaningfulWords(query) {
		if realtimeVoiceDestinationSegmentWords[word] {
			return true
		}
	}
	return false
}

func realtimeVoiceMeaningfulWordsRepresented(query string, represented string) bool {
	represented = " " + strings.ToLower(represented) + " "
	for _, word := range realtimeVoiceMeaningfulWords(query) {
		if realtimeVoiceDestinationSegmentWords[word] && strings.Contains(represented, " "+word+" ") {
			return true
		}
	}
	return false
}

func realtimeVoiceTranscriptHasUnrepresentedDestinationSegment(transcript string, represented string) bool {
	representedWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(represented) {
		representedWords[word] = struct{}{}
	}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		if !realtimeVoiceDestinationSegmentWords[word] {
			continue
		}
		if _, exists := representedWords[word]; !exists {
			return true
		}
	}
	return false
}

func realtimeVoiceLikelyOuterDestinationQuery(transcript string, toolResults []ports.AgentToolResult) string {
	represented := strings.Builder{}
	for _, query := range realtimeVoiceNoMatchQueries(toolResults) {
		represented.WriteString(" ")
		represented.WriteString(query)
	}
	representedText := represented.String()
	for _, phrase := range []string{"living room", "kitchen", "garage", "office", "bedroom", "bathroom", "basement", "attic", "pantry"} {
		if strings.Contains(strings.ToLower(transcript), phrase) && !realtimeVoiceVisibleParentTitlePresent(toolResults, phrase) {
			return phrase
		}
	}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		if realtimeVoiceDestinationSegmentWords[word] && !realtimeVoiceMeaningfulWordsRepresented(word, representedText) {
			return word
		}
	}
	return ""
}

func realtimeVoiceVisibleParentTitlePresent(toolResults []ports.AgentToolResult, title string) bool {
	title = normalizeRealtimeVoiceSourceText(title)
	if title == "" {
		return false
	}
	for _, parent := range realtimeVoiceVisibleParentsInTranscript(toolResults, title) {
		if normalizeRealtimeVoiceSourceText(parent.Title) == title {
			return true
		}
	}
	return false
}

func realtimeVoiceMeaningfulWords(value string) []string {
	words := []string{}
	for _, word := range strings.Fields(strings.ToLower(value)) {
		word = strings.Trim(word, ".,!?;:'\"()[]{}")
		if len(word) < 3 || realtimeVoiceTranscriptStopWords[word] {
			continue
		}
		words = append(words, word)
	}
	return words
}

func (a App) validateRealtimeVoiceNoDuplicateRootCreate(ctx context.Context, session RealtimeVoiceSession, commandKind actionplan.CommandKind, title string, rawKind string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ports.ErrInvalidProviderInput
	}
	kind := strings.TrimSpace(rawKind)
	if commandKind == actionplan.CommandKindCreateLocation {
		kind = asset.KindLocation.String()
	}
	if kind != asset.KindLocation.String() && kind != asset.KindContainer.String() {
		return nil
	}
	results, err := a.SearchAssets(ctx, SearchAssetsInput{
		Principal:      session.Principal,
		TenantID:       session.TenantID,
		InventoryIDs:   []inventory.InventoryID{session.InventoryID},
		Query:          title,
		Mode:           "fuzzy",
		LifecycleState: "active",
		Limit:          5,
	})
	if err != nil {
		return err
	}
	for _, result := range results.Items {
		if strings.EqualFold(result.Asset.Title.String(), title) && result.Asset.Kind.String() == kind {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func realtimeVoiceActionPlanCommandRecord(command ActionPlanCommandInput) (ports.ActionPlanCommandRecord, error) {
	payload, err := json.Marshal(command.Arguments)
	if err != nil {
		return ports.ActionPlanCommandRecord{}, ports.ErrInvalidProviderInput
	}
	return ports.ActionPlanCommandRecord{Kind: command.Kind, ArgumentsJSON: payload}, nil
}

func (a App) validateRealtimeVoiceMentionedAssetID(ctx context.Context, session RealtimeVoiceSession, rawAssetID string, transcript string) error {
	id, ok := asset.NewID(strings.TrimSpace(rawAssetID))
	if !ok {
		return ports.ErrInvalidProviderInput
	}
	item, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, id)
	if err != nil {
		return err
	}
	if !found {
		return ports.ErrInvalidProviderInput
	}
	if !realtimeVoiceTitleMentionedInTranscript(item.Title.String(), transcript) {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func realtimeVoiceTitleMentionedInTranscript(title string, transcript string) bool {
	transcriptWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		transcriptWords[word] = struct{}{}
	}
	for _, word := range realtimeVoiceMeaningfulWords(title) {
		if _, ok := transcriptWords[word]; ok {
			return true
		}
		if _, ok := transcriptWords[word+"s"]; ok {
			return true
		}
	}
	return false
}

var realtimeVoiceTranscriptStopWords = map[string]bool{
	"the":    true,
	"and":    true,
	"for":    true,
	"with":   true,
	"under":  true,
	"inside": true,
}

var realtimeVoiceDestinationSegmentWords = map[string]bool{
	"box":      true,
	"cabinet":  true,
	"shelf":    true,
	"drawer":   true,
	"bin":      true,
	"basket":   true,
	"closet":   true,
	"counter":  true,
	"kitchen":  true,
	"living":   true,
	"room":     true,
	"garage":   true,
	"office":   true,
	"bedroom":  true,
	"bathroom": true,
	"basement": true,
	"attic":    true,
	"pantry":   true,
}

var realtimeVoiceContainerSegmentWords = map[string]bool{
	"box":     true,
	"cabinet": true,
	"shelf":   true,
	"drawer":  true,
	"bin":     true,
	"basket":  true,
	"closet":  true,
	"counter": true,
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
	if a.checkouts != nil {
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
