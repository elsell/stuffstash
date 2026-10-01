package agentmodel

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func RealtimeVoiceToolLabel(name string) string {
	switch name {
	case RealtimeVoiceToolGetExpirationCalendar:
		return "Check current date"
	case RealtimeVoiceToolQueryExpiringAssets:
		return "Check expiration dates"
	case RealtimeVoiceToolGetInventoryVocabulary:
		return "Read inventory vocabulary"
	case RealtimeVoiceToolGetAssetDetail:
		return RealtimeVoiceGetAssetDetailPublicName
	case RealtimeVoiceToolListAuthorizedAssets:
		return RealtimeVoiceListAuthorizedAssetsPublicName
	case RealtimeVoiceToolListAssetAuditHistory:
		return RealtimeVoiceListAssetAuditHistoryPublicName
	case RealtimeVoiceToolListCheckedOutAssets:
		return RealtimeVoiceListCheckedOutAssetsPublicName
	case RealtimeVoiceToolListAssetCheckoutHistory:
		return RealtimeVoiceListCheckoutHistoryPublicName
	default:
		return RealtimeVoiceSearchAuthorizedAssetsPublicName
	}
}

func RealtimeVoiceToolCompletionStatus(result ports.AgentToolResult) string {
	switch result.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets, RealtimeVoiceToolListAuthorizedAssets:
	default:
		return "completed"
	}
	var output RealtimeVoiceAssetToolOutput
	if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
		return "completed"
	}
	if output.Count == 0 {
		return "no_visible_match"
	}
	return "completed"
}
