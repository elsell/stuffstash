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

func TestGoogleGeminiGeneratesResponseFromGroundedBrief(t *testing.T) {
	t.Parallel()
	var request geminiGenerateContentRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"spokenResponse":"Your tools are probably in the toolbox in the garage.","displayResponse":"Your tools are probably in the Toolbox in the Garage."}`))
	}))
	t.Cleanup(server.Close)
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{BaseURL: server.URL, APIKey: "test-key", Model: "gemini-test"})
	brief := agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeLocate, Operation: agentmodel.OperationLocate,
		Subject: "tools", Confidence: agentmodel.ResponseConfidencePlausible,
		Findings: []agentmodel.ResponseFinding{{FactKey: "finding.0", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "Toolbox"}}},
	}
	result, err := provider.GenerateResponse(context.Background(), ports.VoiceResponseGenerationInput{Brief: brief})
	if err != nil {
		t.Fatalf("generate response: %v", err)
	}
	if !strings.Contains(strings.ToLower(result.SpokenResponse), "toolbox") {
		t.Fatalf("unexpected result: %+v", result)
	}
	prompt := request.Contents[0].Parts[0].Text
	if !strings.Contains(prompt, "untrusted grounded response brief") || !strings.Contains(prompt, `"confidence":"plausible"`) {
		t.Fatalf("expected bounded grounded prompt, got %q", prompt)
	}
	if !strings.Contains(prompt, "place findings may be omitted") || !strings.Contains(prompt, "intentionally supplied a bounded presentation subset") {
		t.Fatalf("expected item-complete and truncation-aware realization policy, got %q", prompt)
	}
	if request.GenerationConfig == nil || request.GenerationConfig.ResponseMimeType != "application/json" || request.GenerationConfig.ResponseJSONSchema == nil {
		t.Fatalf("expected schema-constrained response generation, got %+v", request.GenerationConfig)
	}
	if _, found := request.GenerationConfig.ResponseJSONSchema.Properties["factKeysUsed"]; found {
		t.Fatal("response schema must not ask the model to self-attest grounding")
	}
}
