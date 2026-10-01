package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeRealtimeVoiceAssetDetailTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceAssetDetailTool(ctx, realtimeReadScope(session), call, visibleAssetIDs)
}
