package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func Register(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/printer-profiles", func(ctx context.Context, input *dto.ListProfilesInput) (*dto.ListProfilesOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		profiles, err := application.ListPrinterProfiles(ctx, principal, printing.Scope{TenantID: input.TenantID, InventoryID: input.InventoryID})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ListProfilesOutput{Body: shared.SuccessEnvelope[[]dto.PrinterProfile]{Data: mapper.Profiles(profiles), Meta: shared.Meta{TenantID: input.TenantID}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
}
