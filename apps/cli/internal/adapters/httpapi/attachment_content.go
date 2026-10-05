package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) AttachmentContent(ctx context.Context, s ports.Scope, asset, id string) (ports.BinaryContent, error) {
	return binaryContent(c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdContent(ctx, s.Tenant, s.Inventory, asset, id, nil))
}
func (c *Client) AttachmentThumbnail(ctx context.Context, s ports.Scope, asset, id, variant string) (ports.BinaryContent, error) {
	p := &generated.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdThumbnailParams{}
	switch variant {
	case "":
	case "small", "medium", "large":
		v := generated.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdThumbnailParamsVariant(variant)
		p.Variant = &v
	default:
		return ports.BinaryContent{}, ports.Failure("usage", "Use thumbnail size small, medium, or large.")
	}
	return binaryContent(c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdThumbnail(ctx, s.Tenant, s.Inventory, asset, id, p))
}
