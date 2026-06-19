package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/audit/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterListTenant(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/audit-records", func(ctx context.Context, input *dto.ListTenantAuditRecordsInput) (*dto.ListAuditRecordsOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		result, err := application.ListTenantAuditRecords(ctx, app.ListAuditRecordsInput{
			Principal: principal,
			TenantID:  tenant.ID(input.TenantID),
			Limit:     input.Limit,
			Cursor:    input.Cursor,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.ListAuditRecordsOutput{
			Body: shared.SuccessEnvelope[[]dto.RecordResponse]{
				Data: mapper.RecordsToResponse(result.Items),
				Meta: shared.PaginatedMeta(input.TenantID, result.Limit, result.NextCursor, result.HasMore),
			},
		}, nil
	}, huma.OperationTags("audit records"), shared.SecuredOperation)
}
