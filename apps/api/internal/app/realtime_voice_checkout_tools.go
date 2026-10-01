package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (a App) executeRealtimeVoiceCheckedOutAssetsTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceCheckedOutAssetsTool(ctx, realtimeReadScope(session), call)
}

func (a App) executeRealtimeVoiceAssetCheckoutHistoryTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceAssetCheckoutHistoryTool(ctx, realtimeReadScope(session), call, visibleAssetIDs)
}

func optionalRealtimeVoiceCheckoutTime(value time.Time) string {
	return agentapp.OptionalRealtimeVoiceCheckoutTime(value)
}
