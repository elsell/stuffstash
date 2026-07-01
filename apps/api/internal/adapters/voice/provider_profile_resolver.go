package voice

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ProviderProfileProviderConfig struct {
	Profile           agentmodel.ProviderProfile
	CredentialPurpose ports.ProviderCredentialPurpose
	Credential        []byte
}

type ProviderProfileProviderFactory interface {
	SpeechToTextProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error)
	LanguageInferenceProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error)
	TextToSpeechProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error)
}

type ProviderProfileResolver struct {
	profiles     ports.ProviderProfileRepository
	voiceConfigs ports.VoiceProviderConfigurationRepository
	vault        ports.ProviderCredentialVault
	factory      ProviderProfileProviderFactory
}

func NewProviderProfileResolver(profiles ports.ProviderProfileRepository, voiceConfigs ports.VoiceProviderConfigurationRepository, vault ports.ProviderCredentialVault, factory ProviderProfileProviderFactory) ProviderProfileResolver {
	return ProviderProfileResolver{
		profiles:     profiles,
		voiceConfigs: voiceConfigs,
		vault:        vault,
		factory:      factory,
	}
}

func (r ProviderProfileResolver) ResolveRealtimeVoiceProviders(ctx context.Context, input ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	if r.profiles == nil || r.vault == nil || r.factory == nil {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	profiles, err := r.profiles.ListProviderProfiles(ctx, input.TenantID)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	config, hasExplicitConfig, err := r.voiceProviderConfiguration(ctx, input.TenantID)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	sttProfile, ok := r.selectConfiguredProviderProfile(profiles, config.SpeechToTextProfileID, hasExplicitConfig, agentmodel.ProviderCapabilitySpeechToText)
	if !ok {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	languageProfile, ok := r.selectConfiguredProviderProfile(profiles, config.LanguageInferenceProfileID, hasExplicitConfig, agentmodel.ProviderCapabilityLanguageInference)
	if !ok {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	ttsProfile, ok := r.selectConfiguredProviderProfile(profiles, config.TextToSpeechProfileID, hasExplicitConfig, agentmodel.ProviderCapabilityTextToSpeech)
	if !ok {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}

	sttConfig, err := r.providerConfig(ctx, input.TenantID, sttProfile)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	languageConfig, err := r.providerConfig(ctx, input.TenantID, languageProfile)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	ttsConfig, err := r.providerConfig(ctx, input.TenantID, ttsProfile)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}

	stt, err := r.factory.SpeechToTextProvider(ctx, sttConfig)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	language, err := r.factory.LanguageInferenceProvider(ctx, languageConfig)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	tts, err := r.factory.TextToSpeechProvider(ctx, ttsConfig)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	if stt == nil || language == nil || tts == nil {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	return ports.RealtimeVoiceProviderSet{
		SpeechToTextProfileID:      sttProfile.ID.String(),
		LanguageInferenceProfileID: languageProfile.ID.String(),
		TextToSpeechProfileID:      ttsProfile.ID.String(),
		LanguagePromptTemplate:     languageProfile.PromptTemplate.String(),
		SpeechToText:               stt,
		LanguageInference:          language,
		TextToSpeech:               tts,
	}, nil
}

func (r ProviderProfileResolver) voiceProviderConfiguration(ctx context.Context, tenantID tenant.ID) (ports.VoiceProviderConfigurationRecord, bool, error) {
	if r.voiceConfigs == nil {
		return ports.VoiceProviderConfigurationRecord{TenantID: tenantID}, false, nil
	}
	record, found, err := r.voiceConfigs.VoiceProviderConfiguration(ctx, tenantID)
	if err != nil {
		return ports.VoiceProviderConfigurationRecord{}, false, err
	}
	if !found {
		return ports.VoiceProviderConfigurationRecord{TenantID: tenantID}, false, nil
	}
	return record, true, nil
}

func (r ProviderProfileResolver) selectConfiguredProviderProfile(profiles []agentmodel.ProviderProfile, selectedID string, explicit bool, capability agentmodel.ProviderCapability) (agentmodel.ProviderProfile, bool) {
	if explicit && selectedID != "" {
		for _, profile := range profiles {
			if profile.ID.String() == selectedID && profile.Capability == capability && providerProfileRuntimeReady(profile) {
				return profile, true
			}
		}
		return agentmodel.ProviderProfile{}, false
	}
	return selectProviderProfile(profiles, capability)
}

func (r ProviderProfileResolver) providerConfig(ctx context.Context, tenantID tenant.ID, profile agentmodel.ProviderProfile) (ProviderProfileProviderConfig, error) {
	for _, purpose := range providerCredentialPurposes(profile) {
		scope := ports.ProviderCredentialScope{
			TenantID:          tenantID,
			ProviderProfileID: profile.ID.String(),
			Capability:        ports.ProviderCapability(profile.Capability.String()),
			ProviderKind:      ports.ProviderKind(profile.ProviderKind.String()),
			Purpose:           purpose,
		}
		raw, found, err := r.vault.ActiveProviderCredentialMaterial(ctx, scope)
		if err != nil {
			return ProviderProfileProviderConfig{}, err
		}
		if !found {
			continue
		}
		if len(raw) == 0 {
			return ProviderProfileProviderConfig{}, ports.ErrInvalidProviderInput
		}
		return ProviderProfileProviderConfig{Profile: profile, CredentialPurpose: purpose, Credential: raw}, nil
	}
	return ProviderProfileProviderConfig{}, ports.ErrInvalidProviderInput
}

func providerCredentialPurposes(profile agentmodel.ProviderProfile) []ports.ProviderCredentialPurpose {
	if profile.ProviderKind == agentmodel.ProviderKindGemini {
		if profile.Capability == agentmodel.ProviderCapabilityTextToSpeech {
			return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeServerADC, ports.ProviderCredentialPurposeOAuthBearer}
		}
		return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeServerADC, ports.ProviderCredentialPurposeOAuthBearer}
	}
	return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer}
}

func selectProviderProfile(profiles []agentmodel.ProviderProfile, capability agentmodel.ProviderCapability) (agentmodel.ProviderProfile, bool) {
	for _, profile := range profiles {
		if profile.Capability == capability && providerProfileRuntimeReady(profile) {
			return profile, true
		}
	}
	return agentmodel.ProviderProfile{}, false
}

func providerProfileRuntimeReady(profile agentmodel.ProviderProfile) bool {
	return profile.LifecycleState == agentmodel.ProviderProfileEnabled &&
		profile.CredentialStatus == agentmodel.CredentialStatusConfigured &&
		profile.LastTestedAt != nil
}
