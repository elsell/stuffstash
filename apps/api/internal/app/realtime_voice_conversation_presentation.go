package app

import (
	"bytes"
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeConversationPresentTool = "present_answer"

func (e *realtimeConversationTools) present(call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	encoded, err := json.Marshal(call.Arguments)
	if err != nil {
		return ports.ConversationToolOutcome{}, ports.ErrInvalidProviderInput
	}
	var input struct {
		Spoken   string   `json:"spoken"`
		Display  string   `json:"display"`
		AssetIDs []string `json:"assetIds"`
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || input.Display == "" || len(input.AssetIDs) > maxRealtimeVoiceResponseArtifacts {
		return ports.ConversationToolOutcome{}, ports.ErrInvalidProviderInput
	}
	answer := &ports.ConversationAnswer{Spoken: input.Spoken, Display: input.Display, AssetIDs: input.AssetIDs}
	if _, err := realtimeConversationResponse(answer, e.items); err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	return ports.ConversationToolOutcome{Answer: answer, Result: ports.AgentToolResult{Content: `{"status":"answer_prepared"}`}}, nil
}
func realtimeConversationPresentationTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{ResponseTool: true, Name: realtimeConversationPresentTool, Description: "Finish your answer with natural speech, display text and optional clickable item cards. Use assetIds from authorized tool results; summarize naturally without reading every title. This prepares an answer, never an inventory change. No further tool calls run after this one.", Parameters: json.RawMessage(`{"type":"object","properties":{"spoken":{"type":"string","maxLength":500},"display":{"type":"string","maxLength":1000},"assetIds":{"type":"array","maxItems":16,"items":{"type":"string"}}},"required":["spoken","display"],"additionalProperties":false}`)}
}
