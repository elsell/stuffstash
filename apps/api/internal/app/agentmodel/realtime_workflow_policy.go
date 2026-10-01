package agentmodel

import "math"

// RealtimeVoiceSessionTurnLimit counts the initial request plus permitted follow-ups.
func RealtimeVoiceSessionTurnLimit(workflow *PreparedWorkflow) int {
	if workflow == nil {
		return 3
	}
	followUps := workflow.Revision().Snapshot().Definition.Settings().Budget.FollowUpTurns
	if followUps == math.MaxInt {
		return math.MaxInt
	}
	return followUps + 1
}
