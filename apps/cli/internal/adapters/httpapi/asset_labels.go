package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func labelResult(r generated.SuccessEnvelopeLabelResponse) ports.Result[ports.ResolvedLabel] {
	v := r.Data
	return ports.Result[ports.ResolvedLabel]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.ResolvedLabel{InstanceID: v.InstanceId, LabelID: v.LabelId, URL: v.Url, TenantID: v.TenantId, InventoryID: v.InventoryId, AssetID: v.AssetId, Lifecycle: v.LifecycleState}}
}
func assetLabelResult(response *http.Response, err error) (ports.Result[ports.ResolvedLabel], error) {
	r, err := read[generated.SuccessEnvelopeLabelResponse](response, err)
	if err != nil {
		return ports.Result[ports.ResolvedLabel]{}, err
	}
	return labelResult(r), nil
}
func (c *Client) AssetLabel(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.ResolvedLabel], error) {
	return assetLabelResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdLabel(ctx, s.Tenant, s.Inventory, id, nil))
}
func (c *Client) AssignLabel(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.ResolvedLabel], error) {
	return assetLabelResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdLabel(ctx, s.Tenant, s.Inventory, id, nil))
}
