package voice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2/google"

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

func TestGoogleGeminiLanguageInferenceReportsSafeTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"final\":{\"kind\":\"answer\",\"spokenResponse\":\"ready\",\"displayResponse\":\"ready\"}}"}]}}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPTimeout: time.Millisecond,
	})

	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true})
	if err == nil {
		t.Fatalf("expected provider timeout")
	}
	var safe interface{ SafeRealtimeVoiceDiagnostic() string }
	if !errors.As(err, &safe) || safe.SafeRealtimeVoiceDiagnostic() != "provider_timeout" {
		t.Fatalf("expected safe timeout diagnostic, got %T %v", err, err)
	}
}

func TestGoogleGeminiLiveMinimalStructuredAndToolTurns(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("set STUFF_STASH_GOOGLE_LIVE_TESTS=1 to run the live Gemini probe")
	}
	projectID := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		t.Skip("set STUFF_STASH_GOOGLE_CLOUD_PROJECT to run the live Gemini probe")
	}
	location := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_LOCATION"))
	if location == "" {
		location = "us-central1"
	}
	model := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_GEMINI_MODEL"))
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	tokenSource, err := google.DefaultTokenSource(ctx, googleCloudPlatformScope)
	if err != nil {
		t.Skipf("Google ADC unavailable for live Gemini probe: %v", err)
	}
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:    projectID,
		Location:     location,
		Model:        model,
		QuotaProject: projectID,
		TokenSource:  tokenSource,
		HTTPTimeout:  120 * time.Second,
	})

	finalTurn, err := provider.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript: "Say the provider is ready.",
		FinalOnly:  true,
	})
	if err != nil {
		t.Fatalf("structured final live turn failed: %v", err)
	}
	if finalTurn.Final == nil || strings.TrimSpace(finalTurn.Final.SpokenResponse) == "" {
		t.Fatalf("expected structured final response, got %+v", finalTurn)
	}

	toolTurn, err := provider.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript:      "Where is my water bottle?",
		RequireToolCall: true,
		Tools: []ports.AgentToolDescriptor{{
			Name:        "search_authorized_assets",
			Description: "Search visible assets.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"query"},
				Properties: map[string]ports.AgentToolParameter{
					"query": {Type: ports.AgentToolParameterTypeString, Description: "Search keywords."},
				},
			},
		}},
	})
	if err != nil {
		t.Fatalf("forced tool live turn failed: %v", err)
	}
	if len(toolTurn.ToolCalls) != 1 || toolTurn.ToolCalls[0].Name != "search_authorized_assets" {
		t.Fatalf("expected forced search tool call, got %+v", toolTurn)
	}
}
