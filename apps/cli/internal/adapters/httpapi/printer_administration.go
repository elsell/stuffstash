package httpapi

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func printerResult(response *http.Response, err error) (ports.Result[ports.RegisteredPrinter], error) {
	r, err := read[printerEnvelope[ports.RegisteredPrinter]](response, err)
	if err != nil {
		return ports.Result[ports.RegisteredPrinter]{}, err
	}
	return ports.Result[ports.RegisteredPrinter]{Data: completePrinter(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) CreatePrinter(ctx context.Context, s ports.Scope, retryKey string, body []byte) (ports.Result[ports.RegisteredPrinter], error) {
	return printerResult(c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintersWithBody(ctx, s.Tenant, s.Inventory, &generated.PostTenantsByTenantIdInventoriesByInventoryIdPrintersParams{IdempotencyKey: retryKey}, "application/json", bytes.NewReader(body)))
}
func (c *Client) UpdatePrinter(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.RegisteredPrinter], error) {
	return printerResult(c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterIdWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
}
func (c *Client) UpdatePrintConnector(ctx context.Context, s ports.Scope, id string, body []byte) (ports.Result[ports.PrintConnector], error) {
	r, err := read[printerEnvelope[ports.PrintConnector]](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdPrintConnectorsByConnectorIdWithBody(ctx, s.Tenant, s.Inventory, id, nil, "application/json", bytes.NewReader(body)))
	if err != nil {
		return ports.Result[ports.PrintConnector]{}, err
	}
	return ports.Result[ports.PrintConnector]{Data: r.Data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) UpdatePrintSettings(ctx context.Context, s ports.Scope, body []byte) (ports.Result[ports.PrintSettings], error) {
	return printSettingsResult(c.sdk.PutTenantsByTenantIdInventoriesByInventoryIdPrintSettingsWithBody(ctx, s.Tenant, s.Inventory, nil, "application/json", bytes.NewReader(body)))
}
