package voice

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleProviderProfileFactoryBuildsOAuthBackedProviders(t *testing.T) {
	t.Parallel()

	options := `{"projectId":"project","location":"us-central1","quotaProject":"quota-project","languageCode":"en-US","voiceName":"en-US-Standard-C"}`
	config := ProviderProfileProviderConfig{
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("access-token"),
	}
	factory := GoogleProviderProfileFactory{}

	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilitySpeechToText, options)
	stt, err := factory.SpeechToTextProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build speech-to-text provider: %v", err)
	}
	if provider, ok := stt.(GoogleGeminiSpeechToText); !ok || provider.path != "/v1/projects/project/locations/us-central1/publishers/google/models/gemini-test:generateContent" {
		t.Fatalf("unexpected speech-to-text provider: %#v", stt)
	}
	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, options)
	language, err := factory.LanguageInferenceProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build language provider: %v", err)
	}
	if provider, ok := language.(GoogleGeminiLanguageInference); !ok || provider.path != "/v1/projects/project/locations/us-central1/publishers/google/models/gemini-test:generateContent" {
		t.Fatalf("unexpected language provider: %#v", language)
	}
	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilityTextToSpeech, options)
	tts, err := factory.TextToSpeechProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build text-to-speech provider: %v", err)
	}
	if provider, ok := tts.(GoogleTextToSpeech); !ok || provider.languageCode != "en-US" || provider.voiceName != "en-US-Standard-C" {
		t.Fatalf("unexpected text-to-speech provider: %#v", tts)
	}
}

func TestGoogleProviderProfileFactoryRejectsAPIKeyUntilSupported(t *testing.T) {
	t.Parallel()

	factory := GoogleProviderProfileFactory{}
	_, err := factory.LanguageInferenceProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, `{"projectId":"project","location":"us-central1"}`),
		CredentialPurpose: ports.ProviderCredentialPurposeAPIKey,
		Credential:        []byte("api-key"),
	})
	if err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
}

func googleFactoryProfile(t *testing.T, capability agentmodel.ProviderCapability, runtimeOptions string) agentmodel.ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := agentmodel.NewProviderProfile(agentmodel.ProviderProfileInput{
		ID:                 agentmodel.ProviderProfileID("profile-one"),
		TenantID:           agentmodel.TenantID("tenant-home"),
		Capability:         capability,
		ProviderKind:       agentmodel.ProviderKindGemini,
		DisplayName:        agentmodel.DisplayName("Google"),
		ModelName:          agentmodel.ModelName("gemini-test"),
		RuntimeOptionsJSON: []byte(runtimeOptions),
		CapabilityJSON:     []byte(`{}`),
		CredentialStatus:   agentmodel.CredentialStatusConfigured,
		LifecycleState:     agentmodel.ProviderProfileEnabled,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("expected valid profile")
	}
	return profile
}
