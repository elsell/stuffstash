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

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
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
	_, err := provider.NextTurn(context.Background(), googleInitialInvestigationInput("Where are my tools?"))
	if err == nil || !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("expected safe provider status, got %v", err)
	}
	if strings.Contains(err.Error(), "secret-project") || strings.Contains(err.Error(), "should-not-leak") {
		t.Fatalf("provider error leaked response body: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected bounded retry before surfacing 429, got %d calls", calls)
	}
}

func TestGoogleRetryAfterDurationParsesSecondsAndHTTPDate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	if got := googleRetryAfterDuration("2", now); got != 2*time.Second {
		t.Fatalf("expected seconds retry-after, got %s", got)
	}
	date := now.Add(1500 * time.Millisecond).UTC().Format(http.TimeFormat)
	if got := googleRetryAfterDuration(date, now); got != time.Second {
		t.Fatalf("expected rounded HTTP-date retry-after, got %s", got)
	}
	if got := googleRetryAfterDuration("nonsense", now); got != 0 {
		t.Fatalf("expected invalid retry-after to be ignored, got %s", got)
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
	_, err := provider.NextTurn(context.Background(), googleInitialInvestigationInput("Where are my tools?"))
	if err == nil {
		t.Fatal("expected provider timeout")
	}
	var safe interface{ SafeRealtimeVoiceDiagnostic() string }
	if !errors.As(err, &safe) || safe.SafeRealtimeVoiceDiagnostic() != "provider_timeout" {
		t.Fatalf("expected safe timeout diagnostic, got %T %v", err, err)
	}
}

func TestGoogleGeminiLiveStructuredInvestigationTurn(t *testing.T) {
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
		ProjectID: projectID, Location: location, Model: model, QuotaProject: projectID,
		TokenSource: tokenSource, HTTPTimeout: 120 * time.Second,
	})
	turn, err := provider.NextTurn(ctx, googleInitialInvestigationInput("Where might Sarah's winter coat be?"))
	if err != nil {
		t.Fatalf("structured investigation live turn failed: %v", err)
	}
	if turn.Investigation == nil || turn.Investigation.Validate() != nil {
		t.Fatalf("expected valid structured investigation, got %+v", turn)
	}
	t.Logf("live structured investigation: decision=%s operation=%s probes=%d", turn.Investigation.Decision, turn.Investigation.Intent.Operation, len(turn.Investigation.SearchRequests))
}

func googleInitialInvestigationInput(transcript string) ports.LanguageInferenceInput {
	return ports.LanguageInferenceInput{
		Transcript: transcript,
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1", SchemaVersion: "voice-investigation-v1",
			Transcript: transcript, MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
		},
	}
}
