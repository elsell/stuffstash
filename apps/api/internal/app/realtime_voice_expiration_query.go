package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const RealtimeVoiceToolQueryExpiringAssets = agentapp.RealtimeVoiceToolQueryExpiringAssets

func realtimeConversationExpirationTool() ports.ConversationToolDefinition {
	return agentapp.RealtimeConversationExpirationTool()
}

func (a App) executeRealtimeVoiceExpirationQuery(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceExpirationQuery(ctx, realtimeReadScope(session), call)
}
