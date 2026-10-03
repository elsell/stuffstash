package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"net/http"
)

func printerActor(ctx context.Context, application app.App, input dto.PrinterScope) (printregistry.Actor, error) {
	p, err := shared.Authenticate(ctx, application, input.Authorization)
	return printregistry.Actor{Principal: p, Scope: printing.Scope{TenantID: input.TenantID, InventoryID: input.InventoryID}, RequestID: input.RequestID}, err
}
func registerPrinters(api huma.API, application app.App) {
	const collection = "/tenants/{tenantId}/inventories/{inventoryId}/printers"
	huma.Post(api, collection, func(ctx context.Context, input *dto.CreatePrinterInput) (*dto.PrinterOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		p, created, err := application.PrinterRegistry().Register(ctx, printregistry.RegisterPrinter{Actor: actor, RequestKey: input.IdempotencyKey, Name: input.Body.Name, AdapterID: input.Body.AdapterID, PresetID: input.Body.PresetID, PresetVersion: input.Body.PresetVersion})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		return &dto.PrinterOutput{Status: status, Body: shared.SuccessEnvelope[dto.Printer]{Data: mapper.Printer(p), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Get(api, collection+"/{printerId}", func(ctx context.Context, input *dto.PrinterInput) (*dto.PrinterOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		p, err := application.PrinterRegistry().Get(ctx, actor, printing.PrinterID(input.PrinterID))
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PrinterOutput{Status: http.StatusOK, Body: shared.SuccessEnvelope[dto.Printer]{Data: mapper.Printer(p), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Patch(api, collection+"/{printerId}", func(ctx context.Context, input *dto.UpdatePrinterInput) (*dto.PrinterOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		p, err := application.PrinterRegistry().Update(ctx, printregistry.UpdatePrinter{Actor: actor, ID: printing.PrinterID(input.PrinterID), Revision: input.Body.Revision, Name: input.Body.Name, PresetID: input.Body.PresetID, PresetVersion: input.Body.PresetVersion, Retired: input.Body.Retired})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PrinterOutput{Status: http.StatusOK, Body: shared.SuccessEnvelope[dto.Printer]{Data: mapper.Printer(p), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Get(api, collection, func(ctx context.Context, input *dto.ListPrintersInput) (*dto.PrintersOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		result, err := application.PrinterRegistry().List(ctx, actor, input.Limit, input.Cursor)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PrintersOutput{Body: shared.SuccessEnvelope[[]dto.Printer]{Data: mapper.Printers(result.Items), Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore)}}, nil

	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
