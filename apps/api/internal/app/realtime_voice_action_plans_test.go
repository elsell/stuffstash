package app

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceCanProposePersistedActionPlanForMobileReview(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-living-room",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "living room"},
			}},
		},
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
	resolver.providers.TextToSpeech = tts
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
	if len(language.seenToolResults) != 2 {
		t.Fatalf("expected loop to stop after proposal without requesting final turn, got %d turns", len(language.seenToolResults))
	}
	if tts.lastText != "" {
		t.Fatalf("expected no speech synthesis while review is pending, got %q", tts.lastText)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted || event.Type == RealtimeVoiceEventSessionCompleted {
			t.Fatalf("expected review proposal to pause before final completion, got %+v in %+v", event, events)
		}
	}
}

func TestRealtimeVoiceActionPlanProposalPersistsNativeObjectArguments(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-living-room",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "living room"},
			}},
		},
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
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	location := assetItem("location-living-room", "tenant-home", "inventory-home", asset.KindLocation, "")
	locationTitle, _ := asset.NewTitle("Living room")
	location.Title = locationTitle
	if err := store.CreateAsset(context.Background(), location, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "location-living-room", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
	}

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

func TestRealtimeVoiceActionPlanProposalSupportsDependentCreateCommands(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-living-room",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "living room"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Create an Apple TV remote in the box underneath the TV in the living room.",
					"modelInterpretationSummary": "The user wants a new container and a new item placed inside it.",
					"confirmationSummary":        "Create a box underneath the TV and add an Apple TV remote inside it?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-box",
							"kind":    "create_asset",
							"summary": "Create Box underneath the TV in Living room",
							"arguments": map[string]any{
								"title":         "Box underneath the TV",
								"kind":          "container",
								"parentAssetId": "location-1",
							},
						},
						map[string]any{
							"id":      "cmd-remote",
							"kind":    "create_asset",
							"summary": "Create Apple TV remote inside Box underneath the TV",
							"arguments": map[string]any{
								"title":           "Apple TV remote",
								"kind":            "item",
								"parentCommandId": "cmd-box",
							},
						},
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
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add my Apple TV remote to the box underneath the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	location := assetItem("location-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	locationTitle, _ := asset.NewTitle("Living room")
	location.Title = locationTitle
	if err := store.CreateAsset(context.Background(), location, audit.Record{ID: audit.ID("audit-location"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "location-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed location: %v", err)
	}

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
	if proposed == nil || len(proposed.Commands) != 2 {
		t.Fatalf("expected two proposed commands, got %+v", proposed)
	}
	if proposed.Commands[0].ID != "cmd-box" || proposed.Commands[0].Operation != "create" || proposed.Commands[0].Title != "Box underneath the TV" || proposed.Commands[0].AssetKind != "container" || proposed.Commands[0].ParentAssetID != "location-1" || proposed.Commands[0].ParentTitle != "Living room" {
		t.Fatalf("unexpected first command detail: %+v", proposed.Commands[0])
	}
	if proposed.Commands[1].ID != "cmd-remote" || proposed.Commands[1].Title != "Apple TV remote" || proposed.Commands[1].ParentCommandID != "cmd-box" {
		t.Fatalf("unexpected second command detail: %+v", proposed.Commands[1])
	}
}

func TestRealtimeVoiceActionPlanProposalCanonicalizesDependentParentAssetIDCommandReference(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-water-bottle",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "water bottle", "limit": float64(10)},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to a new Kitchen location.",
					"modelInterpretationSummary": "The user wants the visible water bottle moved into Kitchen, which should be created.",
					"confirmationSummary":        "Create Kitchen and move the water bottle there?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_asset",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"title": "Kitchen",
								"kind":  "location",
							},
						},
						map[string]any{
							"id":      "cmd-move-water-bottle",
							"kind":    "move_asset",
							"summary": "Move Water bottle to Kitchen",
							"arguments": map[string]any{
								"assetId":       "water-bottle-1",
								"parentAssetId": "cmd-kitchen",
							},
						},
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I should not be called after the proposal.",
				DisplayResponse: "I should not be called after the proposal.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	if err := store.CreateAsset(context.Background(), waterBottle, audit.Record{ID: audit.ID("audit-water-bottle"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed water bottle: %v", err)
	}

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
	if proposed == nil || len(proposed.Commands) != 2 {
		t.Fatalf("expected dependent proposed plan, got %+v", proposed)
	}
	if proposed.Commands[0].Kind != string(actionplan.CommandKindCreateLocation) || proposed.Commands[0].AssetKind != asset.KindLocation.String() {
		t.Fatalf("expected create_asset location command to canonicalize to create_location, got %+v", proposed.Commands[0])
	}
	if proposed.Commands[1].ParentCommandID != "cmd-kitchen" || proposed.Commands[1].ParentAssetID != "" {
		t.Fatalf("expected parentAssetId command reference to canonicalize to parentCommandId, got %+v", proposed.Commands[1])
	}
	if len(language.seenToolResults) != 2 {
		t.Fatalf("expected loop to pause after canonicalized proposal, got %d model turns", len(language.seenToolResults))
	}
}

func TestRealtimeVoiceRejectsActionPlanIDsNotReturnedByReadTools(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-move",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "move_asset",
					"intentSummary":              "Move the drill to the living room.",
					"modelInterpretationSummary": "The user wants to move a drill.",
					"confirmationSummary":        "Move Drill to Living room?",
					"commandSummary":             "Move Drill to Living room",
					"arguments": map[string]any{
						"assetId":       "drill-1",
						"parentAssetId": "living-room-1",
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I need to find the drill and destination before preparing that move.",
				DisplayResponse: "I need to find the drill and destination before preparing that move.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my drill to the living room."}
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
	if proposed != nil {
		t.Fatalf("expected unprovenanced asset IDs to be rejected before proposal, got %+v", proposed)
	}
	if len(language.seenToolResults) < 2 || !strings.Contains(language.seenToolResults[1][0].Content, `"code":"invalid_tool_request"`) {
		t.Fatalf("expected model-visible invalid tool result, got %+v", language.seenToolResults)
	}
}

func TestRealtimeVoiceCanGatherContextAndProposeMoveActionPlan(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{
				{
					ID:   "tool-search-drill",
					Name: RealtimeVoiceToolSearchAuthorizedAssets,
					Arguments: map[string]any{
						"query": "drill",
						"limit": 5,
					},
				},
				{
					ID:   "tool-search-living-room",
					Name: RealtimeVoiceToolSearchAuthorizedAssets,
					Arguments: map[string]any{
						"query": "living room",
						"limit": 5,
					},
				},
			},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-move",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "move_asset",
					"intentSummary":              "Move the drill to the living room.",
					"modelInterpretationSummary": "The user wants the visible drill moved into the visible Living room location.",
					"confirmationSummary":        "Move Drill to Living room?",
					"commandSummary":             "Move Drill to Living room",
					"arguments": map[string]any{
						"assetId":       "drill-1",
						"parentAssetId": "living-room-1",
					},
					"risks": []any{"Moves an existing item to a different location."},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I prepared that move for review.",
				DisplayResponse: "I prepared that move for review.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my drill to the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	drillTitle, _ := asset.NewTitle("Drill")
	drill.Title = drillTitle
	livingRoom := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	livingRoomTitle, _ := asset.NewTitle("Living room")
	livingRoom.Title = livingRoomTitle
	if err := store.CreateAsset(context.Background(), drill, audit.Record{ID: audit.ID("audit-drill"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "drill-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed drill: %v", err)
	}
	if err := store.CreateAsset(context.Background(), livingRoom, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "living-room-1", OccurredAt: time.Date(2026, 6, 26, 15, 1, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
	}

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
	if len(language.seenToolResults) < 2 || len(language.seenToolResults[1]) != 2 {
		t.Fatalf("expected first turn tool results to be sent back to model, got %+v", language.seenToolResults)
	}
	if !strings.Contains(language.seenToolResults[1][0].Content, `"assetId":"drill-1"`) || !strings.Contains(language.seenToolResults[1][1].Content, `"assetId":"living-room-1"`) {
		t.Fatalf("expected authorized read tools to expose opaque IDs for planning, got %+v", language.seenToolResults[1])
	}
	if proposed == nil || len(proposed.Commands) != 1 {
		t.Fatalf("expected move plan proposal, got %+v", proposed)
	}
	if proposed.Commands[0].Kind != string(actionplan.CommandKindMoveAsset) || proposed.Commands[0].Operation != "move" || proposed.Commands[0].Summary != "Move Drill to Living room" {
		t.Fatalf("unexpected move command proposal: %+v", proposed.Commands[0])
	}
}

func TestRealtimeVoiceRequestsFinalOnlyTurnWhenToolBudgetIsExhausted(t *testing.T) {
	t.Parallel()

	invalidToolTurn := func(id string) ports.LanguageInferenceTurn {
		return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{{
			ID:   id,
			Name: "unknown_tool",
			Arguments: map[string]any{
				"query": id,
			},
		}}}
	}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		invalidToolTurn("bad-tool-1"),
		invalidToolTurn("bad-tool-2"),
		invalidToolTurn("bad-tool-3"),
		invalidToolTurn("bad-tool-4"),
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindUnsupportedAction,
				SpokenResponse:  "I could not complete that with the available tools.",
				DisplayResponse: "I could not complete that with the available tools.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my drill somewhere."}
	resolver.providers.LanguageInference = language
	tts := &resolvedTextToSpeech{}
	resolver.providers.TextToSpeech = tts
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenFinalOnly) != 5 || !language.seenFinalOnly[4] {
		t.Fatalf("expected fifth turn to be final-only, got %+v", language.seenFinalOnly)
	}
	if len(language.seenTools[4]) != 0 {
		t.Fatalf("expected no tools on final-only turn, got %+v", language.seenTools[4])
	}
	if tts.lastText != "I could not complete that with the available tools." {
		t.Fatalf("expected final-only response to be spoken, got %q", tts.lastText)
	}
}

type scriptedRealtimeLanguageInference struct {
	turns           []ports.LanguageInferenceTurn
	seenTools       [][]ports.AgentToolDescriptor
	seenToolResults [][]ports.AgentToolResult
	seenFinalOnly   []bool
}

func (s *scriptedRealtimeLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	s.seenTools = append(s.seenTools, append([]ports.AgentToolDescriptor{}, input.Tools...))
	s.seenToolResults = append(s.seenToolResults, append([]ports.AgentToolResult{}, input.ToolResults...))
	s.seenFinalOnly = append(s.seenFinalOnly, input.FinalOnly)
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
