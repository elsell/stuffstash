package httpserver

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2/google"
)

func liveVoiceRequired(t *testing.T, key string) string {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		t.Fatalf("%s is required for the enabled live test", key)
	}
	return value
}
func liveGoogleVoiceProviders(t *testing.T, ctx context.Context) ports.RealtimeVoiceProviderSet {
	t.Helper()
	project := liveVoiceRequired(t, "STUFF_STASH_GOOGLE_CLOUD_PROJECT")
	tokenSource, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		t.Fatal("ADC unavailable")
	}
	config := voice.GoogleGeminiConfig{ProjectID: project, Location: liveVoiceRequired(t, "STUFF_STASH_GOOGLE_CLOUD_LOCATION"), Model: liveVoiceRequired(t, "STUFF_STASH_GOOGLE_GEMINI_MODEL"), QuotaProject: project, TokenSource: tokenSource, HTTPTimeout: 45 * time.Second, HTTPClient: &http.Client{Transport: liveVoiceTraceTransport{t: t, next: http.DefaultTransport}}}
	return ports.RealtimeVoiceProviderSet{ConversationModel: voice.NewGoogleGeminiLanguageInference(config), SpeechToText: voice.NewGoogleGeminiSpeechToText(config), TextToSpeech: voice.NewGoogleTextToSpeech(voice.GoogleTextToSpeechConfig{LanguageCode: liveVoiceRequired(t, "STUFF_STASH_GOOGLE_TTS_LANGUAGE"), VoiceName: liveVoiceRequired(t, "STUFF_STASH_GOOGLE_TTS_VOICE"), QuotaProject: project, TokenSource: tokenSource, HTTPTimeout: 45 * time.Second})}
}
