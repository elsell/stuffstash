package agentmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/app/audithistory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
	"time"
)

func (a RealtimeReadTools) ExecuteRealtimeVoiceAssetAuditHistoryTool(ctx context.Context, session RealtimeReadToolScope, call ports.AgentToolCall, visibleAssetIDs map[string]struct{}) (ports.AgentToolResult, error) {
	args, err := tools.ParseRealtimeVoiceAssetAuditHistoryArgs(call.Arguments)
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
	toolItem, err := a.RealtimeVoiceAuditHistoryAssetToolItem(ctx, session, item, inventoryItem.Name.String())
	if err != nil {
		return ports.AgentToolResult{}, err
	}

	history, err := a.ListAssetAuditHistory(ctx, audithistory.ListAssetAuditHistoryInput{
		Principal:   session.Principal,
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		AssetID:     args.AssetID,
		Limit:       args.Limit,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	entries := make([]RealtimeVoiceAssetAuditHistoryEntry, 0, len(history.Items))
	for _, record := range history.Items {
		entries = append(entries, a.RealtimeVoiceAssetAuditHistoryEntry(ctx, session, toolItem.Title, record))
	}
	note := ""
	if len(entries) == 0 {
		note = "No safe audit history entries were returned for this visible asset."
	}
	payload, err := json.Marshal(RealtimeVoiceAssetAuditHistoryToolOutput{
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

func (a RealtimeReadTools) RealtimeVoiceAssetAuditHistoryEntry(ctx context.Context, session RealtimeReadToolScope, currentTitle string, record audit.Record) RealtimeVoiceAssetAuditHistoryEntry {
	previousParentTitle := ""
	newParentTitle := ""
	if record.Action == audit.ActionAssetMoved {
		previousParentTitle = a.RealtimeVoiceAuditParentTitle(ctx, session, record.Metadata["previous_parent"])
		newParentTitle = a.RealtimeVoiceAuditParentTitle(ctx, session, record.Metadata["new_parent"])
	}
	entry := RealtimeVoiceAssetAuditHistoryEntry{
		Action:              record.Action.String(),
		Source:              record.Source.String(),
		OccurredAt:          record.OccurredAt.UTC().Format(time.RFC3339),
		Actor:               RealtimeVoiceAuditActor(session, record),
		TargetType:          record.TargetType.String(),
		AssetKind:           record.Metadata["asset_kind"],
		PreviousParentTitle: previousParentTitle,
		NewParentTitle:      newParentTitle,
		PreviousState:       record.Metadata["previous_state"],
		LifecycleState:      record.Metadata["lifecycle_state"],
	}
	entry.Summary = RealtimeVoiceAssetAuditHistorySummary(currentTitle, entry)
	return entry
}

func (a RealtimeReadTools) RealtimeVoiceAuditParentTitle(ctx context.Context, session RealtimeReadToolScope, rawAssetID string) string {
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

func (a RealtimeReadTools) RealtimeVoiceAuditHistoryAssetToolItem(ctx context.Context, session RealtimeReadToolScope, item asset.Asset, inventoryName string) (RealtimeVoiceAssetToolItem, error) {
	ancestors, err := a.RealtimeVoiceAuditHistoryAncestors(ctx, session, item)
	if err != nil {
		return RealtimeVoiceAssetToolItem{}, err
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
	return RealtimeVoiceAssetToolItem{
		AssetID:         item.ID.String(),
		Title:           item.Title.String(),
		Kind:            item.Kind.String(),
		Description:     item.Description.String(),
		InventoryName:   inventoryName,
		LifecycleState:  item.LifecycleState.String(),
		ParentAssetID:   item.ParentAssetID.String(),
		ParentTitle:     parentTitle,
		ParentKind:      parentKind,
		LocationTitle:   locationTitle,
		ContainmentPath: path,
	}, nil
}

func (a RealtimeReadTools) RealtimeVoiceAuditHistoryAncestors(ctx context.Context, session RealtimeReadToolScope, item asset.Asset) ([]asset.Asset, error) {
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

func RealtimeVoiceAuditActor(session RealtimeReadToolScope, record audit.Record) string {
	if record.PrincipalID.String() == "" {
		return ""
	}
	if record.PrincipalID.String() == session.Principal.ID.String() {
		return "you"
	}
	return "another authorized user"
}

func RealtimeVoiceAssetAuditHistorySummary(title string, entry RealtimeVoiceAssetAuditHistoryEntry) string {
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
