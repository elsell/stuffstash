package app

import (
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceToolLabel(name string) string { return agentmodelapp.RealtimeVoiceToolLabel(name) }
func realtimeVoiceToolCompletionStatus(result ports.AgentToolResult) string {
	return agentmodelapp.RealtimeVoiceToolCompletionStatus(result)
}

func emitRealtimeVoiceProgress(session RealtimeVoiceSession, status, message string, emit RealtimeVoiceEventSink) error {
	return agentmodelapp.EmitRealtimeVoiceProgress(session.ID, status, message, emit)
}
func safeRealtimeVoiceProgressMessage(message string) string {
	return agentmodelapp.SafeRealtimeVoiceProgressMessage(message)
}
func safeRealtimeVoiceProgressStatus(status string) string {
	return agentmodelapp.SafeRealtimeVoiceProgressStatus(status)
}
