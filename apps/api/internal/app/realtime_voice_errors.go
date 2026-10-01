package app

import (
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var errRealtimeVoiceToolCallTimedOut = agentmodelapp.ErrRealtimeVoiceToolCallTimedOut

type realtimeVoiceProviderStageError = agentmodelapp.RealtimeVoiceProviderStageError

func validateRealtimeVoiceFinalResponse(response ports.StructuredAgentResponse) error {
	return agentmodelapp.ValidateRealtimeVoiceFinalResponse(response)
}
func safeRealtimeVoiceFinalText(value string, limit int) bool {
	return agentmodelapp.SafeRealtimeVoiceFinalText(value, limit)
}
func realtimeVoiceErrorCode(err error) string { return agentmodelapp.RealtimeVoiceErrorCode(err) }
func safeRealtimeVoiceErrorDetail(err error) string {
	return agentmodelapp.SafeRealtimeVoiceErrorDetail(err)
}
