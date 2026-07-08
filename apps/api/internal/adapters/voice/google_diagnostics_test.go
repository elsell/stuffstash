package voice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceExposesSafeDiagnostics(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]any{{
						"functionCall": map[string]any{
							"name": "search_authorized_assets",
							"args": map[string]any{"query": "water bottle"},
						},
					}},
				},
			}},
		})
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
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript:         "Move my water bottle to the kitchen.",
		IncludeDiagnostics: true,
		Tools: []ports.AgentToolDescriptor{{
			Name:     "search_authorized_assets",
			ReadOnly: true,
			Parameters: ports.AgentToolParameters{
				Properties: map[string]ports.AgentToolParameter{"query": {Type: ports.AgentToolParameterTypeString}},
			},
		}},
	})
	if err != nil {
		t.Fatalf("language inference: %v", err)
	}
	if len(turn.Diagnostics) != 3 {
		t.Fatalf("expected prompt, tool catalog, and model turn diagnostics, got %+v", turn.Diagnostics)
	}
	if turn.Diagnostics[0].Title != "Language prompt (turn 1)" || !strings.Contains(turn.Diagnostics[0].Detail, "Move my water bottle to the kitchen.") {
		t.Fatalf("unexpected prompt diagnostic: %+v", turn.Diagnostics[0])
	}
	if turn.Diagnostics[1].Title != "Language tool catalog (turn 1)" ||
		!strings.Contains(turn.Diagnostics[1].Detail, `"requireToolCall": false`) ||
		!strings.Contains(turn.Diagnostics[1].Detail, `"name": "search_authorized_assets"`) ||
		!strings.Contains(turn.Diagnostics[1].Detail, `"readOnly": true`) ||
		!strings.Contains(turn.Diagnostics[1].Detail, `"providerCallable": true`) {
		t.Fatalf("unexpected tool catalog diagnostic: %+v", turn.Diagnostics[1])
	}
	if turn.Diagnostics[2].Title != "Language model turn (turn 1)" || !strings.Contains(turn.Diagnostics[2].Detail, "search_authorized_assets") {
		t.Fatalf("unexpected model turn diagnostic: %+v", turn.Diagnostics[2])
	}
}

func TestGoogleGeminiLanguageInferenceElidesRepeatedPromptDiagnostics(t *testing.T) {
	t.Parallel()

	diagnostics := languageInferenceDiagnostics(ports.LanguageInferenceInput{
		PreviousTurns: 2,
		Tools: []ports.AgentToolDescriptor{{
			Name:     "search_authorized_assets",
			ReadOnly: true,
		}},
	}, "Full repeated prompt.", `{"final":{"kind":"answer"}}`)
	if len(diagnostics) != 3 {
		t.Fatalf("expected prompt marker, tool catalog, and model turn diagnostics, got %+v", diagnostics)
	}
	if diagnostics[0].Title != "Language prompt (turn 3)" || strings.Contains(diagnostics[0].Detail, "Full repeated prompt.") {
		t.Fatalf("expected repeated prompt to be elided, got %+v", diagnostics[0])
	}
	if diagnostics[1].Title != "Language tool catalog (turn 3)" || !strings.Contains(diagnostics[1].Detail, "search_authorized_assets") {
		t.Fatalf("expected turn-labeled tool catalog diagnostic, got %+v", diagnostics[1])
	}
	if !strings.Contains(diagnostics[2].Title, "turn 3") || !strings.Contains(diagnostics[2].Detail, "answer") {
		t.Fatalf("expected turn-labeled model diagnostic, got %+v", diagnostics[2])
	}
}

func TestGoogleGeminiLanguageInferenceCatalogDiagnosticsUsePhaseEffectiveCallability(t *testing.T) {
	t.Parallel()

	diagnostics := languageInferenceDiagnostics(ports.LanguageInferenceInput{
		PreviousTurns: 1,
		PlanOnly:      true,
		Tools: []ports.AgentToolDescriptor{{
			Name:     "search_authorized_assets",
			ReadOnly: true,
		}},
	}, "Plan prompt.", `{"actionPlan":{"commands":[]}}`)
	if len(diagnostics) != 3 {
		t.Fatalf("expected prompt, tool catalog, and model turn diagnostics, got %+v", diagnostics)
	}
	if diagnostics[1].Title != "Language tool catalog (turn 2)" ||
		!strings.Contains(diagnostics[1].Detail, `"planOnly": true`) ||
		!strings.Contains(diagnostics[1].Detail, `"name": "search_authorized_assets"`) ||
		!strings.Contains(diagnostics[1].Detail, `"providerCallable": false`) {
		t.Fatalf("expected planner diagnostics to mark native tools unavailable, got %+v", diagnostics[1])
	}
}
