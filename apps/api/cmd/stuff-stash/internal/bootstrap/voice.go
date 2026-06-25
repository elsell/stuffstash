package bootstrap

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func buildRealtimeVoiceProviders(cfg config.Config) (ports.SpeechToTextProvider, ports.LanguageInferenceProvider, ports.TextToSpeechProvider) {
	if !cfg.VoiceDevFakeEnabled {
		return nil, nil, nil
	}
	return voice.DevFakeSpeechToText{}, voice.DevFakeLanguageInference{}, voice.DevFakeTextToSpeech{}
}
