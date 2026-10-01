package mcpserver

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type arguments struct {
	TenantID    string `json:"tenantId,omitempty"`
	InventoryID string `json:"inventoryId,omitempty"`
	AssetID     string `json:"assetId,omitempty"`
	Query       string `json:"query,omitempty"`
	Mode        string `json:"mode,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
}

func (h *handler) call(ctx context.Context, principal identity.Principal, definition tools.Definition, input arguments) (output map[string]any, returned error) {
	started := h.options.Clock.Now()
	defer func() {
		if h.options.Observer != nil {
			outcome := "succeeded"
			if returned != nil {
				outcome = "failed"
			}
			h.options.Observer.Record(ctx, ports.Event{Name: ports.EventMCPToolCompleted, Message: "inventory agent read completed", Fields: map[string]string{"tool": string(definition.Name), "outcome": outcome, "principal_id": principal.ID.String(), "elapsed_ms": strconv.FormatInt(h.options.Clock.Now().Sub(started).Milliseconds(), 10)}})
		}
	}()
	if (definition.TenantScoped && strings.TrimSpace(input.TenantID) == "") || (definition.InventoryScoped && strings.TrimSpace(input.InventoryID) == "") || (definition.AssetScoped && strings.TrimSpace(input.AssetID) == "") {
		return nil, errors.New("Invalid tool arguments")
	}
	result, err := h.execute(ctx, principal, definition.Name, input)
	if err != nil {
		return nil, safeError(err)
	}
	if err = ctx.Err(); err != nil {
		return nil, safeError(err)
	}
	return result, nil
}
func (h *handler) execute(ctx context.Context, principal identity.Principal, name tools.Name, in arguments) (map[string]any, error) {
	tenantID := tenant.ID(strings.TrimSpace(in.TenantID))
	inventoryID := inventory.InventoryID(strings.TrimSpace(in.InventoryID))
	assetID := asset.ID(strings.TrimSpace(in.AssetID))
	switch name {
	case tools.ListTenants:
		result, err := h.application.ListMyTenants(ctx, app.ListMyTenantsInput{Principal: principal, Source: audit.SourceMCP, Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, err
		}
		items := []any{}
		for _, item := range result.Items {
			items = append(items, map[string]any{"id": item.Tenant.ID.String(), "name": item.Tenant.Name.String()})
		}
		return page(items, result.Limit, result.HasMore, result.NextCursor), nil
	case tools.ListInventories:
		result, err := h.application.ListInventories(ctx, app.ListInventoriesInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, err
		}
		items := []any{}
		for _, item := range result.Items {
			items = append(items, map[string]any{"id": item.ID.String(), "tenantId": item.TenantID.String(), "name": item.Name.String()})
		}
		return page(items, result.Limit, result.HasMore, result.NextCursor), nil
	case tools.GetAsset:
		result, err := h.application.GetAssetDetail(ctx, app.GetAssetInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, InventoryID: inventoryID, AssetID: assetID})
		if err != nil {
			return nil, err
		}
		value := assetValue(result.Item)
		value["tags"] = tagValues(result.Tags)
		if result.CurrentCheckout != nil {
			value["checkout"] = checkoutValue(*result.CurrentCheckout)
		}
		return map[string]any{"asset": value}, nil
	case tools.ListRootAssets, tools.ListLocationAssets:
		parent := ports.AssetParentFilter{Applied: true}
		if name == tools.ListLocationAssets {
			parent.ID = assetID
		}
		result, err := h.application.ListAssets(ctx, app.ListAssetsInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, InventoryID: inventoryID, Parent: parent, Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, err
		}
		items := []any{}
		for _, item := range result.Items {
			value := assetValue(item)
			value["tags"] = tagValues(result.Tags[item.ID])
			items = append(items, value)
		}
		return page(items, result.Limit, result.HasMore, result.NextCursor), nil
	case tools.SearchAssets:
		if _, err := h.application.GetInventory(ctx, app.GetInventoryInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, InventoryID: inventoryID}); err != nil {
			return nil, err
		}
		result, err := h.application.SearchAssets(ctx, app.SearchAssetsInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, InventoryIDs: []inventory.InventoryID{inventoryID}, Query: in.Query, Mode: in.Mode, Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, err
		}
		items := []any{}
		for _, item := range result.Items {
			value := assetValue(item.Asset)
			value["tags"] = tagValues(item.AssignedTags)
			items = append(items, value)
		}
		return page(items, result.Limit, result.HasMore, result.NextCursor), nil
	case tools.ListCheckedOutAssets:
		result, err := h.application.ListCheckedOutAssets(ctx, app.ListCheckedOutAssetsInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, InventoryID: inventoryID, Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, err
		}
		items := []any{}
		for _, item := range result.Items {
			value := assetValue(item.Asset)
			value["checkout"] = checkoutValue(item.Checkout)
			items = append(items, value)
		}
		return page(items, result.Limit, result.HasMore, result.NextCursor), nil
	case tools.ListAssetCheckoutHistory:
		result, err := h.application.ListAssetCheckoutHistory(ctx, app.ListAssetCheckoutHistoryInput{Principal: principal, Source: audit.SourceMCP, TenantID: tenantID, InventoryID: inventoryID, AssetID: assetID, Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, err
		}
		items := []any{}
		for _, item := range result.Items {
			items = append(items, checkoutValue(item))
		}
		return page(items, result.Limit, result.HasMore, result.NextCursor), nil
	}
	return nil, apperrors.ErrInvalidInput
}
func safeError(err error) error {
	if errors.Is(err, apperrors.ErrInvalidInput) {
		return errors.New("Invalid tool arguments")
	}
	if errors.Is(err, ports.ErrForbidden) || errors.Is(err, apperrors.ErrNotFound) {
		return errors.New("Resource unavailable")
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return errors.New("Request canceled")
	}
	return errors.New("Inventory read failed")
}
func page(items []any, limit int, hasMore bool, cursor *string) map[string]any {
	return map[string]any{"items": items, "pagination": map[string]any{"limit": limit, "hasMore": hasMore, "nextCursor": cursor}}
}
