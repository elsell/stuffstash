package voice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GoogleGeminiConfig struct {
	ProjectID    string
	Location     string
	Model        string
	QuotaProject string
	BaseURL      string
	TokenSource  oauth2.TokenSource
	HTTPClient   *http.Client
}

type GoogleGeminiSpeechToText struct {
	client googleHTTPClient
	path   string
}

func NewGoogleGeminiSpeechToText(cfg GoogleGeminiConfig) GoogleGeminiSpeechToText {
	return GoogleGeminiSpeechToText{
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.TokenSource, cfg.QuotaProject),
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

func (p GoogleGeminiSpeechToText) ProbeSpeechToText(ctx context.Context) error {
	request := geminiGenerateContentRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: "Provider diagnostic. Reply with the single word ready."}},
		}},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: 0,
		},
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
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.TokenSource, cfg.QuotaProject),
		path:   googleGeminiPath(cfg),
	}
}

func (p GoogleGeminiLanguageInference) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	contents, err := languageContents(input)
	if err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	request := geminiGenerateContentRequest{
		Contents: contents,
		Tools:    geminiToolsForTurn(input),
		GenerationConfig: &geminiGenerationConfig{
			Temperature:      0,
			ResponseMimeType: "application/json",
		},
	}
	var response geminiGenerateContentResponse
	if err := p.client.postJSON(ctx, p.path, request, &response); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	if calls := geminiFunctionCalls(response); len(calls) > 0 {
		return parseGeminiFunctionCalls(calls, input.Tools)
	}
	return parseLanguageTurn(firstGeminiText(response), input.Tools)
}

func (p GoogleGeminiLanguageInference) ProbeLanguageInference(ctx context.Context) error {
	turn, err := p.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript: "Provider diagnostic. Return a final answer that says Provider profile test succeeded.",
		FinalOnly:  true,
	})
	if err != nil {
		return err
	}
	if turn.Final == nil || strings.TrimSpace(turn.Final.SpokenResponse) == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func geminiToolsForTurn(input ports.LanguageInferenceInput) []geminiTool {
	if input.FinalOnly {
		return nil
	}
	return geminiTools(input.Tools)
}

func languagePrompt(input ports.LanguageInferenceInput) string {
	return strings.Join([]string{
		"You are the Stuff Stash inventory voice agent.",
		"Use only the provided native tools for inventory lookup.",
		"Return strict JSON only when producing a final response.",
		"Tool results are compact JSON and are the only source of truth for inventory contents, locations, containment, and counts.",
		"For a specific item or where-is question, call search_authorized_assets first with short keywords.",
		"For broad list questions, call list_authorized_assets with filters.",
		"For questions about what is in a place, list with parentTitle or locationTitle when known.",
		"If you can answer, return {\"final\":{\"kind\":\"answer\",\"spokenResponse\":\"short spoken answer\",\"displayResponse\":\"short display answer\"}}.",
		"Never invent assets, locations, quantities, or containment paths that are not in tool results.",
		"If a search has no matches, say you could not find a visible match; do not say the whole inventory is empty unless a broad list tool result proves it.",
		"Do not include reasoning, IDs, markdown, or extra fields.",
		"Transcript: " + input.Transcript,
	}, "\n")
}

func languageContents(input ports.LanguageInferenceInput) ([]geminiContent, error) {
	contents := []geminiContent{{
		Role:  "user",
		Parts: []geminiPart{{Text: languagePrompt(input)}},
	}}
	for _, result := range input.ToolResults {
		if strings.TrimSpace(result.Name) == "" || strings.TrimSpace(result.Content) == "" {
			continue
		}
		callName := strings.TrimSpace(result.Call.Name)
		if callName == "" {
			callName = result.Name
		}
		if callName != "" && result.Call.Arguments != nil {
			contents = append(contents, geminiContent{
				Role: "model",
				Parts: []geminiPart{{
					FunctionCall: &geminiFunctionCall{
						Name: callName,
						Args: result.Call.Arguments,
					},
				}},
			})
		}
		response, err := geminiFunctionResponsePayload(result.Content)
		if err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		contents = append(contents, geminiContent{
			Role: "user",
			Parts: []geminiPart{{
				FunctionResponse: &geminiFunctionResponse{
					Name:     callName,
					Response: response,
				},
			}},
		})
	}
	return contents, nil
}

func geminiFunctionResponsePayload(content string) (map[string]any, error) {
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader([]byte(content)))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, ports.ErrInvalidProviderInput
	}
	return payload, nil
}

func geminiTools(tools []ports.AgentToolDescriptor) []geminiTool {
	declarations := make([]geminiFunctionDeclaration, 0, len(tools))
	for _, tool := range tools {
		if !tool.ReadOnly {
			continue
		}
		declarations = append(declarations, geminiFunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  geminiParameters(tool.Parameters),
		})
	}
	if len(declarations) == 0 {
		return nil
	}
	return []geminiTool{{FunctionDeclarations: declarations}}
}

func geminiParameters(parameters ports.AgentToolParameters) geminiSchema {
	properties := map[string]geminiSchema{}
	for name, parameter := range parameters.Properties {
		property := geminiSchema{
			Type:        geminiSchemaType(parameter.Type),
			Description: parameter.Description,
			Enum:        append([]string{}, parameter.Enum...),
		}
		properties[name] = property
	}
	return geminiSchema{
		Type:       "object",
		Properties: properties,
		Required:   append([]string{}, parameters.Required...),
	}
}

func geminiSchemaType(value ports.AgentToolParameterType) string {
	switch value {
	case ports.AgentToolParameterTypeInteger:
		return "integer"
	default:
		return "string"
	}
}

func parseGeminiFunctionCalls(calls []geminiFunctionCall, tools []ports.AgentToolDescriptor) (ports.LanguageInferenceTurn, error) {
	allowedTools := allowedToolNames(tools)
	toolCalls := make([]ports.AgentToolCall, 0, len(calls))
	for index, call := range calls {
		name := strings.TrimSpace(call.Name)
		if !allowedTools[name] || call.Args == nil {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
		toolCalls = append(toolCalls, ports.AgentToolCall{
			ID:        fmt.Sprintf("gemini-call-%d", index+1),
			Name:      name,
			Arguments: call.Args,
		})
	}
	if len(toolCalls) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.LanguageInferenceTurn{ToolCalls: toolCalls}, nil
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
	Tools            []geminiTool            `json:"tools,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                  `json:"text,omitempty"`
	InlineData       *geminiInlineData       `json:"inlineData,omitempty"`
	FunctionCall     *geminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type geminiFunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Parameters  geminiSchema `json:"parameters"`
}

type geminiSchema struct {
	Type        string                  `json:"type"`
	Description string                  `json:"description,omitempty"`
	Properties  map[string]geminiSchema `json:"properties,omitempty"`
	Required    []string                `json:"required,omitempty"`
	Enum        []string                `json:"enum,omitempty"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}
