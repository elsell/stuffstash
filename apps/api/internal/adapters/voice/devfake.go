package voice

import (
	"context"
	"encoding/json"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type DevFakeSpeechToText struct{}

func (DevFakeSpeechToText) Transcribe(_ context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	return ports.SpeechToTextResult{Transcript: "Where are my tools?"}, nil
}

type DevFakeLanguageInference struct{}

// Converse is the fixed development demonstration, never a fallback model.
func (DevFakeLanguageInference) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if err := ctx.Err(); err != nil {
		return ports.ConversationModelTurn{}, err
	}
	if len(input.Messages) == 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	last := input.Messages[len(input.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "dev-tools", Name: "search_authorized_assets", Arguments: map[string]any{"query": "tools"}}}}, nil
	}
	if len(last.ToolResults) != 1 || last.ToolResults[0].CallID != "dev-tools" || last.ToolResults[0].Name != "search_authorized_assets" {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	var evidence struct {
		Error json.RawMessage `json:"error"`
		Items []struct {
			AssetID     string `json:"assetId"`
			Title       string `json:"title"`
			ParentTitle string `json:"parentTitle"`
		} `json:"items"`
	}
	if json.Unmarshal([]byte(last.ToolResults[0].Content), &evidence) != nil || len(evidence.Error) != 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	text := "I couldn't find matching tools in this inventory."
	var ids []string
	if len(evidence.Items) > 0 {
		item := evidence.Items[0]
		if item.AssetID == "" || item.Title == "" {
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
		text = "I found " + item.Title
		if item.ParentTitle != "" {
			text += " in " + item.ParentTitle
		}
		text += "."
		ids = []string{item.AssetID}
	}
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: text, Display: text, AssetIDs: ids}}, nil
}

func (DevFakeLanguageInference) ProbeLanguageInference(context.Context) error { return nil }

type DevFakeTextToSpeech struct{}

func (DevFakeTextToSpeech) Synthesize(_ context.Context, input ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	if input.Text == "" {
		return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
	}
	return ports.TextToSpeechResult{
		MimeType: "audio/mpeg",
		Chunks:   [][]byte{[]byte(input.Text)},
	}, nil
}
