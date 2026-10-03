package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func selection(in dto.PrintJobSelection) printingapp.JobSelection {
	return printingapp.JobSelection{PrinterID: p.PrinterID(in.PrinterID), ExpectedMediaFingerprint: in.ExpectedMediaFingerprint, Template: p.TemplateSelection{ID: p.TemplateID(in.TemplateID), Version: in.TemplateVersion, Options: p.TemplateOptions{ShowReference: in.TemplateOptions.ShowReference}}, Copies: in.Copies, PreviewFingerprint: in.PreviewFingerprint}
}
func registerJobRequests(api huma.API, a app.App) {
	huma.Post(api, path+"/print-jobs/{jobId}/reprints", func(ctx context.Context, in *dto.ReprintInput) (*dto.Output, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		j, created, err := a.PrintJobs().Reprint(ctx, scope, p.JobID(in.JobID), in.IdempotencyKey, selection(in.Body))
		status := 200
		if created {
			status = 201
		}
		return output(j, status, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, path+"/printers/{printerId}/test-jobs", func(ctx context.Context, in *dto.TestJobInput) (*dto.Output, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		if in.Body.PrinterID != in.PrinterID {
			return nil, huma.Error400BadRequest("The selected printer does not match the destination.")
		}
		j, created, err := a.PrintJobs().TestPrinter(ctx, scope, in.IdempotencyKey, selection(in.Body))
		status := 200
		if created {
			status = 201
		}
		return output(j, status, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
