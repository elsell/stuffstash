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
	if _, hasToolConfig := request["toolConfig"]; hasToolConfig {
		t.Fatalf("planner turn must not send tool config, got %+v", request["toolConfig"])
	}
	config := objectAt(t, request, "generationConfig")
	if config["responseMimeType"] != "application/json" || !generationConfigHasActionPlanSchema(config) {
		t.Fatalf("planner turn should request action-plan structured output, got %+v", config)
	}
	if _, hasOpenAPISchema := config["responseSchema"]; hasOpenAPISchema {
		t.Fatalf("planner turn must use responseJsonSchema for branchy action-plan output, got %+v", config)
	}
	if _, hasJSONSchema := config["responseJsonSchema"]; !hasJSONSchema {
		t.Fatalf("planner turn should send responseJsonSchema, got %+v", config)
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
	if len(commands.Items.AnyOf) != 7 {
		t.Fatalf("expected command-specific schema branches, got %+v", commands.Items)
	}
	createAsset := commands.Items.AnyOf[0]
	if !sameStrings(createAsset.Required, []string{"id", "kind", "summary", "arguments"}) ||
		!sameStrings(createAsset.Properties["kind"].Enum, []string{"create_asset"}) {
		t.Fatalf("expected create_asset branch, got %+v", createAsset)
	}
	createArgs := createAsset.Properties["arguments"]
	if !sameStrings(createArgs.Required, []string{"title", "kind", "parentAssetId", "parentCommandId"}) ||
		len(createArgs.AnyOf) != 0 ||
		createArgs.Properties["assetId"].Type != "" ||
		createArgs.Properties["parentCommandId"].Type != "string" {
		t.Fatalf("expected constrained create_asset arguments, got %+v", createArgs)
	}
	moveAsset := commands.Items.AnyOf[2]
	if !sameStrings(moveAsset.Properties["kind"].Enum, []string{"move_asset"}) {
		t.Fatalf("expected move_asset branch, got %+v", moveAsset)
	}
	moveArgs := moveAsset.Properties["arguments"]
	if !sameStrings(moveArgs.Required, []string{"assetId", "parentAssetId", "parentCommandId"}) ||
		len(moveArgs.AnyOf) != 0 ||
		len(moveArgs.Properties["parentAssetId"].AnyOf) != 2 ||
		moveArgs.Properties["title"].Type != "" {
		t.Fatalf("expected constrained move_asset arguments, got %+v", moveArgs)
	}
	checkoutAsset := commands.Items.AnyOf[5]
	if !sameStrings(checkoutAsset.Properties["kind"].Enum, []string{"checkout_asset"}) {
		t.Fatalf("expected checkout_asset branch, got %+v", checkoutAsset)
	}
	checkoutArgs := checkoutAsset.Properties["arguments"]
	if !sameStrings(checkoutArgs.Required, []string{"assetId"}) ||
		checkoutArgs.Properties["assetId"].Type != "string" ||
		checkoutArgs.Properties["details"].Type != "string" ||
		checkoutArgs.Properties["parentAssetId"].Type != "" {
		t.Fatalf("expected constrained checkout_asset arguments, got %+v", checkoutArgs)
	}
	returnAsset := commands.Items.AnyOf[6]
	if !sameStrings(returnAsset.Properties["kind"].Enum, []string{"return_asset"}) {
		t.Fatalf("expected return_asset branch, got %+v", returnAsset)
	}
	returnArgs := returnAsset.Properties["arguments"]
	if !sameStrings(returnArgs.Required, []string{"assetId"}) ||
		returnArgs.Properties["assetId"].Type != "string" ||
		returnArgs.Properties["details"].Type != "string" ||
		returnArgs.Properties["parentAssetId"].Type != "" {
		t.Fatalf("expected constrained return_asset arguments, got %+v", returnArgs)
	}
}

func TestGoogleGeminiLanguageInferencePassesMalformedPlannerCommandsToAppRepair(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "semantically invalid move id",
			body: `{"actionPlan":{"intentSummary":"Move a bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[{"id":"cmd-1","kind":"move_asset","summary":"Move it","arguments":{"assetId":"water bottle","parentAssetId":"kitchen"}}]}}`,
		},
		{
			name: "checkout asset with visible id",
			body: `{"actionPlan":{"intentSummary":"Check out the drill.","modelInterpretationSummary":"Mark the visible drill as checked out.","confirmationSummary":"Check out the drill?","commands":[{"id":"cmd-checkout-drill","kind":"checkout_asset","summary":"Check out drill","arguments":{"assetId":"drill-1","details":"using it at the bench"}}]}}`,
		},
		{
			name: "return asset with visible id",
			body: `{"actionPlan":{"intentSummary":"Return the drill.","modelInterpretationSummary":"Mark the visible drill as returned.","confirmationSummary":"Return the drill?","commands":[{"id":"cmd-return-drill","kind":"return_asset","summary":"Return drill","arguments":{"assetId":"drill-1","details":"back in the tool bin"}}]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			turn, err := parseLanguageTurn(tt.body, []ports.AgentToolDescriptor{testGeminiActionPlanToolDescriptor()}, true)
			if err != nil {
				t.Fatalf("planner output should reach app repair path: %v", err)
			}
			if len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "propose_action_plan" {
				t.Fatalf("expected app action-plan tool call, got %+v", turn)
			}
		})
	}
}

func TestGoogleGeminiLanguageInferenceRejectsPlannerOutputWithoutCommands(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		`{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?"}}`,
		`{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[]}}`,
		`{"actionPlan":{"intentSummary":"Move a water bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[{"id":"cmd-1","kind":"move_asset","summary":"Move it"}]}}`,
		`{"actionPlan":{"intentSummary":"Rename a bottle.","modelInterpretationSummary":"Rename it.","confirmationSummary":"Rename it?","commands":[{"id":"cmd-1","kind":"update_asset","summary":"Rename it","arguments":{"assetId":"asset-water-bottle","title":"Bottle"}}]}}`,
		`{"actionPlan":{"intentSummary":"Move a bottle.","modelInterpretationSummary":"Move it.","confirmationSummary":"Move it?","commands":[{"kind":"move_asset","summary":"Move it","arguments":{"assetId":"asset-water-bottle"}}]}}`,
	} {
		if _, err := parseLanguageTurn(body, []ports.AgentToolDescriptor{testGeminiActionPlanToolDescriptor()}, true); err == nil {
			t.Fatalf("expected planner output without executable command envelopes to be rejected: %s", body)
		}
	}
}

func TestGoogleGeminiLanguageInferenceRetriesMalformedPlannerOutput(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"actionPlan":{"commands":[]}}`))
			return
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"actionPlan":{"intentSummary":"Move the drill.","modelInterpretationSummary":"Move it to Garage.","confirmationSummary":"Move the drill to Garage?","commands":[{"id":"cmd-move","kind":"move_asset","summary":"Move drill","arguments":{"assetId":"drill-1","parentAssetId":"garage-1"}}]}}`))
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
		Transcript: "Move the drill out to the garage.",
		PlanOnly:   true,
	})
	if err != nil {
		t.Fatalf("planner retry: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected one retry, got %d calls", calls)
	}
	if len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "propose_action_plan" {
		t.Fatalf("expected recovered planner tool call, got %+v", turn)
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
