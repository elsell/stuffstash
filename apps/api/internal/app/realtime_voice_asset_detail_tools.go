package app

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeRealtimeVoiceAssetDetailTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := parseRealtimeVoiceAssetDetailArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if _, visible := visibleAssetIDs[args.AssetID]; !visible {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	assetID, _ := asset.NewID(args.AssetID)
	detail, err := a.GetAssetDetail(ctx, GetAssetInput{
		Principal:   session.Principal,
		Source:      audit.SourceAPI,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		AssetID:     assetID,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	inventoryItem, err := a.ensureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	toolItem, err := a.realtimeVoiceAssetToolItemWithoutCheckoutLookup(ctx, session, detail.Item, inventoryItem.Name.String(), nil, true)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if detail.CurrentCheckout != nil {
		toolItem.CheckoutState = &realtimeVoiceCheckoutState{
			State:        detail.CurrentCheckout.State.String(),
			CheckedOut:   detail.CurrentCheckout.State == asset.CheckoutStateOpen,
			CheckedOutAt: detail.CurrentCheckout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
		}
	}
	return realtimeVoiceToolResult(call, realtimeVoiceAssetToolOutput{
		Tool:  call.Name,
		Count: 1,
		Items: []realtimeVoiceAssetToolItem{toolItem},
	})
}
