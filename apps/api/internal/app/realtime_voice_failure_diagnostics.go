package app

import (
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func emitRealtimeVoiceConversationFailureDiagnostic(session RealtimeVoiceSession, modelCalls, toolCalls int, results []ports.AgentToolResult, err error, emit RealtimeVoiceEventSink) error {
	return agentmodelapp.EmitRealtimeConversationFailureDiagnostic(session.ID, session.DeveloperDiagnostics, modelCalls, toolCalls, results, err, emit)
}
func safeRealtimeVoiceProviderDiagnosticError(err error) string {
	return agentmodelapp.SafeRealtimeVoiceProviderDiagnosticError(err)
}
