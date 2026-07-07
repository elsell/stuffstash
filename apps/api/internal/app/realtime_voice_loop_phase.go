package app

import (
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceShouldRequireReadTool(transcript string, turn int, toolResults []ports.AgentToolResult) bool {
	if turn == 0 {
		return true
	}
	if realtimeVoiceShouldRequireContentsList(transcript, toolResults) {
		return true
	}
	if realtimeVoiceShouldRequireCreateDestinationRead(transcript, turn, toolResults) {
		return true
	}
	if realtimeVoiceShouldRequireNestedCreateParentRead(transcript, turn, toolResults) {
		return true
	}
	if realtimeVoiceShouldRequireCheckoutHistory(transcript, toolResults) {
		return true
	}
	return realtimeVoiceLooksLikeMoveRequest(transcript) && turn == 1 && realtimeVoiceReadToolResultCount(toolResults) < 2
}

func realtimeVoiceReadToolsForTurn(transcript string, turn int, toolResults []ports.AgentToolResult) []ports.AgentToolDescriptor {
	if turn == 0 {
		return realtimeVoiceInitialReadToolDescriptors()
	}
	if realtimeVoiceShouldRequireContentsList(transcript, toolResults) {
		return []ports.AgentToolDescriptor{realtimeVoiceListAuthorizedAssetsToolDescriptor()}
	}
	if realtimeVoiceShouldRequireCreateDestinationRead(transcript, turn, toolResults) {
		return []ports.AgentToolDescriptor{realtimeVoiceSearchAuthorizedAssetsToolDescriptor()}
	}
	if realtimeVoiceShouldRequireNestedCreateParentRead(transcript, turn, toolResults) {
		return []ports.AgentToolDescriptor{realtimeVoiceSearchAuthorizedAssetsToolDescriptor()}
	}
	if realtimeVoiceShouldRequireCheckoutHistory(transcript, toolResults) {
		return []ports.AgentToolDescriptor{realtimeVoiceListAssetCheckoutHistoryToolDescriptor()}
	}
	return realtimeVoiceReadToolDescriptors()
}

func realtimeVoiceServerSelectedReadCall(transcript string, turn int, toolResults []ports.AgentToolResult, call ports.AgentToolCall) (ports.AgentToolCall, string) {
	if args, ok := realtimeVoiceContentsListArgs(transcript, toolResults); ok {
		call.Name = RealtimeVoiceToolListAuthorizedAssets
		call.Arguments = args
		return call, "Server-selected contents list"
	}
	if query := realtimeVoiceRequiredCreateDestinationQuery(transcript, turn, toolResults); query != "" {
		return realtimeVoiceSearchCallWithQuery(call, query), "Server-selected destination read"
	}
	if query := realtimeVoiceRequiredNestedCreateParentQuery(transcript, turn, toolResults); query != "" {
		return realtimeVoiceSearchCallWithQuery(call, query), "Server-selected parent read"
	}
	if args, ok := realtimeVoiceCheckoutHistoryArgs(transcript, toolResults); ok {
		call.Name = RealtimeVoiceToolListAssetCheckoutHistory
		call.Arguments = args
		return call, "Server-selected checkout history"
	}
	return call, ""
}

func realtimeVoiceShouldRequireContentsList(transcript string, toolResults []ports.AgentToolResult) bool {
	if realtimeVoiceLooksLikeWriteRequest(transcript) || !realtimeVoiceLooksLikeContentsQuestion(transcript) || realtimeVoiceHasListResult(toolResults) {
		return false
	}
	_, ok := realtimeVoiceContentsListArgs(transcript, toolResults)
	return ok
}

func realtimeVoiceLooksLikeContentsQuestion(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	return strings.Contains(text, " what s in ") ||
		strings.Contains(text, " what is in ") ||
		strings.Contains(text, " what are in ") ||
		strings.Contains(text, " what do i have in ") ||
		strings.Contains(text, " what do we have in ") ||
		strings.Contains(text, " inside ")
}

func realtimeVoiceContentsListArgs(transcript string, toolResults []ports.AgentToolResult) (map[string]any, bool) {
	if !realtimeVoiceLooksLikeContentsQuestion(transcript) {
		return nil, false
	}
	best := realtimeVoiceBestContentsTarget(transcript, realtimeVoiceVisibleReadItems(toolResults))
	if best.Title == "" {
		return nil, false
	}
	switch best.Kind {
	case "location":
		return map[string]any{"locationTitle": best.Title}, true
	case "container":
		return map[string]any{"parentTitle": best.Title}, true
	default:
		return nil, false
	}
}

func realtimeVoiceBestContentsTarget(transcript string, items []realtimeVoiceAssetToolItem) realtimeVoiceAssetToolItem {
	best := realtimeVoiceAssetToolItem{}
	bestScore := 0
	for _, item := range items {
		if !realtimeVoiceTitleMentionedInTranscript(item.Title, transcript) {
			continue
		}
		if item.Kind != "location" && item.Kind != "container" {
			continue
		}
		score := realtimeVoiceContentsTargetScore(transcript, item)
		if score > bestScore {
			best = item
			bestScore = score
		}
	}
	return best
}

func realtimeVoiceContentsTargetScore(transcript string, item realtimeVoiceAssetToolItem) int {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		return 0
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	normalizedTitle := strings.TrimSpace(normalizedRealtimeVoiceVerbText(title))
	score := 1
	if strings.Contains(text, " "+normalizedTitle+" ") {
		score += 100
	}
	matchedWords := 0
	for _, word := range strings.Fields(normalizedTitle) {
		if len(word) < 2 {
			continue
		}
		if strings.Contains(text, " "+word+" ") {
			matchedWords++
			score += 10
		}
	}
	if matchedWords == len(strings.Fields(normalizedTitle)) {
		score += 30
	}
	if item.Kind == "container" {
		score += 5
	}
	score += len(strings.Fields(normalizedTitle))
	return score
}

func realtimeVoiceShouldRequireCreateDestinationRead(transcript string, turn int, toolResults []ports.AgentToolResult) bool {
	return realtimeVoiceRequiredCreateDestinationQuery(transcript, turn, toolResults) != ""
}

func realtimeVoiceRequiredCreateDestinationQuery(transcript string, turn int, toolResults []ports.AgentToolResult) string {
	if turn != 1 || !realtimeVoiceLooksLikeCreateRequest(transcript) || realtimeVoiceReadToolResultCount(toolResults) >= 2 {
		return ""
	}
	if len(realtimeVoiceNoMatchQueries(toolResults)) == 0 {
		return ""
	}
	return realtimeVoiceLikelyOuterDestinationQuery(transcript, toolResults)
}

func realtimeVoiceLooksLikeCreateRequest(transcript string) bool {
	if realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, token := range []string{" add ", " create ", " put ", " place ", " store ", " stash "} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceShouldRequireNestedCreateParentRead(transcript string, turn int, toolResults []ports.AgentToolResult) bool {
	if turn != 1 || !realtimeVoiceLooksLikeWriteRequest(transcript) || realtimeVoiceReadToolResultCount(toolResults) >= 2 {
		return false
	}
	for _, query := range realtimeVoiceNoMatchQueries(toolResults) {
		if !realtimeVoiceQueryLooksLikeDestinationSegment(query, transcript) {
			continue
		}
		if realtimeVoiceTranscriptHasUnrepresentedDestinationSegment(transcript, query) {
			return true
		}
	}
	return false
}

func realtimeVoiceRequiredNestedCreateParentQuery(transcript string, turn int, toolResults []ports.AgentToolResult) string {
	if !realtimeVoiceShouldRequireNestedCreateParentRead(transcript, turn, toolResults) {
		return ""
	}
	return realtimeVoiceLikelyOuterDestinationQuery(transcript, toolResults)
}

func realtimeVoiceSearchCallWithQuery(call ports.AgentToolCall, query string) ports.AgentToolCall {
	if strings.TrimSpace(query) == "" {
		return call
	}
	call.Name = RealtimeVoiceToolSearchAuthorizedAssets
	call.Arguments = map[string]any{"query": query}
	return call
}

func realtimeVoiceShouldUseConstrainedPlanner(transcript string, turn int, toolResults []ports.AgentToolResult) bool {
	if !realtimeVoiceLooksLikeWriteRequest(transcript) || turn == 0 {
		return false
	}
	if realtimeVoiceShouldRequireCreateDestinationRead(transcript, turn, toolResults) {
		return false
	}
	if realtimeVoiceShouldRequireNestedCreateParentRead(transcript, turn, toolResults) {
		return false
	}
	if realtimeVoiceLooksLikeMoveRequest(transcript) && turn < 2 && realtimeVoiceReadToolResultCount(toolResults) < 2 {
		return false
	}
	return true
}

func realtimeVoiceShouldFinalizeReadOnlyAfterToolTurn(transcript string, toolResults []ports.AgentToolResult) bool {
	if realtimeVoiceLooksLikeWriteRequest(transcript) || realtimeVoiceReadToolResultCount(toolResults) == 0 {
		return false
	}
	if realtimeVoiceLooksLikeHistoryQuestion(transcript) && !realtimeVoiceHasAssetAuditHistoryResult(toolResults) {
		return false
	}
	if realtimeVoiceLooksLikeCheckoutHistoryQuestion(transcript) && !realtimeVoiceHasAssetCheckoutHistoryResult(toolResults) {
		return false
	}
	if realtimeVoiceShouldRequireContentsList(transcript, toolResults) {
		return false
	}
	normalized := strings.ToLower(transcript)
	if strings.Contains(normalized, " and ") && realtimeVoiceReadToolResultCount(toolResults) < 2 {
		return false
	}
	return true
}

func realtimeVoiceLooksLikeHistoryQuestion(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, token := range []string{" when ", " history ", " audit ", " log ", " moved ", " move ", " changed ", " created ", " updated ", " archived ", " restored ", " before "} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceHasAssetAuditHistoryResult(toolResults []ports.AgentToolResult) bool {
	for _, result := range toolResults {
		if result.Name == RealtimeVoiceToolListAssetAuditHistory {
			return true
		}
	}
	return false
}

func realtimeVoiceShouldRequireCheckoutHistory(transcript string, toolResults []ports.AgentToolResult) bool {
	if realtimeVoiceLooksLikeWriteRequest(transcript) || !realtimeVoiceLooksLikeCheckoutHistoryQuestion(transcript) || realtimeVoiceHasAssetCheckoutHistoryResult(toolResults) {
		return false
	}
	_, ok := realtimeVoiceCheckoutHistoryArgs(transcript, toolResults)
	return ok
}

func realtimeVoiceLooksLikeCheckoutHistoryQuestion(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	if !realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	for _, token := range []string{" checked out ", " check out ", " checkout ", " returned ", " return ", " checked in ", " check in ", " who has ", " who had ", " borrowed ", " borrow ", " loaned ", " lent "} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceCheckoutHistoryArgs(transcript string, toolResults []ports.AgentToolResult) (map[string]any, bool) {
	if !realtimeVoiceLooksLikeCheckoutHistoryQuestion(transcript) {
		return nil, false
	}
	items := realtimeVoiceVisibleReadItems(toolResults)
	if len(items) == 0 {
		return nil, false
	}
	best := realtimeVoiceBestMentionedVisibleItem(transcript, items)
	if strings.TrimSpace(best.AssetID) == "" {
		return nil, false
	}
	return map[string]any{"assetId": best.AssetID}, true
}

func realtimeVoiceBestMentionedVisibleItem(transcript string, items []realtimeVoiceAssetToolItem) realtimeVoiceAssetToolItem {
	if len(items) == 1 {
		return items[0]
	}
	for _, item := range items {
		if realtimeVoiceTitleMentionedInTranscript(item.Title, transcript) {
			return item
		}
	}
	return realtimeVoiceAssetToolItem{}
}

func realtimeVoiceHasAssetCheckoutHistoryResult(toolResults []ports.AgentToolResult) bool {
	for _, result := range toolResults {
		if result.Name == RealtimeVoiceToolListAssetCheckoutHistory && realtimeVoiceToolResultSucceeded(result) {
			return true
		}
	}
	return false
}

func realtimeVoiceReadToolResultCount(toolResults []ports.AgentToolResult) int {
	count := 0
	for _, result := range toolResults {
		if result.Name == RealtimeVoiceToolSearchAuthorizedAssets || result.Name == RealtimeVoiceToolListAuthorizedAssets {
			count++
		}
	}
	return count
}

func realtimeVoiceHasListResult(toolResults []ports.AgentToolResult) bool {
	for _, result := range toolResults {
		if result.Name == RealtimeVoiceToolListAuthorizedAssets {
			return true
		}
	}
	return false
}

func realtimeVoiceVisibleReadItems(toolResults []ports.AgentToolResult) []realtimeVoiceAssetToolItem {
	items := []realtimeVoiceAssetToolItem{}
	for _, result := range toolResults {
		if !realtimeVoiceToolReturnsVisibleAssetItems(result.Name) {
			continue
		}
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			continue
		}
		items = append(items, output.Items...)
	}
	return items
}

func realtimeVoiceToolReturnsVisibleAssetItems(name string) bool {
	return name == RealtimeVoiceToolSearchAuthorizedAssets ||
		name == RealtimeVoiceToolListAuthorizedAssets ||
		name == RealtimeVoiceToolListCheckedOutAssets
}

func realtimeVoiceToolResultSucceeded(result ports.AgentToolResult) bool {
	var envelope struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(result.Content), &envelope); err != nil {
		return false
	}
	return envelope.Status != "error"
}
