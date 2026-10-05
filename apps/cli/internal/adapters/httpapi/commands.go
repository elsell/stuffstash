package httpapi

import (
	"bytes"
	"context"
	"github.com/oapi-codegen/nullable"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) CreateAsset(ctx context.Context, s ports.Scope, in ports.AssetInput, idempotency string) (ports.Result[ports.Asset], error) {
	if len(in.RequestBody) > 0 {
		return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(in.RequestBody), key(idempotency))))
	}
	body := generated.CreateAssetBody{Title: in.Title, Kind: generated.CreateAssetBodyKind(in.Kind)}
	if p := in.PrintLabel; p != nil {
		body.PrintLabel = &generated.AssetPrintSelection{PrinterId: p.PrinterID, ExpectedMediaFingerprint: p.ExpectedMediaFingerprint, TemplateId: p.TemplateID, TemplateVersion: int32(p.TemplateVersion), TemplateOptions: generated.AssetPrintTemplateOptions{ShowReference: p.ShowReference}, Copies: int64(p.Copies)}
	}
	if in.Parent != "" {
		body.ParentAssetId = &in.Parent
	}
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssets(ctx, s.Tenant, s.Inventory, nil, body, key(idempotency))))
}
func (c *Client) UpdateAsset(ctx context.Context, s ports.Scope, id string, in ports.AssetChange, idempotency string) (ports.Result[ports.Asset], error) {
	if idempotency != "" {
		return ports.Result[ports.Asset]{}, ports.Failure("usage", "Asset updates do not support retry keys. Remove --idempotency-key.")
	}
	if len(in.RequestBody) > 0 {
		return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(in.RequestBody))))
	}
	body := generated.UpdateAssetBody{Title: in.Title}
	if in.MoveToRoot {
		body.ParentAssetId = nullable.NewNullNullable[string]()
	} else if in.Parent != nil {
		body.ParentAssetId = nullable.NewNullableWithValue(*in.Parent)
	}
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetId(ctx, s.Tenant, s.Inventory, id, nil, body)))
}
func (c *Client) SetArchived(ctx context.Context, s ports.Scope, id string, archived bool, idempotency string) (ports.Result[ports.Asset], error) {
	if idempotency != "" {
		return ports.Result[ports.Asset]{}, ports.Failure("usage", "Asset lifecycle commands do not support retry keys. Remove --idempotency-key.")
	}
	if archived {
		return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdArchive(ctx, s.Tenant, s.Inventory, id, nil)))
	}
	return assetResult(read[generated.SuccessEnvelopeAssetResponse](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdRestore(ctx, s.Tenant, s.Inventory, id, nil)))
}

func (c *Client) DeleteAsset(ctx context.Context, s ports.Scope, id string) error {
	return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetId(ctx, s.Tenant, s.Inventory, id, nil))
}
