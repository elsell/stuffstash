package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceAmbiguousDestinationTranscriptResponse(transcript string) (ports.StructuredAgentResponse, bool) {
	text := normalizedRealtimeVoiceVerbText(transcript)
	if !realtimeVoiceLooksLikeMoveRequest(transcript) {
		return ports.StructuredAgentResponse{}, false
	}
	if strings.Contains(text, " side yard ") || strings.Contains(text, " side room ") {
		return ports.StructuredAgentResponse{}, false
	}
	for _, phrase := range []string{" over there ", " to the side ", " to side ", " on the side ", " onto the side "} {
		if strings.Contains(text, phrase) {
			return ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I need to know where to move it before I can prepare that move.",
				DisplayResponse: "I need to know where to move it before I can prepare that move.",
			}, true
		}
	}
	return ports.StructuredAgentResponse{}, false
}
