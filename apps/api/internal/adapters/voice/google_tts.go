package voice

import (
	"context"
	"encoding/base64"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GoogleTextToSpeechConfig struct {
	LanguageCode string
	VoiceName    string
	QuotaProject string
	BaseURL      string
	TokenSource  oauth2.TokenSource
	HTTPClient   *http.Client
}

type GoogleTextToSpeech struct {
	client       googleHTTPClient
	languageCode string
	voiceName    string
}

func NewGoogleTextToSpeech(cfg GoogleTextToSpeechConfig) GoogleTextToSpeech {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://texttospeech.googleapis.com"
	}
	return GoogleTextToSpeech{
		client:       newGoogleHTTPClient(baseURL, cfg.HTTPClient, cfg.TokenSource, cfg.QuotaProject),
		languageCode: cfg.LanguageCode,
		voiceName:    cfg.VoiceName,
	}
}

func (p GoogleTextToSpeech) Synthesize(ctx context.Context, input ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	if input.Text == "" {
		return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
	}
	var response struct {
		AudioContent string `json:"audioContent"`
	}
	request := map[string]any{
		"input": map[string]any{"text": input.Text},
		"voice": map[string]any{
			"languageCode": p.languageCode,
			"name":         p.voiceName,
		},
		"audioConfig": map[string]any{"audioEncoding": "MP3"},
	}
	if err := p.client.postJSON(ctx, "/v1/text:synthesize", request, &response); err != nil {
		return ports.TextToSpeechResult{}, err
	}
	audio, err := base64.StdEncoding.DecodeString(response.AudioContent)
	if err != nil || len(audio) == 0 {
		return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
	}
	return ports.TextToSpeechResult{
		MimeType: "audio/mpeg",
		Chunks:   [][]byte{audio},
	}, nil
}
