package agentmodel

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (a RealtimeReadTools) ExecuteRealtimeVoiceCheckedOutAssetsTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	args, err := tools.ParseRealtimeVoiceCheckedOutAssetsArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	inventoryItem, err := a.EnsureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	result, err := a.ListCheckedOutAssets(ctx, assetapp.ListCheckedOutAssetsInput{
		Principal:   session.Principal,
		Source:      audit.SourceAPI,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		Limit:       args.Limit,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	items := make([]RealtimeVoiceAssetToolItem, 0, len(result.Items))
	for _, checkedOut := range result.Items {
		toolItem, err := a.RealtimeVoiceAssetToolItem(ctx, session, checkedOut.Asset, inventoryItem.Name.String(), nil, true)
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		toolItem.CurrentCheckout = &RealtimeVoiceCurrentCheckoutEntry{
			ID:                      checkedOut.Checkout.ID.String(),
			CheckedOutAt:            checkedOut.Checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
			CheckedOutByPrincipalID: checkedOut.Checkout.CheckedOutByPrincipal,
		}
		items = append(items, toolItem)
	}
	return RealtimeVoiceToolResult(call, RealtimeVoiceAssetToolOutput{
		Tool:    call.Name,
		Count:   len(items),
		HasMore: result.HasMore,
		Filters: map[string]string{
			"checkoutState": "checked_out",
		},
		Items: items,
	})
}

func (a RealtimeReadTools) ExecuteRealtimeVoiceAssetCheckoutHistoryTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := tools.ParseRealtimeVoiceAssetCheckoutHistoryArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if _, visible := visibleAssetIDs[args.AssetID]; !visible {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	assetID, _ := asset.NewID(args.AssetID)
	item, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, assetID)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if !found {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	inventoryItem, err := a.EnsureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	toolItem, err := a.RealtimeVoiceAssetToolItem(ctx, session, item, inventoryItem.Name.String(), nil, true)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	history, err := a.ListAssetCheckoutHistory(ctx, assetapp.ListAssetCheckoutHistoryInput{
		Principal:   session.Principal,
		Source:      audit.SourceAPI,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		AssetID:     assetID,
		Limit:       args.Limit,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	entries := make([]RealtimeVoiceAssetCheckoutHistoryEntry, 0, len(history.Items))
	for _, checkout := range history.Items {
		entries = append(entries, RealtimeVoiceAssetCheckoutHistoryEntry{
			ID:                      checkout.ID.String(),
			State:                   checkout.State.String(),
			CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
			CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
			CheckoutDetails:         checkout.CheckoutDetails.String(),
			ReturnedAt:              OptionalRealtimeVoiceCheckoutTime(checkout.ReturnedAt),
			ReturnedByPrincipalID:   checkout.ReturnedByPrincipal,
			ReturnDetails:           checkout.ReturnDetails.String(),
		})
	}
	note := ""
	if len(entries) == 0 {
		note = "No checkout history entries were returned for this visible asset."
	}
	payload, err := json.Marshal(RealtimeVoiceAssetCheckoutHistoryToolOutput{
		Tool:    call.Name,
		Asset:   toolItem,
		Order:   "newest_first",
		Count:   len(entries),
		HasMore: history.HasMore,
		Note:    note,
		Entries: entries,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	return ports.AgentToolResult{
		CallID:  call.ID,
		Name:    call.Name,
		Call:    call,
		Content: string(payload),
	}, nil
}

func OptionalRealtimeVoiceCheckoutTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
