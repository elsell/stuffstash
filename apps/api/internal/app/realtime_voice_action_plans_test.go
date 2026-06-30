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

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), sessionInput)
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
	if len(language.seenTools) < 2 || containsRealtimeTool(language.seenTools[0], RealtimeVoiceToolProposeActionPlan) || !containsRealtimeTool(language.seenTools[1], RealtimeVoiceToolProposeActionPlan) {
		t.Fatalf("expected first turn to gather context and second turn to include propose action plan tool, got %+v", language.seenTools)
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

func TestRealtimeVoiceCanCreateNewItemInsideExistingLocationFromVoice(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-office",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "office"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-phone-charger",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Add a phone charger to the office",
					"modelInterpretationSummary": "Add a phone charger to the office.",
					"confirmationSummary":        "Add phone charger to Office?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-phone-charger",
							"kind":    "create_asset",
							"summary": "Add phone charger",
							"arguments": map[string]any{
								"description":   "phone charger",
								"kind":          "item",
								"parentAssetId": "office-1",
								"title":         "phone charger",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add a phone charger to the office."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	if err := store.CreateAsset(context.Background(), office, audit.Record{ID: audit.ID("audit-office-phone"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "office-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed office: %v", err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var proposed *RealtimeVoiceActionPlanProposal
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v events=%+v", err, events)
	}
	if proposed == nil || len(proposed.Commands) != 1 {
		t.Fatalf("expected one proposed create command, got %+v", proposed)
	}
	command := proposed.Commands[0]
	if command.Title != "phone charger" || command.AssetKind != asset.KindItem.String() || command.ParentAssetID != "office-1" || command.ParentTitle != "Office" {
		t.Fatalf("expected phone charger inside Office, got %+v", command)
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

func TestRealtimeVoiceActionPlanProposalOrdersDependentCommandsBeforeReview(t *testing.T) {
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
							"id":      "cmd-move-water-bottle",
							"kind":    "move_asset",
							"summary": "Move Water bottle to Kitchen",
							"arguments": map[string]any{
								"assetId":         "water-bottle-1",
								"parentCommandId": "cmd-kitchen",
							},
						},
						map[string]any{
							"id":      "cmd-unrelated-box",
							"kind":    "create_asset",
							"summary": "Create unrelated box",
							"arguments": map[string]any{
								"name": "Unrelated box",
								"kind": "container",
							},
						},
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"name": "Kitchen",
								"kind": "location",
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
	if err := store.CreateAsset(context.Background(), waterBottle, audit.Record{ID: audit.ID("audit-water-bottle-order"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
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
	if proposed == nil || len(proposed.Commands) != 3 {
		t.Fatalf("expected dependent proposed plan, got %+v", proposed)
	}
	if proposed.Commands[0].ID != "cmd-unrelated-box" || proposed.Commands[0].Kind != string(actionplan.CommandKindCreateAsset) {
		t.Fatalf("expected unrelated ready command to keep stable order before later dependency parent, got %+v", proposed.Commands)
	}
	if proposed.Commands[1].ID != "cmd-kitchen" || proposed.Commands[1].Kind != string(actionplan.CommandKindCreateLocation) {
		t.Fatalf("expected create command to be ordered before dependent move, got %+v", proposed.Commands)
	}
	if proposed.Commands[2].ID != "cmd-move-water-bottle" || proposed.Commands[2].ParentCommandID != "cmd-kitchen" {
		t.Fatalf("expected dependent move after create command, got %+v", proposed.Commands)
	}
	if len(language.seenToolResults) != 2 {
		t.Fatalf("expected loop to pause after reordered proposal, got %d model turns", len(language.seenToolResults))
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

func TestRealtimeVoiceActionPlanProposalCanonicalizesProviderCreateContainerCommand(t *testing.T) {
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
					"intentSummary":              "Create a box under the TV in the living room.",
					"modelInterpretationSummary": "The user wants a new container inside an existing location.",
					"confirmationSummary":        "Create a box under the TV in the living room?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-box",
							"kind":    "create_container",
							"summary": "Create Box under the TV in Living room",
							"arguments": map[string]any{
								"title":         "Box under the TV",
								"parentAssetId": "location-1",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add a box under the TV to the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	location := assetItem("location-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	locationTitle, _ := asset.NewTitle("Living room")
	location.Title = locationTitle
	if err := store.CreateAsset(context.Background(), location, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "location-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed location: %v", err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var proposed *RealtimeVoiceActionPlanProposal
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if proposed == nil || len(proposed.Commands) != 1 {
		t.Fatalf("expected create container proposal, got %+v", proposed)
	}
	command := proposed.Commands[0]
	if command.Kind != string(actionplan.CommandKindCreateAsset) || command.AssetKind != asset.KindContainer.String() || command.Title != "Box under the TV" || command.ParentAssetID != "location-1" {
		t.Fatalf("expected create_container to persist as create_asset container, got %+v", command)
	}
}

func TestRealtimeVoiceRejectsRootMoveWhenTranscriptNamesMissingDestination(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-drill",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "drill"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-side",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "side"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-root",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the drill to the root of the inventory.",
					"modelInterpretationSummary": "The requested destination was not found, so the drill will be moved to root.",
					"confirmationSummary":        "Move Drill to the root of the inventory?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-move",
							"kind":    "move_asset",
							"summary": "Move Drill to root",
							"arguments": map[string]any{
								"assetId":       "drill-1",
								"parentAssetId": nil,
							},
						},
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I found the drill, but I could not understand the destination. Please tell me the room or container.",
				DisplayResponse: "I found the drill, but I could not understand the destination. Please tell me the room or container.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my drill to the side."}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	drillTitle, _ := asset.NewTitle("Drill")
	drill.Title = drillTitle
	if err := store.CreateAsset(context.Background(), drill, audit.Record{ID: audit.ID("audit-drill"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "drill-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed drill: %v", err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v events=%+v", err, events)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			t.Fatalf("expected unsafe root move to be rejected before proposal, got %+v", event.ActionPlan)
		}
	}
	if tts.lastText != "I need to know where to move it before I can prepare that move." {
		t.Fatalf("expected clarification after rejected root move, got %q", tts.lastText)
	}
}

func containsRealtimeTool(tools []ports.AgentToolDescriptor, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}
