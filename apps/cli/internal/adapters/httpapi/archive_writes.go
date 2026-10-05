package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) CreateArchiveJob(ctx context.Context, tenant, key string, body []byte) (ports.Result[ports.ArchiveJob], error) {
	return archiveJobResult(c.sdk.PostTenantsByTenantIdArchiveJobsWithBody(ctx, tenant, &generated.PostTenantsByTenantIdArchiveJobsParams{IdempotencyKey: key}, "application/json", bytes.NewReader(body)))
}
func (c *Client) ApproveArchiveJob(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.ArchiveJob], error) {
	return archiveJobResult(c.sdk.PostTenantsByTenantIdArchiveJobsByJobIdApproveWithBody(ctx, s.Tenant, id, &generated.PostTenantsByTenantIdArchiveJobsByJobIdApproveParams{InventoryId: archiveInventory(s)}, "application/json", bytes.NewReader(body)))
}
