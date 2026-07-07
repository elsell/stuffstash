package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeRealtimeVoiceCheckedOutAssetsTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	args, err := parseRealtimeVoiceCheckedOutAssetsArgs(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	inventoryItem, err := a.ensureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	result, err := a.ListCheckedOutAssets(ctx, ListCheckedOutAssetsInput{
		Principal:   session.Principal,
		Source:      audit.SourceAPI,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		Limit:       args.Limit,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	items := make([]realtimeVoiceAssetToolItem, 0, len(result.Items))
	for _, checkedOut := range result.Items {
		toolItem, err := a.realtimeVoiceAssetToolItem(ctx, session, checkedOut.Asset, inventoryItem.Name.String(), nil, true)
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		toolItem.CurrentCheckout = &realtimeVoiceCurrentCheckoutEntry{
			ID:                      checkedOut.Checkout.ID.String(),
			CheckedOutAt:            checkedOut.Checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
			CheckedOutByPrincipalID: checkedOut.Checkout.CheckedOutByPrincipal,
		}
		items = append(items, toolItem)
	}
	return realtimeVoiceToolResult(call, realtimeVoiceAssetToolOutput{
		Tool:    call.Name,
		Count:   len(items),
		HasMore: result.HasMore,
		Filters: map[string]string{
			"checkoutState": "checked_out",
		},
		Items: items,
	})
}

func (a App) executeRealtimeVoiceAssetCheckoutHistoryTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := parseRealtimeVoiceAssetCheckoutHistoryArgs(call.Arguments)
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
	inventoryItem, err := a.ensureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	toolItem, err := a.realtimeVoiceAssetToolItem(ctx, session, item, inventoryItem.Name.String(), nil, true)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	history, err := a.ListAssetCheckoutHistory(ctx, ListAssetCheckoutHistoryInput{
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
	entries := make([]realtimeVoiceAssetCheckoutHistoryEntry, 0, len(history.Items))
	for _, checkout := range history.Items {
		entries = append(entries, realtimeVoiceAssetCheckoutHistoryEntry{
			ID:                      checkout.ID.String(),
			State:                   checkout.State.String(),
			CheckedOutAt:            checkout.CheckedOutAt.UTC().Format(time.RFC3339Nano),
			CheckedOutByPrincipalID: checkout.CheckedOutByPrincipal,
			CheckoutDetails:         checkout.CheckoutDetails.String(),
			ReturnedAt:              optionalRealtimeVoiceCheckoutTime(checkout.ReturnedAt),
			ReturnedByPrincipalID:   checkout.ReturnedByPrincipal,
			ReturnDetails:           checkout.ReturnDetails.String(),
		})
	}
	note := ""
	if len(entries) == 0 {
		note = "No checkout history entries were returned for this visible asset."
	}
	payload, err := json.Marshal(realtimeVoiceAssetCheckoutHistoryToolOutput{
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

func optionalRealtimeVoiceCheckoutTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
