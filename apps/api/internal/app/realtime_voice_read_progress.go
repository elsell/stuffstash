package app

import (
	"encoding/json"

	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolGetExpirationCalendar:
		return "Check current date"
	case RealtimeVoiceToolQueryExpiringAssets:
		return "Check expiration dates"
	case RealtimeVoiceToolGetInventoryVocabulary:
		return "Read inventory vocabulary"
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

func emitRealtimeVoiceProgress(session RealtimeVoiceSession, status, message string, emit RealtimeVoiceEventSink) error {
	return agentmodelapp.EmitRealtimeVoiceProgress(session.ID, status, message, emit)
}
func safeRealtimeVoiceProgressMessage(message string) string {
	return agentmodelapp.SafeRealtimeVoiceProgressMessage(message)
}
func safeRealtimeVoiceProgressStatus(status string) string {
	return agentmodelapp.SafeRealtimeVoiceProgressStatus(status)
}
