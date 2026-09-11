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

func TestGoogleConversationTranslatesBoundedUnionArrayWithoutChangingToolContract(t *testing.T) {
	parameters := json.RawMessage(`{"type":"object","properties":{"commands":{"type":"array","minItems":1,"maxItems":10,"items":{"anyOf":[{"type":"object","properties":{"kind":{"type":"string","enum":["create"]},"title":{"type":"string"}},"required":["kind","title"],"additionalProperties":false},{"type":"object","properties":{"kind":{"type":"string","enum":["move"]},"assetId":{"type":"string"}},"required":["kind","assetId"],"additionalProperties":false}]}},"maxItems":{"type":"integer"},"notes":{"type":"array","maxItems":3,"items":{"type":"string"}}},"required":["commands"],"additionalProperties":false}`)
	original := string(parameters)
	requestSchemas := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Tools []struct {
				FunctionDeclarations []struct {
					Schema map[string]any `json:"parametersJsonSchema"`
				} `json:"functionDeclarations"`
			} `json:"tools"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		// Controlled provider implements the observed rejection of bounded union arrays.
		encoded, _ := json.Marshal(request.Tools[0].FunctionDeclarations[0].Schema)
		if strings.Contains(string(encoded), `"maxItems":10`) {
			http.Error(w, "schema complexity", 400)
			return
		}
		requestSchemas <- request.Tools[0].FunctionDeclarations[0].Schema
		_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"finishReason": "STOP", "content": map[string]any{"role": "model", "parts": []any{map[string]any{"functionCall": map[string]any{"name": "propose", "args": map[string]any{"commands": []any{map[string]any{"kind": "create", "title": "Toolbox"}}}}}}}}}})
	}))
	defer server.Close()
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{BaseURL: server.URL, APIKey: "fixture-key", Model: "fixture-model"})
	input := ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Create a toolbox."}}, Tools: []ports.ConversationToolDefinition{{Name: "propose", Parameters: parameters, ResponseTool: true}}}
	result, err := provider.Converse(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ToolCalls) != 1 || result.ToolCalls[0].Name != "propose" {
		t.Fatalf("tool response changed: %+v", result)
	}
	if string(parameters) != original {
		t.Fatal("shared catalog mutated")
	}
	args := <-requestSchemas
	actualProps := args["properties"].(map[string]any)
	commands := actualProps["commands"].(map[string]any)
	if !strings.Contains(commands["description"].(string), "10") {
		t.Fatal("maximum lost from provider guidance")
	}
	delete(commands, "description")
	commands["maxItems"] = float64(10)
	restored, _ := json.Marshal(args)
	var wanted map[string]any
	_ = json.Unmarshal(parameters, &wanted)
	expected, _ := json.Marshal(wanted)
	if string(restored) != string(expected) {
		t.Fatalf("argument shape or unrelated bounds changed: %s", restored)
	}
}
