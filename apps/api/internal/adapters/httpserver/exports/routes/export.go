package routes

import (
	"context"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/exports/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Register(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/inventories/{inventoryId}/export", func(ctx context.Context, in *dto.ExportInput) (*dto.ExportOutput, error) {
		principal, err := shared.Authenticate(ctx, application, in.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ExportInventory(ctx, app.ExportInventoryInput{Principal: principal, TenantID: tenant.ID(in.TenantID), InventoryID: inventory.InventoryID(in.InventoryID), Format: ports.InventoryExportFormat(in.Format), Source: audit.SourceAPI, RequestID: in.RequestID})
		if errors.Is(err, ports.ErrInventoryExportLimit) {
			return nil, huma.Error422UnprocessableEntity("Inventory exceeds the configured export size limit. Ask your administrator to increase the limit.")
		}
		if errors.Is(err, ports.ErrInventoryExportFormat) {
			return nil, huma.Error422UnprocessableEntity("Choose JSON or CSV for export.")
		}
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		contentType := "application/json; charset=utf-8"
		if result.Format == ports.InventoryExportCSV {
			contentType = "text/csv; charset=utf-8"
		}
		return &dto.ExportOutput{ContentType: contentType, ContentDisposition: "attachment; filename=\"stuff-stash-inventory." + string(result.Format) + "\"", CacheControl: "private, no-store", Body: result.Content}, nil
	}, huma.OperationTags("exports"), shared.SecuredOperation)
}
