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

func responseCatalog() []ports.ConversationToolDefinition {
	return []ports.ConversationToolDefinition{
		{Name: "search", Description: "Search inventory.", Parameters: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"],"additionalProperties":false}`)},
		{Name: "deliver", Description: "Answer naturally.", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"speech":{"type":"string"}},"required":["speech"],"additionalProperties":false}`)},
	}
}

func TestGoogleStructuredConversationKeepsModelToolChoiceAndResults(t *testing.T) {
	requests := make(chan map[string]any, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		requests <- body
		text := `{"toolCalls":[{"name":"search","arguments":{"query":"clothes"}}]}`
		if len(body["contents"].([]any)) > 1 {
			text = `{"toolCalls":[{"name":"deliver","arguments":{"speech":"They are in the hall closet."}}]}`
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"finishReason": "STOP", "content": map[string]any{"role": "model", "parts": []any{map[string]any{"text": text}}}}}})
	}))
	defer server.Close()
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{BaseURL: server.URL, APIKey: "fixture-key", Model: "fixture-model"})
	input := ports.ConversationModelInput{Instructions: "Tenant guidance.", Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Where are my clothes?"}}, Tools: responseCatalog()}
	first, err := provider.Converse(context.Background(), input)
	if err != nil || len(first.ToolCalls) != 1 || first.ToolCalls[0].ID == "" || first.ToolCalls[0].Name != "search" || first.ToolCalls[0].Arguments["query"] != "clothes" {
		t.Fatalf("structured read not translated: %+v %v", first, err)
	}
	call := first.ToolCalls[0]
	input.Messages = append(input.Messages, ports.ConversationMessage{Role: ports.ConversationRoleAssistant, ToolCalls: first.ToolCalls, ProviderState: first.ProviderState}, ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: call.ID, Name: call.Name, Call: call, Content: `{"items":[{"assetId":"clothes-1","parentTitle":"Hall Closet"}]}`}}})
	final, err := provider.Converse(context.Background(), input)
	if err != nil || len(final.ToolCalls) != 1 || final.ToolCalls[0].Name != "deliver" || final.ToolCalls[0].Arguments["speech"] != "They are in the hall closet." {
		t.Fatalf("model answer lost: %+v %v", final, err)
	}
	firstRequest := <-requests
	config := firstRequest["generationConfig"].(map[string]any)
	if config["responseMimeType"] != "application/json" || config["responseJsonSchema"] == nil || firstRequest["tools"] != nil || firstRequest["toolConfig"] != nil {
		t.Fatalf("mixed or unconstrained protocol: %+v", firstRequest)
	}
	secondRequest := <-requests
	encoded, _ := json.Marshal(secondRequest["contents"])
	for _, value := range []string{call.ID, "Hall Closet", "clothes-1", "toolResults"} {
		if !strings.Contains(string(encoded), value) {
			t.Fatalf("missing correlated evidence %q", value)
		}
	}
	if strings.Contains(string(encoded), "functionResponse") {
		t.Fatal("JSON tool results emitted as native function responses")
	}
}

func TestGoogleConversationWithoutResponseToolKeepsNativeTextChoice(t *testing.T) {
	input := ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Hello"}}, Tools: responseCatalog()[:1]}
	request, err := googleConversationRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(request)
	var wire map[string]any
	_ = json.Unmarshal(encoded, &wire)
	config := wire["generationConfig"].(map[string]any)
	if config["responseMimeType"] != nil || wire["toolConfig"] != nil || wire["tools"] == nil {
		t.Fatal("native text/tool choice changed")
	}
}

func TestGoogleEnvelopeSchemaBindsNamesToArgumentShapes(t *testing.T) {
	request, err := googleConversationRequest(ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Find clothes"}}, Tools: responseCatalog()})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(request)
	var wire struct {
		GenerationConfig struct {
			ResponseJSONSchema struct {
				Properties map[string]struct {
					Items struct {
						AnyOf []struct {
							Properties map[string]json.RawMessage `json:"properties"`
						} `json:"anyOf"`
					} `json:"items"`
				} `json:"properties"`
			} `json:"responseJsonSchema"`
		} `json:"generationConfig"`
	}
	if json.Unmarshal(encoded, &wire) != nil {
		t.Fatal("invalid JSON schema")
	}
	choices := wire.GenerationConfig.ResponseJSONSchema.Properties["toolCalls"].Items.AnyOf
	if len(choices) != 2 {
		t.Fatalf("expected distinct tool schemas, got %d", len(choices))
	}
	expected := map[string]string{"search": "query", "deliver": "speech"}
	for _, choice := range choices {
		var name struct {
			Enum []string `json:"enum"`
		}
		var args struct {
			Properties map[string]any `json:"properties"`
			Required   []string       `json:"required"`
		}
		if json.Unmarshal(choice.Properties["name"], &name) != nil || len(name.Enum) != 1 || json.Unmarshal(choice.Properties["arguments"], &args) != nil {
			t.Fatal("invalid tool alternative")
		}
		field, ok := expected[name.Enum[0]]
		if !ok || len(args.Properties) != 1 || args.Properties[field] == nil || len(args.Required) != 1 || args.Required[0] != field {
			t.Fatalf("tool name not bound to its arguments: %s %+v", name.Enum[0], args)
		}
		delete(expected, name.Enum[0])
	}
	if len(expected) != 0 {
		t.Fatal("tool schema missing")
	}
}
