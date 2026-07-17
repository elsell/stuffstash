package voice

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProviderProfileFactory struct {
	DefaultTokenSource    func(context.Context) (oauth2.TokenSource, error)
	ServerADCProjectID    string
	ServerADCLocation     string
	ServerADCQuotaProject string
}

func (f GoogleProviderProfileFactory) SpeechToTextProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.SpeechToTextProvider, error) {
	if config.Profile.Capability != agentmodel.ProviderCapabilitySpeechToText {
		return nil, ports.ErrInvalidProviderInput
	}
	geminiConfig, err := f.googleGeminiConfigFromProfile(ctx, config)
	if err != nil {
		return nil, err
	}
	return NewGoogleGeminiSpeechToText(geminiConfig), nil
}

func (f GoogleProviderProfileFactory) LanguageInferenceProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.LanguageInferenceProvider, error) {
	return f.RealtimeLanguageProvider(ctx, config)
}

func (f GoogleProviderProfileFactory) RealtimeLanguageProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.RealtimeLanguageProvider, error) {
	if config.Profile.Capability != agentmodel.ProviderCapabilityLanguageInference {
		return nil, ports.ErrInvalidProviderInput
	}
	geminiConfig, err := f.googleGeminiConfigFromProfile(ctx, config)
	if err != nil {
		return nil, err
	}
	return NewGoogleGeminiLanguageInference(geminiConfig), nil
}

func (f GoogleProviderProfileFactory) TextToSpeechProvider(ctx context.Context, config ProviderProfileProviderConfig) (ports.TextToSpeechProvider, error) {
	if config.Profile.Capability != agentmodel.ProviderCapabilityTextToSpeech {
		return nil, ports.ErrInvalidProviderInput
	}
	if config.Profile.ProviderKind != agentmodel.ProviderKindGemini {
		return nil, ports.ErrInvalidProviderInput
	}
	tokenSource, err := f.googleTokenSource(ctx, config)
	if err != nil {
		return nil, err
	}
	options, err := providerRuntimeOptions(config.Profile)
	if err != nil {
		return nil, err
	}
	if err := f.validateGoogleQuotaProjectOption(config.CredentialPurpose, options); err != nil {
		return nil, err
	}
	languageCode := stringOption(options, "languageCode")
	voiceName := stringOption(options, "voiceName")
	if languageCode == "" || voiceName == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	httpTimeout, err := httpTimeoutOption(options)
	if err != nil {
		return nil, err
	}
	baseURL := config.Profile.EndpointURL.String()
	return NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: languageCode,
		VoiceName:    voiceName,
		QuotaProject: f.googleQuotaProjectOption(config.CredentialPurpose, options),
		BaseURL:      baseURL,
		TokenSource:  tokenSource,
		HTTPTimeout:  httpTimeout,
	}), nil
}

func (f GoogleProviderProfileFactory) googleTokenSource(ctx context.Context, config ProviderProfileProviderConfig) (oauth2.TokenSource, error) {
	switch config.CredentialPurpose {
	case ports.ProviderCredentialPurposeOAuthBearer:
		token := strings.TrimSpace(string(config.Credential))
		if token == "" {
			return nil, ports.ErrInvalidProviderInput
		}
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token, TokenType: "Bearer"}), nil
	case ports.ProviderCredentialPurposeServerADC:
		tokenSource := f.DefaultTokenSource
		if tokenSource == nil {
			tokenSource = func(ctx context.Context) (oauth2.TokenSource, error) {
				return google.DefaultTokenSource(ctx, googleCloudPlatformScope)
			}
		}
		return tokenSource(ctx)
	default:
		return nil, ports.ErrInvalidProviderInput
	}
}

func googleGeminiConfigFromProfile(config ProviderProfileProviderConfig) (GoogleGeminiConfig, error) {
	return GoogleProviderProfileFactory{}.googleGeminiConfigFromProfile(context.Background(), config)
}

func (f GoogleProviderProfileFactory) googleGeminiConfigFromProfile(ctx context.Context, config ProviderProfileProviderConfig) (GoogleGeminiConfig, error) {
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
	httpTimeout, err := httpTimeoutOption(options)
	if err != nil {
		return GoogleGeminiConfig{}, err
	}
	if err := f.validateGoogleQuotaProjectOption(config.CredentialPurpose, options); err != nil {
		return GoogleGeminiConfig{}, err
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
			HTTPTimeout: httpTimeout,
		}, nil
	}
	if config.CredentialPurpose != ports.ProviderCredentialPurposeOAuthBearer && config.CredentialPurpose != ports.ProviderCredentialPurposeServerADC {
		return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
	}
	projectID, err := f.googleProjectOption(config.CredentialPurpose, options)
	if err != nil {
		return GoogleGeminiConfig{}, err
	}
	location, err := f.googleLocationOption(config.CredentialPurpose, options)
	if err != nil {
		return GoogleGeminiConfig{}, err
	}
	if projectID == "" || location == "" {
		return GoogleGeminiConfig{}, ports.ErrInvalidProviderInput
	}
	tokenSource, err := f.googleTokenSource(ctx, config)
	if err != nil {
		return GoogleGeminiConfig{}, err
	}
	return GoogleGeminiConfig{
		ProjectID:    projectID,
		Location:     location,
		Model:        model,
		QuotaProject: f.googleQuotaProjectOption(config.CredentialPurpose, options),
		BaseURL:      config.Profile.EndpointURL.String(),
		TokenSource:  tokenSource,
		HTTPTimeout:  httpTimeout,
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

func (f GoogleProviderProfileFactory) googleProjectOption(purpose ports.ProviderCredentialPurpose, options map[string]any) (string, error) {
	if purpose != ports.ProviderCredentialPurposeServerADC {
		return stringOption(options, "projectId"), nil
	}
	return boundedGoogleOption(stringOption(options, "projectId"), f.ServerADCProjectID)
}

func (f GoogleProviderProfileFactory) googleLocationOption(purpose ports.ProviderCredentialPurpose, options map[string]any) (string, error) {
	if purpose != ports.ProviderCredentialPurposeServerADC {
		return stringOption(options, "location"), nil
	}
	return boundedGoogleOption(stringOption(options, "location"), f.ServerADCLocation)
}

func (f GoogleProviderProfileFactory) googleQuotaProjectOption(purpose ports.ProviderCredentialPurpose, options map[string]any) string {
	if purpose != ports.ProviderCredentialPurposeServerADC {
		return quotaProjectOption(options)
	}
	bound := strings.TrimSpace(f.ServerADCQuotaProject)
	if bound != "" {
		return bound
	}
	return strings.TrimSpace(f.ServerADCProjectID)
}

func (f GoogleProviderProfileFactory) validateGoogleQuotaProjectOption(purpose ports.ProviderCredentialPurpose, options map[string]any) error {
	if purpose != ports.ProviderCredentialPurposeServerADC {
		return nil
	}
	_, err := boundedGoogleOption(stringOption(options, "quotaProject"), f.googleQuotaProjectOption(purpose, options))
	return err
}

func boundedGoogleOption(profileValue string, boundValue string) (string, error) {
	profileValue = strings.TrimSpace(profileValue)
	boundValue = strings.TrimSpace(boundValue)
	if profileValue != "" && boundValue == "" {
		return "", ports.ErrInvalidProviderInput
	}
	if profileValue != "" && profileValue != boundValue {
		return "", ports.ErrInvalidProviderInput
	}
	if boundValue != "" {
		return boundValue, nil
	}
	return profileValue, nil
}

func httpTimeoutOption(options map[string]any) (time.Duration, error) {
	timeout := stringOption(options, "httpTimeout")
	if timeout == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(timeout)
	if err != nil || parsed <= 0 {
		return 0, ports.ErrInvalidProviderInput
	}
	return parsed, nil
}
