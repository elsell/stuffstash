package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const RealtimeVoiceToolGetExpirationCalendar = agentapp.RealtimeVoiceToolGetExpirationCalendar

func realtimeConversationExpirationCalendarTool() ports.ConversationToolDefinition {
	return agentapp.RealtimeConversationExpirationCalendarTool()
}

func (a App) executeRealtimeVoiceExpirationCalendar(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceExpirationCalendar(ctx, realtimeReadScope(session), call)
}
