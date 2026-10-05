package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func printConnector(v generated.Connector) ports.PrintConnector {
	p := ports.PrintConnector{ID: v.Id, Name: v.Name, Generation: v.Generation, AuthorizationPending: v.AuthorizationPending, Availability: string(v.Availability), State: v.State, LastSeenAt: v.LastSeenAt, ReportReceivedAt: v.ReportReceivedAt, PrinterIDs: v.PrinterIds.GetOrEmpty()}
	if v.Report != nil {
		r := v.Report
		p.Report = &ports.PrintConnectorReport{Architecture: r.Architecture, Platform: r.Platform, Commit: r.Commit, Version: r.Version}
		if r.Adapters.GetOrEmpty() != nil {
			p.Report.Adapters = make([]ports.PrintConnectorAdapter, 0, len(r.Adapters.GetOrEmpty()))
		}
		for _, a := range r.Adapters.GetOrEmpty() {
			adapter := ports.PrintConnectorAdapter{ID: a.Id, CompletionEvidence: a.CompletionEvidence, ContractVersions: a.ContractVersions.GetOrEmpty(), Formats: a.Formats.GetOrEmpty(), Wake: a.Wake}
			if a.Media.GetOrEmpty() != nil {
				adapter.Media = make([]ports.PrintConnectorMedia, 0, len(a.Media.GetOrEmpty()))
			}
			for _, m := range a.Media.GetOrEmpty() {
				adapter.Media = append(adapter.Media, ports.PrintConnectorMedia{ID: m.Id, Version: m.Version})
			}
			p.Report.Adapters = append(p.Report.Adapters, adapter)
		}
	}
	return p
}
func (c *Client) PrintConnectors(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.PrintConnector], error) {
	r, err := read[generated.SuccessEnvelopeListConnector](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintConnectors(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdPrintConnectorsParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.PrintConnector]{}, err
	}
	var values []ports.PrintConnector
	if r.Data.GetOrEmpty() != nil {
		values = make([]ports.PrintConnector, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		values = append(values, printConnector(v))
	}
	return ports.Result[[]ports.PrintConnector]{Data: values, Schema: r.Schema, Meta: metadata(r.Meta), Pagination: page(r.Meta)}, nil
}
func (c *Client) PrintConnector(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.PrintConnector], error) {
	r, err := read[generated.SuccessEnvelopeConnector](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintConnectorsByConnectorId(ctx, s.Tenant, s.Inventory, id, nil))
	if err != nil {
		return ports.Result[ports.PrintConnector]{}, err
	}
	return ports.Result[ports.PrintConnector]{Data: printConnector(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
