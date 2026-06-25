package voice

import (
	"context"

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

func (DevFakeLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "dev-search-assets",
				Name: "search_authorized_assets",
				Arguments: map[string]any{
					"query": "tools",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  "I checked your inventory for tools. Open the voice details to see the result.",
			DisplayResponse: "I checked your inventory for tools. Open the voice details to see the result.",
		},
	}, nil
}

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
