package agentmodel

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func EmitRealtimeConversationFailureDiagnostic(sessionID string, enabled bool, modelCalls, toolCalls int, results []ports.AgentToolResult, err error, emit RealtimeVoiceEventSink) error {
	if !enabled {
		return nil
	}
	allowed := map[string]bool{}
	for _, tool := range RealtimeConversationReadTools() {
		allowed[tool.Name] = true
	}
	names := []string{}
	for _, result := range results {
		if allowed[result.Name] {
			names = append(names, result.Name)
		}
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{
		"stage": "conversation", "safeCode": RealtimeVoiceFailureLanguageInference,
		"safeError":      SafeRealtimeVoiceProviderDiagnosticError(err),
		"modelCallCount": modelCalls, "toolCallCount": toolCalls, "toolResultCount": len(results), "toolNames": names,
	}, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}
	return EmitRealtimeVoiceDiagnostic(sessionID, "Language provider failed", string(payload), emit)
}
