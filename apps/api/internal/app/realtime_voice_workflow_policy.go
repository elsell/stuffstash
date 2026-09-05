package app

import (
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"math"
)

func realtimeVoiceEvidenceRoundLimit(session RealtimeVoiceSession) int {
	if session.workflow == nil {
		return agentmodel.DefaultEvidenceRounds
	}
	return session.workflow.Revision().Snapshot().Definition.Settings().Budget.EvidenceRounds
}

func realtimeVoiceSearchModes(session RealtimeVoiceSession) []search.Mode {
	if session.workflow != nil && session.workflow.Revision().Snapshot().Definition.Settings().Retrieval == agentmodel.WorkflowRetrievalPreciseFirst {
		return []search.Mode{search.ModeExact, search.ModeFuzzy}
	}
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
	return session.workflow == nil || session.workflow.CanContinue()
}
