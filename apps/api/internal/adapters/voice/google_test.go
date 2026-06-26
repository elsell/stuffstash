package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"

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
			if _, err := parseLanguageTurn(payload, tools); err == nil {
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

func roleAt(t *testing.T, content any) string {
	t.Helper()
	item := objectFromAny(t, content)
	role, ok := item["role"].(string)
	if !ok {
		t.Fatalf("content role missing or wrong type: %+v", item)
	}
	return role
}

func partObjectAt(t *testing.T, content any, partIndex int, key string) map[string]any {
	t.Helper()
	item := objectFromAny(t, content)
	parts, ok := item["parts"].([]any)
	if !ok {
		t.Fatalf("content parts missing or wrong type: %+v", item)
	}
	part := objectFromAny(t, parts[partIndex])
	return objectAt(t, part, key)
}

func objectAt(t *testing.T, item map[string]any, key string) map[string]any {
	t.Helper()
	return objectFromAny(t, item[key])
}

func objectFromAny(t *testing.T, value any) map[string]any {
	t.Helper()
	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is not an object: %+v", value)
	}
	return item
}

func requestHasFunctionDeclaration(request map[string]any, name string) bool {
	tools, ok := request["tools"].([]any)
	if !ok {
		return false
	}
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]any)
		if !ok {
			continue
		}
		declarations, ok := tool["functionDeclarations"].([]any)
		if !ok {
			continue
		}
		for _, rawDeclaration := range declarations {
			declaration, ok := rawDeclaration.(map[string]any)
			if ok && declaration["name"] == name {
				return true
			}
		}
	}
	return false
}

func geminiFunctionCallResponse(name string, args map[string]any) map[string]any {
	return map[string]any{
		"candidates": []map[string]any{{
			"content": map[string]any{
				"parts": []map[string]any{{
					"functionCall": map[string]any{
						"name": name,
						"args": args,
					},
				}},
			},
		}},
	}
}

type staticTokenSource struct{}

func (staticTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"}, nil
}
