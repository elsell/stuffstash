package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Keep UseNumber active for custom fields nested in search result lists.
type searchResponse struct {
	generated.SuccessEnvelopeListAssetSearchResultResponse
	Data []generated.AssetSearchResultResponse `json:"data"`
}

func (c *Client) SearchAssets(ctx context.Context, s ports.Scope, q ports.SearchQuery) (ports.Result[[]ports.SearchResult], error) {
	p := &generated.GetTenantsByTenantIdSearchAssetsParams{Limit: &q.Page.Limit, Cursor: &q.Page.Cursor}
	if s.Inventory != "" {
		p.InventoryId = &s.Inventory
	}
	if q.Query != "" {
		p.Q = &q.Query
	}
	if q.TypeID != "" {
		p.CustomAssetTypeId = &q.TypeID
	}
	if q.TagIDs != nil {
		p.TagIds = &q.TagIDs
	}
	if q.Mode != "" {
		v := generated.GetTenantsByTenantIdSearchAssetsParamsMode(q.Mode)
		p.Mode = &v
	}
	if q.Lifecycle != "" {
		v := generated.GetTenantsByTenantIdSearchAssetsParamsLifecycleState(q.Lifecycle)
		p.LifecycleState = &v
	}
	if q.CheckoutState != "" {
		v := generated.GetTenantsByTenantIdSearchAssetsParamsCheckoutState(q.CheckoutState)
		p.CheckoutState = &v
	}
	r, err := read[searchResponse](c.sdk.GetTenantsByTenantIdSearchAssets(ctx, s.Tenant, p))
	if err != nil {
		return ports.Result[[]ports.SearchResult]{}, err
	}
	var items []ports.SearchResult
	if r.Data != nil {
		items = make([]ports.SearchResult, 0, len(r.Data))
	}
	for _, v := range r.Data {
		items = append(items, searchResult(v))
	}
	return ports.Result[[]ports.SearchResult]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func searchResult(v generated.AssetSearchResultResponse) ports.SearchResult {
	result := ports.SearchResult{Type: v.Type, TenantID: v.TenantId, Inventory: ports.SearchInventory{ID: v.Inventory.Id, Name: v.Inventory.Name}, Asset: searchAsset(v.Asset)}
	if matches := v.Matches.GetOrEmpty(); matches != nil {
		result.Matches = make([]ports.SearchMatch, 0, len(matches))
		for _, m := range matches {
			result.Matches = append(result.Matches, ports.SearchMatch{Field: m.Field, Value: m.Value})
		}
	}
	if path := v.AncestorPath.GetOrEmpty(); path != nil {
		result.AncestorPath = make([]ports.SearchAncestor, 0, len(path))
		for _, a := range path {
			result.AncestorPath = append(result.AncestorPath, ports.SearchAncestor{ID: a.Id, Title: a.Title})
		}
	}
	return result
}
