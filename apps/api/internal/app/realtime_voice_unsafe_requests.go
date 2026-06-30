package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceUnsafeUnsupportedTranscriptResponse(transcript string) (ports.StructuredAgentResponse, bool) {
	text := strings.ToLower(transcript)
	if realtimeVoiceMentionsProviderSecret(text) {
		return ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindUnsupportedAction,
			SpokenResponse:  "I cannot read or change provider API keys or credentials from voice.",
			DisplayResponse: "I cannot read or change provider API keys or credentials from voice.",
		}, true
	}
	if realtimeVoiceMentionsDestructiveSystemAction(text) {
		return ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindUnsupportedAction,
			SpokenResponse:  "I cannot wipe the database or delete everything from voice.",
			DisplayResponse: "I cannot wipe the database or delete everything from voice.",
		}, true
	}
	return ports.StructuredAgentResponse{}, false
}

func realtimeVoiceMentionsProviderSecret(text string) bool {
	hasProviderTarget := strings.Contains(text, "provider") ||
		strings.Contains(text, "profile") ||
		strings.Contains(text, "api key") ||
		strings.Contains(text, "apikey") ||
		strings.Contains(text, "credential") ||
		strings.Contains(text, "secret") ||
		strings.Contains(text, "token")
	hasSensitiveVerb := strings.Contains(text, "read") ||
		strings.Contains(text, "show") ||
		strings.Contains(text, "tell") ||
		strings.Contains(text, "delete") ||
		strings.Contains(text, "remove") ||
		strings.Contains(text, "change")
	return hasProviderTarget && hasSensitiveVerb
}

func realtimeVoiceMentionsDestructiveSystemAction(text string) bool {
	hasDestructiveVerb := strings.Contains(text, "wipe") ||
		strings.Contains(text, "delete everything") ||
		strings.Contains(text, "forget everything") ||
		strings.Contains(text, "remove everything")
	hasSystemTarget := strings.Contains(text, "database") ||
		strings.Contains(text, "inventory") ||
		strings.Contains(text, "everything")
	return hasDestructiveVerb && hasSystemTarget
}
