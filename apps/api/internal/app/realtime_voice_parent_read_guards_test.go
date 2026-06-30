package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceRepairsRootContainerCreateWhenVisibleParentWasRead(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-box",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "box under the TV"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-living-room",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "living room"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-root-box",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Add an Apple TV remote to the box under the TV in the living room.",
					"modelInterpretationSummary": "The model found Living room but forgot to parent the new box there.",
					"confirmationSummary":        "Create box under TV and Apple TV remote?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-box",
							"kind":    "create_asset",
							"summary": "Create Box under the TV",
							"arguments": map[string]any{
								"title": "Box under the TV",
								"kind":  "container",
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
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-parented-box",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Add an Apple TV remote to the box under the TV in the living room.",
					"modelInterpretationSummary": "Create the missing box under the visible Living room, then create the Apple TV remote inside it.",
					"confirmationSummary":        "Create box under TV in Living room and add Apple TV remote?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-box",
							"kind":    "create_asset",
							"summary": "Create Box under the TV in Living room",
							"arguments": map[string]any{
								"title":         "Box under the TV",
								"kind":          "container",
								"parentAssetId": "living-room-1",
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
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an Apple TV remote to the box under the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedParentReadLivingRoom(t, store)

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
	if proposed == nil {
		t.Fatalf("expected repaired planner output to produce a proposal")
	}
	for _, command := range proposed.Commands {
		if command.Kind == string(actionplan.CommandKindCreateAsset) && command.AssetKind == asset.KindContainer.String() && command.Title == "Box under the TV" && command.ParentAssetID == "living-room-1" {
			if len(language.seenToolResults) < 4 || !strings.Contains(language.seenToolResults[3][2].Content, "invalid_tool_request") {
				t.Fatalf("expected invalid first plan to be returned as repair feedback, got %+v", language.seenToolResults)
			}
			return
		}
	}
	t.Fatalf("expected repaired planner output to parent new container to visible Living room, got %+v", proposed)
}

func TestRealtimeVoiceRequiresParentReadBeforePlanningNestedCreate(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-box",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "box under the TV"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-living-room",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "living room"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I checked the parent before planning.",
				DisplayResponse: "I checked the parent before planning.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an Apple TV remote to the box under the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedParentReadLivingRoom(t, store)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenPlanOnly) < 2 {
		t.Fatalf("expected at least two model turns, got %+v", language.seenPlanOnly)
	}
	if language.seenPlanOnly[1] {
		t.Fatalf("expected second turn to remain read-only before planner mode, got %+v", language.seenPlanOnly)
	}
	if len(language.seenTools[1]) != 1 || language.seenTools[1][0].Name != RealtimeVoiceToolSearchAuthorizedAssets {
		t.Fatalf("expected second turn to expose only search, got %+v", language.seenTools[1])
	}
}

func TestRealtimeVoiceRequiredParentReadExecutesServerSelectedSearch(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-box",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "box under the TV"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "list-instead",
				Name:      RealtimeVoiceToolListAuthorizedAssets,
				Arguments: map[string]any{"parentTitle": "box under the TV"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I checked the parent before planning.",
				DisplayResponse: "I checked the parent before planning.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an Apple TV remote to the box under the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedParentReadLivingRoom(t, store)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenTools) < 2 || len(language.seenTools[1]) != 1 || language.seenTools[1][0].Name != RealtimeVoiceToolSearchAuthorizedAssets {
		t.Fatalf("expected required parent-read turn to expose only search, got %+v", language.seenTools)
	}
	if len(language.seenToolResults) < 3 || len(language.seenToolResults[2]) < 2 {
		t.Fatalf("expected final turn to receive two read results, got %+v", language.seenToolResults)
	}
	parentRead := language.seenToolResults[2][1]
	if parentRead.Name != RealtimeVoiceToolSearchAuthorizedAssets || !strings.Contains(parentRead.Content, `"query":"living room"`) || !strings.Contains(parentRead.Content, "living-room-1") {
		t.Fatalf("expected server-selected parent search result, got %+v", parentRead)
	}
}

func seedParentReadLivingRoom(t *testing.T, store *memory.Store) {
	t.Helper()
	livingRoom := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	livingRoomTitle, _ := asset.NewTitle("Living room")
	livingRoom.Title = livingRoomTitle
	if err := store.CreateAsset(context.Background(), livingRoom, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "living-room-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
	}
}
