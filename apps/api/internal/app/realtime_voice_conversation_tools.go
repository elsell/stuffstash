package app

import (
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) newRealtimeConversationTools(session RealtimeVoiceSession, emit RealtimeVoiceEventSink) *agentmodelapp.RealtimeConversationTools {
	return a.realtimeConversationService().NewTools(a.realtimeConversationSession(session), emit)
}
func realtimeConversationReadTools() []ports.ConversationToolDefinition {
	return agentmodelapp.RealtimeConversationReadTools()
}
