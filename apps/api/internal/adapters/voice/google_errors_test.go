package voice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceReportsSafeHTTPStatusWithoutBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"quota exhausted for secret-project and bearer should-not-leak"}}`))
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

	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true})
	if err == nil {
		t.Fatalf("expected provider error")
	}
	if !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("expected safe status in error, got %v", err)
	}
	if strings.Contains(err.Error(), "secret-project") || strings.Contains(err.Error(), "should-not-leak") {
		t.Fatalf("provider error leaked response body: %v", err)
	}
}
