package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) executeRealtimeVoiceAssetAuditHistoryTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := parseRealtimeVoiceAssetAuditHistoryArgs(call.Arguments)
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
	toolItem, err := a.realtimeVoiceAuditHistoryAssetToolItem(ctx, session, item, inventoryItem.Name.String())
	if err != nil {
		return ports.AgentToolResult{}, err
	}

	history, err := a.ListAssetAuditHistory(ctx, ListAssetAuditHistoryInput{
		Principal:   session.Principal,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		AssetID:     args.AssetID,
		Limit:       args.Limit,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	entries := make([]realtimeVoiceAssetAuditHistoryEntry, 0, len(history.Items))
	for _, record := range history.Items {
		entries = append(entries, a.realtimeVoiceAssetAuditHistoryEntry(ctx, session, toolItem.Title, record))
	}
	note := ""
	if len(entries) == 0 {
		note = "No safe audit history entries were returned for this visible asset."
	}
	payload, err := json.Marshal(realtimeVoiceAssetAuditHistoryToolOutput{
		Tool:    call.Name,
		Asset:   toolItem,
		Order:   "newest_first",
		Count:   len(entries),
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

func (a App) realtimeVoiceAssetAuditHistoryEntry(ctx context.Context, session RealtimeVoiceSession, currentTitle string, record audit.Record) realtimeVoiceAssetAuditHistoryEntry {
	previousParentTitle := ""
	newParentTitle := ""
	if record.Action == audit.ActionAssetMoved {
		previousParentTitle = a.realtimeVoiceAuditParentTitle(ctx, session, record.Metadata["previous_parent"])
		newParentTitle = a.realtimeVoiceAuditParentTitle(ctx, session, record.Metadata["new_parent"])
	}
	entry := realtimeVoiceAssetAuditHistoryEntry{
		Action:              record.Action.String(),
		Source:              record.Source.String(),
		OccurredAt:          record.OccurredAt.UTC().Format(time.RFC3339),
		Actor:               realtimeVoiceAuditActor(session, record),
		TargetType:          record.TargetType.String(),
		AssetKind:           record.Metadata["asset_kind"],
		PreviousParentTitle: previousParentTitle,
		NewParentTitle:      newParentTitle,
		PreviousState:       record.Metadata["previous_state"],
		LifecycleState:      record.Metadata["lifecycle_state"],
	}
	entry.Summary = realtimeVoiceAssetAuditHistorySummary(currentTitle, entry)
	return entry
}

func (a App) realtimeVoiceAuditParentTitle(ctx context.Context, session RealtimeVoiceSession, rawAssetID string) string {
	rawAssetID = strings.TrimSpace(rawAssetID)
	if rawAssetID == "" {
		return "Inventory root"
	}
	parentID, ok := asset.NewID(rawAssetID)
	if !ok {
		return "Unknown or removed parent"
	}
	parent, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, parentID)
	if err != nil || !found {
		return "Unknown or removed parent"
	}
	return parent.Title.String()
}

func (a App) realtimeVoiceAuditHistoryAssetToolItem(ctx context.Context, session RealtimeVoiceSession, item asset.Asset, inventoryName string) (realtimeVoiceAssetToolItem, error) {
	ancestors, err := a.realtimeVoiceAuditHistoryAncestors(ctx, session, item)
	if err != nil {
		return realtimeVoiceAssetToolItem{}, err
	}
	path := make([]string, 0, len(ancestors)+1)
	locationTitle := ""
	for _, ancestor := range ancestors {
		path = append(path, ancestor.Title.String())
		if ancestor.Kind == asset.KindLocation {
			locationTitle = ancestor.Title.String()
		}
	}
	path = append(path, item.Title.String())
	parentTitle := ""
	parentKind := ""
	if len(ancestors) > 0 {
		parent := ancestors[len(ancestors)-1]
		parentTitle = parent.Title.String()
		parentKind = parent.Kind.String()
	}
	if item.Kind == asset.KindLocation {
		locationTitle = item.Title.String()
	}
	return realtimeVoiceAssetToolItem{
		Title:           item.Title.String(),
		Kind:            item.Kind.String(),
		Description:     item.Description.String(),
		InventoryName:   inventoryName,
		LifecycleState:  item.LifecycleState.String(),
		ParentTitle:     parentTitle,
		ParentKind:      parentKind,
		LocationTitle:   locationTitle,
		ContainmentPath: path,
	}, nil
}

func (a App) realtimeVoiceAuditHistoryAncestors(ctx context.Context, session RealtimeVoiceSession, item asset.Asset) ([]asset.Asset, error) {
	ancestors := []asset.Asset{}
	seen := map[asset.ID]struct{}{item.ID: {}}
	for parentID := item.ParentAssetID; parentID.String() != ""; {
		if _, duplicate := seen[parentID]; duplicate {
			return nil, ports.ErrInvalidProviderInput
		}
		seen[parentID] = struct{}{}
		parent, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, parentID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, ports.ErrInvalidProviderInput
		}
		ancestors = append([]asset.Asset{parent}, ancestors...)
		parentID = parent.ParentAssetID
	}
	return ancestors, nil
}

func realtimeVoiceAuditActor(session RealtimeVoiceSession, record audit.Record) string {
	if record.PrincipalID.String() == "" {
		return ""
	}
	if record.PrincipalID.String() == session.Principal.ID.String() {
		return "you"
	}
	return "another authorized user"
}

func realtimeVoiceAssetAuditHistorySummary(title string, entry realtimeVoiceAssetAuditHistoryEntry) string {
	switch entry.Action {
	case audit.ActionAssetMoved.String():
		return fmt.Sprintf("%s moved from %s to %s.", title, entry.PreviousParentTitle, entry.NewParentTitle)
	case audit.ActionAssetCreated.String():
		return fmt.Sprintf("%s was created.", title)
	case audit.ActionAssetUpdated.String():
		return fmt.Sprintf("%s was updated.", title)
	case audit.ActionAssetArchived.String():
		return fmt.Sprintf("%s was archived.", title)
	case audit.ActionAssetRestored.String():
		return fmt.Sprintf("%s was restored.", title)
	default:
		return fmt.Sprintf("%s changed with action %s.", title, entry.Action)
	}
}
