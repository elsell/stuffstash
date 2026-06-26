package voice

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2"
)

type GoogleProviderProfileFactory struct{}

func (GoogleProviderProfileFactory) SpeechToTextProvider(_ context.Context, config ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	if config.Profile.Capability != agentmodel.ProviderCapabilitySpeechToText {
		return nil, ports.ErrInvalidProviderInput
	}
	geminiConfig, err := googleGeminiConfigFromProfile(config)
	if err != nil {
		return nil, err
	}
	return NewGoogleGeminiSpeechToText(geminiConfig), nil
}

func (GoogleProviderProfileFactory) LanguageInferenceProvider(_ context.Context, config ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error) {
	if config.Profile.Capability != agentmodel.ProviderCapabilityLanguageInference {
		return nil, ports.ErrInvalidProviderInput
	}
	geminiConfig, err := googleGeminiConfigFromProfile(config)
	if err != nil {
		return nil, err
	}
	return NewGoogleGeminiLanguageInference(geminiConfig), nil
}

func (GoogleProviderProfileFactory) TextToSpeechProvider(_ context.Context, config ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	if config.Profile.Capability != agentmodel.ProviderCapabilityTextToSpeech {
		return nil, ports.ErrInvalidProviderInput
	}
	if config.Profile.ProviderKind != agentmodel.ProviderKindGemini || config.CredentialPurpose != ports.ProviderCredentialPurposeOAuthBearer {
		return nil, ports.ErrInvalidProviderInput
	}
	options, err := providerRuntimeOptions(config.Profile)
	if err != nil {
		return nil, err
	}
	languageCode := stringOption(options, "languageCode")
	voiceName := stringOption(options, "voiceName")
	if languageCode == "" || voiceName == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	baseURL := config.Profile.EndpointURL.String()
	return NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: languageCode,
		VoiceName:    voiceName,
		QuotaProject: quotaProjectOption(options),
		BaseURL:      baseURL,
		TokenSource:  oauth2.StaticTokenSource(&oauth2.Token{AccessToken: strings.TrimSpace(string(config.Credential)), TokenType: "Bearer"}),
	}), nil
}

func googleGeminiConfigFromProfile(config ProviderProfileProviderConfig) (GoogleGeminiConfig, error) {
	if config.Profile.ProviderKind != agentmodel.ProviderKindGemini {
		return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
	}
	options, err := providerRuntimeOptions(config.Profile)
	if err != nil {
		return GoogleGeminiConfig{}, err
	}
	model := config.Profile.ModelName.String()
	if model == "" {
		return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
	}
	if config.CredentialPurpose == ports.ProviderCredentialPurposeAPIKey {
		apiKey := strings.TrimSpace(string(config.Credential))
		if apiKey == "" {
			return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
		}
		return GoogleGeminiConfig{
			Model:       model,
			BaseURL:     config.Profile.EndpointURL.String(),
			APIKey:      apiKey,
			TokenSource: nil,
		}, nil
	}
	if config.CredentialPurpose != ports.ProviderCredentialPurposeOAuthBearer {
		return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
	}
	projectID := stringOption(options, "projectId")
	location := stringOption(options, "location")
	if projectID == "" || location == "" {
		return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
	}
	return GoogleGeminiConfig{
		ProjectID:    projectID,
		Location:     location,
		Model:        model,
		QuotaProject: quotaProjectOption(options),
		BaseURL:      config.Profile.EndpointURL.String(),
		TokenSource:  oauth2.StaticTokenSource(&oauth2.Token{AccessToken: strings.TrimSpace(string(config.Credential)), TokenType: "Bearer"}),
	}, nil
}

func providerRuntimeOptions(profile agentmodel.ProviderProfile) (map[string]any, error) {
	options := map[string]any{}
	if err := json.Unmarshal([]byte(profile.RuntimeOptionsJSON.String()), &options); err != nil {
		return nil, ports.ErrInvalidProviderInput
	}
	return options, nil
}

func stringOption(options map[string]any, key string) string {
	value, _ := options[key].(string)
	return strings.TrimSpace(value)
}

func quotaProjectOption(options map[string]any) string {
	if quotaProject := stringOption(options, "quotaProject"); quotaProject != "" {
		return quotaProject
	}
	return stringOption(options, "projectId")
}
