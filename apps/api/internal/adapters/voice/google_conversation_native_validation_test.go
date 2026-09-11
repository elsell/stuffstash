package voice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleConversationRejectsInvalidNativeCompletionsWithoutRetry(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"valid", `{"functionCall":{"name":"deliver","args":{"speech":"Ready."}}}`, true},
		{"plain text", `{"text":"Ready."}`, false},
		{"legacy envelope", `{"text":"{\"toolCalls\":[]}"}`, false},
		{"malformed arguments", `{"functionCall":{"name":"deliver","args":"bad"}}`, false},
		{"empty name", `{"functionCall":{"name":"","args":{}}}`, false},
		{"no parts", `{}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"finishReason": "STOP", "content": map[string]any{"role": "model", "parts": []any{json.RawMessage(tc.body)}}}}})
			}))
			defer server.Close()
			model := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{BaseURL: server.URL, APIKey: "fixture-key", Model: "fixture-model"})
			input := googleConversationTestInput("Hello")
			input.Tools = []ports.ConversationToolDefinition{{Name: "deliver", Description: "Answer.", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"speech":{"type":"string"}},"required":["speech"],"additionalProperties":false}`)}}
			turn, err := model.Converse(context.Background(), input)
			if calls.Load() != 1 {
				t.Fatalf("one invocation made %d requests", calls.Load())
			}
			if tc.valid {
				if err != nil || len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Arguments["speech"] != "Ready." {
					t.Fatalf("valid native call lost: %+v %v", turn, err)
				}
			} else if !errors.Is(err, ports.ErrInvalidProviderInput) || len(turn.ToolCalls) != 0 {
				t.Fatalf("invalid native completion accepted: %+v %v", turn, err)
			}
		})
	}
}
