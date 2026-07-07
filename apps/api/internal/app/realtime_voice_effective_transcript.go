package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func safeRealtimeVoiceConversationTurns(turns []ports.AgentConversationTurn) []ports.AgentConversationTurn {
	if len(turns) == 0 {
		return nil
	}
	const maxTurns = 6
	safe := make([]ports.AgentConversationTurn, 0, min(len(turns), maxTurns))
	start := 0
	if len(turns) > maxTurns {
		start = len(turns) - maxTurns
	}
	for _, turn := range turns[start:] {
		text := safeRealtimeVoiceDiagnosticText(turn.Text, 500)
		if text == "" {
			continue
		}
		role := turn.Role
		if role != ports.AgentConversationRoleUser && role != ports.AgentConversationRoleAssistant {
			continue
		}
		safe = append(safe, ports.AgentConversationTurn{
			Role: role,
			Kind: safeRealtimeVoiceDiagnosticText(turn.Kind, 80),
			Text: text,
		})
	}
	return safe
}

func realtimeVoiceEffectiveTranscript(current string, turns []ports.AgentConversationTurn) string {
	current = safeRealtimeVoiceDiagnosticText(strings.TrimSpace(current), 500)
	if current == "" {
		return ""
	}
	safeTurns := safeRealtimeVoiceConversationTurns(turns)
	if len(safeTurns) == 0 {
		return current
	}
	latestUserIntent := ""
	pendingUserIntent := ""
	for _, turn := range safeTurns {
		switch turn.Role {
		case ports.AgentConversationRoleUser:
			if realtimeVoiceLooksLikeConversationIntent(turn.Text) {
				pendingUserIntent = strings.TrimSpace(turn.Text)
			} else {
				pendingUserIntent = ""
			}
		case ports.AgentConversationRoleAssistant:
			if turn.Kind == string(ports.StructuredAgentResponseKindClarification) && pendingUserIntent != "" {
				latestUserIntent = pendingUserIntent
			}
		}
	}
	if latestUserIntent == "" {
		return current
	}
	return safeRealtimeVoiceDiagnosticText(latestUserIntent+" Follow-up answer: "+current, 1000)
}

func realtimeVoiceLooksLikeConversationIntent(transcript string) bool {
	return realtimeVoiceLooksLikeWriteRequest(transcript) ||
		realtimeVoiceLooksLikeReadQuestion(transcript) ||
		realtimeVoiceLooksLikeContentsQuestion(transcript)
}
