package voice

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

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
	HTTPTimeout  time.Duration
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
		client:       newGoogleHTTPClient(baseURL, cfg.HTTPClient, cfg.HTTPTimeout, cfg.TokenSource, cfg.QuotaProject, ""),
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

func (p GoogleTextToSpeech) ProbeTextToSpeech(ctx context.Context) error {
	result, err := p.Synthesize(ctx, ports.TextToSpeechInput{Text: "Stuff Stash provider test."})
	if err != nil {
		return err
	}
	if len(result.Chunks) == 0 {
		return ports.ErrInvalidProviderInput
	}
	for _, chunk := range result.Chunks {
		if len(chunk) > 0 {
			return nil
		}
	}
	return ports.ErrInvalidProviderInput
}
