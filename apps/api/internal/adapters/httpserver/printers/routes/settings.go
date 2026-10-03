package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func registerPrintSettings(api huma.API, application app.App) {
	const path = "/tenants/{tenantId}/inventories/{inventoryId}/print-settings"
	huma.Get(api, path, func(ctx context.Context, input *dto.PrintSettingsInput) (*dto.PrintSettingsOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		settings, err := application.PrinterRegistry().PrintSettings(ctx, actor)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PrintSettingsOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[dto.InventoryPrintSettings]{Data: mapper.PrintSettings(settings), Meta: shared.Meta{TenantID: actor.Scope.TenantID}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Put(api, path, func(ctx context.Context, input *dto.ReplacePrintSettingsInput) (*dto.PrintSettingsOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		settings, err := application.PrinterRegistry().ReplacePrintSettings(ctx, actor, mapper.PrintSettingsCommand(input.Body))
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PrintSettingsOutput{CacheControl: "private, no-store", Body: shared.SuccessEnvelope[dto.InventoryPrintSettings]{Data: mapper.PrintSettings(settings), Meta: shared.Meta{TenantID: actor.Scope.TenantID}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
