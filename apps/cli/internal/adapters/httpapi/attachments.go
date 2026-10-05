package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func attachment(v generated.AttachmentResponse) ports.Attachment {
	return ports.Attachment{ID: v.Id, TenantID: v.TenantId, InventoryID: v.InventoryId, AssetID: v.AssetId, FileName: v.FileName, ContentType: v.ContentType, SizeBytes: v.SizeBytes, SHA256: v.Sha256, CreatedAt: v.CreatedAt, Lifecycle: v.LifecycleState}
}
func attachmentResult(response *http.Response, err error) (ports.Result[ports.Attachment], error) {
	r, err := read[generated.SuccessEnvelopeAttachmentResponse](response, err)
	if err != nil {
		return ports.Result[ports.Attachment]{}, err
	}
	return ports.Result[ports.Attachment]{Data: attachment(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) Attachments(ctx context.Context, s ports.Scope, asset string, p ports.Page) (ports.Result[[]ports.Attachment], error) {
	r, err := read[generated.SuccessEnvelopeListAttachmentResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachments(ctx, s.Tenant, s.Inventory, asset, &generated.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.Attachment]{}, err
	}
	var items []ports.Attachment
	if r.Data.GetOrEmpty() != nil {
		items = make([]ports.Attachment, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		items = append(items, attachment(v))
	}
	return ports.Result[[]ports.Attachment]{Data: items, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) Attachment(ctx context.Context, s ports.Scope, asset, id string) (ports.Result[ports.Attachment], error) {
	return attachmentResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentId(ctx, s.Tenant, s.Inventory, asset, id, nil))
}
func (c *Client) ChangeAttachment(ctx context.Context, s ports.Scope, asset, id string, action ports.AttachmentAction) (ports.Result[ports.Attachment], error) {
	switch action {
	case ports.ArchiveAttachment:
		return attachmentResult(c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdArchive(ctx, s.Tenant, s.Inventory, asset, id, nil))
	case ports.RestoreAttachment:
		return attachmentResult(c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdRestore(ctx, s.Tenant, s.Inventory, asset, id, nil))
	default:
		return ports.Result[ports.Attachment]{}, ports.Failure("usage", "Unknown attachment action. Use archive or restore.")
	}
}
func (c *Client) DeleteAttachment(ctx context.Context, s ports.Scope, asset, id string) error {
	return noContent(c.sdk.DeleteTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentId(ctx, s.Tenant, s.Inventory, asset, id, nil))
}
