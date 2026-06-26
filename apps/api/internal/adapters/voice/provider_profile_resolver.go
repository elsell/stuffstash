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
	profiles    ports.ProviderProfileRepository
	credentials ports.ProviderCredentialRepository
	sealer      ports.ProviderCredentialSealer
	factory     ProviderProfileProviderFactory
}

func NewProviderProfileResolver(profiles ports.ProviderProfileRepository, credentials ports.ProviderCredentialRepository, sealer ports.ProviderCredentialSealer, factory ProviderProfileProviderFactory) ProviderProfileResolver {
	return ProviderProfileResolver{
		profiles:    profiles,
		credentials: credentials,
		sealer:      sealer,
		factory:     factory,
	}
}

func (r ProviderProfileResolver) ResolveRealtimeVoiceProviders(ctx context.Context, input ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	if r.profiles == nil || r.credentials == nil || r.sealer == nil || r.factory == nil {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	profiles, err := r.profiles.ListProviderProfiles(ctx, input.TenantID)
	if err != nil {
		return ports.RealtimeVoiceProviderSet{}, err
	}
	sttProfile, ok := selectProviderProfile(profiles, agentmodel.ProviderCapabilitySpeechToText)
	if !ok {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	languageProfile, ok := selectProviderProfile(profiles, agentmodel.ProviderCapabilityLanguageInference)
	if !ok {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	ttsProfile, ok := selectProviderProfile(profiles, agentmodel.ProviderCapabilityTextToSpeech)
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
		SpeechToText:               stt,
		LanguageInference:          language,
		TextToSpeech:               tts,
	}, nil
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
		record, found, err := r.credentials.ActiveProviderCredential(ctx, scope)
		if err != nil {
			return ProviderProfileProviderConfig{}, err
		}
		if !found {
			continue
		}
		raw, err := r.sealer.UnsealProviderCredential(ctx, scope, record.Sealed)
		if err != nil {
			return ProviderProfileProviderConfig{}, ports.ErrInvalidProviderInput
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
		return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeOAuthBearer, ports.ProviderCredentialPurposeAPIKey}
	}
	return []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer}
}

func selectProviderProfile(profiles []agentmodel.ProviderProfile, capability agentmodel.ProviderCapability) (agentmodel.ProviderProfile, bool) {
	for _, profile := range profiles {
		if profile.Capability == capability &&
			profile.LifecycleState == agentmodel.ProviderProfileEnabled &&
			profile.CredentialStatus == agentmodel.CredentialStatusConfigured {
			return profile, true
		}
	}
	return agentmodel.ProviderProfile{}, false
}
