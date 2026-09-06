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

func TestGoogleConversationRejectsMalformedEnvelopesWithoutRetry(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"valid", `{"toolCalls":[{"name":"deliver","arguments":{"speech":"Ready."}}]}`, true},
		{"trailing JSON", `{"toolCalls":[{"name":"deliver","arguments":{"speech":"Ready."}}]} {}`, false},
		{"unknown envelope field", `{"toolCalls":[{"name":"deliver","arguments":{"speech":"Ready."}}],"intent":"read"}`, false},
		{"unknown call field", `{"toolCalls":[{"name":"deliver","arguments":{"speech":"Ready."},"resolution":"strong"}]}`, false},
		{"empty calls", `{"toolCalls":[]}`, false},
		{"null arguments", `{"toolCalls":[{"name":"deliver","arguments":null}]}`, false},
		{"missing arguments", `{"toolCalls":[{"name":"deliver"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				_ = json.NewEncoder(w).Encode(geminiTextResponse(tc.body))
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
					t.Fatalf("valid envelope lost: %+v %v", turn, err)
				}
			} else if !errors.Is(err, ports.ErrInvalidProviderInput) || len(turn.ToolCalls) != 0 {
				t.Fatalf("malformed envelope accepted: %+v %v", turn, err)
			}
		})
	}
}
