package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/tenants/mapper"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func Register(api huma.API, application app.App) {
	huma.Post(api, "/tenants", func(ctx context.Context, input *dto.CreateTenantInput) (*dto.CreateTenantOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}

		item, err := application.CreateTenant(ctx, app.CreateTenantInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			Name:      input.Body.Name,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}

		return &dto.CreateTenantOutput{
			Body: shared.SuccessEnvelope[dto.TenantResponse]{
				Data: mapper.TenantToResponse(item),
				Meta: shared.Meta{TenantID: item.ID.String()},
			},
		}, nil
	}, huma.OperationTags("tenants"), shared.CreatedOperation, shared.SecuredOperation)

	huma.Get(api, "/tenants/{tenantId}", func(ctx context.Context, input *dto.GetTenantInput) (*dto.GetTenantOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		item, err := application.GetTenant(ctx, app.GetTenantInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.GetTenantOutput{Body: shared.SuccessEnvelope[dto.TenantResponse]{
			Data: mapper.TenantToResponse(item),
			Meta: shared.Meta{TenantID: item.ID.String()},
		}}, nil
	}, huma.OperationTags("tenants"), shared.SecuredOperation)

	huma.Patch(api, "/tenants/{tenantId}", func(ctx context.Context, input *dto.UpdateTenantInput) (*dto.UpdateTenantOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		item, err := application.UpdateTenant(ctx, app.UpdateTenantInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
			Name:      input.Body.Name,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateTenantOutput{Body: shared.SuccessEnvelope[dto.TenantResponse]{
			Data: mapper.TenantToResponse(item),
			Meta: shared.Meta{TenantID: item.ID.String()},
		}}, nil
	}, huma.OperationTags("tenants"), shared.SecuredOperation)

	huma.Patch(api, "/tenants/{tenantId}/archive", func(ctx context.Context, input *dto.UpdateTenantLifecycleInput) (*dto.UpdateTenantLifecycleOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		item, err := application.ArchiveTenant(ctx, app.UpdateTenantLifecycleInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateTenantLifecycleOutput{Body: shared.SuccessEnvelope[dto.TenantResponse]{
			Data: mapper.TenantToResponse(item),
			Meta: shared.Meta{TenantID: item.ID.String()},
		}}, nil
	}, huma.OperationTags("tenants"), shared.SecuredOperation)

	huma.Patch(api, "/tenants/{tenantId}/restore", func(ctx context.Context, input *dto.UpdateTenantLifecycleInput) (*dto.UpdateTenantLifecycleOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		item, err := application.RestoreTenant(ctx, app.UpdateTenantLifecycleInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.UpdateTenantLifecycleOutput{Body: shared.SuccessEnvelope[dto.TenantResponse]{
			Data: mapper.TenantToResponse(item),
			Meta: shared.Meta{TenantID: item.ID.String()},
		}}, nil
	}, huma.OperationTags("tenants"), shared.SecuredOperation)

	huma.Delete(api, "/tenants/{tenantId}", func(ctx context.Context, input *dto.UpdateTenantLifecycleInput) (*dto.DeleteTenantOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if err := application.DeleteTenant(ctx, app.UpdateTenantLifecycleInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
		}); err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.DeleteTenantOutput{}, nil
	}, huma.OperationTags("tenants"), shared.NoContentOperation, shared.SecuredOperation)
}
