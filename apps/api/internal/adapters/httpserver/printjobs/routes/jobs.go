package routes

import (
	"context"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printjobs/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"net/http"
)

const path = "/tenants/{tenantId}/inventories/{inventoryId}"

func jobError(err error) error {
	if errors.Is(err, printingapp.ErrLabelsUnavailable) {
		return huma.Error503ServiceUnavailable("Printing is not configured.")
	}
	return shared.ToHumaError(err)
}
func authenticate(ctx context.Context, a app.App, in dto.Scope) (printingapp.LabelScope, error) {
	p, e := shared.Authenticate(ctx, a, in.Authorization)
	if e != nil {
		return printingapp.LabelScope{}, e
	}
	return printingapp.LabelScope{Principal: p, TenantID: tenant.ID(in.TenantID), InventoryID: inventory.InventoryID(in.InventoryID), RequestID: in.RequestID}, nil
}
func output(j printing.Job, status int, err error) (*dto.Output, error) {
	if err != nil {
		return nil, jobError(err)
	}
	return &dto.Output{Status: status, CacheControl: "private, no-store", Body: shared.SuccessEnvelope[dto.PrintJob]{Data: mapper.Job(j), Meta: shared.Meta{TenantID: j.Scope.TenantID}}}, nil
}
func Register(api huma.API, a app.App) {
	registerConsumers(api, a)
	huma.Post(api, path+"/assets/{assetId}/print-jobs", func(ctx context.Context, in *dto.CreateInput) (*dto.Output, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		j, created, err := a.PrintJobs().Create(ctx, printingapp.CreateJobInput{Scope: scope, AssetID: asset.ID(in.AssetID), IdempotencyKey: in.IdempotencyKey, Selection: printingapp.JobSelection{PrinterID: printing.PrinterID(in.Body.PrinterID), ExpectedMediaFingerprint: in.Body.ExpectedMediaFingerprint, Template: printing.TemplateSelection{ID: printing.TemplateID(in.Body.TemplateID), Version: in.Body.TemplateVersion, Options: printing.TemplateOptions{ShowReference: in.Body.TemplateOptions.ShowReference}}, Copies: in.Body.Copies, PreviewFingerprint: in.Body.PreviewFingerprint}})
		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		return output(j, status, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Get(api, path+"/print-jobs/{jobId}", func(ctx context.Context, in *dto.JobInput) (*dto.Output, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		j, err := a.PrintJobs().Get(ctx, scope, printing.JobID(in.JobID))
		return output(j, http.StatusOK, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, path+"/print-jobs/{jobId}/cancellation", func(ctx context.Context, in *dto.CancelInput) (*dto.Output, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		j, err := a.PrintJobs().Cancel(ctx, scope, printing.JobID(in.JobID), in.Body.Revision)
		return output(j, http.StatusOK, err)
	}, huma.OperationTags("printing"), shared.SecuredOperation, func(op *huma.Operation) { op.DefaultStatus = http.StatusOK })
	huma.Get(api, path+"/print-jobs", func(ctx context.Context, in *dto.ListInput) (*dto.ListOutput, error) {
		scope, err := authenticate(ctx, a, in.Scope)
		if err != nil {
			return nil, err
		}
		page, err := a.PrintJobs().List(ctx, scope, printing.PrinterID(in.PrinterID), in.Limit, in.Cursor)
		if err != nil {
			return nil, jobError(err)
		}
		return &dto.ListOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[[]dto.PrintJob]{Data: mapper.Jobs(page.Items), Meta: shared.PaginatedMeta(in.TenantID, page.Limit, page.NextCursor, page.HasMore)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
