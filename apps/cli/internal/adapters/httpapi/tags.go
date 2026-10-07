package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func (c *Client) Tags(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.Tag], error) {
	r, err := read[generated.SuccessEnvelopeListAssetTagResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdTags(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdTagsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Tag]{}, err
	}
	var items []ports.Tag
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.Tag, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, tag(v))
	}
	return ports.Result[[]ports.Tag]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) ChangeTag(ctx context.Context, s ports.Scope, action ports.TagAction, id string, body []byte) (ports.Result[ports.Tag], error) {
	var response *http.Response
	var err error
	switch action {
	case ports.CreateTag:
		response, err = c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdTagsWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body))
	case ports.UpdateTag:
		response, err = c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdTagsByTagIdWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body))
	case ports.DeleteTag:
		response, err = c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdTagsByTagId(ctx, s.Tenant, s.Inventory, id, nil)
	default:
		return ports.Result[ports.Tag]{}, ports.Failure("usage", "Unknown tag action. Use --help to select a command.")
	}
	r, err := read[generated.SuccessEnvelopeAssetTagResponse](response, err)
	if err != nil {
		return ports.Result[ports.Tag]{}, err
	}
	return ports.Result[ports.Tag]{Data: tag(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func tag(v generated.AssetTagResponse) ports.Tag {
	return ports.Tag{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, Key: v.Key, DisplayName: v.DisplayName, Color: v.Color, Lifecycle: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
