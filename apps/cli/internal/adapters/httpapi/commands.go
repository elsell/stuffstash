package httpapi

import (
	"context"
	"github.com/oapi-codegen/nullable"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) CreateAsset(ctx context.Context, s ports.Scope, in ports.AssetInput, idempotency string) (ports.Result[ports.Asset], error) {
	body := generated.CreateAssetBody{Title: in.Title, Kind: generated.CreateAssetBodyKind(in.Kind)}
	if in.Parent != "" {
		body.ParentAssetId = &in.Parent
	}
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssets(ctx, s.Tenant, s.Inventory, nil, body, key(idempotency))))
}
func (c *Client) UpdateAsset(ctx context.Context, s ports.Scope, id string, in ports.AssetChange, idempotency string) (ports.Result[ports.Asset], error) {
	body := generated.UpdateAssetBody{Title: in.Title}
	if in.MoveToRoot {
		body.ParentAssetId = nullable.NewNullNullable[string]()
	} else if in.Parent != nil {
		body.ParentAssetId = nullable.NewNullableWithValue(*in.Parent)
	}
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetId(ctx, s.Tenant, s.Inventory, id, nil, body, key(idempotency))))
}
func (c *Client) SetArchived(ctx context.Context, s ports.Scope, id string, archived bool, idempotency string) (ports.Result[ports.Asset], error) {
	if archived {
		return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdArchive(ctx, s.Tenant, s.Inventory, id, nil, key(idempotency))))
	}
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdRestore(ctx, s.Tenant, s.Inventory, id, nil, key(idempotency))))
}
