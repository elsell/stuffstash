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

func TestRealtimeVoiceRestoreActionPlanArgsValidateVisibleArchivedAsset(t *testing.T) {
	t.Parallel()

	for name, arguments := range map[string]map[string]any{
		"legacy single command": {
			"commandKind":                "restore_asset",
			"intentSummary":              "Restore the water bottle.",
			"modelInterpretationSummary": "The user wants the visible archived water bottle restored.",
			"confirmationSummary":        "Restore water bottle?",
			"commandSummary":             "Restore water bottle",
			"argumentsJson":              `{"assetId":"asset-id"}`,
			"riskSummary":                "Restores an item in this inventory.",
		},
		"commands array": {
			"intentSummary":              "Restore the water bottle.",
			"modelInterpretationSummary": "The user wants the visible archived water bottle restored.",
			"confirmationSummary":        "Restore water bottle?",
			"commands": []any{map[string]any{
				"id":      "cmd-restore-water-bottle",
				"kind":    "restore_asset",
				"summary": "Restore water bottle",
				"arguments": map[string]any{
					"assetId": "asset-id",
				},
			}},
			"riskSummary": "Restores an item in this inventory.",
		},
	} {
		t.Run(name, func(t *testing.T) {
			parsed, err := parseRealtimeVoiceActionPlanArgs(arguments, "Restore the water bottle.")
			if err != nil {
				t.Fatalf("parse restore action-plan args: %v", err)
			}
			if len(parsed.Commands) != 1 || parsed.Commands[0].Kind != actionplan.CommandKindRestoreAsset {
				t.Fatalf("unexpected restore commands: %+v", parsed.Commands)
			}
			if err := validateRealtimeVoiceActionPlanVisibleIDs(parsed.Commands, map[string]struct{}{"asset-id": {}}); err != nil {
				t.Fatalf("validate visible restore asset: %v", err)
			}
		})
	}
}

func TestRealtimeVoiceRestoreActionPlanSkipsMoveAndDestinationGuards(t *testing.T) {
	t.Parallel()

	priorResults := []ports.AgentToolResult{{
		Name:    RealtimeVoiceToolListAuthorizedAssets,
		Content: `{"tool":"list_authorized_assets","filters":{"kind":"item","lifecycleState":"archived"},"count":1,"items":[{"assetId":"water-bottle-1","title":"Water bottle","kind":"item","inventoryName":"Home","lifecycleState":"archived","containmentPath":["Water bottle"]}]}`,
	}}
	commands := []ActionPlanCommandInput{{
		ID:      "cmd-restore-water-bottle",
		Kind:    actionplan.CommandKindRestoreAsset,
		Summary: "Restore water bottle",
		Arguments: map[string]any{
			"assetId": "water-bottle-1",
		},
	}}
	if err := validateRealtimeVoiceMoveRequestUsesVisibleSource(commands, "Restore the water bottle.", priorResults); err != nil {
		t.Fatalf("restore should not be rejected by move source guard: %v", err)
	}
	if err := validateRealtimeVoiceMissingDestinationSegmentsAccountedFor(commands, "Restore the water bottle.", priorResults); err != nil {
		t.Fatalf("restore should not be rejected by missing destination guard: %v", err)
	}
	if err := validateRealtimeVoiceMissingDestinationHierarchy(commands, "Restore the water bottle.", priorResults); err != nil {
		t.Fatalf("restore should not be rejected by destination hierarchy guard: %v", err)
	}
}

func TestCreateActionPlanAcceptsRestoreCommandFromVoicePlanner(t *testing.T) {
	t.Parallel()

	application := newActionPlanTestApp(&fakeActionPlanRepository{}, &fakeIDGenerator{ids: []string{"plan-id"}}, nil)
	created, err := application.CreateActionPlan(context.Background(), CreateActionPlanInput{
		Principal:                  defaultRealtimeVoiceSessionInput().Principal,
		TenantID:                   "tenant-home",
		InventoryID:                "inventory-home",
		Source:                     RealtimeVoiceSourceMobile,
		RealtimeSessionID:          "voice-session-id",
		IntentSummary:              "Restore the water bottle.",
		ModelInterpretationSummary: "The user wants the visible archived water bottle restored.",
		ConfirmationSummary:        "Restore water bottle?",
		Commands: []ActionPlanCommandInput{{
			ID:      "cmd-restore-water-bottle",
			Kind:    actionplan.CommandKindRestoreAsset,
			Summary: "Restore water bottle",
			Arguments: map[string]any{
				"assetId": "water-bottle-1",
			},
		}},
		Risks: []string{"Restores an item in this inventory."},
	})
	if err != nil {
		t.Fatalf("create restore action plan: %v", err)
	}
	if created.ID != "plan-id" || len(created.Commands) != 1 || created.Commands[0].Kind != actionplan.CommandKindRestoreAsset {
		t.Fatalf("unexpected restore action plan: %+v", created)
	}
}

func TestRealtimeVoiceCanProposeRestoreActionPlanInPlannerMode(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "list-archived-water-bottle",
				Name: RealtimeVoiceToolListAuthorizedAssets,
				Arguments: map[string]any{
					"kind":           "item",
					"lifecycleState": "archived",
					"limit":          10,
				},
			}},
		},
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-tool-call",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Restore the water bottle.",
					"modelInterpretationSummary": "The user wants the visible archived water bottle restored.",
					"confirmationSummary":        "Restore water bottle?",
					"commands": []any{map[string]any{
						"id":      "cmd-restore-water-bottle",
						"kind":    "restore_asset",
						"summary": "Restore water bottle",
						"arguments": map[string]any{
							"assetId": "water-bottle-1",
						},
					}},
					"riskSummary": "Restores an item in this inventory.",
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Restore the water bottle."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStoreSessionsAndIDs(t, resolver, newFakeRealtimeSessionRepository(), &fakeIDGenerator{ids: []string{"voice-session-id", "plan-id", "command-id"}})

	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	if err := store.CreateAsset(context.Background(), waterBottle, audit.Record{ID: audit.ID("audit-water-bottle"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed water bottle: %v", err)
	}
	archivedWaterBottle := waterBottle
	archivedWaterBottle.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(context.Background(), archivedWaterBottle, audit.Record{ID: audit.ID("archive-audit-id"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetArchived, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 1, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("archive water bottle: %v", err)
	}

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), sessionInput)
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var proposed *RealtimeVoiceActionPlanProposal
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v events=%+v", err, events)
	}
	if len(language.seenPlanOnly) < 2 || !language.seenPlanOnly[1] {
		t.Fatalf("expected restore proposal turn to use constrained planner mode, got %+v", language.seenPlanOnly)
	}
	if proposed == nil || len(proposed.Commands) != 1 || proposed.Commands[0].Kind != string(actionplan.CommandKindRestoreAsset) {
		t.Fatalf("expected restore action plan proposal, got %+v", proposed)
	}
}
