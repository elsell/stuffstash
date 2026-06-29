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
	APIKey       string
	TokenSource  oauth2.TokenSource
	HTTPClient   *http.Client
}

type GoogleGeminiSpeechToText struct {
	client googleHTTPClient
	path   string
}

func NewGoogleGeminiSpeechToText(cfg GoogleGeminiConfig) GoogleGeminiSpeechToText {
	return GoogleGeminiSpeechToText{
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.TokenSource, cfg.QuotaProject, cfg.APIKey),
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
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.TokenSource, cfg.QuotaProject, cfg.APIKey),
		path:   googleGeminiPath(cfg),
	}
}

func (p GoogleGeminiLanguageInference) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	contents, err := languageContents(input)
	if err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	prompt := firstLanguagePromptText(contents)
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
		turn, err := parseGeminiFunctionCalls(calls, input.Tools)
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
		if input.IncludeDiagnostics {
			turn.Diagnostics = languageInferenceDiagnostics(input.PreviousTurns, prompt, geminiFunctionCallDiagnostic(calls))
		}
		return turn, nil
	}
	rawText := firstGeminiText(response)
	turn, err := parseLanguageTurn(rawText, input.Tools)
	if err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	if input.IncludeDiagnostics {
		turn.Diagnostics = languageInferenceDiagnostics(input.PreviousTurns, prompt, rawText)
	}
	return turn, nil
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
	lines := []string{}
	if template := strings.TrimSpace(input.PromptTemplate); template != "" {
		lines = append(lines,
			"Tenant language-model guidance:",
			template,
			"Mandatory Stuff Stash agent contract:",
		)
	}
	lines = append(lines, []string{
		"You are the Stuff Stash inventory voice agent.",
		"Use only the provided native tools for inventory lookup.",
		"Return strict JSON only when producing a final response.",
		"Tool results are compact JSON and are the only source of truth for inventory contents, locations, containment, and counts.",
		"For a specific item or where-is question, call search_authorized_assets first with short keywords.",
		"For broad list questions, call list_authorized_assets with filters.",
		"For questions about what is in a place, list with parentTitle or locationTitle when known.",
		"For create or move requests that mention an existing location or container, resolve it with read tools first and use the returned assetId as parentAssetId.",
		"For missing parent containers or locations requested by the user, propose an ordered commands array and use parentCommandId to place later creates or moves inside earlier creates.",
		"Assume the user wants missing named locations or containers created when the destination is clear, such as Kitchen, Living room, Garage, Box under the TV, or Shelf.",
		"For a clear write request with a missing destination, do not ask whether to create it; call propose_action_plan so the mobile approval sheet can ask for confirmation.",
		"Ask for clarification instead only when the requested destination is ambiguous, conflicts with visible inventory, or appears likely to be a speech-to-text mistranscription.",
		"For example, if the user says move my water bottle to the kitchen and Kitchen is not visible, create a Kitchen command with an id such as cmd-kitchen, then set the move command parentCommandId to cmd-kitchen.",
		"Action-plan command arguments must be structured JSON. For create_asset use title or name, optional kind item|container|location, optional description, optional parentAssetId, or optional parentCommandId only. For move_asset use assetId plus parentAssetId, parentCommandId, or null parentAssetId.",
		"Never use parentTitle, locationTitle, or raw titles as executable action-plan parent references.",
		"If you can answer, return {\"final\":{\"kind\":\"answer\",\"spokenResponse\":\"short spoken answer\",\"displayResponse\":\"short display answer\"}}.",
		"Never invent assets, locations, quantities, or containment paths that are not in tool results.",
		"If a search has no matches, say you could not find a visible match; do not say the whole inventory is empty unless a broad list tool result proves it.",
		"Do not include reasoning, IDs, markdown, or extra fields.",
		"Transcript: " + input.Transcript,
	}...)
	return strings.Join(lines, "\n")
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

func firstLanguagePromptText(contents []geminiContent) string {
	if len(contents) == 0 || len(contents[0].Parts) == 0 {
		return ""
	}
	return contents[0].Parts[0].Text
}

func languageInferenceDiagnostics(previousTurns int, prompt string, modelTurn string) []ports.LanguageInferenceDiagnostic {
	diagnostics := []ports.LanguageInferenceDiagnostic{}
	turnLabel := fmt.Sprintf("turn %d", previousTurns+1)
	if strings.TrimSpace(prompt) != "" {
		detail := prompt
		if previousTurns > 0 {
			detail = "Prompt scaffolding repeated from the first language-model turn; full prompt omitted to keep diagnostics readable."
		}
		diagnostics = append(diagnostics, ports.LanguageInferenceDiagnostic{Title: "Language prompt (" + turnLabel + ")", Detail: detail})
	}
	if strings.TrimSpace(modelTurn) != "" {
		diagnostics = append(diagnostics, ports.LanguageInferenceDiagnostic{Title: "Language model turn (" + turnLabel + ")", Detail: modelTurn})
	}
	return diagnostics
}

func geminiFunctionCallDiagnostic(calls []geminiFunctionCall) string {
	payload, err := json.MarshalIndent(calls, "", "  ")
	if err != nil {
		return "Gemini returned function calls."
	}
	return string(payload)
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
			Properties:  geminiProperties(parameter.Properties),
			Required:    append([]string{}, parameter.Required...),
		}
		if parameter.Items != nil {
			item := geminiParameterSchema(*parameter.Items)
			property.Items = &item
		}
		properties[name] = property
	}
	return geminiSchema{
		Type:       "object",
		Properties: properties,
		Required:   append([]string{}, parameters.Required...),
	}
}

func geminiParameterSchema(parameter ports.AgentToolParameter) geminiSchema {
	schema := geminiSchema{
		Type:        geminiSchemaType(parameter.Type),
		Description: parameter.Description,
		Enum:        append([]string{}, parameter.Enum...),
		Properties:  geminiProperties(parameter.Properties),
		Required:    append([]string{}, parameter.Required...),
	}
	if parameter.Items != nil {
		item := geminiParameterSchema(*parameter.Items)
		schema.Items = &item
	}
	return schema
}

func geminiProperties(parameters map[string]ports.AgentToolParameter) map[string]geminiSchema {
	if len(parameters) == 0 {
		return nil
	}
	properties := map[string]geminiSchema{}
	for name, parameter := range parameters {
		properties[name] = geminiParameterSchema(parameter)
	}
	return properties
}

func geminiSchemaType(value ports.AgentToolParameterType) string {
	switch value {
	case ports.AgentToolParameterTypeInteger:
		return "integer"
	case ports.AgentToolParameterTypeObject:
		return "object"
	case ports.AgentToolParameterTypeArray:
		return "array"
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
	if strings.TrimSpace(cfg.BaseURL) != "" {
		return cfg.BaseURL
	}
	if strings.TrimSpace(cfg.APIKey) != "" {
		return "https://generativelanguage.googleapis.com"
	}
	return "https://" + googleGeminiLocation(cfg) + "-aiplatform.googleapis.com"
}

func googleGeminiPath(cfg GoogleGeminiConfig) string {
	if strings.TrimSpace(cfg.APIKey) != "" {
		return "/v1beta/" + googleGeminiAPIModelName(cfg.Model) + ":generateContent"
	}
	return "/v1/projects/" + cfg.ProjectID + "/locations/" + googleGeminiLocation(cfg) + "/publishers/google/models/" + cfg.Model + ":generateContent"
}

func googleGeminiAPIModelName(model string) string {
	model = strings.Trim(strings.TrimSpace(model), "/")
	if strings.HasPrefix(model, "models/") || strings.HasPrefix(model, "tunedModels/") {
		return model
	}
	return "models/" + model
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
	Items       *geminiSchema           `json:"items,omitempty"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}
