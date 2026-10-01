package app

import (
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
)

type realtimeVoiceSearchArgs = tools.RealtimeVoiceSearchArgs
type realtimeVoiceListArgs = tools.RealtimeVoiceListArgs
type realtimeVoiceAssetAuditHistoryArgs = tools.RealtimeVoiceAssetAuditHistoryArgs
type realtimeVoiceAssetDetailArgs = tools.RealtimeVoiceAssetDetailArgs
type realtimeVoiceCheckedOutAssetsArgs = tools.RealtimeVoiceCheckedOutAssetsArgs
type realtimeVoiceAssetCheckoutHistoryArgs = tools.RealtimeVoiceAssetCheckoutHistoryArgs

const (
	realtimeVoiceParentScopeAny  = tools.RealtimeVoiceParentScopeAny
	realtimeVoiceParentScopeRoot = tools.RealtimeVoiceParentScopeRoot
)

func parseRealtimeVoiceSearchArgs(args map[string]any) (realtimeVoiceSearchArgs, error) {
	return tools.ParseRealtimeVoiceSearchArgs(args)
}

func parseRealtimeVoiceListArgs(args map[string]any) (realtimeVoiceListArgs, error) {
	return tools.ParseRealtimeVoiceListArgs(args)
}

func parseRealtimeVoiceAssetAuditHistoryArgs(args map[string]any) (realtimeVoiceAssetAuditHistoryArgs, error) {
	return tools.ParseRealtimeVoiceAssetAuditHistoryArgs(args)
}

func parseRealtimeVoiceAssetDetailArgs(args map[string]any) (realtimeVoiceAssetDetailArgs, error) {
	return tools.ParseRealtimeVoiceAssetDetailArgs(args)
}

func parseRealtimeVoiceCheckedOutAssetsArgs(args map[string]any) (realtimeVoiceCheckedOutAssetsArgs, error) {
	return tools.ParseRealtimeVoiceCheckedOutAssetsArgs(args)
}

func parseRealtimeVoiceAssetCheckoutHistoryArgs(args map[string]any) (realtimeVoiceAssetCheckoutHistoryArgs, error) {
	return tools.ParseRealtimeVoiceAssetCheckoutHistoryArgs(args)
}

func rejectUnknownRealtimeVoiceArgs(args map[string]any, allowed ...string) error {
	return tools.RejectUnknownRealtimeVoiceArgs(args, allowed...)
}

func optionalRealtimeVoiceTitle(raw any) (string, error) {
	return tools.OptionalRealtimeVoiceTitle(raw)
}

func realtimeVoiceOptionalParentScope(raw any) (string, error) {
	return tools.RealtimeVoiceOptionalParentScope(raw)
}

type realtimeVoiceAssetToolOutput = agentapp.RealtimeVoiceAssetToolOutput
type realtimeVoiceAssetToolItem = agentapp.RealtimeVoiceAssetToolItem
type realtimeVoiceCurrentCheckoutEntry = agentapp.RealtimeVoiceCurrentCheckoutEntry
type realtimeVoiceCheckoutState = agentapp.RealtimeVoiceCheckoutState
type realtimeVoiceAssetAuditHistoryToolOutput = agentapp.RealtimeVoiceAssetAuditHistoryToolOutput
type realtimeVoiceAssetAuditHistoryEntry = agentapp.RealtimeVoiceAssetAuditHistoryEntry
type realtimeVoiceAssetCheckoutHistoryToolOutput = agentapp.RealtimeVoiceAssetCheckoutHistoryToolOutput
type realtimeVoiceAssetCheckoutHistoryEntry = agentapp.RealtimeVoiceAssetCheckoutHistoryEntry
