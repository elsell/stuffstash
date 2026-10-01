package app

import (
	"context"
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoiceToolMaxResults = tools.RealtimeVoiceToolMaxResults

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceTool(ctx, realtimeReadScope(session), call, visibleAssetIDs)
}

func realtimeVoiceToolDeadlineError(parentCtx context.Context, toolCtx context.Context, err error) error {
	return agentapp.RealtimeVoiceToolDeadlineError(parentCtx, toolCtx, err)
}

func (a App) executeRealtimeVoiceSearchTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceSearchTool(ctx, realtimeReadScope(session), call)
}

func (a App) executeRealtimeVoiceListTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	return a.realtimeReadTools().ExecuteRealtimeVoiceListTool(ctx, realtimeReadScope(session), call, visibleAssetIDs)
}

func (a App) realtimeVoiceAssetToolItem(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (realtimeVoiceAssetToolItem, error) {
	return a.realtimeReadTools().RealtimeVoiceAssetToolItem(ctx, realtimeReadScope(session), item, inventoryName, matchFields, includeAssetID)
}

func (a App) realtimeVoiceAssetToolItemWithoutCheckoutLookup(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool) (realtimeVoiceAssetToolItem, error) {
	return a.realtimeReadTools().RealtimeVoiceAssetToolItemWithoutCheckoutLookup(ctx, realtimeReadScope(session), item, inventoryName, matchFields, includeAssetID)
}

func (a App) realtimeVoiceAssetToolItemWithCheckout(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string, matchFields []string, includeAssetID bool, includeCheckout bool) (realtimeVoiceAssetToolItem, error) {
	return a.realtimeReadTools().RealtimeVoiceAssetToolItemWithCheckout(ctx, realtimeReadScope(session), item, inventoryName, matchFields, includeAssetID, includeCheckout)
}

func (a App) realtimeVoiceAncestors(ctx context.Context, session RealtimeVoiceSession, item asset.Asset) ([]asset.Asset, error) {
	return a.realtimeReadTools().RealtimeVoiceAncestors(ctx, realtimeReadScope(session), item)
}

func realtimeVoiceToolResult(call ports.AgentToolCall, output realtimeVoiceAssetToolOutput) (ports.AgentToolResult, error) {
	return agentapp.RealtimeVoiceToolResult(call, output)
}

func realtimeVoiceMatchFields(matches []search.Match) []string {
	return agentapp.RealtimeVoiceMatchFields(matches)
}

func stringArg(raw any) string {
	return agentapp.StringArg(raw)
}

func realtimeVoiceToolLimit(raw any) (int, error) {
	return agentapp.RealtimeVoiceToolLimit(raw)
}

func realtimeVoiceOptionalAssetKind(raw any) (asset.Kind, error) {
	return agentapp.RealtimeVoiceOptionalAssetKind(raw)
}

func realtimeVoiceOptionalLifecycleState(raw any) (string, error) {
	return agentapp.RealtimeVoiceOptionalLifecycleState(raw)
}
