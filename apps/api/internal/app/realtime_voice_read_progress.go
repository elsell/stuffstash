package app

import (
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolGetAssetDetail:
		return realtimeVoiceGetAssetDetailPublicName
	case RealtimeVoiceToolListAuthorizedAssets:
		return realtimeVoiceListAuthorizedAssetsPublicName
	case RealtimeVoiceToolListAssetAuditHistory:
		return realtimeVoiceListAssetAuditHistoryPublicName
	case RealtimeVoiceToolListCheckedOutAssets:
		return realtimeVoiceListCheckedOutAssetsPublicName
	case RealtimeVoiceToolListAssetCheckoutHistory:
		return realtimeVoiceListCheckoutHistoryPublicName
	default:
		return realtimeVoiceSearchAuthorizedAssetsPublicName
	}
}

func realtimeVoiceToolCompletionStatus(result ports.AgentToolResult) string {
	switch result.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets, RealtimeVoiceToolListAuthorizedAssets:
	default:
		return "completed"
	}
	var output realtimeVoiceAssetToolOutput
	if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
		return "completed"
	}
	if output.Count == 0 {
		return "no_visible_match"
	}
	return "completed"
}

func emitRealtimeVoiceProgress(session RealtimeVoiceSession, status string, message string, emit RealtimeVoiceEventSink) error {
	safeMessage := safeRealtimeVoiceProgressMessage(message)
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAgentProgress, SessionID: session.ID, Status: safeRealtimeVoiceProgressStatus(status), Message: safeMessage})
}

func safeRealtimeVoiceProgressMessage(message string) string {
	if realtimeVoiceDiagnosticUnsafePhrasePattern.MatchString(message) {
		return "Working safely."
	}
	safeMessage := safeRealtimeVoiceDiagnosticText(message, 160)
	if safeMessage == "" {
		return "Working safely."
	}
	return safeMessage
}

func safeRealtimeVoiceProgressStatus(status string) string {
	switch status {
	case realtimeVoiceProgressUnderstanding,
		realtimeVoiceProgressExploring,
		realtimeVoiceProgressPlanning,
		realtimeVoiceProgressReviewing,
		realtimeVoiceProgressAnswering,
		realtimeVoiceProgressRecovering:
		return status
	default:
		return realtimeVoiceProgressRecovering
	}
}
