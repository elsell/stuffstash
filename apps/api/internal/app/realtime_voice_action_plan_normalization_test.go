package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceActionPlanProposalCanonicalizesCreateAssetIDAsParentAssetID(t *testing.T) {
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
					"modelInterpretationSummary": "The model put the visible parent id in assetId on a create command.",
					"confirmationSummary":        "Create a box under the TV in Living room?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-box",
							"kind":    "create_asset",
							"summary": "Create Box under the TV in Living room",
							"arguments": map[string]any{
								"assetId": "living-room-1",
								"title":   "Box under the TV",
								"kind":    "container",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Create a box under the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	livingRoom := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	livingRoomTitle, _ := asset.NewTitle("Living room")
	livingRoom.Title = livingRoomTitle
	if err := store.CreateAsset(context.Background(), livingRoom, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "living-room-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
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
		t.Fatalf("expected canonical create proposal, got %+v", proposed)
	}
	if proposed.Commands[0].ParentAssetID != "living-room-1" || proposed.Commands[0].ParentTitle != "Living room" {
		t.Fatalf("expected create assetId to be treated as parentAssetId, got %+v", proposed.Commands[0])
	}
}

func TestRealtimeVoiceActionPlanProposalFoldsCreatedItemMoveIntoCreateCommand(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-toolbox",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "toolbox"},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Create AA batteries and put them in the toolbox.",
					"modelInterpretationSummary": "The model represented the new item as a create followed by a move.",
					"confirmationSummary":        "Create AA batteries in the Toolbox?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-batteries",
							"kind":    "create_asset",
							"summary": "Create AA batteries",
							"arguments": map[string]any{
								"title": "AA batteries",
								"kind":  "container",
							},
						},
						map[string]any{
							"id":      "cmd-move",
							"kind":    "move_asset",
							"summary": "Move AA batteries into Toolbox",
							"arguments": map[string]any{
								"assetId":         "toolbox-1",
								"parentCommandId": "cmd-batteries",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "I got a pack of AA batteries. Put it in the toolbox."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	toolbox := assetItem("toolbox-1", "tenant-home", "inventory-home", asset.KindContainer, "")
	toolboxTitle, _ := asset.NewTitle("Toolbox")
	toolbox.Title = toolboxTitle
	if err := store.CreateAsset(context.Background(), toolbox, audit.Record{ID: audit.ID("audit-toolbox"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "toolbox-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed toolbox: %v", err)
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
		t.Fatalf("expected folded create proposal, got %+v", proposed)
	}
	command := proposed.Commands[0]
	if command.Kind != string(actionplan.CommandKindCreateAsset) || command.AssetKind != asset.KindItem.String() || command.ParentAssetID != "toolbox-1" || command.ParentTitle != "Toolbox" {
		t.Fatalf("expected created item to be placed directly in Toolbox, got %+v", command)
	}
}

func TestRealtimeVoiceActionPlanProposalFoldsGuessedCreatedItemMoveIntoCreateCommand(t *testing.T) {
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
					"intentSummary":              "Create an HDMI cable and put it in a drawer in the living room.",
					"modelInterpretationSummary": "The provider guessed an asset id for the newly-created HDMI cable.",
					"confirmationSummary":        "Create an HDMI cable in a new drawer?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-drawer",
							"kind":    "create_asset",
							"summary": "Create drawer",
							"arguments": map[string]any{
								"title":         "Drawer",
								"kind":          "container",
								"parentAssetId": "living-room-1",
							},
						},
						map[string]any{
							"id":      "cmd-cable",
							"kind":    "create_asset",
							"summary": "Create HDMI cable",
							"arguments": map[string]any{
								"title": "spare HDMI cable",
								"kind":  "item",
							},
						},
						map[string]any{
							"id":      "cmd-move",
							"kind":    "move_asset",
							"summary": "Move HDMI cable to drawer",
							"arguments": map[string]any{
								"assetId":         "hdmi-cable-1",
								"parentCommandId": "cmd-drawer",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Put a spare HDMI cable in the drawer under the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	livingRoom := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	livingRoomTitle, _ := asset.NewTitle("Living room")
	livingRoom.Title = livingRoomTitle
	if err := store.CreateAsset(context.Background(), livingRoom, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "living-room-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
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
	if proposed == nil || len(proposed.Commands) != 2 {
		t.Fatalf("expected folded nested create proposal, got %+v", proposed)
	}
	if proposed.Commands[1].Title != "spare HDMI cable" || proposed.Commands[1].ParentCommandID != "cmd-drawer" {
		t.Fatalf("expected HDMI cable create to point at drawer command, got %+v", proposed.Commands)
	}
}

func TestRealtimeVoiceActionPlanProposalPlacesSingleCreatedItemInsideSingleCreatedContainer(t *testing.T) {
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
					"intentSummary":              "Put a spare HDMI cable in a drawer in the living room.",
					"modelInterpretationSummary": "The provider created the item and drawer but attached them incorrectly.",
					"confirmationSummary":        "Create a drawer and HDMI cable?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-cable",
							"kind":    "create_asset",
							"summary": "Create HDMI cable",
							"arguments": map[string]any{
								"title":         "HDMI cable",
								"kind":          "item",
								"parentAssetId": "living-room-1",
							},
						},
						map[string]any{
							"id":      "cmd-drawer",
							"kind":    "create_asset",
							"summary": "Create drawer",
							"arguments": map[string]any{
								"title":           "Drawer",
								"kind":            "container",
								"parentAssetId":   "living-room-1",
								"parentCommandId": "cmd-cable",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Put a spare HDMI cable in the drawer under the TV in the living room."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	livingRoom := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	livingRoomTitle, _ := asset.NewTitle("Living room")
	livingRoom.Title = livingRoomTitle
	if err := store.CreateAsset(context.Background(), livingRoom, audit.Record{ID: audit.ID("audit-living-room"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "living-room-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed living room: %v", err)
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
	if proposed == nil || len(proposed.Commands) != 2 {
		t.Fatalf("expected corrected create proposal, got %+v", proposed)
	}
	if proposed.Commands[0].ID != "cmd-drawer" || proposed.Commands[0].ParentAssetID != "living-room-1" {
		t.Fatalf("expected drawer under living room first, got %+v", proposed.Commands)
	}
	if proposed.Commands[1].ID != "cmd-cable" || proposed.Commands[1].ParentCommandID != "cmd-drawer" {
		t.Fatalf("expected HDMI cable inside drawer, got %+v", proposed.Commands)
	}
}
