package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func checkout(v generated.AssetCheckoutResponse) ports.Checkout {
	return ports.Checkout{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, AssetID: v.AssetId, CheckedOutAt: v.CheckedOutAt, CheckedOutByPrincipalID: v.CheckedOutByPrincipalId, CheckoutDetails: v.CheckoutDetails, ReturnDetails: v.ReturnDetails, ReturnedAt: v.ReturnedAt, ReturnedByPrincipalID: v.ReturnedByPrincipalId, State: v.State, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, UndoableOperationID: v.UndoableOperationId}
}
func (c *Client) Checkouts(ctx context.Context, s ports.Scope, id string, p ports.Page) (ports.Result[[]ports.Checkout], error) {
	r, err := read[generated.SuccessEnvelopeListAssetCheckoutResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdCheckouts(ctx, s.Tenant, s.Inventory, id, &generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdCheckoutsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Checkout]{}, err
	}
	var items []ports.Checkout
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.Checkout, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, checkout(v))
	}
	return ports.Result[[]ports.Checkout]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) ChangeCheckout(ctx context.Context, s ports.Scope, id, checkoutID string, action ports.CheckoutAction, body []byte) (ports.Result[ports.Checkout], error) {
	var response *http.Response
	var err error
	switch action {
	case ports.CheckOut:
		response, err = c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdCheckoutWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body))
	case ports.Return:
		response, err = c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdReturnWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body))
	case ports.UpdateReturnDetails:
		response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdCheckoutsByCheckoutIdReturnDetailsWithBody(ctx, s.Tenant, s.Inventory, id, checkoutID, nil, "application/json", bytes.NewReader(body))
	default:
		return ports.Result[ports.Checkout]{}, ports.Failure("usage", "Unknown checkout action. Use --help to choose a command.")
	}
	r, err := read[generated.SuccessEnvelopeAssetCheckoutResponse](response, err)
	if err != nil {
		return ports.Result[ports.Checkout]{}, err
	}
	return ports.Result[ports.Checkout]{Data: checkout(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
