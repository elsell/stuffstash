package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceCreationEvidenceGrounded(intent agentmodel.Intent, transcript string, turns []ports.AgentConversationTurn) bool {
	if intent.CreationMode != agentmodel.CreationModeAdditional {
		return true
	}
	normalize := func(text string) string { return strings.ToLower(strings.Join(strings.Fields(text), " ")) }
	quote := normalize(intent.CreationEvidence)
	if quote == "" {
		return false
	}
	if strings.Contains(normalize(transcript), quote) {
		return true
	}
	for _, turn := range safeRealtimeVoiceConversationTurns(turns) {
		if turn.Role == ports.AgentConversationRoleUser && strings.Contains(normalize(turn.Text), quote) {
			return true
		}
	}
	return false
}
