package voice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	return parseLanguageTurn(firstGeminiText(response), input.Tools)
}

func languagePrompt(input ports.LanguageInferenceInput) string {
	toolResults, _ := json.Marshal(input.ToolResults)
	tools := make([]string, 0, len(input.Tools))
	for _, tool := range input.Tools {
		tools = append(tools, fmt.Sprintf("%s: %s Read-only: %t.", tool.Name, tool.Description, tool.ReadOnly))
	}
	return strings.Join([]string{
		"You are the Stuff Stash inventory voice agent.",
		"Use only these tools:",
		strings.Join(tools, "\n"),
		"Return strict JSON only.",
		"If you need to search, return {\"toolCalls\":[{\"id\":\"call-1\",\"name\":\"search_authorized_assets\",\"arguments\":{\"query\":\"short query\"}}]}.",
		"If you can answer, return {\"final\":{\"kind\":\"answer\",\"spokenResponse\":\"short spoken answer\",\"displayResponse\":\"short display answer\"}}.",
		"Do not include reasoning, hidden IDs, markdown, or extra fields.",
		"Transcript: " + input.Transcript,
		"Tool results: " + string(toolResults),
	}, "\n")
}

func parseLanguageTurn(raw string, tools []ports.AgentToolDescriptor) (ports.LanguageInferenceTurn, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	var decoded languageTurnJSON
	decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	if decoder.More() {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.Final != nil && len(decoded.ToolCalls) > 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.Final != nil {
		kind := ports.StructuredAgentResponseKind(decoded.Final.Kind)
		if !isAllowedStructuredResponseKind(kind) ||
			!boundedNonEmpty(decoded.Final.SpokenResponse, 500) ||
			!boundedOptional(decoded.Final.DisplayResponse, 1000) {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
		return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{
			Kind:            kind,
			SpokenResponse:  decoded.Final.SpokenResponse,
			DisplayResponse: decoded.Final.DisplayResponse,
		}}, nil
	}
	allowedTools := allowedToolNames(tools)
	toolCalls := make([]ports.AgentToolCall, 0, len(decoded.ToolCalls))
	for _, call := range decoded.ToolCalls {
		if !boundedNonEmpty(call.ID, 100) || !allowedTools[call.Name] || call.Arguments == nil {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
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

type languageTurnJSON struct {
	ToolCalls []languageToolCallJSON `json:"toolCalls,omitempty"`
	Final     *languageFinalJSON     `json:"final,omitempty"`
}

type languageToolCallJSON struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type languageFinalJSON struct {
	Kind            string `json:"kind"`
	SpokenResponse  string `json:"spokenResponse"`
	DisplayResponse string `json:"displayResponse"`
}

func allowedToolNames(tools []ports.AgentToolDescriptor) map[string]bool {
	allowed := make(map[string]bool, len(tools))
	for _, tool := range tools {
		if tool.ReadOnly {
			allowed[tool.Name] = true
		}
	}
	return allowed
}

func isAllowedStructuredResponseKind(kind ports.StructuredAgentResponseKind) bool {
	switch kind {
	case ports.StructuredAgentResponseKindAnswer,
		ports.StructuredAgentResponseKindClarification,
		ports.StructuredAgentResponseKindUnsupportedAction,
		ports.StructuredAgentResponseKindSafeFailure:
		return true
	default:
		return false
	}
}

func boundedNonEmpty(value string, max int) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && len(trimmed) <= max
}

func boundedOptional(value string, max int) bool {
	return strings.TrimSpace(value) == "" || len(value) <= max
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
