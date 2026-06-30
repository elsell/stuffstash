package voice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceRequestsJSONForFinalOnlyTurns(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		config := objectAt(t, request, "generationConfig")
		if config["responseMimeType"] != "application/json" {
			t.Fatalf("expected final-only turn to request JSON response mime type, got %+v", config)
		}
		if !generationConfigHasFinalResponseSchema(config) {
			t.Fatalf("expected final-only turn to request final response schema, got %+v", config)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Ready.","displayResponse":"Ready."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})
	if _, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true}); err != nil {
		t.Fatalf("final-only turn: %v", err)
	}
}

func TestGoogleGeminiLanguageInferenceRetriesRateLimitedFinalOnlyTurn(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Ready.","displayResponse":"Ready."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true})
	if err != nil {
		t.Fatalf("final-only retry: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected one retry, got %d calls", calls)
	}
	if turn.Final == nil || turn.Final.SpokenResponse != "Ready." {
		t.Fatalf("unexpected final turn: %+v", turn)
	}
}
