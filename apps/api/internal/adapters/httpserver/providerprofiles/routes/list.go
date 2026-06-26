package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterList(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/provider-profiles", func(ctx context.Context, input *dto.ListProviderProfilesInput) (*dto.ListProviderProfilesOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		profiles, err := application.ListProviderProfiles(ctx, app.ListProviderProfilesInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ListProviderProfilesOutput{
			Body: shared.SuccessEnvelope[[]dto.ProviderProfileResponse]{
				Data: mapper.ProviderProfilesToResponse(profiles),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("provider profiles"), shared.SecuredOperation)
}
