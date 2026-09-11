package app

import (
	"bytes"
	"context"
	"encoding/json"
	expirationapp "github.com/stuffstash/stuff-stash/internal/app/expiration"
	notificationapp "github.com/stuffstash/stuff-stash/internal/app/notifications"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const RealtimeVoiceToolQueryExpiringAssets = "query_expiring_assets"
const expirationQueryPageBudget = 10

func realtimeConversationExpirationTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{Name: RealtimeVoiceToolQueryExpiringAssets, Description: "Find assets by recorded expiration. upcoming uses personal type windows; expired is separate; all includes all recorded dates. Optional type ID and tag key match either category when both supplied. Delivery switches do not hide explicit results. Continue with nextCursor and identical filters while hasMore is true, including empty pages. Date bounds include the recorded last-valid day; month labels use month end for comparison. Do not claim no matches or complete coverage from a partial scan.", Parameters: json.RawMessage(`{"type":"object","properties":{"status":{"type":"string","enum":["upcoming","expired","all"]},"customAssetTypeId":{"type":"string"},"tagKey":{"type":"string"},"fromDate":{"type":"string"},"throughDate":{"type":"string"},"cursor":{"type":"string"}},"required":["status"],"additionalProperties":false}`)}
}
func (a App) executeRealtimeVoiceExpirationQuery(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	var args struct {
		expirationapp.Query
		Cursor string `json:"cursor"`
	}
	encoded, err := json.Marshal(call.Arguments)
	if err != nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&args) != nil || args.Query.Validate() != nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	scope := notificationapp.ScopeInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation}
	scopeJSON, _ := json.Marshal(scope.Scope())
	cursor, err := expirationapp.DecodeContinuation(args.Cursor, string(scopeJSON), args.Query)
	if err != nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	preferences, err := a.notificationService.GetPreferences(ctx, scope)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	if preferences.Settings.Validate() != nil || a.clock == nil {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}
	inventoryItem, err := a.ensureInventoryAccessItem(ctx, session.Principal, session.TenantID, session.InventoryID, ports.InventoryPermissionView)
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	now := a.clock.Now()
	items := []realtimeVoiceAssetToolItem{}
	hasMore := false
	for pageIndex := 0; pageIndex < expirationQueryPageBudget && len(items) < realtimeVoiceToolMaxResults; pageIndex++ {
		page, err := a.ListAssets(ctx, ListAssetsInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, Source: audit.SourceConversation, LifecycleState: "active", Limit: realtimeVoiceToolMaxResults - len(items), Cursor: cursor})
		if err != nil {
			return ports.AgentToolResult{}, err
		}
		tagKeys := map[asset.ID][]string{}
		if args.TagKey != "" {
			if a.assetTags == nil {
				return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
			}
			ids := make([]asset.ID, 0, len(page.Items))
			for _, item := range page.Items {
				ids = append(ids, item.ID)
			}
			tags, err := a.assetTags.AssetTagsByAssets(ctx, session.TenantID, session.InventoryID, ids)
			if err != nil {
				return ports.AgentToolResult{}, err
			}
			for id, values := range tags {
				for _, tag := range values {
					tagKeys[id] = append(tagKeys[id], tag.Key.String())
				}
			}
		}
		for _, item := range page.Items {
			matched, err := args.Query.Matches(item.Expiration, notification.AssetTypeID(item.CustomAssetTypeID), tagKeys[item.ID], preferences.Settings, now)
			if err != nil {
				return ports.AgentToolResult{}, err
			}
			if !matched {
				continue
			}
			value, err := a.realtimeVoiceAssetToolItem(ctx, session, item, inventoryItem.Name.String(), nil, true)
			if err != nil {
				return ports.AgentToolResult{}, err
			}
			if value.Expiration != nil {
				description, err := expirationapp.Describe(item.Expiration, notification.AssetTypeID(item.CustomAssetTypeID), value.Expiration.TrackingEnabled, preferences.Settings, now)
				if err != nil {
					return ports.AgentToolResult{}, err
				}
				value.Expiration.State = description.State
				value.Expiration.AdvanceDays = description.AdvanceDays
				value.Expiration.Timezone = description.Timezone
			}
			items = append(items, value)
		}
		hasMore = page.HasMore
		if !hasMore {
			cursor = ""
			break
		}
		if page.NextCursor == nil || *page.NextCursor == "" {
			return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
		}
		if *page.NextCursor == cursor {
			return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
		}
		cursor = *page.NextCursor
	}
	if err := a.ensureRealtimeVoiceAccess(ctx, session.Principal, session.TenantID, session.InventoryID); err != nil {
		return ports.AgentToolResult{}, err
	}
	return realtimeVoiceToolResult(call, realtimeVoiceAssetToolOutput{Tool: call.Name, Count: len(items), Items: items, HasMore: hasMore, NextCursor: expirationapp.EncodeContinuation(cursor, string(scopeJSON), args.Query)})
}
