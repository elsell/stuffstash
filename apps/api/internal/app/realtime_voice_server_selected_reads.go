package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceServerSelectedReadCallWithoutModel(transcript string, turn int, toolResults []ports.AgentToolResult, id string) (ports.AgentToolCall, string, bool) {
	if args, ok := realtimeVoiceContentsListArgs(transcript, toolResults); ok {
		return ports.AgentToolCall{ID: id, Name: RealtimeVoiceToolListAuthorizedAssets, Arguments: args}, "Server-selected contents list", true
	}
	if turn == 0 && realtimeVoiceLooksLikeContentsQuestion(transcript) {
		if query := realtimeVoiceContentsQuestionTargetQuery(transcript); query != "" {
			return ports.AgentToolCall{ID: id, Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": query}}, "Server-selected contents target read", true
		}
	}
	if turn == 0 && realtimeVoiceLooksLikeSimpleCreateOrAddRequest(transcript) {
		if query := realtimeVoiceSimpleCreateDestinationQuery(transcript); query != "" {
			return ports.AgentToolCall{ID: id, Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": query}}, "Server-selected destination read", true
		}
	}
	return ports.AgentToolCall{}, "", false
}

func realtimeVoiceLooksLikeSimpleCreateOrAddRequest(transcript string) bool {
	if realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	return strings.Contains(text, " add ") || strings.Contains(text, " create ")
}

func realtimeVoiceContentsQuestionTargetQuery(transcript string) string {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, marker := range []string{" in ", " inside "} {
		if index := strings.LastIndex(text, marker); index >= 0 {
			return normalizeRealtimeVoiceSourceText(text[index+len(marker):])
		}
	}
	return ""
}

func realtimeVoiceSimpleCreateDestinationQuery(transcript string) string {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, marker := range []string{" under ", " inside ", " within ", " beneath ", " below "} {
		if strings.Contains(text, marker) {
			return ""
		}
	}
	best := ""
	bestIndex := -1
	for _, marker := range []string{" into ", " inside ", " onto ", " to ", " in ", " on "} {
		if index := strings.LastIndex(text, marker); index > bestIndex {
			bestIndex = index
			best = text[index+len(marker):]
		}
	}
	return normalizeRealtimeVoiceSourceText(best)
}

func realtimeVoiceServerSelectedExplorationCall(transcript string, turn int, toolResults []ports.AgentToolResult, id string) (ports.AgentToolCall, string, bool) {
	if turn != 0 || !realtimeVoiceLooksLikeSpecificSingularLookup(transcript) || realtimeVoiceLooksLikePluralCategoryQuery(transcript) {
		return ports.AgentToolCall{}, "", false
	}
	query := realtimeVoiceSpecificLookupObjectQuery(transcript)
	if query == "" || realtimeVoiceSearchQueryAlreadyAttempted(query, toolResults) {
		return ports.AgentToolCall{}, "", false
	}
	for _, miss := range realtimeVoiceNoMatchQueries(toolResults) {
		if realtimeVoiceQueriesMeaningfullyOverlap(query, miss) {
			return ports.AgentToolCall{
				ID:        id,
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": query},
			}, "Server-selected narrow retry", true
		}
	}
	return ports.AgentToolCall{}, "", false
}

func realtimeVoiceLooksLikeSpecificSingularLookup(transcript string) bool {
	if !realtimeVoiceLooksLikeReadQuestion(transcript) || realtimeVoiceLooksLikeWriteRequest(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, pluralMarker := range []string{" where are ", " what are ", " do i have any ", " do we have any "} {
		if strings.Contains(text, pluralMarker) {
			return false
		}
	}
	for _, marker := range []string{" where is ", " where's ", " find my ", " find the ", " can you find ", " what is "} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func realtimeVoiceSpecificLookupObjectQuery(transcript string) string {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, marker := range []string{" where is ", " where's ", " can you find ", " find my ", " find the ", " what is "} {
		if index := strings.LastIndex(text, marker); index >= 0 {
			query := normalizeRealtimeVoiceSourceText(text[index+len(marker):])
			query = strings.TrimSpace(strings.TrimPrefix(query, "my "))
			query = strings.TrimSpace(strings.TrimPrefix(query, "the "))
			if followUpQuery := realtimeVoiceFollowUpAnswerLookupQuery(query); followUpQuery != "" {
				return followUpQuery
			}
			return query
		}
	}
	words := realtimeVoiceMeaningfulWords(transcript)
	filtered := make([]string, 0, len(words))
	for _, word := range words {
		switch word {
		case "where", "what", "find", "have", "does":
			continue
		default:
			filtered = append(filtered, word)
		}
	}
	return strings.Join(filtered, " ")
}

func realtimeVoiceFollowUpAnswerLookupQuery(query string) string {
	const marker = " follow-up answer "
	if !strings.Contains(query, marker) {
		return ""
	}
	parts := strings.SplitN(query, marker, 2)
	if len(parts) != 2 || !realtimeVoiceLooksLikeGenericLookupTarget(parts[0]) {
		return ""
	}
	answer := normalizeRealtimeVoiceSourceText(parts[1])
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "my "))
	answer = strings.TrimSpace(strings.TrimPrefix(answer, "the "))
	return answer
}

func realtimeVoiceLooksLikeGenericLookupTarget(value string) bool {
	switch normalizeRealtimeVoiceSourceText(value) {
	case "it", "that", "this", "one", "that one", "this one", "the one", "item", "thing":
		return true
	default:
		return false
	}
}

func realtimeVoiceLooksLikePluralCategoryQuery(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, marker := range []string{" where are ", " what are ", " any ", " all ", " list "} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	query := realtimeVoiceSpecificLookupObjectQuery(transcript)
	words := strings.Fields(query)
	if len(words) == 0 {
		return false
	}
	last := words[len(words)-1]
	return strings.HasSuffix(last, "s") && !strings.HasSuffix(last, "ss")
}

func realtimeVoiceSearchQueryAlreadyAttempted(query string, toolResults []ports.AgentToolResult) bool {
	query = normalizeRealtimeVoiceSourceText(query)
	for _, result := range toolResults {
		if result.Name != RealtimeVoiceToolSearchAuthorizedAssets {
			continue
		}
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			continue
		}
		if normalizeRealtimeVoiceSourceText(output.Query) == query {
			return true
		}
	}
	return false
}

func realtimeVoiceQueriesMeaningfullyOverlap(left string, right string) bool {
	leftWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(left) {
		leftWords[word] = struct{}{}
	}
	if len(leftWords) == 0 {
		return false
	}
	matches := 0
	for _, word := range realtimeVoiceMeaningfulWords(right) {
		if _, ok := leftWords[word]; ok {
			matches++
		}
	}
	return matches > 0
}

func (a App) executeRealtimeVoiceServerSelectedRead(ctx context.Context, session RealtimeVoiceSession, transcript string, call ports.AgentToolCall, diagnosticTitle string, toolResults *[]ports.AgentToolResult, toolCallIDs *[]string, executedToolCalls map[string]struct{}, visibleAssetIDs map[string]struct{}, emit RealtimeVoiceEventSink) (*RealtimeVoiceActionPlanProposal, error) {
	if session.DeveloperDiagnostics {
		if err := emitRealtimeVoiceDiagnostic(session.ID, diagnosticTitle, realtimeVoiceToolCallDiagnosticDetail(call), emit); err != nil {
			return nil, err
		}
	}
	signature, err := realtimeVoiceToolCallSignature(call)
	if err != nil {
		return nil, err
	}
	if _, duplicate := executedToolCalls[signature]; duplicate {
		return nil, nil
	}
	toolCallID := strings.TrimSpace(call.ID)
	if toolCallID == "" {
		toolCallID = a.newRealtimeVoiceID()
	}
	toolLabel := realtimeVoiceToolLabel(call.Name)
	*toolCallIDs = append(*toolCallIDs, toolCallID)
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallStarted, SessionID: session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Status: "searching"}); err != nil {
		return nil, err
	}
	executableCall := ports.AgentToolCall{ID: toolCallID, Name: call.Name, Arguments: call.Arguments}
	result, proposal, err := a.executeRealtimeVoiceTool(ctx, session, transcript, *toolResults, executableCall, visibleAssetIDs)
	if err != nil {
		if !recoverableRealtimeVoiceToolError(err) {
			_ = emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Code: "tool_failed", Message: "I could not check that safely."})
			return nil, err
		}
		if session.DeveloperDiagnostics {
			if diagnosticErr := emitRealtimeVoiceDiagnostic(session.ID, "Tool validation failed", safeRealtimeVoiceErrorDetail(err), emit); diagnosticErr != nil {
				return nil, diagnosticErr
			}
		}
		if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Code: "invalid_tool_request", Message: "I need a little more detail to do that safely."}); err != nil {
			return nil, err
		}
		result, err = realtimeVoiceToolErrorResult(executableCall, "invalid_tool_request", realtimeVoiceInvalidToolRequestRepairMessage(executableCall.Name), true)
		if err != nil {
			return nil, err
		}
		if session.DeveloperDiagnostics {
			if err := emitRealtimeVoiceDiagnostic(session.ID, "Tool result received", realtimeVoiceToolResultDiagnosticDetail(result), emit); err != nil {
				return nil, err
			}
		}
		executedToolCalls[signature] = struct{}{}
		*toolResults = append(*toolResults, result)
		return nil, nil
	}
	executedToolCalls[signature] = struct{}{}
	if call.Name == RealtimeVoiceToolSearchAuthorizedAssets || call.Name == RealtimeVoiceToolListAuthorizedAssets {
		if err := collectRealtimeVoiceVisibleAssetIDs(result, visibleAssetIDs); err != nil {
			return nil, err
		}
	}
	if session.DeveloperDiagnostics {
		if err := emitRealtimeVoiceDiagnostic(session.ID, "Tool result received", realtimeVoiceToolResultDiagnosticDetail(result), emit); err != nil {
			return nil, err
		}
	}
	*toolResults = append(*toolResults, result)
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallCompleted, SessionID: session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Status: "completed"}); err != nil {
		return nil, err
	}
	return proposal, nil
}

func emitRealtimeVoiceProgress(session RealtimeVoiceSession, status string, message string, emit RealtimeVoiceEventSink) error {
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAgentProgress, SessionID: session.ID, Status: status, Message: message})
}
