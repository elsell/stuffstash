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
)

func registerConnectors(api huma.API, application app.App) {
	registerCredentialRotation(api, application)
	const detail = "/tenants/{tenantId}/inventories/{inventoryId}/print-connectors/{connectorId}"
	huma.Patch(api, detail, func(ctx context.Context, input *dto.UpdateConnectorInput) (*dto.ConnectorOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		var ids *[]printing.PrinterID
		if input.Body.PrinterIDs != nil {
			values := make([]printing.PrinterID, len(*input.Body.PrinterIDs))
			for i, id := range *input.Body.PrinterIDs {
				values[i] = printing.PrinterID(id)
			}
			ids = &values
		}
		r, err := application.PrintConnectors().Update(ctx, printregistry.UpdateConnector{Actor: actor, ID: printing.ConnectorID(input.ConnectorID), Generation: input.Body.Generation, Name: input.Body.Name, Revoked: input.Body.Revoked, PrinterIDs: ids})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ConnectorOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.Connector]{Data: mapper.ConnectorRegistration(r)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Get(api, detail, func(ctx context.Context, input *dto.ConnectorInput) (*dto.ConnectorOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		r, err := application.PrintConnectors().Get(ctx, actor, printing.ConnectorID(input.ConnectorID))
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ConnectorOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.Connector]{Data: mapper.ConnectorRegistration(r)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/print-connectors", func(ctx context.Context, input *dto.ListConnectorsInput) (*dto.ConnectorsOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		page, err := application.PrintConnectors().List(ctx, actor, input.Limit, input.Cursor)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		result := make([]dto.Connector, 0, len(page.Items))
		for _, r := range page.Items {
			result = append(result, mapper.ConnectorRegistration(r))
		}
		return &dto.ConnectorsOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[[]dto.Connector]{Data: result, Meta: shared.PaginatedMeta(input.TenantID, page.Limit, page.NextCursor, page.HasMore)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)

}
