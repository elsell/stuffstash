package agentmodel

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (a RealtimeReadTools) ExecuteRealtimeVoiceAssetDetailTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := tools.ParseRealtimeVoiceAssetDetailArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if _, visible := visibleAssetIDs[args.AssetID]; !visible {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	assetID, _ := asset.NewID(args.AssetID)
	detail, err := a.GetAssetDetail(ctx, assetapp.GetAssetInput{
		Principal:   session.Principal,
		Source:      audit.SourceConversation,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		AssetID:     assetID,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	inventoryItem, err := a.EnsureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	toolItem, err := a.RealtimeVoiceAssetToolItemWithoutCheckoutLookup(ctx, session, detail.Item, inventoryItem.Name.String(), nil, true)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	toolItem.CustomFields = detail.Item.CustomFields.Values()
	toolItem.CustomAssetTypeID = string(detail.Item.CustomAssetTypeID)
	if detail.CurrentCheckout != nil {
		toolItem.CheckoutState = &RealtimeVoiceCheckoutState{
			State:        detail.CurrentCheckout.State.String(),
			CheckedOut:   detail.CurrentCheckout.State == asset.CheckoutStateOpen,
			CheckedOutAt: detail.CurrentCheckout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
		}
	}
	return RealtimeVoiceToolResult(call, RealtimeVoiceAssetToolOutput{
		Tool:  call.Name,
		Count: 1,
		Items: []RealtimeVoiceAssetToolItem{toolItem},
	})
}
