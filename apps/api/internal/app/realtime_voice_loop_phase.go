package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceShouldRequireReadTool(transcript string, turn int, toolResults []ports.AgentToolResult) bool {
	if turn == 0 {
		return true
	}
	if realtimeVoiceShouldRequireNestedCreateParentRead(transcript, turn, toolResults) {
		return true
	}
	return realtimeVoiceLooksLikeMoveRequest(transcript) && turn == 1 && realtimeVoiceReadToolResultCount(toolResults) < 2
}

func realtimeVoiceReadToolsForTurn(transcript string, turn int, toolResults []ports.AgentToolResult) []ports.AgentToolDescriptor {
	if realtimeVoiceShouldRequireNestedCreateParentRead(transcript, turn, toolResults) {
		return []ports.AgentToolDescriptor{realtimeVoiceSearchAuthorizedAssetsToolDescriptor()}
	}
	return realtimeVoiceReadToolDescriptors()
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
	normalized := strings.ToLower(transcript)
	if strings.Contains(normalized, " and ") && realtimeVoiceReadToolResultCount(toolResults) < 2 {
		return false
	}
	return true
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
