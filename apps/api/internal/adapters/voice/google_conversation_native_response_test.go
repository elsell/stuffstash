package voice

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestGoogleResponseToolsUseNativeRequiredFunctionCalls(t *testing.T) {
	input := ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Find item"}}, Tools: []ports.ConversationToolDefinition{{Name: "present_answer", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"display":{"type":"string"}},"required":["display"],"additionalProperties":false}`)}}}
	request, err := googleConversationRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(request)
	var wire map[string]any
	_ = json.Unmarshal(raw, &wire)
	config, ok := wire["toolConfig"].(map[string]any)
	if !ok {
		t.Fatal("missing native tool config")
	}
	if config["functionCallingConfig"].(map[string]any)["mode"] != "ANY" {
		t.Fatal("plain response allowed")
	}
	if wire["generationConfig"].(map[string]any)["responseJsonSchema"] != nil {
		t.Fatal("still using output envelope")
	}
	declaration := wire["tools"].([]any)[0].(map[string]any)["functionDeclarations"].([]any)[0].(map[string]any)
	if declaration["parametersJsonSchema"] == nil || declaration["parameters"] != nil {
		t.Fatal("strict JSON schema not carried by native declaration")
	}
}
