package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeConversationVocabularyTool() ports.ConversationToolDefinition {
	return agentapp.RealtimeConversationVocabularyTool()
}

func (a App) executeRealtimeVoiceVocabularyTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceVocabularyTool(ctx, realtimeReadScope(session), call)
}
