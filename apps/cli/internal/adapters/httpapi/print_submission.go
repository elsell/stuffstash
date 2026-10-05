package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func humanSelection(s ports.LabelPrintSelection) ([]byte, error) {
	if s.RequestBody != nil {
		return s.RequestBody, nil
	}
	// The server contract accepts uint32; generated PrintJobSelection uses int32.
	// Encode the project-owned selection to avoid narrowing the template version.
	return json.Marshal(struct {
		PrinterID                string                            `json:"printerId"`
		ExpectedMediaFingerprint string                            `json:"expectedMediaFingerprint"`
		TemplateID               string                            `json:"templateId"`
		TemplateVersion          uint32                            `json:"templateVersion"`
		TemplateOptions          generated.PrintJobTemplateOptions `json:"templateOptions"`
		Copies                   int                               `json:"copies"`
		PreviewFingerprint       string                            `json:"previewFingerprint,omitempty"`
	}{s.PrinterID, s.ExpectedMediaFingerprint, s.TemplateID, s.TemplateVersion, generated.PrintJobTemplateOptions{ShowReference: s.ShowReference}, s.Copies, s.PreviewFingerprint})
}
func (c *Client) QueueLabel(ctx context.Context, s ports.Scope, id string, selection ports.LabelPrintSelection, key string) (ports.Result[ports.PrintJobSummary], error) {
	body, err := humanSelection(selection)
	if err != nil {
		return ports.Result[ports.PrintJobSummary]{}, err
	}
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdPrintJobsWithBody(ctx, s.Tenant, s.Inventory, id, &generated.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdPrintJobsParams{IdempotencyKey: key}, "application/json", bytes.NewReader(body))))
}
func (c *Client) TestPrinter(ctx context.Context, s ports.Scope, selection ports.LabelPrintSelection, key string) (ports.Result[ports.PrintJobSummary], error) {
	body, err := humanSelection(selection)
	if err != nil {
		return ports.Result[ports.PrintJobSummary]{}, err
	}
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterIdTestJobsWithBody(ctx, s.Tenant, s.Inventory, selection.PrinterID, &generated.PostTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterIdTestJobsParams{IdempotencyKey: key}, "application/json", bytes.NewReader(body))))
}
func (c *Client) Reprint(ctx context.Context, s ports.Scope, id string, selection ports.LabelPrintSelection, key string) (ports.Result[ports.PrintJobSummary], error) {
	body, err := humanSelection(selection)
	if err != nil {
		return ports.Result[ports.PrintJobSummary]{}, err
	}
	return humanJobResult(read[generated.SuccessEnvelopePrintJob](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdReprintsWithBody(ctx, s.Tenant, s.Inventory, id, &generated.PostTenantsByTenantIdInventoriesByInventoryIdPrintJobsByJobIdReprintsParams{IdempotencyKey: key}, "application/json", bytes.NewReader(body))))
}
