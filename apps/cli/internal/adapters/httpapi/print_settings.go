package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrintSettings(ctx context.Context, s ports.Scope) (ports.Result[ports.PrintSettings], error) {
	r, err := read[generated.SuccessEnvelopeInventoryPrintSettings](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintSettings(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[ports.PrintSettings]{}, err
	}
	v := r.Data
	var printer *string
	if !v.DefaultPrinterId.IsNull() && v.DefaultPrinterId.IsSpecified() {
		value := v.DefaultPrinterId.GetOrEmpty()
		printer = &value
	}
	return ports.Result[ports.PrintSettings]{Schema: r.Schema, Meta: metadata(r.Meta), Data: ports.PrintSettings{Schema: v.Schema, DefaultPrinterID: printer, PrintOnCreateDefault: v.PrintOnCreateDefault, Revision: v.Revision, Template: ports.PrintSettingsTemplate{ID: v.Template.Id, Version: v.Template.Version, Options: ports.PrintSettingsOptions{ShowReference: v.Template.Options.ShowReference}}}}, nil
}
