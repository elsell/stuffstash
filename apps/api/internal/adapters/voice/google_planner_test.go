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

func TestGoogleGeminiActionPlanSchemaDoesNotUseLegacyToolArgumentShape(t *testing.T) {
	t.Parallel()

	schema := geminiActionPlanResponseSchema()
	if !sameStrings(schema.Required, []string{"actionPlan"}) {
		t.Fatalf("expected root actionPlan required field, got %+v", schema.Required)
	}
	actionPlan, ok := schema.Properties["actionPlan"]
	if !ok {
		t.Fatalf("expected actionPlan property in schema: %+v", schema)
	}
	if !sameStrings(actionPlan.Required, []string{"intentSummary", "modelInterpretationSummary", "confirmationSummary", "commands"}) {
		t.Fatalf("expected action plan required fields, got %+v", actionPlan.Required)
	}
	if _, hasCommandKind := actionPlan.Properties["commandKind"]; hasCommandKind {
		t.Fatalf("planner schema must not expose legacy commandKind field: %+v", actionPlan.Properties)
	}
	if _, hasArguments := actionPlan.Properties["arguments"]; hasArguments {
		t.Fatalf("planner schema must not expose legacy top-level arguments field: %+v", actionPlan.Properties)
	}
	commands := actionPlan.Properties["commands"]
	if commands.Items == nil {
		t.Fatalf("expected typed command items in planner schema: %+v", commands)
	}
	if !sameStrings(commands.Items.Required, []string{"id", "kind", "summary", "arguments"}) {
		t.Fatalf("expected command item required fields, got %+v", commands.Items.Required)
	}
	kindEnum := commands.Items.Properties["kind"].Enum
	if !sameStrings(kindEnum, []string{"create_asset", "create_location", "move_asset", "archive_asset", "restore_asset"}) {
		t.Fatalf("expected supported command kind enum, got %+v", kindEnum)
	}
	arguments := commands.Items.Properties["arguments"]
	if arguments.Properties["title"].Type != "string" || arguments.Properties["assetId"].Type != "string" || arguments.Properties["parentCommandId"].Type != "string" {
		t.Fatalf("expected typed command argument properties, got %+v", arguments.Properties)
	}
}

func TestGoogleGeminiLanguageInferenceRejectsMalformedStructuredActionPlan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing commands",
			body: `{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?"}}`,
		},
		{
			name: "empty commands",
			body: `{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[]}}`,
		},
		{
			name: "unsupported command kind",
			body: `{"actionPlan":{"intentSummary":"Rename a bottle.","modelInterpretationSummary":"Rename it.","confirmationSummary":"Rename it?","commands":[{"id":"cmd-1","kind":"update_asset","summary":"Rename it","arguments":{"assetId":"asset-water-bottle","title":"Bottle"}}]}}`,
		},
		{
			name: "missing command arguments",
			body: `{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[{"id":"cmd-1","kind":"move_asset","summary":"Move it"}]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseLanguageTurn(tt.body, []ports.AgentToolDescriptor{testGeminiActionPlanToolDescriptor()}, true)
			if err == nil {
				t.Fatalf("expected malformed planner output to be rejected")
			}
		})
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

func sameStrings(actual []string, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return false
		}
	}
	return true
}
