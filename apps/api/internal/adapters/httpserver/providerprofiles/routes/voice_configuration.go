package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/providerprofiles/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func RegisterVoiceConfiguration(api huma.API, application app.App) {
	huma.Get(api, "/tenants/{tenantId}/voice-provider-configuration", func(ctx context.Context, input *dto.GetVoiceProviderConfigurationInput) (*dto.VoiceProviderConfigurationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		view, err := application.GetVoiceProviderConfiguration(ctx, app.VoiceProviderConfigurationInput{
			Principal: principal,
			Source:    audit.SourceAPI,
			RequestID: input.RequestID,
			TenantID:  tenant.ID(input.TenantID),
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.VoiceProviderConfigurationOutput{
			Body: shared.SuccessEnvelope[dto.VoiceProviderConfigurationResponse]{
				Data: voiceProviderConfigurationToResponse(view),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("provider profiles"), shared.SecuredOperation)

	huma.Put(api, "/tenants/{tenantId}/voice-provider-configuration", func(ctx context.Context, input *dto.UpdateVoiceProviderConfigurationInput) (*dto.VoiceProviderConfigurationOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		view, err := application.UpdateVoiceProviderConfiguration(ctx, app.UpdateVoiceProviderConfigurationInput{
			Principal:                  principal,
			Source:                     audit.SourceAPI,
			RequestID:                  input.RequestID,
			TenantID:                   tenant.ID(input.TenantID),
			SpeechToTextProfileID:      input.Body.SpeechToTextProfileID,
			LanguageInferenceProfileID: input.Body.LanguageInferenceProfileID,
			TextToSpeechProfileID:      input.Body.TextToSpeechProfileID,
		})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.VoiceProviderConfigurationOutput{
			Body: shared.SuccessEnvelope[dto.VoiceProviderConfigurationResponse]{
				Data: voiceProviderConfigurationToResponse(view),
				Meta: shared.Meta{TenantID: input.TenantID},
			},
		}, nil
	}, huma.OperationTags("provider profiles"), shared.SecuredOperation)
}

func voiceProviderConfigurationToResponse(view app.VoiceProviderConfigurationView) dto.VoiceProviderConfigurationResponse {
	return dto.VoiceProviderConfigurationResponse{
		TenantID:  view.TenantID,
		Readiness: view.Readiness,
		UpdatedAt: view.UpdatedAt,
		ProfileIDs: dto.VoiceProviderConfigurationProfileID{
			SpeechToText:      view.ProfileIDs.SpeechToText,
			LanguageInference: view.ProfileIDs.LanguageInference,
			TextToSpeech:      view.ProfileIDs.TextToSpeech,
		},
		Slots: voiceProviderSlotsToResponse(view.Slots),
	}
}

func voiceProviderSlotsToResponse(slots []app.VoiceProviderSlotView) []dto.VoiceProviderSlotResponse {
	data := make([]dto.VoiceProviderSlotResponse, 0, len(slots))
	for _, slot := range slots {
		data = append(data, voiceProviderSlotToResponse(slot))
	}
	return data
}

func voiceProviderSlotToResponse(slot app.VoiceProviderSlotView) dto.VoiceProviderSlotResponse {
	var selected *dto.ProviderProfileSummaryResponse
	if slot.SelectedProfile != nil {
		value := providerProfileSummaryToResponse(*slot.SelectedProfile)
		selected = &value
	}
	return dto.VoiceProviderSlotResponse{
		Capability:        slot.Capability,
		Label:             slot.Label,
		SelectedProfileID: slot.SelectedProfileID,
		SelectedProfile:   selected,
		SelectionSource:   slot.SelectionSource,
		Readiness:         slot.Readiness,
		Issues:            slot.Issues,
		RecommendedAction: slot.RecommendedAction,
		DuplicateProfiles: providerProfileSummariesToResponse(slot.DuplicateProfiles),
	}
}

func providerProfileSummariesToResponse(profiles []app.ProviderProfileSummary) []dto.ProviderProfileSummaryResponse {
	data := make([]dto.ProviderProfileSummaryResponse, 0, len(profiles))
	for _, profile := range profiles {
		data = append(data, providerProfileSummaryToResponse(profile))
	}
	return data
}

func providerProfileSummaryToResponse(profile app.ProviderProfileSummary) dto.ProviderProfileSummaryResponse {
	return dto.ProviderProfileSummaryResponse{
		ID:                profile.ID,
		Capability:        profile.Capability,
		ProviderKind:      profile.ProviderKind,
		DisplayName:       profile.DisplayName,
		ModelName:         profile.ModelName,
		CredentialStatus:  profile.CredentialStatus,
		CredentialPurpose: profile.CredentialPurpose,
		LifecycleState:    profile.LifecycleState,
		LastTestedAt:      profile.LastTestedAt,
	}
}
