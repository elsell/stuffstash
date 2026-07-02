package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/legacy-homebox/preview", func(ctx context.Context, input *dto.LegacyHomeboxPreviewInput) (*dto.LegacyHomeboxPreviewOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		preview, err := application.PreviewLegacyHomeboxImport(ctx, app.PreviewImportInput{
			Principal:   principal,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Source:      sourceInput(input.Body),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.LegacyHomeboxPreviewOutput{
			Body: shared.SuccessEnvelope[dto.ImportPreviewResponse]{
				Data: mapper.PreviewToResponse(preview.Plan),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)

	huma.Post(api, "/tenants/{tenantId}/inventories/{inventoryId}/imports/legacy-homebox/apply", func(ctx context.Context, input *dto.LegacyHomeboxApplyInput) (*dto.LegacyHomeboxApplyOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		result, err := application.ApplyLegacyHomeboxImport(ctx, app.ApplyImportInput{
			Principal:   principal,
			RequestID:   input.RequestID,
			TenantID:    tenant.ID(input.TenantID),
			InventoryID: inventory.InventoryID(input.InventoryID),
			Source:      sourceInput(input.Body),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.LegacyHomeboxApplyOutput{
			Body: shared.SuccessEnvelope[dto.ImportApplyResponse]{
				Data: mapper.ApplyToResponse(mapper.ApplyCounts{
					FieldsCreated:      result.Counts.FieldsCreated,
					FieldsExisting:     result.Counts.FieldsExisting,
					LocationsCreated:   result.Counts.LocationsCreated,
					AssetsCreated:      result.Counts.AssetsCreated,
					AssetsSkipped:      result.Counts.AssetsSkipped,
					AttachmentsCreated: result.Counts.AttachmentsCreated,
					AttachmentsSkipped: result.Counts.AttachmentsSkipped,
				}, result.Messages),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("imports"), shared.SecuredOperation)
}

func sourceInput(body dto.LegacyHomeboxImportRequest) app.ImportSourceInput {
	return app.ImportSourceInput{
		SourceType:          body.SourceType,
		BaseURL:             body.BaseURL,
		Username:            body.Username,
		Password:            body.Password,
		IncludeImages:       body.IncludeImages,
		AllowInsecureTLS:    body.AllowInsecureTLS,
		AllowPrivateNetwork: body.AllowPrivateNetwork,
		FileName:            body.FileName,
		ContentBase64:       body.ContentBase64,
	}
}
