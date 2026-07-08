package app

import (
	"regexp"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

var realtimeVoiceUnsafeWordSeparator = regexp.MustCompile(`[^a-z0-9]+`)

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
	normalized := " " + strings.Join(strings.Fields(realtimeVoiceUnsafeWordSeparator.ReplaceAllString(strings.ToLower(text), " ")), " ") + " "
	hasDestructiveVerb := strings.Contains(normalized, " wipe ") ||
		strings.Contains(normalized, " delete ") ||
		strings.Contains(normalized, " erase ") ||
		strings.Contains(normalized, " remove ") ||
		strings.Contains(normalized, " clear ") ||
		strings.Contains(normalized, " empty ") ||
		strings.Contains(normalized, " reset ") ||
		strings.Contains(normalized, " purge ") ||
		strings.Contains(normalized, " forget ")
	hasBroadScope := strings.Contains(normalized, " everything ") ||
		strings.Contains(normalized, " all ") ||
		strings.Contains(normalized, " every ")
	hasSystemTarget := strings.Contains(normalized, " database ") ||
		strings.Contains(normalized, " inventory ") ||
		strings.Contains(normalized, " assets ") ||
		strings.Contains(normalized, " asset ") ||
		strings.Contains(normalized, " items ") ||
		strings.Contains(normalized, " item ") ||
		strings.Contains(normalized, " things ") ||
		strings.Contains(normalized, " stuff ") ||
		strings.Contains(normalized, " records ") ||
		strings.Contains(normalized, " record ") ||
		strings.Contains(normalized, " entries ") ||
		strings.Contains(normalized, " entry ") ||
		strings.Contains(normalized, " contents ") ||
		strings.Contains(normalized, " content ") ||
		strings.Contains(normalized, " everything ")
	if strings.Contains(normalized, " database ") && hasDestructiveVerb {
		return true
	}
	if strings.Contains(normalized, " inventory ") && strings.Contains(normalized, " wipe ") {
		return true
	}
	return hasDestructiveVerb && hasBroadScope && hasSystemTarget
}
