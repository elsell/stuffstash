package routes

import (
	"context"
	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterCreate(api huma.API, application app.App) {
	huma.Post(api, "/tenants/{tenantId}/provider-profiles", func(ctx context.Context, input *dto.CreateProviderProfileInput) (*dto.ProviderProfileOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		runtimeOptions, err := marshalObject(input.Body.RuntimeOptions)
		if err != nil {
			return nil, shared.ToHumaError(app.ErrValidation)
		}
		capabilityMetadata, err := marshalObject(input.Body.CapabilityMetadata)
		if err != nil {
			return nil, shared.ToHumaError(app.ErrValidation)
		}
		profile, err := application.CreateProviderProfile(ctx, app.CreateProviderProfileInput{
			Principal:          principal,
			Source:             audit.SourceAPI,
			RequestID:          input.RequestID,
			TenantID:           tenant.ID(input.TenantID),
			Capability:         input.Body.Capability,
			ProviderKind:       input.Body.ProviderKind,
			DisplayName:        input.Body.DisplayName,
			EndpointURL:        input.Body.EndpointURL,
			ModelName:          input.Body.ModelName,
			RuntimeOptionsJSON: runtimeOptions,
			CapabilityJSON:     capabilityMetadata,
			PromptTemplate:     input.Body.PromptTemplate,
			Enable:             input.Body.Enable,
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
	}, huma.OperationTags("provider profiles"), shared.CreatedOperation, shared.SecuredOperation)
}

func marshalObject(value map[string]any) ([]byte, error) {
	if value == nil {
		value = map[string]any{}
	}
	return json.Marshal(value)
}
