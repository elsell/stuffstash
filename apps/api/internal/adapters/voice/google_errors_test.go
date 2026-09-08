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

func TestGoogleProviderBillingFailureClassification(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"billing", `{"error":{"message":"private project secret","details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","reason":"BILLING_DISABLED","domain":"googleapis.com","metadata":{"consumer":"private-project"}}]}}`, "provider_billing_disabled"},
		{"permission", `{"error":{"message":"Billing is disabled","details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","reason":"IAM_PERMISSION_DENIED","domain":"googleapis.com"}]}}`, "provider_http_status_403"},
		{"untrusted domain", `{"error":{"details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","reason":"BILLING_DISABLED","domain":"untrusted"}]}}`, "provider_http_status_403"},
		{"malformed", `{`, "provider_http_status_403"},
		{"oversized", strings.Repeat(" ", 65536) + `{"error":{"details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","reason":"BILLING_DISABLED","domain":"googleapis.com"}]}}`, "provider_http_status_403"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403); _, _ = w.Write([]byte(tc.body)) }))
			defer server.Close()
			client := newGoogleHTTPClient(server.URL, server.Client(), 0, staticTokenSource{}, "", "")
			err := client.postJSON(context.Background(), "/", map[string]string{}, &struct{}{})
			var safe interface{ SafeRealtimeVoiceDiagnostic() string }
			if !errors.As(err, &safe) || safe.SafeRealtimeVoiceDiagnostic() != tc.want {
				t.Fatalf("unexpected classification: %v", err)
			}
			if strings.Contains(err.Error(), "private") {
				t.Fatal("response data leaked")
			}
		})
	}
}
