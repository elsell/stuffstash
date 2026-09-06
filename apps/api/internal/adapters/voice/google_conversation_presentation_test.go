package voice

import (
	"encoding/json"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleUsesStructuredToolOutputOnlyWithAResponseTool(t *testing.T) {
	for _, responseTool := range []bool{false, true} {
		input := ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Where are my clothes?"}}, Tools: []ports.ConversationToolDefinition{
			{Name: "search", Parameters: json.RawMessage(`{"type":"object"}`)},
			{Name: "deliver", ResponseTool: responseTool, Parameters: json.RawMessage(`{"type":"object","properties":{"speech":{"type":"string"}},"required":["speech"]}`)},
		}}
		request, err := googleConversationRequest(input)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		var wire struct {
			ToolConfig *struct {
				FunctionCallingConfig struct {
					Mode    string   `json:"mode"`
					Allowed []string `json:"allowedFunctionNames"`
				} `json:"functionCallingConfig"`
			} `json:"toolConfig"`
		}
		if json.Unmarshal(data, &wire) != nil {
			t.Fatal("invalid request JSON")
		}
		if responseTool {
			if wire.ToolConfig == nil || wire.ToolConfig.FunctionCallingConfig.Mode != "ANY" || len(wire.ToolConfig.FunctionCallingConfig.Allowed) != 0 {
				t.Fatal("response shape was not constrained across the full tool catalog")
			}
		} else if wire.ToolConfig != nil {
			t.Fatal("text answer prevented without a response tool")
		}
	}
}
