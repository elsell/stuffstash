package voice

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GoogleGeminiConfig struct {
	ProjectID    string
	Location     string
	Model        string
	QuotaProject string
	BaseURL      string
	APIKey       string
	TokenSource  oauth2.TokenSource
	HTTPClient   *http.Client
	HTTPTimeout  time.Duration
}

type GoogleGeminiSpeechToText struct {
	client googleHTTPClient
	path   string
}

func NewGoogleGeminiSpeechToText(cfg GoogleGeminiConfig) GoogleGeminiSpeechToText {
	return GoogleGeminiSpeechToText{
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.HTTPTimeout, cfg.TokenSource, cfg.QuotaProject, cfg.APIKey),
		path:   googleGeminiPath(cfg),
	}
}

func (p GoogleGeminiSpeechToText) Transcribe(ctx context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 || strings.TrimSpace(input.AudioFormat.MimeType) == "" {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	audio := []byte{}
	for _, chunk := range input.AudioChunks {
		audio = append(audio, chunk...)
	}
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{{
			Role: "user",
			Parts: []geminiPart{
				{Text: "Transcribe this audio. Return only the user's spoken words, with no commentary."},
				{InlineData: &geminiInlineData{
					MimeType: input.AudioFormat.MimeType,
					Data:     base64.StdEncoding.EncodeToString(audio),
				}},
			},
		}},
		GenerationConfig: &geminiGenerationConfig{Temperature: 0},
	}
	var response geminiGenerateContentResponse
	if err := p.client.postJSON(ctx, p.path, request, &response); err != nil {
		return ports.SpeechToTextResult{}, err
	}
	transcript := firstGeminiText(response)
	if transcript == "" {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	return ports.SpeechToTextResult{Transcript: transcript}, nil
}

func (p GoogleGeminiSpeechToText) ProbeSpeechToText(ctx context.Context) error {
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: "Provider diagnostic. Reply with the single word ready."}},
		}},
		GenerationConfig: &geminiGenerationConfig{Temperature: 0},
	}
	var response geminiGenerateContentResponse
	if err := p.client.postJSON(ctx, p.path, request, &response); err != nil {
		return err
	}
	if strings.TrimSpace(firstGeminiText(response)) == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

type GoogleGeminiLanguageInference struct {
	client googleHTTPClient
	path   string
}

func NewGoogleGeminiLanguageInference(cfg GoogleGeminiConfig) GoogleGeminiLanguageInference {
	return GoogleGeminiLanguageInference{
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.HTTPTimeout, cfg.TokenSource, cfg.QuotaProject, cfg.APIKey),
		path:   googleGeminiPath(cfg),
	}
}

// NextTurn accepts only the project-owned structured investigation contract.
// Gemini never receives provider-callable inventory tools and cannot author a
// final response, executable command, or action plan through this adapter.
func (p GoogleGeminiLanguageInference) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if input.Investigation == nil || input.Investigation.Validate() != nil {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	prompt := geminiInvestigationPrompt(input)
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{{Role: "user", Parts: []geminiPart{{Text: prompt}}}},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:        0,
			ResponseMimeType:   "application/json",
			ResponseJSONSchema: geminiInvestigationResponseSchema(*input.Investigation),
		},
	}
	var lastErr error
	for attempt := 0; attempt < googleStructuredInferenceAttempts; attempt++ {
		var response geminiGenerateContentResponse
		if err := p.client.postJSON(ctx, p.path, request, &response); err != nil {
			lastErr = err
			if !retryableGoogleLanguageInferenceError(err) || attempt+1 >= googleStructuredInferenceAttempts {
				return ports.LanguageInferenceTurn{}, err
			}
			if err := sleepGoogleLanguageRetry(ctx, attempt, err); err != nil {
				return ports.LanguageInferenceTurn{}, err
			}
			continue
		}

		rawText := firstGeminiText(response)
		turn, err := parseGeminiInvestigationTurn(rawText)
		if err != nil {
			lastErr = err
			if attempt+1 >= googleStructuredInferenceAttempts {
				return ports.LanguageInferenceTurn{}, err
			}
			if err := sleepGoogleLanguageRetry(ctx, attempt, err); err != nil {
				return ports.LanguageInferenceTurn{}, err
			}
			continue
		}
		return turn, nil
	}
	if lastErr != nil {
		return ports.LanguageInferenceTurn{}, lastErr
	}
	return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
}

// ProbeLanguageInference is deliberately separate from NextTurn. Provider
// readiness does not need, and must not reopen, a legacy final-response mode in
// the production language-inference port.
func (p GoogleGeminiLanguageInference) ProbeLanguageInference(ctx context.Context) error {
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: "Provider diagnostic. Reply with the single word ready."}},
		}},
		GenerationConfig: &geminiGenerationConfig{Temperature: 0},
	}
	var response geminiGenerateContentResponse
	if err := p.client.postJSON(ctx, p.path, request, &response); err != nil {
		return err
	}
	if strings.TrimSpace(firstGeminiText(response)) == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiSchema struct {
	Type        string                  `json:"type,omitempty"`
	Description string                  `json:"description,omitempty"`
	Properties  map[string]geminiSchema `json:"properties,omitempty"`
	Required    []string                `json:"required,omitempty"`
	Enum        []string                `json:"enum,omitempty"`
	Items       *geminiSchema           `json:"items,omitempty"`
	AnyOf       []geminiSchema          `json:"anyOf,omitempty"`
	MinItems    int                     `json:"minItems,omitempty"`
	MaxItems    int                     `json:"maxItems,omitempty"`
}

type geminiGenerationConfig struct {
	Temperature        float64       `json:"temperature"`
	ResponseMimeType   string        `json:"responseMimeType,omitempty"`
	ResponseJSONSchema *geminiSchema `json:"responseJsonSchema,omitempty"`
}
