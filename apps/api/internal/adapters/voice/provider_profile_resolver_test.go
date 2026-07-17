package voice

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProviderProfileResolverBuildsProvidersFromEnabledConfiguredProfiles(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "disabled-stt", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "stt-profile", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts-profile", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	vault := newProviderResolverCredentialVault(profiles[1], profiles[2], profiles[3])
	factory := &providerResolverFactory{}
	resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: profiles}, providerResolverVoiceConfigurationRepository{}, vault, factory)

	set, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
	})
	if err != nil {
		t.Fatalf("resolve providers: %v", err)
	}
	if set.SpeechToTextProfileID != "stt-profile" || set.LanguageInferenceProfileID != "lm-profile" || set.TextToSpeechProfileID != "tts-profile" {
		t.Fatalf("unexpected selected profile IDs: %+v", set)
	}
	if set.LanguagePromptTemplate != "Prefer concise spoken answers." {
		t.Fatalf("expected language prompt template from selected language profile, got %q", set.LanguagePromptTemplate)
	}
	if len(vault.scopes) != 4 {
		t.Fatalf("expected four credential vault reads, got %+v", vault.scopes)
	}
	if vault.scopes[2].Purpose != ports.ProviderCredentialPurposeServerADC || vault.scopes[3].Purpose != ports.ProviderCredentialPurposeOAuthBearer {
		t.Fatalf("expected tts credential lookup to try server ADC before oauth bearer, got %+v", vault.scopes)
	}
	if got := factory.configs["stt-profile"]; string(got.Credential) != "raw-stt-profile" || got.CredentialPurpose != ports.ProviderCredentialPurposeAPIKey {
		t.Fatalf("unexpected stt factory config: %+v", got)
	}
	if got := factory.configs["tts-profile"]; string(got.Credential) != "raw-tts-profile" || got.CredentialPurpose != ports.ProviderCredentialPurposeOAuthBearer {
		t.Fatalf("unexpected tts factory config: %+v", got)
	}
}

func TestProviderProfileResolverUsesExplicitVoiceProviderConfiguration(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt-implicit", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "stt-explicit", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts-profile", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	vault := newProviderResolverCredentialVault(profiles...)
	factory := &providerResolverFactory{}
	resolver := NewProviderProfileResolver(
		providerResolverProfileRepository{profiles: profiles},
		providerResolverVoiceConfigurationRepository{record: ports.VoiceProviderConfigurationRecord{
			TenantID:                   tenant.ID("tenant-home"),
			SpeechToTextProfileID:      "stt-explicit",
			LanguageInferenceProfileID: "lm-profile",
			TextToSpeechProfileID:      "tts-profile",
		}, found: true},
		vault,
		factory,
	)

	set, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: tenant.ID("tenant-home")})
	if err != nil {
		t.Fatalf("resolve providers: %v", err)
	}
	if set.SpeechToTextProfileID != "stt-explicit" {
		t.Fatalf("expected explicit speech profile, got %+v", set)
	}
}

func TestProviderProfileResolverUsesServerADCProviderCredential(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt-profile", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts-profile", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	vault := newProviderResolverCredentialVault(profiles[0], profiles[1])
	ttsScope := ports.ProviderCredentialScope{
		TenantID:          tenant.ID("tenant-home"),
		ProviderProfileID: "tts-profile",
		Capability:        ports.ProviderCapabilityTextToSpeech,
		ProviderKind:      ports.ProviderKindGemini,
		Purpose:           ports.ProviderCredentialPurposeServerADC,
	}
	vault.credentials[ttsScope] = []byte("server_adc")
	factory := &providerResolverFactory{}
	resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: profiles}, providerResolverVoiceConfigurationRepository{}, vault, factory)

	if _, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: tenant.ID("tenant-home")}); err != nil {
		t.Fatalf("resolve providers: %v", err)
	}
	if got := factory.configs["tts-profile"]; got.CredentialPurpose != ports.ProviderCredentialPurposeServerADC || string(got.Credential) != "server_adc" {
		t.Fatalf("unexpected tts factory config: %+v", got)
	}
}

func TestProviderProfileResolverFallsBackPerEmptyExplicitSlot(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt-implicit", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-explicit", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts-implicit", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	resolver := NewProviderProfileResolver(
		providerResolverProfileRepository{profiles: profiles},
		providerResolverVoiceConfigurationRepository{record: ports.VoiceProviderConfigurationRecord{
			TenantID:                   tenant.ID("tenant-home"),
			LanguageInferenceProfileID: "lm-explicit",
		}, found: true},
		newProviderResolverCredentialVault(profiles...),
		&providerResolverFactory{},
	)

	set, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: tenant.ID("tenant-home")})
	if err != nil {
		t.Fatalf("resolve providers: %v", err)
	}
	if set.SpeechToTextProfileID != "stt-implicit" || set.LanguageInferenceProfileID != "lm-explicit" || set.TextToSpeechProfileID != "tts-implicit" {
		t.Fatalf("expected per-slot explicit fallback, got %+v", set)
	}
}

func TestProviderProfileResolverFailsExplicitConfigurationWithoutFallback(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt-implicit", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "stt-disabled", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileDisabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts-profile", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	resolver := NewProviderProfileResolver(
		providerResolverProfileRepository{profiles: profiles},
		providerResolverVoiceConfigurationRepository{record: ports.VoiceProviderConfigurationRecord{
			TenantID:                   tenant.ID("tenant-home"),
			SpeechToTextProfileID:      "stt-disabled",
			LanguageInferenceProfileID: "lm-profile",
			TextToSpeechProfileID:      "tts-profile",
		}, found: true},
		newProviderResolverCredentialVault(profiles...),
		&providerResolverFactory{},
	)

	if _, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: tenant.ID("tenant-home")}); err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
}

func TestProviderProfileResolverFailsWhenRequiredCapabilityMissing(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt-profile", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	resolver := NewProviderProfileResolver(
		providerResolverProfileRepository{profiles: profiles},
		providerResolverVoiceConfigurationRepository{},
		newProviderResolverCredentialVault(profiles...),
		&providerResolverFactory{},
	)

	if _, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: tenant.ID("tenant-home")}); err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
}

func TestProviderProfileResolverFailsWhenCredentialCannotUnseal(t *testing.T) {
	t.Parallel()

	profiles := []agentmodel.ProviderProfile{
		providerResolverProfile(t, "stt-profile", agentmodel.ProviderCapabilitySpeechToText, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "lm-profile", agentmodel.ProviderCapabilityLanguageInference, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
		providerResolverProfile(t, "tts-profile", agentmodel.ProviderCapabilityTextToSpeech, agentmodel.ProviderProfileEnabled, agentmodel.CredentialStatusConfigured),
	}
	resolver := NewProviderProfileResolver(
		providerResolverProfileRepository{profiles: profiles},
		providerResolverVoiceConfigurationRepository{},
		newProviderResolverCredentialVaultWithFailure("lm-profile", profiles...),
		&providerResolverFactory{},
	)

	if _, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{TenantID: tenant.ID("tenant-home")}); err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
}

func providerResolverProfile(t *testing.T, id string, capability agentmodel.ProviderCapability, lifecycle agentmodel.ProviderProfileLifecycleState, credentialStatus agentmodel.CredentialStatus) agentmodel.ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := agentmodel.NewProviderProfile(agentmodel.ProviderProfileInput{
		ID:                 agentmodel.ProviderProfileID(id),
		TenantID:           agentmodel.TenantID("tenant-home"),
		Capability:         capability,
		ProviderKind:       agentmodel.ProviderKindGemini,
		DisplayName:        agentmodel.DisplayName("Gemini " + id),
		EndpointURL:        agentmodel.EndpointURL("https://example.test"),
		ModelName:          agentmodel.ModelName("gemini-test"),
		RuntimeOptionsJSON: []byte(`{"temperature":0}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     providerResolverPromptTemplate(capability),
		CredentialStatus:   credentialStatus,
		LifecycleState:     lifecycle,
		LastTestedAt:       &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("expected valid provider profile")
	}
	return profile
}

func providerResolverPromptTemplate(capability agentmodel.ProviderCapability) string {
	if capability == agentmodel.ProviderCapabilityLanguageInference {
		return "Prefer concise spoken answers."
	}
	return ""
}

type providerResolverProfileRepository struct {
	profiles []agentmodel.ProviderProfile
}

func (r providerResolverProfileRepository) ProviderProfileByID(context.Context, tenant.ID, agentmodel.ProviderProfileID) (agentmodel.ProviderProfile, bool, error) {
	return agentmodel.ProviderProfile{}, false, nil
}

func (r providerResolverProfileRepository) ListProviderProfiles(_ context.Context, tenantID tenant.ID) ([]agentmodel.ProviderProfile, error) {
	profiles := []agentmodel.ProviderProfile{}
	for _, profile := range r.profiles {
		if profile.TenantID.String() == tenantID.String() {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

type providerResolverVoiceConfigurationRepository struct {
	record ports.VoiceProviderConfigurationRecord
	found  bool
}

func (r providerResolverVoiceConfigurationRepository) VoiceProviderConfiguration(context.Context, tenant.ID) (ports.VoiceProviderConfigurationRecord, bool, error) {
	return r.record, r.found, nil
}

func (r providerResolverVoiceConfigurationRepository) SaveVoiceProviderConfiguration(context.Context, ports.VoiceProviderConfigurationRecord, audit.Record) error {
	return nil
}

type providerResolverCredentialVault struct {
	credentials   map[ports.ProviderCredentialScope][]byte
	failProfileID string
	scopes        []ports.ProviderCredentialScope
}

func newProviderResolverCredentialVault(profiles ...agentmodel.ProviderProfile) *providerResolverCredentialVault {
	return newProviderResolverCredentialVaultWithFailure("", profiles...)
}

func newProviderResolverCredentialVaultWithFailure(failProfileID string, profiles ...agentmodel.ProviderProfile) *providerResolverCredentialVault {
	vault := &providerResolverCredentialVault{
		credentials:   map[ports.ProviderCredentialScope][]byte{},
		failProfileID: failProfileID,
	}
	for _, profile := range profiles {
		for _, purpose := range []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer} {
			scope := ports.ProviderCredentialScope{
				TenantID:          tenant.ID(profile.TenantID.String()),
				ProviderProfileID: profile.ID.String(),
				Capability:        ports.ProviderCapability(profile.Capability.String()),
				ProviderKind:      ports.ProviderKind(profile.ProviderKind.String()),
				Purpose:           purpose,
			}
			vault.credentials[scope] = []byte("raw-" + profile.ID.String())
		}
	}
	return vault
}

func (v *providerResolverCredentialVault) PrepareProviderCredential(context.Context, ports.PrepareProviderCredentialInput) (ports.ProviderCredentialRecord, error) {
	return ports.ProviderCredentialRecord{}, nil
}

func (v *providerResolverCredentialVault) ActiveProviderCredentialMaterial(_ context.Context, scope ports.ProviderCredentialScope) ([]byte, bool, error) {
	v.scopes = append(v.scopes, scope)
	if scope.ProviderProfileID == v.failProfileID {
		return nil, false, ports.ErrInvalidProviderInput
	}
	raw, ok := v.credentials[scope]
	return append([]byte{}, raw...), ok, nil
}

type providerResolverFactory struct {
	configs map[string]ProviderProfileProviderConfig
}

func (f *providerResolverFactory) SpeechToTextProvider(_ context.Context, config ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	f.record(config)
	return providerResolverSpeechToText{}, nil
}

func (f *providerResolverFactory) LanguageInferenceProvider(_ context.Context, config ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error) {
	f.record(config)
	return providerResolverLanguageInference{}, nil
}

func (f *providerResolverFactory) TextToSpeechProvider(_ context.Context, config ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	f.record(config)
	return providerResolverTextToSpeech{}, nil
}

func (f *providerResolverFactory) record(config ProviderProfileProviderConfig) {
	if f.configs == nil {
		f.configs = map[string]ProviderProfileProviderConfig{}
	}
	f.configs[config.Profile.ID.String()] = config
}

type providerResolverSpeechToText struct{}

func (providerResolverSpeechToText) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	return ports.SpeechToTextResult{Transcript: "transcript"}, nil
}

type providerResolverLanguageInference struct{}

func (providerResolverLanguageInference) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
}

func (providerResolverLanguageInference) ProbeLanguageInference(context.Context) error { return nil }

type providerResolverTextToSpeech struct{}

func (providerResolverTextToSpeech) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{[]byte("speech")}}, nil
}
