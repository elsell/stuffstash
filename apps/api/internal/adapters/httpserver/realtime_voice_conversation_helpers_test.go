package httpserver

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Controlled conversation implementations for transport tests. These select
// fixed reads for their fixture; production request interpretation belongs to the model.
func (m *capturingLanguageModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationRead(input, app.RealtimeVoiceToolSearchAuthorizedAssets, map[string]any{"query": "tools"}, &m.lastToolResult)
}
func (scriptedLanguageModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationRead(input, app.RealtimeVoiceToolSearchAuthorizedAssets, map[string]any{"query": "tools"}, nil)
}
func (m *locationAwareLanguageModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationRead(input, app.RealtimeVoiceToolSearchAuthorizedAssets, map[string]any{"query": "water bottle"}, &m.lastToolResult)
}
func (m *itemListingLanguageModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationRead(input, app.RealtimeVoiceToolListAuthorizedAssets, map[string]any{"kind": "item"}, &m.lastToolResult)
}
func (m failingLanguageModel) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return ports.ConversationModelTurn{}, m.err
}

func httpConversationRead(input ports.ConversationModelInput, tool string, arguments map[string]any, capture *string) (ports.ConversationModelTurn, error) {
	if len(input.Messages) == 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	last := input.Messages[len(input.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "fixture-read", Name: tool, Arguments: arguments}}}, nil
	}
	if len(last.ToolResults) != 1 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	result := last.ToolResults[0].Content
	if capture != nil {
		*capture = result
	}
	var evidence struct {
		Items []struct {
			AssetID     string `json:"assetId"`
			Title       string `json:"title"`
			ParentTitle string `json:"parentTitle"`
		} `json:"items"`
	}
	if json.Unmarshal([]byte(result), &evidence) != nil {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	var sentences, ids []string
	for _, item := range evidence.Items {
		text := "I found " + item.Title
		if item.ParentTitle != "" {
			text += " in " + item.ParentTitle
		}
		sentences = append(sentences, text+".")
		ids = append(ids, item.AssetID)
	}
	text := strings.Join(sentences, " ")
	if text == "" {
		text = "I couldn't find matching belongings in this inventory."
	}
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: text, Display: text, AssetIDs: ids}}, nil
}
