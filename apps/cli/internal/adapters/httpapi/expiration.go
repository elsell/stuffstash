package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Override nullable items to keep the outer decoder's exact number handling.
type expirationData struct {
	generated.ExpirationWorkspaceData
	Items []generated.ExpirationWorkspaceAsset `json:"items"`
}
type expirationResponse struct {
	generated.SuccessEnvelopeExpirationWorkspaceData
	Data expirationData `json:"data"`
}

func (c *Client) ExpirationAssets(ctx context.Context, s ports.Scope, q ports.ExpirationQuery) (ports.Result[ports.ExpirationWorkspace], error) {
	p := &generated.GetTenantsByTenantIdInventoriesByInventoryIdExpirationAssetsParams{Limit: &q.Page.Limit, Cursor: &q.Page.Cursor}
	if q.Mode != "" {
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdExpirationAssetsParamsMode(q.Mode)
		p.Mode = &v
	}
	if q.Kind != "" {
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdExpirationAssetsParamsKind(q.Kind)
		p.Kind = &v
	}
	if q.CheckoutState != "" {
		v := generated.GetTenantsByTenantIdInventoriesByInventoryIdExpirationAssetsParamsCheckoutState(q.CheckoutState)
		p.CheckoutState = &v
	}
	if q.Query != "" {
		p.Q = &q.Query
	}
	if q.TypeID != "" {
		p.CustomAssetTypeId = &q.TypeID
	}
	if q.LocationID != "" {
		p.LocationId = &q.LocationID
	}
	if q.FromDate != "" {
		p.FromDate = &q.FromDate
	}
	if q.ThroughDate != "" {
		p.ThroughDate = &q.ThroughDate
	}
	if q.TagIDs != nil {
		p.TagIds = &q.TagIDs
	}
	r, err := read[expirationResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdExpirationAssets(ctx, s.Tenant, s.Inventory, p))
	if err != nil {
		return ports.Result[ports.ExpirationWorkspace]{}, err
	}
	data := ports.ExpirationWorkspace{Timezone: r.Data.Timezone, Counts: ports.ExpirationCounts{All: r.Data.Counts.All, Expired: r.Data.Counts.Expired, Soon: r.Data.Counts.Soon}}
	if r.Data.Items != nil {
		data.Items = make([]ports.ExpirationItem, 0, len(r.Data.Items))
	}
	for _, v := range r.Data.Items {
		data.Items = append(data.Items, expirationItem(v))
	}
	return ports.Result[ports.ExpirationWorkspace]{Data: data, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func expirationItem(v generated.ExpirationWorkspaceAsset) ports.ExpirationItem {
	base := generated.AssetResponse{Id: v.Id, TenantId: v.TenantId, InventoryId: v.InventoryId, Title: v.Title, Kind: v.Kind, Description: v.Description, ParentAssetId: v.ParentAssetId, LifecycleState: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, CustomAssetTypeId: v.CustomAssetTypeId, CustomFields: v.CustomFields, Expiration: v.Expiration, ExpirationContext: v.ExpirationContext, Tags: v.Tags, CurrentCheckout: v.CurrentCheckout, PrimaryPhoto: v.PrimaryPhoto, PrintJobId: v.PrintJobId, UndoableOperationId: v.UndoableOperationId}
	item := ports.ExpirationItem{Asset: asset(base)}
	if path := v.AncestorPath.GetOrEmpty(); path != nil {
		item.AncestorPath = make([]ports.ExpirationAncestor, 0, len(path))
		for _, a := range path {
			item.AncestorPath = append(item.AncestorPath, ports.ExpirationAncestor{ID: a.Id, Title: a.Title})
		}
	}
	return item
}
