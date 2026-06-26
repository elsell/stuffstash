package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterDetail(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/provider-profiles/{providerProfileId}", func(ctx context.Context, input *dto.GetProviderProfileInput) (*dto.ProviderProfileOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		profile, err := application.GetProviderProfile(ctx, app.GetProviderProfileInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
			ProfileID: agentmodel.ProviderProfileID(input.ProviderProfileID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ProviderProfileOutput{
			Body: shared.SuccessEnvelope[dto.ProviderProfileResponse]{
				Data: mapper.ProviderProfileToResponse(profile),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("provider profiles"), shared.SecuredOperation)
}
