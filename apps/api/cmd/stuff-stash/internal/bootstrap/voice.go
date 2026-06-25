package bootstrap

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

func buildRealtimeVoiceProviders(ctx context.Context, cfg config.Config) (ports.SpeechToTextProvider, ports.LanguageInferenceProvider, ports.TextToSpeechProvider, error) {
	if cfg.VoiceGoogleEnabled {
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
	geminiConfig := voice.GoogleGeminiConfig{
		ProjectID:   strings.TrimSpace(cfg.GoogleCloudProject),
		Location:    cfg.GoogleCloudLocation,
		Model:       cfg.GoogleGeminiModel,
		TokenSource: tokenSource,
	}
	return voice.NewGoogleGeminiSpeechToText(geminiConfig),
		voice.NewGoogleGeminiLanguageInference(geminiConfig),
		voice.NewGoogleTextToSpeech(voice.GoogleTextToSpeechConfig{
			LanguageCode: cfg.GoogleTTSLanguageCode,
			VoiceName:    cfg.GoogleTTSVoiceName,
			TokenSource:  tokenSource,
		}), nil
}
