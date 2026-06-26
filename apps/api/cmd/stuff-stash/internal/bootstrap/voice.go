package bootstrap

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

var (
	googleProjectPattern   = regexp.MustCompile(`^[a-z][a-z0-9-]{4,61}[a-z0-9]$`)
	googleLocationPattern  = regexp.MustCompile(`^[a-z0-9-]+$`)
	googleModelPattern     = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	googleLanguagePattern  = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]+)*$`)
	googleVoiceNamePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
)

type staticRealtimeVoiceProviderResolver struct {
	providers ports.RealtimeVoiceProviderSet
}

func (r staticRealtimeVoiceProviderResolver) ResolveRealtimeVoiceProviders(context.Context, ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	return r.providers, nil
}

func buildRealtimeVoiceProviders(ctx context.Context, cfg config.Config) (ports.SpeechToTextProvider, ports.LanguageInferenceProvider, ports.TextToSpeechProvider, error) {
	if cfg.VoiceGoogleEnabled {
		if token := strings.TrimSpace(cfg.GoogleAccessToken); token != "" {
			return buildRealtimeVoiceProvidersWithTokenSource(cfg, oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: token,
				TokenType:   "Bearer",
			}))
		}
		tokenSource, err := google.DefaultTokenSource(ctx, googleCloudPlatformScope)
		if err != nil {
			return nil, nil, nil, err
		}
		return buildRealtimeVoiceProvidersWithTokenSource(cfg, tokenSource)
	}
	if cfg.VoiceDevFakeEnabled {
		return voice.DevFakeSpeechToText{}, voice.DevFakeLanguageInference{}, voice.DevFakeTextToSpeech{}, nil
	}
	return nil, nil, nil, nil
}

func buildRealtimeVoiceProvidersWithTokenSource(cfg config.Config, tokenSource oauth2.TokenSource) (ports.SpeechToTextProvider, ports.LanguageInferenceProvider, ports.TextToSpeechProvider, error) {
	if !cfg.VoiceGoogleEnabled {
		if cfg.VoiceDevFakeEnabled {
			return voice.DevFakeSpeechToText{}, voice.DevFakeLanguageInference{}, voice.DevFakeTextToSpeech{}, nil
		}
		return nil, nil, nil, nil
	}
	if strings.TrimSpace(cfg.GoogleCloudProject) == "" {
		return nil, nil, nil, errors.New("google cloud project is required for realtime voice providers")
	}
	if err := validateGoogleVoiceConfig(cfg); err != nil {
		return nil, nil, nil, err
	}
	geminiConfig := voice.GoogleGeminiConfig{
		ProjectID:    strings.TrimSpace(cfg.GoogleCloudProject),
		Location:     cfg.GoogleCloudLocation,
		Model:        cfg.GoogleGeminiModel,
		QuotaProject: strings.TrimSpace(cfg.GoogleCloudProject),
		TokenSource:  tokenSource,
	}
	return voice.NewGoogleGeminiSpeechToText(geminiConfig),
		voice.NewGoogleGeminiLanguageInference(geminiConfig),
		voice.NewGoogleTextToSpeech(voice.GoogleTextToSpeechConfig{
			LanguageCode: cfg.GoogleTTSLanguageCode,
			VoiceName:    cfg.GoogleTTSVoiceName,
			QuotaProject: strings.TrimSpace(cfg.GoogleCloudProject),
			TokenSource:  tokenSource,
		}), nil
}

func buildRealtimeVoiceProviderResolver(cfg config.Config, repositories repositories, sealer ports.ProviderCredentialSealer, stt ports.SpeechToTextProvider, lm ports.LanguageInferenceProvider, tts ports.TextToSpeechProvider) ports.RealtimeVoiceProviderResolver {
	if stt != nil && lm != nil && tts != nil {
		return staticRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{
			SpeechToText:      stt,
			LanguageInference: lm,
			TextToSpeech:      tts,
		}}
	}
	if repositories.providerProfiles == nil || repositories.providerCredentials == nil || sealer == nil {
		return nil
	}
	return voice.NewProviderProfileResolver(repositories.providerProfiles, repositories.providerCredentials, sealer, voice.GoogleProviderProfileFactory{})
}

func validateGoogleVoiceConfig(cfg config.Config) error {
	fields := []struct {
		name    string
		value   string
		pattern *regexp.Regexp
	}{
		{name: "google cloud project", value: strings.TrimSpace(cfg.GoogleCloudProject), pattern: googleProjectPattern},
		{name: "google cloud location", value: strings.TrimSpace(cfg.GoogleCloudLocation), pattern: googleLocationPattern},
		{name: "google gemini model", value: strings.TrimSpace(cfg.GoogleGeminiModel), pattern: googleModelPattern},
		{name: "google tts language code", value: strings.TrimSpace(cfg.GoogleTTSLanguageCode), pattern: googleLanguagePattern},
		{name: "google tts voice name", value: strings.TrimSpace(cfg.GoogleTTSVoiceName), pattern: googleVoiceNamePattern},
	}
	for _, field := range fields {
		if !field.pattern.MatchString(field.value) {
			return errors.New(field.name + " is invalid for realtime voice providers")
		}
	}
	return nil
}
