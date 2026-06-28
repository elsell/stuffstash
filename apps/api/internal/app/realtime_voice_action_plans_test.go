package app

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceCanProposePersistedActionPlanForMobileReview(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create a water bottle item.",
					"modelInterpretationSummary": "The user wants to add a water bottle to the selected inventory.",
					"confirmationSummary":        "Create item water bottle?",
					"commandSummary":             "Create item water bottle",
					"arguments": map[string]any{
						"name": "water bottle",
						"kind": "item",
					},
					"risks": []any{"Adds a new item to this inventory."},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I prepared that change for review.",
				DisplayResponse: "I prepared that change for review.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add a water bottle."}
	resolver.providers.LanguageInference = language
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}

	var proposed *RealtimeVoiceActionPlanProposal
	for _, event := range events {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
			break
		}
	}
	if proposed == nil {
		t.Fatalf("expected proposed action plan event, got %+v", events)
	}
	if proposed.PlanID == "" || proposed.ConfirmationSummary != "Create item water bottle?" {
		t.Fatalf("unexpected proposed plan: %+v", proposed)
	}
	if len(proposed.Commands) != 1 || proposed.Commands[0].Kind != string(actionplan.CommandKindCreateAsset) || proposed.Commands[0].Summary != "Create item water bottle" {
		t.Fatalf("unexpected proposed commands: %+v", proposed.Commands)
	}
	if len(proposed.Risks) != 1 || proposed.Risks[0] != "Adds a new item to this inventory." {
		t.Fatalf("unexpected proposed risks: %+v", proposed.Risks)
	}
	if len(language.seenTools) == 0 || !containsRealtimeTool(language.seenTools[0], RealtimeVoiceToolProposeActionPlan) {
		t.Fatalf("expected language provider to receive propose action plan tool, got %+v", language.seenTools)
	}
}

func TestRealtimeVoiceActionPlanProposalPersistsNativeObjectArguments(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create an Apple TV remote in the living room.",
					"modelInterpretationSummary": "The user wants to add an Apple TV remote item inside the existing Living room location.",
					"confirmationSummary":        "Create an Apple TV remote in the living room?",
					"commandSummary":             "Create an Apple TV remote in Living room",
					"arguments": map[string]any{
						"title":         "Apple TV remote",
						"kind":          "item",
						"parentAssetId": "location-living-room",
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I prepared that change for review.",
				DisplayResponse: "I prepared that change for review.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "I'd like to add an Apple TV remote to the living room."}
	resolver.providers.LanguageInference = language
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var proposed *RealtimeVoiceActionPlanProposal
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if proposed == nil || len(proposed.Commands) != 1 {
		t.Fatalf("expected one proposed command, got %+v", proposed)
	}
	record, found, err := application.actionPlans.ActionPlanByID(context.Background(), session.TenantID, session.InventoryID, proposed.PlanID)
	if err != nil {
		t.Fatalf("read proposed action plan: %v", err)
	}
	if !found || len(record.Commands) != 1 {
		t.Fatalf("expected persisted proposed command, found=%v record=%+v", found, record)
	}
	var arguments map[string]any
	if err := json.Unmarshal(record.Commands[0].ArgumentsJSON, &arguments); err != nil {
		t.Fatalf("decode persisted command arguments: %v", err)
	}
	if arguments["title"] != "Apple TV remote" || arguments["kind"] != "item" || arguments["parentAssetId"] != "location-living-room" {
		t.Fatalf("expected structured action-plan arguments to be preserved, got %+v", arguments)
	}
}

func TestRealtimeVoiceRejectsUnsafeActionPlanProposalArguments(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create an item.",
					"modelInterpretationSummary": "The user wants to add an item.",
					"confirmationSummary":        "Create item?",
					"commandSummary":             "Create item",
					"arguments": map[string]any{
						"apiKey": "secret",
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an item."}
	resolver.providers.LanguageInference = language
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "validation") {
		t.Fatalf("expected unsafe proposal to fail validation, got %v", err)
	}
}

type scriptedRealtimeLanguageInference struct {
	turns     []ports.LanguageInferenceTurn
	seenTools [][]ports.AgentToolDescriptor
}

func (s *scriptedRealtimeLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	s.seenTools = append(s.seenTools, append([]ports.AgentToolDescriptor{}, input.Tools...))
	if len(s.turns) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	turn := s.turns[0]
	s.turns = s.turns[1:]
	return turn, nil
}

func containsRealtimeTool(tools []ports.AgentToolDescriptor, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}
