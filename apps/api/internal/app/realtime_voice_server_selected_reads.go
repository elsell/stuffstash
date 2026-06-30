package app

import (
	"context"
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
