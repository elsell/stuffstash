package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
)

func (c *Client) UploadArchive(ctx context.Context, tenant, key string, body io.Reader) (ports.Result[ports.ArchiveJob], error) {
	// Hide Close and seek/replay capabilities: ownership stays with the caller,
	// and net/http cannot replay a request containing this stream.
	stream := struct{ io.Reader }{body}
	return archiveJobResult(c.sdk.UploadArchiveRestoreWithBody(ctx, tenant, &generated.UploadArchiveRestoreParams{IdempotencyKey: key}, "application/zip", stream))
}
func (c *Client) DownloadArchive(ctx context.Context, s ports.Scope, id string) (ports.BinaryContent, error) {
	return binaryContent(c.sdk.DownloadInventoryArchive(ctx, s.Tenant, id, &generated.DownloadInventoryArchiveParams{InventoryId: archiveInventory(s)}))
}
