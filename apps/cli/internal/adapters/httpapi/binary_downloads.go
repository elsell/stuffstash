package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) LabelContent(ctx context.Context, s ports.Scope, id string) (ports.BinaryContent, error) {
	return binaryContent(c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdLabelRendersByRenderIdContent(ctx, s.Tenant, s.Inventory, id, nil))
}
func (c *Client) ExportInventory(ctx context.Context, s ports.Scope, format string) (ports.BinaryContent, error) {
	v := generated.ListTenantsByTenantIdInventoriesByInventoryIdExportParamsFormat(format)
	return binaryContent(c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdExport(ctx, s.Tenant, s.Inventory, &generated.ListTenantsByTenantIdInventoriesByInventoryIdExportParams{Format: &v}))
}
