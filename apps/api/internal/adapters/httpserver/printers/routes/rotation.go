package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func registerCredentialRotation(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/print-connectors/{connectorId}/credential-rotation", func(ctx context.Context, input *dto.RotateCredentialInput) (*dto.PairingStatusOutput, error) {
		actor, err := printerActor(ctx, application, input.PrinterScope)
		if err != nil {
			return nil, err
		}
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		p, err := application.PrintConnectors().RotateCredential(ctx, printregistry.RotateConnectorCredential{Actor: actor, ConnectorID: printing.ConnectorID(input.ConnectorID), Generation: input.Body.Generation, PairingID: printing.PairingID(input.Body.PairingID), UserCode: input.Body.UserCode})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PairingStatusOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.PairingStatus]{Data: dto.PairingStatus{ID: string(p.ID), State: string(p.State), ExpiresAt: p.ExpiresAt}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
