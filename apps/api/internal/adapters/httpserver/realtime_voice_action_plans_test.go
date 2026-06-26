package httpserver

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryStreamsActionPlanProposalForReview(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "plan-id", "command-id", "response-id"},
	}, fakeSpeechToText{transcript: "Add a water bottle."}, actionPlanProposalLanguageModel{}, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	proposed := findRealtimeEvent(t, events, "action.plan.proposed")
	actionPlan, ok := proposed["actionPlan"].(map[string]any)
	if !ok {
		t.Fatalf("expected action plan payload, got %+v", proposed)
	}
	if actionPlan["planId"] != "plan-id" || actionPlan["confirmationSummary"] != "Create item water bottle?" {
		t.Fatalf("unexpected action plan payload: %+v", actionPlan)
	}
	commands, ok := actionPlan["commands"].([]any)
	if !ok || len(commands) != 1 {
		t.Fatalf("expected one safe command, got %+v", actionPlan["commands"])
	}
	command, ok := commands[0].(map[string]any)
	if !ok || command["kind"] != "create_asset" || command["summary"] != "Create item water bottle" {
		t.Fatalf("unexpected command payload: %+v", commands[0])
	}
	assertSafeRealtimeEvents(t, events, []string{"fake-audio", "apiKey", "Bearer", "provider_session_id"})
}

type actionPlanProposalLanguageModel struct{}

func (m actionPlanProposalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if len(input.ToolResults) == 0 {
		return ports.LanguageInferenceTurn{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: "propose_action_plan",
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create a water bottle item.",
					"modelInterpretationSummary": "The user wants to add a water bottle to this inventory.",
					"confirmationSummary":        "Create item water bottle?",
					"commandSummary":             "Create item water bottle",
					"argumentsJson":              `{"kind":"item","name":"water bottle"}`,
					"riskSummary":                "Adds a new item to this inventory.",
				},
			}},
		}, nil
	}
	return ports.LanguageInferenceTurn{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindClarification,
			SpokenResponse:  "I prepared that change for review.",
			DisplayResponse: "I prepared that change for review.",
		},
	}, nil
}
