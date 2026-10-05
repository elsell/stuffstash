package httpapi

import (
	"context"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrintDefaults(ctx context.Context, s ports.Scope) (ports.InventoryPrintDefaults, error) {
	r, err := read[generated.SuccessEnvelopeInventoryPrintSettings](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintSettings(ctx, s.Tenant, s.Inventory, nil))
	return ports.InventoryPrintDefaults{PrinterID: r.Data.DefaultPrinterId.GetOrEmpty(), TemplateID: r.Data.Template.Id, TemplateVersion: uint32(r.Data.Template.Version), ShowReference: r.Data.Template.Options.ShowReference}, err
}
func humanPrinter(p generated.Printer) ports.RegisteredPrinter {
	return ports.RegisteredPrinter{AdapterID: p.AdapterId, Revision: uint64(p.Revision), MediaName: p.Media.Name, MediaPreset: p.Media.PresetId, ID: p.Id, Name: p.Name, Readiness: p.Readiness, Retired: p.Retired, MediaFingerprint: p.MediaFingerprint}
}
func (c *Client) RegisteredPrinter(ctx context.Context, s ports.Scope, id string) (ports.RegisteredPrinter, error) {
	r, err := read[generated.SuccessEnvelopePrinter](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterId(ctx, s.Tenant, s.Inventory, id, nil))
	return humanPrinter(r.Data), err
}
func (c *Client) RegisteredPrinters(ctx context.Context, s ports.Scope, p ports.Page) (ports.Result[[]ports.RegisteredPrinter], error) {
	r, err := read[generated.SuccessEnvelopeListPrinter](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrinters(ctx, s.Tenant, s.Inventory, &generated.GetTenantsByTenantIdInventoriesByInventoryIdPrintersParams{Limit: &p.Limit, Cursor: &p.Cursor}))
	if err != nil {
		return ports.Result[[]ports.RegisteredPrinter]{}, err
	}
	values := []ports.RegisteredPrinter{}
	for _, v := range r.Data.GetOrEmpty() {
		values = append(values, humanPrinter(v))
	}
	return ports.Result[[]ports.RegisteredPrinter]{Data: values, Pagination: page(r.Meta)}, nil
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
func humanSelection(s ports.LabelPrintSelection) generated.PrintJobSelection {
	return generated.PrintJobSelection{PrinterId: s.PrinterID, ExpectedMediaFingerprint: s.ExpectedMediaFingerprint, TemplateId: s.TemplateID, TemplateVersion: int32(s.TemplateVersion), TemplateOptions: generated.PrintJobTemplateOptions{ShowReference: s.ShowReference}, Copies: int64(s.Copies)}
}
func (c *Client) QueueLabel(ctx context.Context, s ports.Scope, id string, selection ports.LabelPrintSelection, key string) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdPrintJobs(ctx, s.Tenant, s.Inventory, id, &generated.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdPrintJobsParams{IdempotencyKey: key}, humanSelection(selection))))
}
func (c *Client) TestPrinter(ctx context.Context, s ports.Scope, selection ports.LabelPrintSelection, key string) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterIdTestJobs(ctx, s.Tenant, s.Inventory, selection.PrinterID, &generated.PostTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterIdTestJobsParams{IdempotencyKey: key}, humanSelection(selection))))
}
func (c *Client) Reprint(ctx context.Context, s ports.Scope, id string, selection ports.LabelPrintSelection, key string) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdReprints(ctx, s.Tenant, s.Inventory, id, &generated.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdReprintsParams{IdempotencyKey: key}, humanSelection(selection))))
}
func (c *Client) CancelPrint(ctx context.Context, s ports.Scope, id string, revision uint64) (ports.Result[ports.PrintJobSummary], error) {
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdCancellation(ctx, s.Tenant, s.Inventory, id, nil, generated.PrintJobRevision{Revision: int64(revision)})))
}

var _ ports.HumanPrintingAPI = (*Client)(nil)
