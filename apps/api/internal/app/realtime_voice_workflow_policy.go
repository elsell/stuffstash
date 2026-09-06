package app

import (
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"math"
)

func realtimeVoiceSearchModes(RealtimeVoiceSession) []search.Mode {
	return []search.Mode{search.ModeFuzzy}
}

// RealtimeVoiceSessionTurnLimit counts the initial request plus permitted follow-ups.
func RealtimeVoiceSessionTurnLimit(session RealtimeVoiceSession) int {
	if session.workflow == nil {
		return 3
	}
	followUps := session.workflow.Revision().Snapshot().Definition.Settings().Budget.FollowUpTurns
	if followUps == math.MaxInt {
		return math.MaxInt
	}
	return followUps + 1
}

func RealtimeVoiceCanContinue(session RealtimeVoiceSession) bool {
	return session.conversationModel != nil
}
