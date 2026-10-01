package app

import (
	"encoding/json"

	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func emitRealtimeVoiceConversationFailureDiagnostic(session RealtimeVoiceSession, modelCalls, toolCalls int, results []ports.AgentToolResult, err error, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	allowed := map[string]bool{}
	for _, tool := range realtimeConversationReadTools() {
		allowed[tool.Name] = true
	}
	names := []string{}
	for _, result := range results {
		if allowed[result.Name] {
			names = append(names, result.Name)
		}
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{
		"stage": "conversation", "safeCode": realtimeVoiceFailureLanguageInference,
		"safeError":      safeRealtimeVoiceProviderDiagnosticError(err),
		"modelCallCount": modelCalls, "toolCallCount": toolCalls, "toolResultCount": len(results), "toolNames": names,
	}, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}
	return emitRealtimeVoiceDiagnostic(session.ID, "Language provider failed", string(payload), emit)
}

func safeRealtimeVoiceProviderDiagnosticError(err error) string {
	return agentmodelapp.SafeRealtimeVoiceProviderDiagnosticError(err)
}
