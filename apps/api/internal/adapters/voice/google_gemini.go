package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GoogleGeminiConfig struct {
	ProjectID   string
	Location    string
	Model       string
	BaseURL     string
	TokenSource oauth2.TokenSource
	HTTPClient  *http.Client
}

type GoogleGeminiSpeechToText struct {
	client googleHTTPClient
	path   string
}

func NewGoogleGeminiSpeechToText(cfg GoogleGeminiConfig) GoogleGeminiSpeechToText {
	return GoogleGeminiSpeechToText{
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.TokenSource),
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
		GenerationConfig: &geminiGenerationConfig{
			Temperature: 0,
		},
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

type GoogleGeminiLanguageInference struct {
	client googleHTTPClient
	path   string
}

func NewGoogleGeminiLanguageInference(cfg GoogleGeminiConfig) GoogleGeminiLanguageInference {
	return GoogleGeminiLanguageInference{
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.TokenSource),
		path:   googleGeminiPath(cfg),
	}
}

func (p GoogleGeminiLanguageInference) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: languagePrompt(input)}},
		}},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:      0,
			ResponseMimeType: "application/json",
		},
	}
	var response geminiGenerateContentResponse
	if err := p.client.postJSON(ctx, p.path, request, &response); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	return parseLanguageTurn(firstGeminiText(response))
}

func languagePrompt(input ports.LanguageInferenceInput) string {
	toolResults, _ := json.Marshal(input.ToolResults)
	return strings.Join([]string{
		"You are the Stuff Stash inventory voice agent.",
		"Use only these tools: search_authorized_assets.",
		"Return strict JSON only.",
		"If you need to search, return {\"toolCalls\":[{\"id\":\"call-1\",\"name\":\"search_authorized_assets\",\"arguments\":{\"query\":\"short query\"}}]}.",
		"If you can answer, return {\"final\":{\"kind\":\"answer\",\"spokenResponse\":\"short spoken answer\",\"displayResponse\":\"short display answer\"}}.",
		"Do not include reasoning, hidden IDs, markdown, or extra fields.",
		"Transcript: " + input.Transcript,
		"Tool results: " + string(toolResults),
	}, "\n")
}

func parseLanguageTurn(raw string) (ports.LanguageInferenceTurn, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	var decoded struct {
		ToolCalls []struct {
			ID        string         `json:"id"`
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		} `json:"toolCalls"`
		Final *struct {
			Kind            string `json:"kind"`
			SpokenResponse  string `json:"spokenResponse"`
			DisplayResponse string `json:"displayResponse"`
		} `json:"final"`
	}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	if decoded.Final != nil {
		return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKind(decoded.Final.Kind),
			SpokenResponse:  decoded.Final.SpokenResponse,
			DisplayResponse: decoded.Final.DisplayResponse,
		}}, nil
	}
	toolCalls := make([]ports.AgentToolCall, 0, len(decoded.ToolCalls))
	for _, call := range decoded.ToolCalls {
		toolCalls = append(toolCalls, ports.AgentToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		})
	}
	if len(toolCalls) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.LanguageInferenceTurn{ToolCalls: toolCalls}, nil
}

func googleGeminiBaseURL(cfg GoogleGeminiConfig) string {
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return "https://" + googleGeminiLocation(cfg) + "-aiplatform.googleapis.com"
}

func googleGeminiPath(cfg GoogleGeminiConfig) string {
	return "/v1/projects/" + cfg.ProjectID + "/locations/" + googleGeminiLocation(cfg) + "/publishers/google/models/" + cfg.Model + ":generateContent"
}

func googleGeminiLocation(cfg GoogleGeminiConfig) string {
	location := strings.TrimSpace(cfg.Location)
	if location == "" {
		return "us-central1"
	}
	return location
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

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}
