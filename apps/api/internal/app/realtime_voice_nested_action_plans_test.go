package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceRepairsCreateClarificationIntoNestedActionPlan(t *testing.T) {
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
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I can't find the second shelf in the big cabinet in the kitchen. Do you want me to create it?",
				DisplayResponse: "I can't find the second shelf in the big cabinet in the kitchen. Do you want me to create it?",
			},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to the second shelf in the big cabinet in the kitchen.",
					"modelInterpretationSummary": "The user wants missing destination parents created and the visible water bottle moved there.",
					"confirmationSummary":        "Create Kitchen, Big cabinet, Second shelf, and move the water bottle there?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"title": "Kitchen",
								"kind":  "location",
							},
						},
						map[string]any{
							"id":      "cmd-big-cabinet",
							"kind":    "create_asset",
							"summary": "Create Big cabinet inside Kitchen",
							"arguments": map[string]any{
								"title":           "Big cabinet",
								"kind":            "container",
								"parentCommandId": "cmd-kitchen",
							},
						},
						map[string]any{
							"id":      "cmd-second-shelf",
							"kind":    "create_asset",
							"summary": "Create Second shelf inside Big cabinet",
							"arguments": map[string]any{
								"title":           "Second shelf",
								"kind":            "container",
								"parentCommandId": "cmd-big-cabinet",
							},
						},
						map[string]any{
							"id":      "cmd-move-water-bottle",
							"kind":    "move_asset",
							"summary": "Move Water bottle to Second shelf",
							"arguments": map[string]any{
								"assetId":         "water-bottle-1",
								"parentCommandId": "cmd-second-shelf",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the second shelf in the big cabinet in the kitchen."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	if err := store.CreateAsset(context.Background(), waterBottle, audit.Record{ID: audit.ID("audit-water-bottle-nested"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
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
	if proposed == nil || len(proposed.Commands) != 4 {
		t.Fatalf("expected nested proposed plan, got %+v", proposed)
	}
	if len(language.seenToolResults) < 3 || len(language.seenToolResults[2]) < 2 {
		t.Fatalf("expected repair feedback before proposal turn, got %+v", language.seenToolResults)
	}
	repair := language.seenToolResults[2][1]
	if repair.Name != RealtimeVoiceToolProposeActionPlan || !strings.Contains(repair.Content, "final_clarification_rejected") || !strings.Contains(repair.Content, "parentCommandId") {
		t.Fatalf("expected retryable action-plan repair feedback, got %+v", repair)
	}
	if proposed.Commands[0].Title != "Kitchen" || proposed.Commands[1].Title != "Big cabinet" || proposed.Commands[1].ParentCommandID != "cmd-kitchen" || proposed.Commands[2].Title != "Second shelf" || proposed.Commands[2].ParentCommandID != "cmd-big-cabinet" || proposed.Commands[3].ParentCommandID != "cmd-second-shelf" {
		t.Fatalf("unexpected nested proposal: %+v", proposed.Commands)
	}
}

func TestRealtimeVoiceRejectsCreatedDestinationPlanWhenMoveDoesNotReferenceCreate(t *testing.T) {
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
				ID:   "tool-plan-bad",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to the kitchen.",
					"modelInterpretationSummary": "Create Kitchen and move the visible water bottle there.",
					"confirmationSummary":        "Create Kitchen and move the water bottle there?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-move-water-bottle",
							"kind":    "move_asset",
							"summary": "Move Water bottle to root",
							"arguments": map[string]any{
								"assetId":       "water-bottle-1",
								"parentAssetId": nil,
							},
						},
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
							"summary": "Create Kitchen",
							"arguments": map[string]any{
								"title": "Kitchen",
								"kind":  "location",
							},
						},
					},
				},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-good",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Move the water bottle to the kitchen.",
					"modelInterpretationSummary": "Create Kitchen and move the visible water bottle there.",
					"confirmationSummary":        "Create Kitchen and move the water bottle there?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-kitchen",
							"kind":    "create_location",
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
								"assetId":         "water-bottle-1",
								"parentCommandId": "cmd-kitchen",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	if err := store.CreateAsset(context.Background(), waterBottle, audit.Record{ID: audit.ID("audit-water-bottle-root-reject"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
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
	if proposed == nil || len(proposed.Commands) != 2 || proposed.Commands[1].ParentCommandID != "cmd-kitchen" {
		t.Fatalf("expected repaired create-and-move proposal, got %+v", proposed)
	}
	if len(language.seenToolResults) < 3 || !strings.Contains(language.seenToolResults[2][1].Content, "invalid_tool_request") {
		t.Fatalf("expected invalid tool feedback for unlinked move, got %+v", language.seenToolResults)
	}
}

func TestRealtimeVoiceDoesNotRepairCreateClarificationWhenSourceIsMissing(t *testing.T) {
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
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I can't find the water bottle. Do you want me to add it?",
				DisplayResponse: "I can't find the water bottle. Do you want me to add it?",
			},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "tool-plan-unreachable",
				Name:      RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	tts := &resolvedTextToSpeech{}
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

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
		t.Fatalf("expected no action plan for missing source, got %+v", proposed)
	}
	if len(language.seenToolResults) != 2 {
		t.Fatalf("expected loop to stop on missing-source clarification, got %d calls", len(language.seenToolResults))
	}
	if tts.lastText != "I can't find the water bottle. Do you want me to add it?" {
		t.Fatalf("expected final clarification to be spoken, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceDoesNotRepairCreateClarificationForUnrelatedVisibleAsset(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-drill",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "drill", "limit": float64(10)},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I can't find the water bottle. Do you want me to add it?",
				DisplayResponse: "I can't find the water bottle. Do you want me to add it?",
			},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "tool-plan-unreachable",
				Name:      RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	tts := &resolvedTextToSpeech{}
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	drillTitle, _ := asset.NewTitle("Cordless drill")
	drill.Title = drillTitle
	if err := store.CreateAsset(context.Background(), drill, audit.Record{ID: audit.ID("audit-unrelated-drill"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "drill-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed unrelated drill: %v", err)
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
	if proposed != nil {
		t.Fatalf("expected no action plan for unrelated visible asset, got %+v", proposed)
	}
	if len(language.seenToolResults) != 2 {
		t.Fatalf("expected loop to stop on unrelated-source clarification, got %d calls", len(language.seenToolResults))
	}
	if tts.lastText != "I can't find the water bottle. Do you want me to add it?" {
		t.Fatalf("expected final clarification to be spoken, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceAllowsRootMoveWithUnrelatedCreate(t *testing.T) {
	t.Parallel()

	args := map[string]any{
		"intentSummary":              "Create a spare box and move the water bottle to root.",
		"modelInterpretationSummary": "The user wants a new spare box, and separately wants the visible water bottle moved out of its parent.",
		"confirmationSummary":        "Create Spare box and move the Water bottle to root?",
		"commands": []any{
			map[string]any{
				"id":      "cmd-spare-box",
				"kind":    "create_asset",
				"summary": "Create Spare box",
				"arguments": map[string]any{
					"title": "Spare box",
					"kind":  "container",
				},
			},
			map[string]any{
				"id":      "cmd-move-water-bottle",
				"kind":    "move_asset",
				"summary": "Move Water bottle to root",
				"arguments": map[string]any{
					"assetId":       "water-bottle-1",
					"parentAssetId": nil,
				},
			},
		},
	}

	parsed, err := parseRealtimeVoiceActionPlanArgs(args, "Move my water bottle to root.")
	if err != nil {
		t.Fatalf("expected unrelated root move to parse: %v", err)
	}
	if len(parsed.Commands) != 2 || stringArg(parsed.Commands[1].Arguments["parentAssetId"]) != "" || stringArg(parsed.Commands[1].Arguments["parentCommandId"]) != "" {
		t.Fatalf("unexpected parsed root move: %+v", parsed.Commands)
	}
}
