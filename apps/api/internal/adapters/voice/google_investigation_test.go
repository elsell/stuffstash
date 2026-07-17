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

func TestGoogleGeminiLanguageInferenceUsesStructuredInvestigationContract(t *testing.T) {
	t.Parallel()

	var request map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{
          "decision":"search",
          "intent":{"kind":"read","operation":"locate","subjectMention":"Sarah winter coat","newAssetKind":"","destinationPath":[],"details":""},
          "searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"Sarah winter coat","kindHint":"","visibleAssetId":"","searchProbes":["Sarah winter coat","Sarah winter clothes","winter clothing"]}],
          "resolutions":[],
          "rationale":"Gather authorized candidates for the remembered title."
        }`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript: "Where are Sarah's winter coat?",
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1",
			SchemaVersion: "voice-investigation-v1", Transcript: "Where are Sarah's winter coat?",
			MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
		},
	})
	if err != nil {
		t.Fatalf("investigation turn: %v", err)
	}
	if turn.Investigation == nil || turn.Investigation.Intent.Operation != agentmodel.OperationLocate {
		t.Fatalf("unexpected investigation turn: %+v", turn)
	}
	if got := turn.Investigation.SearchRequests[0].SearchProbes; len(got) != 3 || got[1] != "Sarah winter clothes" {
		t.Fatalf("expected diverse model-owned probes, got %+v", got)
	}
	if _, exists := request["tools"]; exists {
		t.Fatalf("investigation must not expose provider-callable tools: %+v", request)
	}
	if _, exists := request["toolConfig"]; exists {
		t.Fatalf("investigation must not expose provider tool choice: %+v", request)
	}
	config := objectAt(t, request, "generationConfig")
	if config["responseMimeType"] != "application/json" || config["responseJsonSchema"] == nil {
		t.Fatalf("expected JSON-schema constrained investigation output, got %+v", config)
	}
	contents, ok := request["contents"].([]any)
	if !ok || len(contents) != 1 || !strings.Contains(string(mustJSON(t, contents[0])), "search hypotheses") {
		t.Fatalf("expected bounded investigation prompt, got %+v", request["contents"])
	}
}

func TestGoogleGeminiLanguageInferenceRejectsInvalidInvestigationPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{
          "decision":"finish",
          "intent":{"kind":"change","operation":"move","subjectMention":"drill","newAssetKind":"","destinationPath":["garage"],"details":""},
          "searchRequests":[],
          "resolutions":[{"referenceKey":"subject","status":"strong","candidateIds":["invented-id"],"evidence":"guess"}],
          "commands":[{"kind":"move_asset"}],
          "rationale":""
        }`))
	}))
	t.Cleanup(server.Close)
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript: "Move the drill to the garage",
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1",
			SchemaVersion: "voice-investigation-v1", Transcript: "Move the drill to the garage",
			MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
		},
	})
	if err == nil {
		t.Fatal("expected provider-authored commands and invalid initial finish to be rejected")
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	return payload
}
