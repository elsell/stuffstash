package agentmodel

func EmitRealtimeVoiceProgress(sessionID string, status string, message string, emit RealtimeVoiceEventSink) error {
	safeMessage := SafeRealtimeVoiceProgressMessage(message)
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAgentProgress, SessionID: sessionID, Status: SafeRealtimeVoiceProgressStatus(status), Message: safeMessage})
}

func SafeRealtimeVoiceProgressMessage(message string) string {
	if realtimeVoiceDiagnosticUnsafePhrasePattern.MatchString(message) {
		return "Working safely."
	}
	safeMessage := SafeRealtimeVoiceDiagnosticText(message, 160)
	if safeMessage == "" {
		return "Working safely."
	}
	return safeMessage
}

func SafeRealtimeVoiceProgressStatus(status string) string {
	switch status {
	case RealtimeVoiceProgressUnderstanding,
		RealtimeVoiceProgressExploring,
		RealtimeVoiceProgressPlanning,
		RealtimeVoiceProgressReviewing,
		RealtimeVoiceProgressAnswering,
		RealtimeVoiceProgressRecovering:
		return status
	default:
		return RealtimeVoiceProgressRecovering
	}
}
