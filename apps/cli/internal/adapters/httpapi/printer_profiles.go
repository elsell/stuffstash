package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrinterProfiles(ctx context.Context, s ports.Scope) (ports.Result[[]ports.PrinterProfile], error) {
	r, err := read[printerEnvelope[[]ports.PrinterProfile]](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrinterProfiles(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[[]ports.PrinterProfile]{}, err
	}
	return ports.Result[[]ports.PrinterProfile]{Data: r.Data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
