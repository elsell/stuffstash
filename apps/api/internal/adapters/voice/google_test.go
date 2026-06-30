package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiSpeechToTextTranscribesInlineAudio(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-User-Project") != "project" {
			t.Fatalf("missing quota project: %q", r.Header.Get("X-Goog-User-Project"))
		}
		if !strings.Contains(r.URL.Path, "/publishers/google/models/gemini-test:generateContent") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if !strings.Contains(string(payload), "audio/mp4") || !strings.Contains(string(payload), base64.StdEncoding.EncodeToString([]byte("audio"))) {
			t.Fatalf("request did not include inline audio: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]any{{"text": "Where are my tools?"}},
				},
			}},
		})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		ProjectID:    "project",
		Location:     "us-central1",
		Model:        "gemini-test",
		QuotaProject: "project",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})

	result, err := provider.Transcribe(context.Background(), ports.SpeechToTextInput{
		AudioFormat: ports.RealtimeAudioFormat{MimeType: "audio/mp4"},
		AudioChunks: [][]byte{[]byte("audio")},
	})
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}
	if result.Transcript != "Where are my tools?" {
		t.Fatalf("unexpected transcript %q", result.Transcript)
	}
}

func TestGoogleGeminiSpeechToTextUsesAPIKeyBackend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("api-key backend must not send bearer authorization: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-Api-Key") != "test-api-key" {
			t.Fatalf("missing api key header: %q", r.Header.Get("X-Goog-Api-Key"))
		}
		if r.URL.Path != "/v1beta/models/gemini-test:generateContent" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse("Where are my tools?"))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		Model:      "gemini-test",
		BaseURL:    server.URL,
		APIKey:     "test-api-key",
		HTTPClient: server.Client(),
	})
	result, err := provider.Transcribe(context.Background(), ports.SpeechToTextInput{
		AudioFormat: ports.RealtimeAudioFormat{MimeType: "audio/mp4"},
		AudioChunks: [][]byte{[]byte("audio")},
	})
	if err != nil {
		t.Fatalf("transcribe with api key: %v", err)
	}
	if result.Transcript != "Where are my tools?" {
		t.Fatalf("unexpected transcript %q", result.Transcript)
	}
}

func TestGoogleGeminiSpeechToTextProbeUsesModelEndpointWithoutTenantAudio(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/publishers/google/models/gemini-test:generateContent") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if strings.Contains(string(payload), "inlineData") || !strings.Contains(string(payload), "Provider diagnostic") {
			t.Fatalf("probe request should use safe text-only diagnostic: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse("ready"))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		ProjectID:    "project",
		Location:     "us-central1",
		Model:        "gemini-test",
		QuotaProject: "project",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})

	if err := provider.ProbeSpeechToText(context.Background()); err != nil {
		t.Fatalf("probe speech-to-text: %v", err)
	}
}

func TestGoogleGeminiLanguageInferenceMapsToolAndFinalTurns(t *testing.T) {
	t.Parallel()

	calls := 0
	requests := []map[string]any{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, request)
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_ = json.NewEncoder(w).Encode(geminiFunctionCallResponse("search_authorized_assets", map[string]any{"query": "tools"}))
			return
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Your tools are in Garage.","displayResponse":"Your tools are in Garage."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:    "project",
		Location:     "us-central1",
		Model:        "gemini-test",
		QuotaProject: "project",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})
	tools := []ports.AgentToolDescriptor{{
		Name:        "search_authorized_assets",
		Description: "Search visible assets.",
		ReadOnly:    true,
	}}

	toolTurn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Where are my tools?", Tools: tools})
	if err != nil {
		t.Fatalf("first turn: %v", err)
	}
	if len(toolTurn.ToolCalls) != 1 || toolTurn.ToolCalls[0].Name != "search_authorized_assets" || toolTurn.ToolCalls[0].Arguments["query"] != "tools" {
		t.Fatalf("unexpected tool turn: %+v", toolTurn)
	}

	finalTurn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript: "Where are my tools?",
		Tools:      tools,
		ToolResults: []ports.AgentToolResult{{
			CallID: "call-1",
			Name:   "search_authorized_assets",
			Call: ports.AgentToolCall{
				ID:        "call-1",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "tools"},
			},
			Content: `{"tool":"search_authorized_assets","count":1,"items":[{"title":"Tools","kind":"container","locationTitle":"Garage"}]}`,
		}},
		PreviousTurns: 1,
	})
	if err != nil {
		t.Fatalf("final turn: %v", err)
	}
	if finalTurn.Final == nil || finalTurn.Final.SpokenResponse != "Your tools are in Garage." {
		t.Fatalf("unexpected final turn: %+v", finalTurn)
	}
	requestPayload, _ := json.Marshal(requests[0])
	if !strings.Contains(string(requestPayload), `"functionDeclarations"`) || !strings.Contains(string(requestPayload), "Search visible assets.") {
		t.Fatalf("request did not include native tool declarations: %s", string(requestPayload))
	}
	if strings.Contains(string(requestPayload), `"toolCalls"`) {
		t.Fatalf("request still prompts for text tool-call JSON: %s", string(requestPayload))
	}
	firstConfig := objectAt(t, requests[0], "generationConfig")
	if _, exists := firstConfig["responseMimeType"]; exists {
		t.Fatalf("tool-capable turn should not mix function calling with structured output, got %+v", firstConfig)
	}
	if _, exists := firstConfig["responseSchema"]; exists {
		t.Fatalf("tool-capable turn should not include final response schema, got %+v", firstConfig)
	}
	firstPrompt := requestTextPart(t, requests[0], 0, 0)
	if strings.Contains(firstPrompt, "Tool results:") {
		t.Fatalf("prompt should not inline raw tool results when native function responses are used: %s", firstPrompt)
	}
	secondContents := requestContents(t, requests[1])
	if len(secondContents) != 3 {
		t.Fatalf("expected user prompt, model function call, and user function response contents, got %d: %+v", len(secondContents), secondContents)
	}
	if roleAt(t, secondContents[1]) != "model" || roleAt(t, secondContents[2]) != "user" {
		t.Fatalf("unexpected function calling roles: %+v", secondContents)
	}
	functionCall := partObjectAt(t, secondContents[1], 0, "functionCall")
	if functionCall["name"] != "search_authorized_assets" || objectAt(t, functionCall, "args")["query"] != "tools" {
		t.Fatalf("unexpected function call history: %+v", functionCall)
	}
	functionResponse := partObjectAt(t, secondContents[2], 0, "functionResponse")
	response := objectAt(t, functionResponse, "response")
	if functionResponse["name"] != "search_authorized_assets" || response["tool"] != "search_authorized_assets" || response["count"] != float64(1) {
		t.Fatalf("unexpected structured function response: %+v", functionResponse)
	}
	if !requestHasFunctionDeclaration(requests[1], "search_authorized_assets") {
		t.Fatalf("second turn should keep usable native tool declarations for distinct follow-up calls: %+v", requests[1]["tools"])
	}
	secondConfig := objectAt(t, requests[1], "generationConfig")
	if _, exists := secondConfig["responseMimeType"]; exists {
		t.Fatalf("continuation tool-capable turn should not mix function calling with structured output, got %+v", secondConfig)
	}
	if _, exists := secondConfig["responseSchema"]; exists {
		t.Fatalf("continuation tool-capable turn should not include final response schema, got %+v", secondConfig)
	}
}

func TestGoogleGeminiLanguageInferenceRequestsJSONForFinalOnlyTurns(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		config := objectAt(t, request, "generationConfig")
		if config["responseMimeType"] != "application/json" {
			t.Fatalf("expected final-only turn to request JSON response mime type, got %+v", config)
		}
		if !generationConfigHasFinalResponseSchema(config) {
			t.Fatalf("expected final-only turn to request final response schema, got %+v", config)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Ready.","displayResponse":"Ready."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})
	if _, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true}); err != nil {
		t.Fatalf("final-only turn: %v", err)
	}
}

func TestGoogleGeminiLanguageInferenceCanRequireToolCall(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		toolConfig := objectAt(t, request, "toolConfig")
		functionCalling := objectAt(t, toolConfig, "functionCallingConfig")
		if functionCalling["mode"] != "ANY" {
			t.Fatalf("expected required tool call mode ANY, got %+v", functionCalling)
		}
		names, ok := functionCalling["allowedFunctionNames"].([]any)
		if !ok || len(names) != 1 || names[0] != "search_authorized_assets" {
			t.Fatalf("expected allowed tool names, got %+v", functionCalling)
		}
		config := objectAt(t, request, "generationConfig")
		if _, exists := config["responseMimeType"]; exists {
			t.Fatalf("required function-call turn should not also request textual structured output, got %+v", config)
		}
		if _, exists := config["responseSchema"]; exists {
			t.Fatalf("required function-call turn should not include response schema, got %+v", config)
		}
		_ = json.NewEncoder(w).Encode(geminiFunctionCallResponse("search_authorized_assets", map[string]any{"query": "water bottle"}))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript:      "Where is my water bottle?",
		RequireToolCall: true,
		Tools: []ports.AgentToolDescriptor{{
			Name:        "search_authorized_assets",
			Description: "Search visible assets.",
			ReadOnly:    true,
		}},
	})
	if err != nil {
		t.Fatalf("required tool turn: %v", err)
	}
	if len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "search_authorized_assets" {
		t.Fatalf("expected required tool call, got %+v", turn)
	}
}

func TestGoogleGeminiLanguageInferenceUsesAPIKeyBackend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("api-key backend must not send bearer authorization: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-Api-Key") != "test-api-key" {
			t.Fatalf("missing api key header: %q", r.Header.Get("X-Goog-Api-Key"))
		}
		if r.URL.Path != "/v1beta/models/gemini-test:generateContent" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Ready.","displayResponse":"Ready."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		Model:      "gemini-test",
		BaseURL:    server.URL,
		APIKey:     "test-api-key",
		HTTPClient: server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true})
	if err != nil {
		t.Fatalf("language inference with api key: %v", err)
	}
	if turn.Final == nil || turn.Final.SpokenResponse != "Ready." {
		t.Fatalf("unexpected turn: %+v", turn)
	}
}

func TestGoogleGeminiLanguagePromptIncludesTenantTemplateAndMandatoryRules(t *testing.T) {
	t.Parallel()

	prompt := languagePrompt(ports.LanguageInferenceInput{
		Transcript:     "Where is my water bottle?",
		PromptTemplate: "Prefer concise spoken answers.",
	})

	templateIndex := strings.Index(prompt, "Prefer concise spoken answers.")
	mandatoryIndex := strings.Index(prompt, "Use only the provided native tools for inventory lookup and action-plan proposal.")
	if templateIndex < 0 {
		t.Fatalf("expected tenant prompt template in prompt: %s", prompt)
	}
	if mandatoryIndex < 0 {
		t.Fatalf("expected mandatory server-owned rule in prompt: %s", prompt)
	}
	if templateIndex > mandatoryIndex {
		t.Fatalf("expected mandatory rules to follow tenant template so they cannot be removed: %s", prompt)
	}
	for _, required := range []string{
		"use the returned assetId as parentAssetId",
		"Action-plan command arguments must be structured JSON",
		"For write requests involving an existing asset, call search_authorized_assets for the source asset before propose_action_plan.",
		"Do not call propose_action_plan for moving, archiving, or restoring an existing asset until a read tool result has returned that source asset's assetId.",
		"If a move request names a source item and you do not have that item's assetId from a read tool, call search_authorized_assets for the source item before proposing a plan.",
		"For where-is questions, if search_authorized_assets returns the requested item with locationTitle or containmentPath, answer from that result instead of listing unrelated assets.",
		"containmentPath ends with the returned asset itself; never say an item is inside itself.",
		"Assume the user wants missing named locations, containers, or household surfaces created",
		"do not ask whether to create it; call propose_action_plan",
		"the session is not complete until you either call propose_action_plan or ask a necessary clarification",
		"assetId and parentAssetId must be opaque assetId values copied exactly from successful search_authorized_assets or list_authorized_assets tool results.",
		"Never use titles, lowercase names, or guessed IDs such as water bottle, kitchen, or kitchen-1 as assetId or parentAssetId.",
		"If the destination name is not present as an assetId in tool results, create it in the same commands array and reference it with parentCommandId.",
		"For nested create or move requests, resolve named outer locations and containers as separate search terms",
		"A combined phrase search returning no matches does not prove each path segment is missing",
		"If a tool result contains the requested source asset, do not later say you cannot find that asset.",
		"If propose_action_plan returns an invalid_tool_request error, retry it once with corrected structured arguments instead of giving a final answer.",
		"Never use parentTitle, locationTitle, or raw titles as executable action-plan parent references.",
		"follows the provided response schema",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("expected prompt to include %q, got: %s", required, prompt)
		}
	}
	if strings.Contains(prompt, `{"final"`) {
		t.Fatalf("prompt should not duplicate the provider response schema: %s", prompt)
	}
}

func TestGoogleGeminiLanguageInferenceExposesSafeDiagnostics(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]any{{
						"functionCall": map[string]any{
							"name": "search_authorized_assets",
							"args": map[string]any{"query": "water bottle"},
						},
					}},
				},
			}},
		})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript:         "Move my water bottle to the kitchen.",
		IncludeDiagnostics: true,
		Tools: []ports.AgentToolDescriptor{{
			Name:     "search_authorized_assets",
			ReadOnly: true,
			Parameters: ports.AgentToolParameters{
				Properties: map[string]ports.AgentToolParameter{"query": {Type: ports.AgentToolParameterTypeString}},
			},
		}},
	})
	if err != nil {
		t.Fatalf("language inference: %v", err)
	}
	if len(turn.Diagnostics) != 2 {
		t.Fatalf("expected prompt and model turn diagnostics, got %+v", turn.Diagnostics)
	}
	if turn.Diagnostics[0].Title != "Language prompt (turn 1)" || !strings.Contains(turn.Diagnostics[0].Detail, "Move my water bottle to the kitchen.") {
		t.Fatalf("unexpected prompt diagnostic: %+v", turn.Diagnostics[0])
	}
	if turn.Diagnostics[1].Title != "Language model turn (turn 1)" || !strings.Contains(turn.Diagnostics[1].Detail, "search_authorized_assets") {
		t.Fatalf("unexpected model turn diagnostic: %+v", turn.Diagnostics[1])
	}
}

func TestGoogleGeminiLanguageInferenceElidesRepeatedPromptDiagnostics(t *testing.T) {
	t.Parallel()

	diagnostics := languageInferenceDiagnostics(2, "Full repeated prompt.", `{"final":{"kind":"answer"}}`)
	if len(diagnostics) != 2 {
		t.Fatalf("expected prompt marker and model turn diagnostics, got %+v", diagnostics)
	}
	if diagnostics[0].Title != "Language prompt (turn 3)" || strings.Contains(diagnostics[0].Detail, "Full repeated prompt.") {
		t.Fatalf("expected repeated prompt to be elided, got %+v", diagnostics[0])
	}
	if !strings.Contains(diagnostics[1].Title, "turn 3") || !strings.Contains(diagnostics[1].Detail, "answer") {
		t.Fatalf("expected turn-labeled model diagnostic, got %+v", diagnostics[1])
	}
}

func TestGoogleGeminiLanguageInferenceOmitsDiagnosticsByDefault(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Ready.","displayResponse":"Ready."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Where is my water bottle?", FinalOnly: true})
	if err != nil {
		t.Fatalf("language inference: %v", err)
	}
	if len(turn.Diagnostics) != 0 {
		t.Fatalf("expected diagnostics to be disabled by default, got %+v", turn.Diagnostics)
	}
}

func TestGoogleGeminiToolSchemaSupportsObjectParameters(t *testing.T) {
	t.Parallel()

	tools := geminiTools([]ports.AgentToolDescriptor{{
		Name:     "propose_action_plan",
		ReadOnly: true,
		Parameters: ports.AgentToolParameters{
			Properties: map[string]ports.AgentToolParameter{
				"arguments": {Type: ports.AgentToolParameterTypeObject},
				"commands": {
					Type: ports.AgentToolParameterTypeArray,
					Items: &ports.AgentToolParameter{
						Type:     ports.AgentToolParameterTypeObject,
						Required: []string{"id", "kind", "summary", "arguments"},
						Properties: map[string]ports.AgentToolParameter{
							"id":      {Type: ports.AgentToolParameterTypeString},
							"kind":    {Type: ports.AgentToolParameterTypeString, Enum: []string{"create_asset", "create_location"}},
							"summary": {Type: ports.AgentToolParameterTypeString},
							"arguments": {
								Type: ports.AgentToolParameterTypeObject,
								Properties: map[string]ports.AgentToolParameter{
									"title":           {Type: ports.AgentToolParameterTypeString},
									"parentCommandId": {Type: ports.AgentToolParameterTypeString},
								},
							},
						},
					},
				},
			},
		},
	}})
	if len(tools) != 1 || len(tools[0].FunctionDeclarations) != 1 {
		t.Fatalf("expected gemini tool declaration, got %+v", tools)
	}
	schema := tools[0].FunctionDeclarations[0].Parameters.Properties["arguments"]
	if schema.Type != "object" {
		t.Fatalf("expected object schema for structured arguments, got %+v", schema)
	}
	commands := tools[0].FunctionDeclarations[0].Parameters.Properties["commands"]
	if commands.Type != "array" || commands.Items == nil || commands.Items.Type != "object" {
		t.Fatalf("expected command array with object items, got %+v", commands)
	}
	if strings.Join(commands.Items.Required, ",") != "id,kind,summary,arguments" {
		t.Fatalf("expected command item required fields, got %+v", commands.Items.Required)
	}
	if commands.Items.Properties["kind"].Enum[0] != "create_asset" {
		t.Fatalf("expected nested enum for command kind, got %+v", commands.Items.Properties["kind"])
	}
	arguments := commands.Items.Properties["arguments"]
	if arguments.Type != "object" || arguments.Properties["parentCommandId"].Type != "string" {
		t.Fatalf("expected nested command argument properties, got %+v", arguments)
	}
}

func TestGoogleGeminiLanguageInferenceProbeReturnsStructuredFinalResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if strings.Contains(string(payload), "functionDeclarations") || !strings.Contains(string(payload), "Provider diagnostic") {
			t.Fatalf("probe request should be final-only without tools: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Provider profile test succeeded.","displayResponse":"Provider profile test succeeded."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:    "project",
		Location:     "us-central1",
		Model:        "gemini-test",
		QuotaProject: "project",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})

	if err := provider.ProbeLanguageInference(context.Background()); err != nil {
		t.Fatalf("probe language inference: %v", err)
	}
}

func TestGoogleGeminiLanguageInferenceRejectsMalformedTurns(t *testing.T) {
	t.Parallel()

	tools := []ports.AgentToolDescriptor{{Name: "search_authorized_assets", ReadOnly: true}}
	cases := map[string]string{
		"unknown field":     `{"final":{"kind":"answer","spokenResponse":"ok","displayResponse":"ok","secret":"id"}}`,
		"unknown kind":      `{"final":{"kind":"delete","spokenResponse":"ok","displayResponse":"ok"}}`,
		"mixed turn":        `{"toolCalls":[{"id":"call-1","name":"search_authorized_assets","arguments":{"query":"tools"}}],"final":{"kind":"answer","spokenResponse":"ok","displayResponse":"ok"}}`,
		"unknown tool":      `{"toolCalls":[{"id":"call-1","name":"delete_asset","arguments":{"id":"asset-1"}}]}`,
		"oversized speech":  `{"final":{"kind":"answer","spokenResponse":"` + strings.Repeat("a", 501) + `","displayResponse":"ok"}}`,
		"missing tool args": `{"toolCalls":[{"id":"call-1","name":"search_authorized_assets"}]}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseLanguageTurn(payload, tools, false); err == nil {
				t.Fatalf("expected malformed turn to be rejected")
			}
		})
	}
}

func TestGoogleGeminiLanguageInferenceRejectsMalformedFunctionCalls(t *testing.T) {
	t.Parallel()

	tools := []ports.AgentToolDescriptor{{Name: "search_authorized_assets", ReadOnly: true}, {Name: "write_asset", ReadOnly: false}}
	cases := map[string][]geminiFunctionCall{
		"unknown tool":  {{Name: "delete_asset", Args: map[string]any{"id": "asset-1"}}},
		"non read only": {{Name: "write_asset", Args: map[string]any{"title": "x"}}},
		"missing args":  {{Name: "search_authorized_assets"}},
	}
	for name, calls := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseGeminiFunctionCalls(calls, tools); err == nil {
				t.Fatalf("expected malformed function call to be rejected")
			}
		})
	}
}

func TestGoogleGeminiLanguageInferenceRejectsMalformedFunctionResponses(t *testing.T) {
	t.Parallel()

	if _, err := languageContents(ports.LanguageInferenceInput{
		Transcript: "Where are my tools?",
		ToolResults: []ports.AgentToolResult{{
			Name:    "search_authorized_assets",
			Content: "not-json",
		}},
	}); err == nil {
		t.Fatalf("expected malformed function response content to be rejected")
	}
}

func TestGoogleGeminiLanguageInferenceDoesNotReplayRejectedToolArguments(t *testing.T) {
	t.Parallel()

	contents, err := languageContents(ports.LanguageInferenceInput{
		Transcript: "Add an item.",
		ToolResults: []ports.AgentToolResult{{
			CallID: "bad-call",
			Name:   "propose_action_plan",
			Call: ports.AgentToolCall{
				ID:        "bad-call",
				Name:      "propose_action_plan",
				Arguments: map[string]any{},
			},
			Content: `{"tool":"propose_action_plan","status":"error","code":"invalid_tool_request","message":"The tool request was invalid or incomplete.","retryable":true}`,
		}},
	})
	if err != nil {
		t.Fatalf("build language contents: %v", err)
	}
	payload, err := json.Marshal(contents)
	if err != nil {
		t.Fatalf("marshal contents: %v", err)
	}
	if strings.Contains(string(payload), "apiKey") || strings.Contains(string(payload), "secret") {
		t.Fatalf("rejected tool arguments leaked into provider contents: %s", string(payload))
	}
	functionCall := contents[1].Parts[0].FunctionCall
	if functionCall == nil {
		t.Fatalf("expected function call history, got %+v", contents[1])
	}
	if len(functionCall.Args) != 0 {
		t.Fatalf("expected sanitized empty function call args, got %+v", functionCall.Args)
	}
}

func TestGoogleTextToSpeechSynthesizesMP3(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-User-Project") != "project" {
			t.Fatalf("missing quota project: %q", r.Header.Get("X-Goog-User-Project"))
		}
		if r.URL.Path != "/v1/text:synthesize" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if !strings.Contains(string(payload), "Your tools are in Garage.") || !strings.Contains(string(payload), "MP3") {
			t.Fatalf("unexpected synthesize request: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"audioContent": base64.StdEncoding.EncodeToString([]byte("mp3-bytes")),
		})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: "en-US",
		VoiceName:    "en-US-Neural2-F",
		QuotaProject: "project",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})
	result, err := provider.Synthesize(context.Background(), ports.TextToSpeechInput{Text: "Your tools are in Garage."})
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}
	if result.MimeType != "audio/mpeg" || string(result.Chunks[0]) != "mp3-bytes" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGoogleTextToSpeechProbeSynthesizesSafeDiagnosticPhrase(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/text:synthesize" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if !strings.Contains(string(payload), "Stuff Stash provider test.") {
			t.Fatalf("probe request should synthesize safe diagnostic phrase: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"audioContent": base64.StdEncoding.EncodeToString([]byte("mp3-bytes")),
		})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: "en-US",
		VoiceName:    "en-US-Neural2-F",
		QuotaProject: "project",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})
	if err := provider.ProbeTextToSpeech(context.Background()); err != nil {
		t.Fatalf("probe text-to-speech: %v", err)
	}
}

func geminiTextResponse(text string) map[string]any {
	return map[string]any{
		"candidates": []map[string]any{{
			"content": map[string]any{
				"parts": []map[string]any{{"text": text}},
			},
		}},
	}
}

func requestContents(t *testing.T, request map[string]any) []any {
	t.Helper()
	contents, ok := request["contents"].([]any)
	if !ok {
		t.Fatalf("request contents missing or wrong type: %+v", request)
	}
	return contents
}

func requestTextPart(t *testing.T, request map[string]any, contentIndex int, partIndex int) string {
	t.Helper()
	content := objectFromAny(t, requestContents(t, request)[contentIndex])
	parts, ok := content["parts"].([]any)
	if !ok {
		t.Fatalf("content parts missing or wrong type: %+v", content)
	}
	part := objectFromAny(t, parts[partIndex])
	text, ok := part["text"].(string)
	if !ok {
		t.Fatalf("text part missing or wrong type: %+v", part)
	}
	return text
}
