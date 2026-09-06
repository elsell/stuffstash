package voice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceReportsSafeHTTPStatusWithoutBody(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"quota exhausted for secret-project and bearer should-not-leak"}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	_, err := provider.Converse(context.Background(), googleConversationTestInput("Where are my tools?"))
	if err == nil || !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("expected safe provider status, got %v", err)
	}
	if strings.Contains(err.Error(), "secret-project") || strings.Contains(err.Error(), "should-not-leak") {
		t.Fatalf("provider error leaked response body: %v", err)
	}
	if calls != 1 {
		t.Fatalf("provider retried behind the conversation budget, got %d calls", calls)
	}
}

func TestGoogleGeminiLanguageInferenceReportsSafeTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPTimeout: time.Millisecond,
	})
	_, err := provider.Converse(context.Background(), googleConversationTestInput("Where are my tools?"))
	if err == nil {
		t.Fatal("expected provider timeout")
	}
	var safe interface{ SafeRealtimeVoiceDiagnostic() string }
	if !errors.As(err, &safe) || safe.SafeRealtimeVoiceDiagnostic() != "provider_timeout" {
		t.Fatalf("expected safe timeout diagnostic, got %T %v", err, err)
	}
}

func googleConversationTestInput(text string) ports.ConversationModelInput {
	return ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: text}}}
}
