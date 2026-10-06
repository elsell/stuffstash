package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrintConnectors(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.PrintConnector], error) {
	r, err := read[printerEnvelope[[]ports.PrintConnector]](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintConnectors(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdPrintConnectorsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.PrintConnector]{}, err
	}
	var values []ports.PrintConnector
	if r.Data != nil {
		values = make([]ports.PrintConnector, 0, len(r.Data))
	}
	for _, v := range r.Data {
		values = append(values, v)
	}
	return ports.Result[[]ports.PrintConnector]{Data: values, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) PrintConnector(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.PrintConnector], error) {
	r, err := read[printerEnvelope[ports.PrintConnector]](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintConnectorsByConnectorId(ctx, s.Tenant, s.Inventory, id, nil))
	if err != nil {
		return ports.Result[ports.PrintConnector]{}, err
	}
	return ports.Result[ports.PrintConnector]{Data: r.Data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
