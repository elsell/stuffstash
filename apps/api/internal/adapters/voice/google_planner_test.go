package voice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceRequestsStructuredActionPlanForPlannerTurns(t *testing.T) {
	t.Parallel()

	var request map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"actionPlan":{"intentSummary":"Move the water bottle to Kitchen.","modelInterpretationSummary":"Create Kitchen and move the visible water bottle there.","confirmationSummary":"Create Kitchen and move the water bottle there?","commands":[{"id":"cmd-kitchen","kind":"create_location","summary":"Create Kitchen","arguments":{"title":"Kitchen"}},{"id":"cmd-move-water-bottle","kind":"move_asset","summary":"Move Water bottle to Kitchen","arguments":{"assetId":"asset-water-bottle","parentCommandId":"cmd-kitchen"}}]}}`))
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
		Transcript: "Move my water bottle to the kitchen.",
		Tools:      []ports.AgentToolDescriptor{testGeminiActionPlanToolDescriptor()},
		ToolResults: []ports.AgentToolResult{{
			CallID: "call-1",
			Name:   "search_authorized_assets",
			Call: ports.AgentToolCall{
				ID:        "call-1",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "water bottle"},
			},
			Content: `{"tool":"search_authorized_assets","count":1,"items":[{"assetId":"asset-water-bottle","title":"Water bottle","kind":"item","locationTitle":"Office"}]}`,
		}},
		PreviousTurns: 1,
		PlanOnly:      true,
	})
	if err != nil {
		t.Fatalf("planner turn: %v", err)
	}
	if len(turn.ToolCalls) != 1 {
		t.Fatalf("expected one internal action-plan call, got %+v", turn)
	}
	call := turn.ToolCalls[0]
	if call.Name != "propose_action_plan" || call.ID != "gemini-action-plan" {
		t.Fatalf("unexpected internal action-plan call: %+v", call)
	}
	if call.Arguments["confirmationSummary"] != "Create Kitchen and move the water bottle there?" {
		t.Fatalf("unexpected planner arguments: %+v", call.Arguments)
	}
	if _, hasTools := request["tools"]; hasTools {
		t.Fatalf("planner turn must not expose provider-callable tools, got %+v", request["tools"])
	}
	config := objectAt(t, request, "generationConfig")
	if config["responseMimeType"] != "application/json" || !generationConfigHasActionPlanSchema(config) {
		t.Fatalf("planner turn should request action-plan structured output, got %+v", config)
	}
}

func testGeminiActionPlanToolDescriptor() ports.AgentToolDescriptor {
	return ports.AgentToolDescriptor{
		Name: "propose_action_plan",
		Parameters: ports.AgentToolParameters{
			Required: []string{"intentSummary", "modelInterpretationSummary", "confirmationSummary"},
			Properties: map[string]ports.AgentToolParameter{
				"intentSummary": {
					Type: ports.AgentToolParameterTypeString,
				},
				"modelInterpretationSummary": {
					Type: ports.AgentToolParameterTypeString,
				},
				"confirmationSummary": {
					Type: ports.AgentToolParameterTypeString,
				},
				"commands": {
					Type: ports.AgentToolParameterTypeArray,
					Items: &ports.AgentToolParameter{
						Type:     ports.AgentToolParameterTypeObject,
						Required: []string{"id", "kind", "summary", "arguments"},
						Properties: map[string]ports.AgentToolParameter{
							"id": {
								Type: ports.AgentToolParameterTypeString,
							},
							"kind": {
								Type: ports.AgentToolParameterTypeString,
								Enum: []string{"create_asset", "create_location", "move_asset"},
							},
							"summary": {
								Type: ports.AgentToolParameterTypeString,
							},
							"arguments": {
								Type: ports.AgentToolParameterTypeObject,
							},
						},
					},
				},
			},
		},
	}
}

func TestGoogleGeminiLanguageInferenceRejectsTextActionPlanOutsidePlannerTurns(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[{"id":"cmd-1","kind":"move_asset","summary":"Move it","arguments":{"assetId":"asset-water-bottle"}}]}}`))
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
	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript:      "Move my water bottle.",
		RequireToolCall: true,
		Tools: []ports.AgentToolDescriptor{{
			Name:             "search_authorized_assets",
			Description:      "Search visible assets.",
			ReadOnly:         true,
			ProviderCallable: true,
		}},
	})
	if err == nil {
		t.Fatalf("expected text action plan outside planner mode to be rejected")
	}
}
