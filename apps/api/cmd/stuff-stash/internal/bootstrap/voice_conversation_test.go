package bootstrap

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestStaticVoiceBootstrapResolvesConversationModel(t *testing.T) {
	model := voice.NewGoogleGeminiLanguageInference(voice.GoogleGeminiConfig{})
	resolver := buildRealtimeVoiceProviderResolver(config.Config{}, repositories{}, nil, voice.DevFakeSpeechToText{}, &model, voice.DevFakeTextToSpeech{})
	resolved, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{})
	if err != nil || resolved.ConversationModel != &model {
		t.Fatalf("configured native model was lost during bootstrap: %v", err)
	}
}

func TestDevelopmentVoiceBootstrapProvidesConversationModel(t *testing.T) {
	stt, model, tts, err := buildRealtimeVoiceProvidersWithTokenSource(config.Config{VoiceDevFakeEnabled: true}, nil)
	if err != nil || stt == nil || model == nil || tts == nil {
		t.Fatalf("development voice configuration: %v", err)
	}
	if _, ok := any(model).(ports.ConversationModel); !ok {
		t.Fatal("development provider still requires the retired investigation engine")
	}
}
