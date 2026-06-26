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

func RegisterLifecycle(api huma.API, application app.App) {
	registerLifecycleAction(api, application, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/enable", application.EnableProviderProfile)
	registerLifecycleAction(api, application, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/disable", application.DisableProviderProfile)
	registerLifecycleAction(api, application, "/tenants/{tenantId}/provider-profiles/{providerProfileId}/archive", application.ArchiveProviderProfile)
}

func registerLifecycleAction(api huma.API, application app.App, path string, action func(context.Context, app.ProviderProfileLifecycleInput) (agentmodel.ProviderProfile, error)) {
	huma.Post(api, path, func(ctx context.Context, input *dto.ProviderProfileLifecycleInput) (*dto.ProviderProfileOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		lifecycleInput := app.ProviderProfileLifecycleInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
			ProfileID: agentmodel.ProviderProfileID(input.ProviderProfileID),
		}
		profile, err := action(ctx, lifecycleInput)
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
