package voice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleConversationNativeToolsPreserveSignedHistory(t *testing.T) {
	requests := make(chan map[string]any, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		requests <- body
		contents, _ := body["contents"].([]any)
		if len(contents) == 1 {
			_, _ = w.Write([]byte(`{"candidates":[{"finishReason":"STOP","content":{"role":"model","parts":[{"thought":true,"text":"private reasoning","thoughtSignature":"fixture-signature"},{"functionCall":{"name":"search","args":{"query":"chemicals"}}}]}}]}`))
		} else {
			_, _ = w.Write([]byte(`{"candidates":[{"finishReason":"STOP","content":{"role":"model","parts":[{"text":"Yes, you have chemicals, including Acetone."}]}}]}`))
		}
	}))
	defer server.Close()
	raw := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{BaseURL: server.URL, APIKey: "fixture-key", Model: "fixture-model"})
	provider, ok := any(raw).(ports.ConversationModel)
	if !ok {
		t.Fatal("configured Gemini adapter does not support the model-led conversation port")
	}
	input := ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Do I have any chemicals?"}}, Tools: []ports.ConversationToolDefinition{{Name: "search", Description: "Search authorized inventory names and tags.", Parameters: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`)}}}
	first, err := provider.Converse(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.ToolCalls) != 1 || first.ToolCalls[0].Name != "search" || first.ToolCalls[0].Arguments["query"] != "chemicals" || first.ToolCalls[0].ID == "" {
		t.Fatalf("native tool call lost: %+v", first.ToolCalls)
	}
	if strings.Contains(first.Text, "private reasoning") {
		t.Fatal("thought text exposed as speech")
	}
	input.Messages = append(input.Messages,
		ports.ConversationMessage{Role: ports.ConversationRoleAssistant, Text: first.Text, ToolCalls: first.ToolCalls, ProviderState: first.ProviderState},
		ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: first.ToolCalls[0].ID, Name: "search", Content: `{"items":[{"title":"Acetone"}]}`}}},
	)
	final, err := provider.Converse(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if final.Text != "Yes, you have chemicals, including Acetone." {
		t.Fatalf("natural answer lost: %q", final.Text)
	}
	firstRequest := <-requests
	if _, ok := firstRequest["tools"]; !ok {
		t.Fatal("no native function declarations")
	}
	secondRequest := <-requests
	contents := secondRequest["contents"].([]any)
	parts := contents[1].(map[string]any)["parts"].([]any)
	if parts[0].(map[string]any)["thoughtSignature"] != "fixture-signature" || parts[1].(map[string]any)["functionCall"] == nil {
		t.Fatal("original signed parts lost or reordered")
	}
	if contents[2].(map[string]any)["parts"].([]any)[0].(map[string]any)["functionResponse"] == nil {
		t.Fatal("tool result was not sent as native function response")
	}
}

func TestGoogleConversationPreservesProviderToolResponseID(t *testing.T) {
	request, err := googleConversationRequest(ports.ConversationModelInput{Messages: []ports.ConversationMessage{
		{Role: ports.ConversationRoleUser, Text: "Find clothes."},
		{Role: ports.ConversationRoleAssistant, ToolCalls: []ports.AgentToolCall{{ID: "native-id", Name: "search"}}, ProviderState: []byte(`{"role":"model","parts":[{"functionCall":{"id":"native-id","name":"search","args":{}}}]}`)},
		{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: "native-id", Name: "search", Content: `{"items":[]}`}}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var part map[string]any
	if err := json.Unmarshal(request.Contents[2].Parts[0], &part); err != nil {
		t.Fatal(err)
	}
	if part["functionResponse"].(map[string]any)["id"] != "native-id" {
		t.Fatal("provider tool-call ID lost")
	}
}

func TestGoogleConversationRejectsIncompleteAndOversizedResponses(t *testing.T) {
	for _, tc := range []struct{ name, reason, text string }{
		{"partial", "MAX_TOKENS", "You have"},
		{"oversized continuation", "STOP", strings.Repeat("x", 300*1024)},
		{"oversized response", "STOP", strings.Repeat("x", 2*1024*1024)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"finishReason": tc.reason, "content": map[string]any{"role": "model", "parts": []any{map[string]any{"text": tc.text}}}}}})
			}))
			defer server.Close()
			provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{BaseURL: server.URL, APIKey: "fixture-key", Model: "fixture-model"})
			if _, err := provider.Converse(context.Background(), ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Do I have chemicals?"}}}); err == nil {
				t.Fatal("incomplete or oversized response accepted")
			}
		})
	}
}
func TestGoogleConversationRejectsOversizedHistory(t *testing.T) {
	_, err := googleConversationRequest(ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleAssistant, ProviderState: []byte(`{"role":"model","parts":[{"text":"` + strings.Repeat("x", 300*1024) + `"}]}`)}}})
	if err == nil {
		t.Fatal("oversized continuation accepted")
	}
}
