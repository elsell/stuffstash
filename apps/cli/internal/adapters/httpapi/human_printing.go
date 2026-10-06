package httpapi

import (
	"context"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrintDefaults(ctx context.Context, s ports.Scope) (ports.InventoryPrintDefaults, error) {
	r, err := c.PrintSettings(ctx, s)
	if err != nil {
		return ports.InventoryPrintDefaults{}, err
	}
	printer := ""
	if r.Data.DefaultPrinterID != nil {
		printer = *r.Data.DefaultPrinterID
	}
	return ports.InventoryPrintDefaults{PrinterID: printer, TemplateID: r.Data.Template.ID, TemplateVersion: r.Data.Template.Version, ShowReference: r.Data.Template.Options.ShowReference}, nil
}
func humanPrinter(p generated.Printer) ports.RegisteredPrinter {
	return ports.RegisteredPrinter{Media: printerMedia(p.Media), ReadinessReason: p.ReadinessReason, ReportedAt: p.ReportedAt, AdapterID: p.AdapterId, Revision: uint64(p.Revision), MediaName: p.Media.Name, MediaPreset: p.Media.PresetId, ID: p.Id, Name: p.Name, Readiness: p.Readiness, Retired: p.Retired, MediaFingerprint: p.MediaFingerprint}
}
func (c *Client) Printer(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.RegisteredPrinter], error) {
	return printerResult(c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterId(ctx, s.Tenant, s.Inventory, id, nil))
}
func (c *Client) RegisteredPrinter(ctx context.Context, s ports.Scope, id string) (ports.RegisteredPrinter, error) {
	r, err := c.Printer(ctx, s, id)
	return r.Data, err
}
func (c *Client) RegisteredPrinters(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.RegisteredPrinter], error) {
	r, err := read[printerEnvelope[[]ports.RegisteredPrinter]](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrinters(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdPrintersParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.RegisteredPrinter]{}, err
	}
	var values []ports.RegisteredPrinter
	if r.Data != nil {
		values = make([]ports.RegisteredPrinter, 0, len(r.Data))
	}
	for _, v := range r.Data {
		values = append(values, completePrinter(v))
	}
	return ports.Result[[]ports.RegisteredPrinter]{Data: values, Pagination: page(r.Meta), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func humanJob(j generated.PrintJob) ports.PrintJobSummary {
	r := ports.PrintJobSummary{Kind: j.Kind, MediaFingerprint: j.MediaFingerprint, RequestedBy: j.RequestedBy, UpdatedAt: j.UpdatedAt, ID: j.Id, PrinterID: j.PrinterId, Status: j.Status, Revision: uint64(j.Revision), Copies: int(j.Copies), CreatedAt: j.CreatedAt, Attempts: []ports.PrintAttemptSummary{}}
	if j.Resolution != nil {
		r.Resolution = &ports.PrintJobResolution{ReportedOutcome: j.Resolution.ReportedOutcome, ResolvedAt: j.Resolution.ResolvedAt, ResolvedBy: j.Resolution.ResolvedBy}
	}
	if j.Attempts.GetOrEmpty() == nil {
		r.Attempts = nil
	}
	if j.AssetId != nil {
		r.AssetID = *j.AssetId
	}
	if j.Predecessor != nil {
		r.Predecessor = *j.Predecessor
	}
	for _, a := range j.Attempts.GetOrEmpty() {
		r.Attempts = append(r.Attempts, ports.PrintAttemptSummary{ClaimedAt: a.ClaimedAt, LeaseExpiresAt: a.LeaseExpiresAt, IdleConfirmedAt: a.IdleConfirmedAt, SettledAt: a.SettledAt, StartedAt: a.StartedAt, ID: a.Id, ConnectorID: a.ConnectorId, Outcome: a.Outcome, Reason: a.Reason, CompletedCopies: int(a.CompletedCopies)})
	}
	return r
}
func humanJobResult(r generated.SuccessEnvelopePrintJob, err error) (ports.Result[ports.PrintJobSummary], error) {
	if err != nil {
		return ports.Result[ports.PrintJobSummary]{}, err
	}
	return ports.Result[ports.PrintJobSummary]{Data: humanJob(r.Data), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) PrintJobs(ctx context.Context, s ports.Scope, p ports.Page, printer string) (ports.Result[[]ports.PrintJobSummary], error) {
	r, err := read[generated.SuccessEnvelopeListPrintJob](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintJobs(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdPrintJobsParams{Limit: &p.Limit, Cursor: &p.Cursor, PrinterId: &printer}))
	if err != nil {
		return ports.Result[[]ports.PrintJobSummary]{}, err
	}
	var values []ports.PrintJobSummary
	if r.Data.GetOrEmpty() != nil {
		values = make([]ports.PrintJobSummary, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		values = append(values, humanJob(v))
	}
	return ports.Result[[]ports.PrintJobSummary]{Data: values, Pagination: page(r.Meta), Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
func (c *Client) PrintJob(ctx context.Context, s ports.Scope, id string) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobId(ctx, s.Tenant, s.Inventory, id, nil)))
}
func (c *Client) CancelPrint(ctx context.Context, s ports.Scope, id string, revision uint64) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdCancellation(ctx, s.Tenant, s.Inventory, id, nil, generated.PrintJobRevision{Revision: int64(revision)})))
}

var _ ports.HumanPrintingAPI = (*Client)(nil)
