package voice

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2"
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

func TestGoogleProviderProfileFactoryBuildsAPIKeyBackedGeminiProviders(t *testing.T) {
	t.Parallel()

	factory := GoogleProviderProfileFactory{}
	language, err := factory.LanguageInferenceProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, `{}`),
		CredentialPurpose: ports.ProviderCredentialPurposeAPIKey,
		Credential:        []byte("api-key"),
	})
	if err != nil {
		t.Fatalf("build api-key language provider: %v", err)
	}
	if provider, ok := language.(GoogleGeminiLanguageInference); !ok || provider.path != "/v1beta/models/gemini-test:generateContent" {
		t.Fatalf("unexpected api-key language provider: %#v", language)
	}
	stt, err := factory.SpeechToTextProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilitySpeechToText, `{}`),
		CredentialPurpose: ports.ProviderCredentialPurposeAPIKey,
		Credential:        []byte("api-key"),
	})
	if err != nil {
		t.Fatalf("build api-key speech-to-text provider: %v", err)
	}
	if provider, ok := stt.(GoogleGeminiSpeechToText); !ok || provider.path != "/v1beta/models/gemini-test:generateContent" {
		t.Fatalf("unexpected api-key speech-to-text provider: %#v", stt)
	}
	_, err = factory.TextToSpeechProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilityTextToSpeech, `{"languageCode":"en-US","voiceName":"en-US-Standard-C"}`),
		CredentialPurpose: ports.ProviderCredentialPurposeAPIKey,
		Credential:        []byte("api-key"),
	})
	if err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected api-key text-to-speech rejection, got %v", err)
	}
}

func TestGoogleProviderProfileFactoryBuildsServerADCBackedProviders(t *testing.T) {
	t.Parallel()

	tokenSource := countingTokenSource{}
	factory := GoogleProviderProfileFactory{
		DefaultTokenSource: func(context.Context) (oauth2.TokenSource, error) {
			return tokenSource, nil
		},
		ServerADCProjectID:    "project",
		ServerADCLocation:     "us-central1",
		ServerADCQuotaProject: "quota-project",
	}
	options := `{"projectId":"project","location":"us-central1","quotaProject":"quota-project","languageCode":"en-US","voiceName":"en-US-Standard-C"}`
	config := ProviderProfileProviderConfig{
		CredentialPurpose: ports.ProviderCredentialPurposeServerADC,
		Credential:        []byte("server_adc"),
	}

	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, options)
	language, err := factory.LanguageInferenceProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build ADC language provider: %v", err)
	}
	languageProvider, ok := language.(GoogleGeminiLanguageInference)
	if !ok || languageProvider.client.tokenSource == nil || languageProvider.client.quotaProject != "quota-project" {
		t.Fatalf("unexpected ADC language provider: %#v", language)
	}

	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilityTextToSpeech, options)
	tts, err := factory.TextToSpeechProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build ADC text-to-speech provider: %v", err)
	}
	ttsProvider, ok := tts.(GoogleTextToSpeech)
	if !ok || ttsProvider.client.tokenSource == nil || ttsProvider.client.quotaProject != "quota-project" {
		t.Fatalf("unexpected ADC text-to-speech provider: %#v", tts)
	}
}

func TestGoogleProviderProfileFactoryRejectsServerADCProjectOverride(t *testing.T) {
	t.Parallel()

	factory := GoogleProviderProfileFactory{
		DefaultTokenSource: func(context.Context) (oauth2.TokenSource, error) {
			return countingTokenSource{}, nil
		},
		ServerADCProjectID:    "allowed-project",
		ServerADCLocation:     "us-central1",
		ServerADCQuotaProject: "allowed-project",
	}
	_, err := factory.LanguageInferenceProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, `{"projectId":"other-project","location":"us-central1","quotaProject":"allowed-project"}`),
		CredentialPurpose: ports.ProviderCredentialPurposeServerADC,
		Credential:        []byte("server_adc"),
	})
	if err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected server ADC project override rejection, got %v", err)
	}

	_, err = factory.TextToSpeechProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilityTextToSpeech, `{"languageCode":"en-US","voiceName":"en-US-Standard-C","quotaProject":"other-project"}`),
		CredentialPurpose: ports.ProviderCredentialPurposeServerADC,
		Credential:        []byte("server_adc"),
	})
	if err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected server ADC quota project override rejection, got %v", err)
	}
}

func TestGoogleProviderProfileFactoryAppliesProfileHTTPTimeout(t *testing.T) {
	t.Parallel()

	options := `{"projectId":"project","location":"us-central1","languageCode":"en-US","voiceName":"en-US-Standard-C","httpTimeout":"75s"}`
	config := ProviderProfileProviderConfig{
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("access-token"),
	}
	factory := GoogleProviderProfileFactory{}

	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, options)
	language, err := factory.LanguageInferenceProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build language provider: %v", err)
	}
	languageProvider, ok := language.(GoogleGeminiLanguageInference)
	if !ok || languageProvider.client.httpClient.Timeout != 75*time.Second {
		t.Fatalf("unexpected language timeout provider: %#v", language)
	}

	config.Profile = googleFactoryProfile(t, agentmodel.ProviderCapabilityTextToSpeech, options)
	tts, err := factory.TextToSpeechProvider(context.Background(), config)
	if err != nil {
		t.Fatalf("build text-to-speech provider: %v", err)
	}
	ttsProvider, ok := tts.(GoogleTextToSpeech)
	if !ok || ttsProvider.client.httpClient.Timeout != 75*time.Second {
		t.Fatalf("unexpected text-to-speech timeout provider: %#v", tts)
	}
}

func TestGoogleProviderProfileFactoryRejectsMalformedProfileHTTPTimeout(t *testing.T) {
	t.Parallel()

	factory := GoogleProviderProfileFactory{}
	_, err := factory.LanguageInferenceProvider(context.Background(), ProviderProfileProviderConfig{
		Profile:           googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, `{"projectId":"project","location":"us-central1","httpTimeout":"0s"}`),
		CredentialPurpose: ports.ProviderCredentialPurposeOAuthBearer,
		Credential:        []byte("access-token"),
	})
	if err != ports.ErrInvalidProviderInput {
		t.Fatalf("expected invalid timeout rejection, got %v", err)
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

type countingTokenSource struct{}

func (countingTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "adc-token", TokenType: "Bearer"}, nil
}
