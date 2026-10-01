package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeRealtimeVoiceAssetAuditHistoryTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceAssetAuditHistoryTool(ctx, realtimeReadScope(session), call, visibleAssetIDs)
}

func (a App) realtimeVoiceAssetAuditHistoryEntry(ctx context.Context, session RealtimeVoiceSession, currentTitle string, record audit.Record) realtimeVoiceAssetAuditHistoryEntry {
	return a.realtimeReadTools().RealtimeVoiceAssetAuditHistoryEntry(ctx, realtimeReadScope(session), currentTitle, record)
}

func (a App) realtimeVoiceAuditParentTitle(ctx context.Context, session RealtimeVoiceSession, rawAssetID string) string {
	return a.realtimeReadTools().RealtimeVoiceAuditParentTitle(ctx, realtimeReadScope(session), rawAssetID)
}

func (a App) realtimeVoiceAuditHistoryAssetToolItem(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string) (realtimeVoiceAssetToolItem, error) {
	return a.realtimeReadTools().RealtimeVoiceAuditHistoryAssetToolItem(ctx, realtimeReadScope(session), item, inventoryName)
}

func (a App) realtimeVoiceAuditHistoryAncestors(ctx context.Context, session RealtimeVoiceSession, item asset.Asset) ([]asset.Asset, error) {
	return a.realtimeReadTools().RealtimeVoiceAuditHistoryAncestors(ctx, realtimeReadScope(session), item)
}

func realtimeVoiceAuditActor(session RealtimeVoiceSession, record audit.Record) string {
	return agentapp.RealtimeVoiceAuditActor(realtimeReadScope(session), record)
}

func realtimeVoiceAssetAuditHistorySummary(title string, entry realtimeVoiceAssetAuditHistoryEntry) string {
	return agentapp.RealtimeVoiceAssetAuditHistorySummary(title, entry)
}
