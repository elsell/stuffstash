package voice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceDiagnosticsExposeOnlySafeInvestigationMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{
          "decision":"search",
          "intent":{"kind":"read","operation":"locate","subjectMention":"Secret winter coat","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},
          "searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"Secret winter coat","kindHint":"","visibleAssetId":"","searchProbes":["Secret winter coat"],"lifecycleScope":"active"}],
          "vocabularyRequests":[],"resolutions":[],"rationale":"raw-model-private-rationale"
        }`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript:         "Where is Secret winter coat?",
		IncludeDiagnostics: true,
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1", SchemaVersion: "voice-investigation-v1",
			Transcript: "Where is Secret winter coat?", MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
			Vocabulary: agentmodel.VoiceVocabularyManifest{CustomAssetTypes: []agentmodel.VoiceVocabularyAssetType{{Key: "secret-type", DisplayName: "Private attic garment"}}},
		},
	})
	if err != nil {
		t.Fatalf("language inference: %v", err)
	}
	if len(turn.Diagnostics) != 1 || turn.Diagnostics[0].Title != "Language investigation (turn 1)" {
		t.Fatalf("expected one bounded metadata diagnostic, got %+v", turn.Diagnostics)
	}
	detail := turn.Diagnostics[0].Detail
	for _, unsafe := range []string{"Secret winter coat", "Private attic garment", "secret-type", "raw-model-private-rationale", "search_assets", "subjectMention", "transcript"} {
		if strings.Contains(detail, unsafe) {
			t.Fatalf("diagnostic leaked %q: %s", unsafe, detail)
		}
	}
	for _, safe := range []string{`"phase":"initial"`, `"evidenceRound":0`, `"customAssetTypeCount":1`, `"promptVersion":"voice-investigation-v1"`} {
		if !strings.Contains(detail, safe) {
			t.Fatalf("diagnostic missing safe metadata %q: %s", safe, detail)
		}
	}
}

func TestGoogleGeminiLanguageInferenceDiagnosticsBoundUntrustedVersions(t *testing.T) {
	t.Parallel()

	input := googleInitialInvestigationInput("private transcript")
	input.PreviousTurns = 2
	input.Investigation.PromptVersion = "prompt version with bearer should-not-leak"
	input.Investigation.SchemaVersion = "voice-investigation-v1"
	diagnostics := languageInferenceDiagnostics(input)
	if len(diagnostics) != 1 || diagnostics[0].Title != "Language investigation (turn 3)" {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	if strings.Contains(diagnostics[0].Detail, "should-not-leak") || !strings.Contains(diagnostics[0].Detail, `"promptVersion":"unknown"`) {
		t.Fatalf("unsafe prompt version was not bounded: %s", diagnostics[0].Detail)
	}
}
