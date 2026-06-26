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

func RegisterUpdate(api huma.API, application app.App) {
	huma.Patch(api, "/tenants/{tenantId}/provider-profiles/{providerProfileId}", func(ctx context.Context, input *dto.UpdateProviderProfileInput) (*dto.ProviderProfileOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		runtimeOptions, err := marshalOptionalObject(input.Body.RuntimeOptions)
		if err != nil {
			return nil, shared.ToHumaError(app.ErrValidation)
		}
		capabilityMetadata, err := marshalOptionalObject(input.Body.CapabilityMetadata)
		if err != nil {
			return nil, shared.ToHumaError(app.ErrValidation)
		}
		profile, err := application.UpdateProviderProfile(ctx, app.UpdateProviderProfileInput{
			Principal:          principal,
			Source:             audit.SourceAPI,
			RequestID:          input.RequestID,
			TenantID:           tenant.ID(input.TenantID),
			ProfileID:          agentmodel.ProviderProfileID(input.ProviderProfileID),
			DisplayName:        input.Body.DisplayName,
			EndpointURL:        input.Body.EndpointURL,
			ModelName:          input.Body.ModelName,
			RuntimeOptionsJSON: runtimeOptions,
			CapabilityJSON:     capabilityMetadata,
			PromptTemplate:     input.Body.PromptTemplate,
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
