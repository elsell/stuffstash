package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"net/http"
)

func (c *Client) PrintSettings(ctx context.Context, s ports.Scope) (ports.Result[ports.PrintSettings], error) {
	return printSettingsResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintSettings(ctx, s.Tenant, s.Inventory, nil))
}
func printSettingsResult(response *http.Response, err error) (ports.Result[ports.PrintSettings], error) {
	r, err := read[printerEnvelope[ports.PrintSettings]](response, err)
	if err != nil {
		return ports.Result[ports.PrintSettings]{}, err
	}
	return ports.Result[ports.PrintSettings]{Data: r.Data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
