package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
)

func (c *Client) CreateAttachment(ctx context.Context, s ports.Scope, asset string, body io.Reader) (ports.Result[ports.Attachment], error) {
	return attachmentResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsWithBody(ctx, s.Tenant, s.Inventory, asset, nil, "application/json", body))
}
func (c *Client) StartAttachmentUpload(ctx context.Context, s ports.Scope, asset string, body io.Reader) (ports.Result[ports.DirectUpload], error) {
	r, err := read[generated.SuccessEnvelopeDirectUploadResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsDirectUploadsWithBody(ctx, s.Tenant, s.Inventory, asset, nil, "application/json", body))
	if err != nil {
		return ports.Result[ports.DirectUpload]{}, err
	}
	v := r.Data
	return ports.Result[ports.DirectUpload]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.DirectUpload{UploadID: v.UploadId, AttachmentID: v.AttachmentId, Method: v.Method, URL: v.Url, Headers: v.Headers, FormFields: v.FormFields, ExpiresAt: v.ExpiresAt}}, nil
}
func (c *Client) CompleteAttachmentUpload(ctx context.Context, s ports.Scope, asset, id string) (ports.Result[ports.Attachment], error) {
	return attachmentResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsDirectUploadsByUploadIdComplete(ctx, s.Tenant, s.Inventory, asset, id, nil))
}
