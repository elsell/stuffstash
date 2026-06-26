package voice

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
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
	credentials := newProviderResolverCredentialRepository(profiles[1], profiles[2], profiles[3])
	sealer := &providerResolverSealer{}
	factory := &providerResolverFactory{}
	resolver := NewProviderProfileResolver(providerResolverProfileRepository{profiles: profiles}, credentials, sealer, factory)

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
	if len(sealer.unsealedScopes) != 3 {
		t.Fatalf("expected three credential unseal calls, got %+v", sealer.unsealedScopes)
	}
	if got := factory.configs["stt-profile"]; string(got.Credential) != "raw-stt-profile" || got.CredentialPurpose != ports.ProviderCredentialPurposeAPIKey {
		t.Fatalf("unexpected stt factory config: %+v", got)
	}
	if got := factory.configs["tts-profile"]; string(got.Credential) != "raw-tts-profile" || got.CredentialPurpose != ports.ProviderCredentialPurposeOAuthBearer {
		t.Fatalf("unexpected tts factory config: %+v", got)
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
		newProviderResolverCredentialRepository(profiles...),
		&providerResolverSealer{},
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
		newProviderResolverCredentialRepository(profiles...),
		&providerResolverSealer{failProfileID: "lm-profile"},
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

type providerResolverCredentialRepository struct {
	credentials map[ports.ProviderCredentialScope]ports.ProviderCredentialRecord
}

func newProviderResolverCredentialRepository(profiles ...agentmodel.ProviderProfile) providerResolverCredentialRepository {
	repository := providerResolverCredentialRepository{credentials: map[ports.ProviderCredentialScope]ports.ProviderCredentialRecord{}}
	for _, profile := range profiles {
		for _, purpose := range []ports.ProviderCredentialPurpose{ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer} {
			scope := ports.ProviderCredentialScope{
				TenantID:          tenant.ID(profile.TenantID.String()),
				ProviderProfileID: profile.ID.String(),
				Capability:        ports.ProviderCapability(profile.Capability.String()),
				ProviderKind:      ports.ProviderKind(profile.ProviderKind.String()),
				Purpose:           purpose,
			}
			repository.credentials[scope] = ports.ProviderCredentialRecord{
				ID:     "credential-" + profile.ID.String() + "-" + string(purpose),
				Scope:  scope,
				Sealed: ports.SealedProviderCredential{KeyID: profile.ID.String(), Algorithm: ports.ProviderCredentialAlgorithmAES256GCM, Nonce: []byte("123456789012"), Ciphertext: []byte("sealed")},
			}
		}
	}
	return repository
}

func (r providerResolverCredentialRepository) ReplaceProviderCredential(context.Context, ports.ProviderCredentialRecord) error {
	return nil
}

func (r providerResolverCredentialRepository) ActiveProviderCredential(_ context.Context, scope ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	record, ok := r.credentials[scope]
	return record, ok, nil
}

func (r providerResolverCredentialRepository) ActiveProviderCredentialsExist(context.Context) (bool, error) {
	return len(r.credentials) > 0, nil
}

func (r providerResolverCredentialRepository) SupersedeActiveProviderCredential(context.Context, ports.ProviderCredentialScope, time.Time) error {
	return nil
}

type providerResolverSealer struct {
	failProfileID  string
	unsealedScopes []ports.ProviderCredentialScope
}

func (s *providerResolverSealer) SealProviderCredential(context.Context, ports.ProviderCredentialScope, []byte) (ports.SealedProviderCredential, error) {
	return ports.SealedProviderCredential{}, nil
}

func (s *providerResolverSealer) UnsealProviderCredential(_ context.Context, scope ports.ProviderCredentialScope, _ ports.SealedProviderCredential) ([]byte, error) {
	s.unsealedScopes = append(s.unsealedScopes, scope)
	if scope.ProviderProfileID == s.failProfileID {
		return nil, ports.ErrInvalidProviderCredential
	}
	return []byte("raw-" + scope.ProviderProfileID), nil
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
	return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: "answer"}}, nil
}

type providerResolverTextToSpeech struct{}

func (providerResolverTextToSpeech) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{[]byte("speech")}}, nil
}
