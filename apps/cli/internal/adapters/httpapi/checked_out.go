package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Bypass nullable's custom list decoder so custom-field numbers retain precision.
type checkedOutResponse struct {
	generated.SuccessEnvelopeListCheckedOutAssetResponse
	Data []generated.CheckedOutAssetResponse `json:"data"`
}

func (c *Client) CheckedOutAssets(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.CheckedOutAsset], error) {
	r, err := read[checkedOutResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdCheckedOutAssets(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdCheckedOutAssetsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.CheckedOutAsset]{}, err
	}
	var items []ports.CheckedOutAsset
	if r.Data != nil {
		items = make([]ports.CheckedOutAsset, 0, len(r.Data))
	}
	for _, v := range r.Data {
		items = append(items, ports.CheckedOutAsset{Asset: asset(v.Asset), Checkout: currentCheckout(v.Checkout)})
	}
	return ports.Result[[]ports.CheckedOutAsset]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func currentCheckout(v generated.CurrentCheckout) ports.CurrentCheckout {
	c := ports.CurrentCheckout{ID: v.Id, State: v.State, CheckedOutAt: v.CheckedOutAt, CheckedOutByPrincipalID: v.CheckedOutByPrincipalId}
	if v.CheckedOutByPrincipal != nil {
		c.CheckedOutByPrincipal = &ports.Principal{ID: v.CheckedOutByPrincipal.Id, Email: v.CheckedOutByPrincipal.Email}
	}
	return c
}
