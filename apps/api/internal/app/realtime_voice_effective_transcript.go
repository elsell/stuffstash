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

func realtimeVoiceLooksLikeWriteRequest(transcript string) bool {
	if realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	if realtimeVoiceLooksLikeCasualAcquisitionCreateRequest(text) {
		return true
	}
	for _, token := range []string{" add ", " archive ", " check in ", " check out ", " create ", " move ", " place ", " put ", " restore ", " return ", " stash ", " store ", " update "} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceLooksLikeReadQuestion(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, prefix := range []string{" where ", " what ", " when ", " who ", " which ", " do i ", " do we ", " did i ", " did we ", " can you find ", " find my ", " find the "} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return strings.Contains(text, " where is ") ||
		strings.Contains(text, " where are ") ||
		strings.Contains(text, " what is ") ||
		strings.Contains(text, " what are ")
}

func realtimeVoiceLooksLikeMoveRequest(transcript string) bool {
	if realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, token := range []string{" move ", " put ", " place ", " store ", " stash "} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceLooksLikeContentsQuestion(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	return strings.Contains(text, " what s in ") ||
		strings.Contains(text, " what is in ") ||
		strings.Contains(text, " what are in ") ||
		strings.Contains(text, " what do i have in ") ||
		strings.Contains(text, " what do we have in ") ||
		strings.Contains(text, " inside ")
}

func normalizedRealtimeVoiceVerbText(transcript string) string {
	text := strings.ToLower(transcript)
	for _, replacer := range []string{".", ",", "!", "?", ";", ":", "\"", "'", "(", ")", "[", "]", "{", "}"} {
		text = strings.ReplaceAll(text, replacer, " ")
	}
	return " " + strings.Join(strings.Fields(text), " ") + " "
}

func realtimeVoiceLooksLikeCasualAcquisitionCreateRequest(normalizedText string) bool {
	acquisitionIndex := -1
	for _, marker := range []string{" i got ", " we got ", " i bought ", " we bought ", " i picked up ", " we picked up "} {
		if index := strings.Index(normalizedText, marker); index >= 0 && (acquisitionIndex == -1 || index < acquisitionIndex) {
			acquisitionIndex = index
		}
	}
	if acquisitionIndex == -1 {
		return false
	}
	for _, marker := range []string{" put it in ", " put it into ", " put it inside ", " put it on ", " placed it in ", " placed it into ", " stored it in ", " stored it inside ", " stashed it in ", " stashed it inside "} {
		if index := strings.Index(normalizedText, marker); index > acquisitionIndex {
			return true
		}
	}
	return false
}
