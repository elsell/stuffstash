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
		client: newGoogleHTTPClient(googleGeminiBaseURL(cfg), cfg.HTTPClient, cfg.HTTPTimeout, cfg.TokenSource, cfg.QuotaProject, cfg.APIKey),
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
		Contents:         contents,
		Tools:            geminiToolsForTurn(input),
		ToolConfig:       geminiToolConfigForTurn(input),
		GenerationConfig: geminiGenerationConfigForTurn(input),
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
	turn, err := parseLanguageTurn(rawText, input.Tools, input.PlanOnly)
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
	if input.FinalOnly || input.PlanOnly {
		return nil
	}
	return geminiTools(input.Tools)
}

func geminiToolConfigForTurn(input ports.LanguageInferenceInput) *geminiToolConfig {
	if input.FinalOnly || input.PlanOnly || !input.RequireToolCall || len(input.Tools) == 0 {
		return nil
	}
	names := make([]string, 0, len(input.Tools))
	for _, tool := range input.Tools {
		if googleGeminiProviderCallableTool(tool) {
			names = append(names, tool.Name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return &geminiToolConfig{FunctionCallingConfig: &geminiFunctionCallingConfig{
		Mode:                 "ANY",
		AllowedFunctionNames: names,
	}}
}

func geminiGenerationConfigForTurn(input ports.LanguageInferenceInput) *geminiGenerationConfig {
	config := &geminiGenerationConfig{Temperature: 0}
	if input.PlanOnly {
		config.ResponseMimeType = "application/json"
		config.ResponseSchema = geminiActionPlanResponseSchema()
		return config
	}
	if input.RequireToolCall && !input.FinalOnly {
		return config
	}
	config.ResponseMimeType = "application/json"
	config.ResponseSchema = geminiFinalResponseSchema()
	return config
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
	if input.RequireToolCall && !input.FinalOnly {
		if input.PreviousTurns > 0 {
			lines = append(lines, []string{
				"You are the Stuff Stash inventory voice agent.",
				"This turn must gather missing context with exactly one provided read tool.",
				"For move, put, place, store, or stash requests, search the named destination, outer room, place, or container now. Do not repeat the source-item search unless the source was not found.",
				"Use short search keywords copied from the destination phrase, such as living room, garage, kitchen, cabinet, box, shelf, drawer, or counter.",
				"Do not answer yet and do not propose changes on this turn.",
				"Transcript: " + input.Transcript,
			}...)
			return strings.Join(lines, "\n")
		}
		lines = append(lines, []string{
			"You are the Stuff Stash inventory voice agent.",
			"This turn must gather context with exactly one provided read tool.",
			"Use search_authorized_assets for specific items, where-is questions, resolving named locations or containers, and write requests that mention named inventory things.",
			"For add/create requests into a nested destination, search the outermost room, place, or container separately first, such as living room, garage, office, kitchen, cabinet, box, shelf, drawer, or counter; do not search only the item or the whole destination phrase.",
			"For move/archive/restore requests involving an existing item, this first read turn must search the source item first, not the destination.",
			"Use list_authorized_assets for broad inventory lists or questions about what is inside a known place.",
			"Use short search keywords copied from the transcript. Do not answer yet and do not propose changes on this turn.",
			"Transcript: " + input.Transcript,
		}...)
		return strings.Join(lines, "\n")
	}
	if input.PlanOnly {
		lines = append(lines, []string{
			"You are the Stuff Stash inventory voice planner.",
			"Return strict JSON matching the provided actionPlan response schema.",
			"Do not return a final answer. Do not ask whether to create a clear missing destination; the action plan is the user review step.",
			"Use only tool results as source of truth for existing assets and opaque asset IDs.",
			"For move_asset, assetId must be copied from a read-tool result.",
			"For missing named destinations, create every missing path segment in order and reference earlier create commands with parentCommandId.",
			"Use create_asset with kind container for household containers or surfaces. Use create_location only for true rooms or places.",
			"Use create_asset with kind item for new items. Never include assetId in create_asset arguments.",
			"Transcript: " + input.Transcript,
		}...)
		return strings.Join(lines, "\n")
	}
	lines = append(lines, []string{
		"You are the Stuff Stash inventory voice agent.",
		"Use only the provided native tools for inventory lookup and action-plan proposal.",
		"Return strict JSON only when producing a final response.",
		"Tool results are compact JSON and are the only source of truth for inventory contents, locations, containment, and counts.",
		"For a specific item or where-is question, call search_authorized_assets first with short keywords.",
		"For where-is questions, if search_authorized_assets returns the requested item with locationTitle or containmentPath, answer from that result instead of listing unrelated assets.",
		"containmentPath ends with the returned asset itself; never say an item is inside itself.",
		"For broad list questions, call list_authorized_assets with filters.",
		"For questions about what is in a place, list with parentTitle or locationTitle when known.",
		"For write requests involving an existing asset, call search_authorized_assets for the source asset before propose_action_plan.",
		"Do not call propose_action_plan for moving, archiving, or restoring an existing asset until a read tool result has returned that source asset's assetId.",
		"If a move request names a source item and you do not have that item's assetId from a read tool, call search_authorized_assets for the source item before proposing a plan.",
		"For add/create requests for a new item, use create_asset with title or name and kind item.",
		"Do not invent an assetId for a new item; use move_asset only for an existing asset returned by a read tool.",
		"Never include assetId in create_asset arguments; create_asset creates a new thing and needs title or name plus kind.",
		"When a new item should go inside an existing parent, use one create_asset command with parentAssetId set to the visible parent. Do not create the item and then move it.",
		"For write requests, the session is not complete until you either call propose_action_plan or ask a necessary clarification.",
		"For create or move requests that mention an existing location or container, resolve it with read tools first and use the returned assetId as parentAssetId.",
		"For nested create or move requests, resolve named outer locations and containers as separate search terms before proposing. A combined phrase search returning no matches does not prove each path segment is missing.",
		"If a transcript says a missing container or surface is in a room, search the room before proposing a plan.",
		"If only a combined nested destination phrase has been searched and returned no matches, search a separate outer location or container term before proposing a plan.",
		"For missing parent containers or locations requested by the user, propose an ordered commands array and use parentCommandId to place later creates or moves inside earlier creates.",
		"Assume the user wants missing named locations, containers, or household surfaces created when the destination is clear, such as Kitchen, Living room, Garage, Box under the TV, Big cabinet, Counter, or Second shelf.",
		"For nested missing destinations, create every missing path segment in order, then move or create the requested item into the deepest created command.",
		"For a clear write request with a missing destination, do not ask whether to create it; call propose_action_plan so the mobile approval sheet can ask for confirmation.",
		"Ask for clarification instead only when the requested destination is ambiguous, conflicts with visible inventory, or appears likely to be a speech-to-text mistranscription.",
		"For example, if the user says move my water bottle to the kitchen and Kitchen is not visible, create a Kitchen command with an id such as cmd-kitchen, then set the move command parentCommandId to cmd-kitchen.",
		"For example, if the user says move my water bottle to the second shelf in the big cabinet in the kitchen and none of that path is visible, create Kitchen, create Big cabinet with parentCommandId cmd-kitchen, create Second shelf with parentCommandId cmd-big-cabinet, then move the water bottle with parentCommandId cmd-second-shelf.",
		"For example, if the user adds a new remote to a missing box under an existing Living room, create Box under the TV with kind container and parentAssetId set to Living room's assetId, then create Apple TV remote with kind item and parentCommandId set to the box command id. Do not create the new item first and do not add a move_asset command for the new item.",
		"Action-plan command arguments must be structured JSON. For create_asset use title or name, optional kind item|container|location, optional description, optional parentAssetId, or optional parentCommandId only. For move_asset use assetId plus parentAssetId, parentCommandId, or null parentAssetId.",
		"assetId and parentAssetId must be opaque assetId values copied exactly from successful search_authorized_assets or list_authorized_assets tool results.",
		"Never use titles, lowercase names, or guessed IDs such as water bottle, kitchen, or kitchen-1 as assetId or parentAssetId.",
		"If the destination name is not present as an assetId in tool results, create it in the same commands array and reference it with parentCommandId.",
		"If a missing container belongs inside an existing visible location, create the container with parentAssetId set to that visible location assetId, then create or move the item with parentCommandId set to the container command id.",
		"Use create_asset with kind container for new containers; use create_location only for true locations.",
		"Never use parentTitle, locationTitle, or raw titles as executable action-plan parent references.",
		"If a tool result contains the requested source asset, do not later say you cannot find that asset.",
		"If propose_action_plan returns an invalid_tool_request error, retry it once with corrected structured arguments instead of giving a final answer.",
		"If you can answer without another tool call, return a final response that follows the provided response schema.",
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
		if !googleGeminiProviderCallableTool(tool) {
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

func geminiFinalResponseSchema() *geminiSchema {
	return &geminiSchema{
		Type: "object",
		Properties: map[string]geminiSchema{
			"final": {
				Type: "object",
				Properties: map[string]geminiSchema{
					"kind": {
						Type: "string",
						Enum: []string{
							string(ports.StructuredAgentResponseKindAnswer),
							string(ports.StructuredAgentResponseKindClarification),
							string(ports.StructuredAgentResponseKindUnsupportedAction),
							string(ports.StructuredAgentResponseKindSafeFailure),
						},
					},
					"spokenResponse": {
						Type:        "string",
						Description: "Short user-facing response that will be spoken aloud.",
					},
					"displayResponse": {
						Type:        "string",
						Description: "Short user-facing response for the app to display.",
					},
				},
				Required: []string{"kind", "spokenResponse", "displayResponse"},
			},
		},
		Required: []string{"final"},
	}
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

func parseLanguageTurn(raw string, tools []ports.AgentToolDescriptor, allowActionPlan bool) (ports.LanguageInferenceTurn, error) {
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
	if decoded.Final != nil && (len(decoded.ToolCalls) > 0 || decoded.ActionPlan != nil) {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.ActionPlan != nil && len(decoded.ToolCalls) > 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.ActionPlan != nil && !allowActionPlan {
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
	if decoded.ActionPlan != nil {
		if !validLanguageActionPlan(decoded.ActionPlan) {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
		return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{{
			ID:        "gemini-action-plan",
			Name:      "propose_action_plan",
			Arguments: decoded.ActionPlan,
		}}}, nil
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

func validLanguageActionPlan(plan map[string]any) bool {
	if !boundedNonEmpty(stringValue(plan["intentSummary"]), 500) ||
		!boundedNonEmpty(stringValue(plan["modelInterpretationSummary"]), 1000) ||
		!boundedNonEmpty(stringValue(plan["confirmationSummary"]), 500) {
		return false
	}
	commands, ok := plan["commands"].([]any)
	if !ok || len(commands) == 0 {
		return false
	}
	for _, raw := range commands {
		command, ok := raw.(map[string]any)
		if !ok ||
			!boundedNonEmpty(stringValue(command["id"]), 100) ||
			!allowedLanguageActionPlanCommandKind(stringValue(command["kind"])) ||
			!boundedNonEmpty(stringValue(command["summary"]), 500) {
			return false
		}
		if _, ok := command["arguments"].(map[string]any); !ok {
			return false
		}
	}
	return true
}

func allowedLanguageActionPlanCommandKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "create_asset", "create_location", "move_asset", "archive_asset", "restore_asset":
		return true
	default:
		return false
	}
}

type languageTurnJSON struct {
	ToolCalls  []languageToolCallJSON `json:"toolCalls,omitempty"`
	Final      *languageFinalJSON     `json:"final,omitempty"`
	ActionPlan map[string]any         `json:"actionPlan,omitempty"`
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
		if googleGeminiProviderCallableTool(tool) {
			allowed[tool.Name] = true
		}
	}
	return allowed
}

func googleGeminiProviderCallableTool(tool ports.AgentToolDescriptor) bool {
	return tool.ProviderCallable || tool.ReadOnly
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

func stringValue(value any) string {
	text, _ := value.(string)
	return text
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
	ToolConfig       *geminiToolConfig       `json:"toolConfig,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiToolConfig struct {
	FunctionCallingConfig *geminiFunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

type geminiFunctionCallingConfig struct {
	Mode                 string   `json:"mode"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
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
	Temperature      float64       `json:"temperature"`
	ResponseMimeType string        `json:"responseMimeType,omitempty"`
	ResponseSchema   *geminiSchema `json:"responseSchema,omitempty"`
}
