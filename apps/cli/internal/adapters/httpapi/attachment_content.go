package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func attachmentContent(response *http.Response, err error) (ports.AttachmentContent, error) {
	if err != nil {
		return ports.AttachmentContent{}, ports.Failure("network", "Could not download the file. Check your connection and try again.")
	}
	if response.StatusCode != http.StatusOK {
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			_, err = read[any](response, nil)
			return ports.AttachmentContent{}, err
		}
		response.Body.Close()
		return ports.AttachmentContent{}, ports.Failure("protocol", "The server did not return a complete file. Try the download again.")
	}
	return ports.AttachmentContent{Body: response.Body, ContentType: response.Header.Get("Content-Type"), ContentDisposition: response.Header.Get("Content-Disposition"), ContentLength: response.ContentLength}, nil
}
func (c *Client) AttachmentContent(ctx context.Context, s ports.Scope, asset, id string) (ports.AttachmentContent, error) {
	return attachmentContent(c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdContent(ctx, s.Tenant, s.Inventory, asset, id, nil))
}
func (c *Client) AttachmentThumbnail(ctx context.Context, s ports.Scope, asset, id, variant string) (ports.AttachmentContent, error) {
	p := &generated.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdThumbnailParams{}
	switch variant {
	case "":
	case "small", "medium", "large":
		v := generated.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdThumbnailParamsVariant(variant)
		p.Variant = &v
	default:
		return ports.AttachmentContent{}, ports.Failure("usage", "Use thumbnail size small, medium, or large.")
	}
	return attachmentContent(c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdAttachmentsByAttachmentIdThumbnail(ctx, s.Tenant, s.Inventory, asset, id, p))
}
