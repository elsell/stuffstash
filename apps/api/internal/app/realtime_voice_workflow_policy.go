package app

import agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"

func RealtimeVoiceSessionTurnLimit(session RealtimeVoiceSession) int {
	return agentmodelapp.RealtimeVoiceSessionTurnLimit(session.workflow)
}
func RealtimeVoiceCanContinue(session RealtimeVoiceSession) bool {
	return session.conversationModel != nil
}
