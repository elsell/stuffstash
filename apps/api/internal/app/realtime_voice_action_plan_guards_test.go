package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceRejectsMovePlanToParentNotNamedByTranscript(t *testing.T) {
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
				ID:        "search-office",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "office"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-office",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the drill to Office.",
					"modelInterpretationSummary": "The model substituted Office for the requested destination.",
					"confirmationSummary":        "Move Drill to Office?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-move",
							"kind":    "move_asset",
							"summary": "Move Drill to Office",
							"arguments": map[string]any{
								"assetId":       "drill-1",
								"parentAssetId": "office-1",
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
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	if err := store.CreateAsset(context.Background(), drill, audit.Record{ID: audit.ID("audit-drill"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "drill-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed drill: %v", err)
	}
	if err := store.CreateAsset(context.Background(), office, audit.Record{ID: audit.ID("audit-office"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "office-1", OccurredAt: time.Date(2026, 6, 26, 15, 1, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed office: %v", err)
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
			t.Fatalf("expected parent substitution to be rejected before proposal, got %+v", event.ActionPlan)
		}
	}
	if tts.lastText != "I need to know where to move it before I can prepare that move." {
		t.Fatalf("expected clarification after rejected parent substitution, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceRejectsDuplicateRootCreateWhenVisibleAssetAlreadyExists(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-combined",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "box under the TV in the living room"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-duplicate-room",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Add an Apple TV remote to the box under the TV in the living room.",
					"modelInterpretationSummary": "The whole destination phrase was missing, so the plan creates every segment.",
					"confirmationSummary":        "Create living room, box under the TV, and Apple TV remote?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-living-room",
							"kind":    "create_location",
							"summary": "Create Living room",
							"arguments": map[string]any{
								"title": "Living room",
								"kind":  "location",
							},
						},
						map[string]any{
							"id":      "cmd-box",
							"kind":    "create_asset",
							"summary": "Create box under the TV",
							"arguments": map[string]any{
								"title":           "Box under the TV",
								"kind":            "container",
								"parentCommandId": "cmd-living-room",
							},
						},
						map[string]any{
							"id":      "cmd-remote",
							"kind":    "create_asset",
							"summary": "Create Apple TV remote",
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
				SpokenResponse:  "I found an existing Living room. Please try again and I will place the new box there.",
				DisplayResponse: "I found an existing Living room. Please try again and I will place the new box there.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an Apple TV remote to the box under the TV in the living room."}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	location := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	locationTitle, _ := asset.NewTitle("Living room")
	location.Title = locationTitle
	if err := store.CreateAsset(context.Background(), location, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "living-room-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
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
		t.Fatalf("run realtime voice query: %v", err)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			t.Fatalf("expected duplicate visible create to be rejected before proposal, got %+v", event.ActionPlan)
		}
	}
	if tts.lastText != "I found an existing Living room. Please try again and I will place the new box there." {
		t.Fatalf("expected clarification after duplicate create rejection, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceRejectsMovePlanThatCreatesVisibleSourceInsteadOfMovingIt(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-water-bottle",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "water bottle"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-kitchen",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "kitchen"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-create-duplicate-source",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to the kitchen.",
					"modelInterpretationSummary": "The source was visible, but the model tried to create a new water bottle.",
					"confirmationSummary":        "Create Kitchen and a new water bottle?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"title": "Kitchen",
							},
						},
						map[string]any{
							"id":      "cmd-water-bottle",
							"kind":    "create_asset",
							"summary": "Create Water bottle",
							"arguments": map[string]any{
								"title":           "Water bottle",
								"kind":            "item",
								"parentCommandId": "cmd-kitchen",
							},
						},
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I need to move the existing water bottle, not create a duplicate.",
				DisplayResponse: "I need to move the existing water bottle, not create a duplicate.",
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
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if proposed != nil {
		t.Fatalf("expected duplicate-source create to be rejected, got %+v", proposed)
	}
}

func TestRealtimeVoiceRejectsMovePlanThatCreatesMissingSourceItem(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-passport",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "passport"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-office",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "office"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-create-source",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the passport to the office.",
					"modelInterpretationSummary": "The source search was empty, but the model tried to create the source.",
					"confirmationSummary":        "Create passport in Office?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-passport",
							"kind":    "create_asset",
							"summary": "Create Passport",
							"arguments": map[string]any{
								"title":         "Passport",
								"kind":          "item",
								"parentAssetId": "office-1",
							},
						},
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I could not find a passport to move. Add it first or tell me a different item.",
				DisplayResponse: "I could not find a passport to move. Add it first or tell me a different item.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my passport to the office."}
	resolver.providers.LanguageInference = language
	tts := &resolvedTextToSpeech{}
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	if err := store.CreateAsset(context.Background(), office, audit.Record{ID: audit.ID("audit-office"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "office-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed office: %v", err)
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
	if proposed != nil {
		t.Fatalf("expected missing-source create to be rejected, got %+v", proposed)
	}
	if tts.lastText != "I could not find a passport to move. Add it first or tell me a different item." {
		t.Fatalf("expected fall-forward response, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceRejectsInvertedMissingDestinationHierarchy(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-counter",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "counter"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-kitchen",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "kitchen"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-inverted-path",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to the counter in the kitchen.",
					"modelInterpretationSummary": "The model inverted the spoken destination path.",
					"confirmationSummary":        "Create Counter and Kitchen?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-counter",
							"kind":    "create_asset",
							"summary": "Create Counter",
							"arguments": map[string]any{
								"title": "Counter",
								"kind":  "container",
							},
						},
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"title":           "Kitchen",
								"parentCommandId": "cmd-counter",
							},
						},
						map[string]any{
							"id":      "cmd-move",
							"kind":    "move_asset",
							"summary": "Move Water bottle to Counter",
							"arguments": map[string]any{
								"assetId":         "water-bottle-1",
								"parentCommandId": "cmd-counter",
							},
						},
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I need the counter to be inside the kitchen.",
				DisplayResponse: "I need the counter to be inside the kitchen.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle onto the counter in the kitchen."}
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
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if proposed != nil {
		t.Fatalf("expected inverted hierarchy to be rejected, got %+v", proposed)
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
	plannerTurn := -1
	for index, planOnly := range language.seenPlanOnly {
		if planOnly {
			plannerTurn = index
			break
		}
	}
	if plannerTurn < 0 {
		t.Fatalf("expected a constrained planner turn, got %+v", language.seenPlanOnly)
	}
	if len(language.seenToolResults) <= plannerTurn || len(language.seenToolResults[plannerTurn]) != 2 {
		t.Fatalf("expected read tool results to be sent to planner turn, got %+v", language.seenToolResults)
	}
	if !strings.Contains(language.seenToolResults[plannerTurn][0].Content, `"assetId":"drill-1"`) || !strings.Contains(language.seenToolResults[plannerTurn][1].Content, `"assetId":"living-room-1"`) {
		t.Fatalf("expected authorized read tools to expose opaque IDs for planning, got %+v", language.seenToolResults[plannerTurn])
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
		invalidToolTurn("bad-tool-5"),
		invalidToolTurn("bad-tool-6"),
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
	if len(language.seenFinalOnly) != 7 || !language.seenFinalOnly[6] {
		t.Fatalf("expected seventh turn to be final-only, got %+v", language.seenFinalOnly)
	}
	if len(language.seenTools[6]) != 0 {
		t.Fatalf("expected no tools on final-only turn, got %+v", language.seenTools[6])
	}
	if tts.lastText != "I could not complete that with the available tools." {
		t.Fatalf("expected final-only response to be spoken, got %q", tts.lastText)
	}
}
