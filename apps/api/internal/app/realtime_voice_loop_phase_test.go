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

func TestRealtimeVoiceFinalizesReadOnlyAnswerWithConstrainedFinalTurn(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-water-bottle",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "water bottle"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "Your water bottle is in the Office.",
				DisplayResponse: "Your water bottle is in the Office.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is my water bottle?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "office-1")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	if err := store.CreateAsset(context.Background(), office, audit.Record{ID: audit.ID("audit-office"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "office-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed office: %v", err)
	}
	if err := store.CreateAsset(context.Background(), waterBottle, audit.Record{ID: audit.ID("audit-water-bottle"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 26, 15, 1, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed water bottle: %v", err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}

	if len(language.seenFinalOnly) != 2 || !language.seenFinalOnly[1] {
		t.Fatalf("expected second turn to be final-only, got %+v", language.seenFinalOnly)
	}
	if len(language.seenTools) != 2 || len(language.seenTools[1]) != 0 {
		t.Fatalf("expected final-only turn to omit tools, got %+v", language.seenTools)
	}
	if tts.lastText != "Your water bottle is in the Office." {
		t.Fatalf("expected final response to be spoken, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceSelectsProactiveReadOnlyCallsForUnambiguousTranscripts(t *testing.T) {
	t.Parallel()

	call, _, ok := realtimeVoiceServerSelectedReadCallWithoutModel("What's in the toolbox?", 0, nil, "read-1")
	if !ok || call.Name != RealtimeVoiceToolSearchAuthorizedAssets || call.Arguments["query"] != "toolbox" {
		t.Fatalf("expected proactive toolbox search, got ok=%v call=%+v", ok, call)
	}
	call, _, ok = realtimeVoiceServerSelectedReadCallWithoutModel("Add a phone charger to the office.", 0, nil, "read-2")
	if !ok || call.Name != RealtimeVoiceToolSearchAuthorizedAssets || call.Arguments["query"] != "office" {
		t.Fatalf("expected proactive office search, got ok=%v call=%+v", ok, call)
	}
	_, _, ok = realtimeVoiceServerSelectedReadCallWithoutModel("Add an Apple TV remote to the box under the TV in the living room.", 0, nil, "read-3")
	if ok {
		t.Fatalf("did not expect proactive read for nested create path")
	}
	for _, transcript := range []string{
		"Put my water bottle in the kitchen.",
		"Store the drill in the garage.",
		"Stash it in the toolbox.",
		"Place my passport in the office.",
		"Move my water bottle to the kitchen.",
	} {
		if call, _, ok := realtimeVoiceServerSelectedReadCallWithoutModel(transcript, 0, nil, "read-move-like"); ok {
			t.Fatalf("did not expect proactive destination read for %q, got %+v", transcript, call)
		}
	}
}

func TestRealtimeVoiceListsContentsAfterContainerSearchForContentsQuestion(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "The toolbox contains the cordless drill.",
				DisplayResponse: "The toolbox contains the cordless drill.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "What's in the toolbox?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	garage := assetItem("garage-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	garageTitle, _ := asset.NewTitle("Garage")
	garage.Title = garageTitle
	toolbox := assetItem("toolbox-1", "tenant-home", "inventory-home", asset.KindContainer, "garage-1")
	toolboxTitle, _ := asset.NewTitle("Toolbox")
	toolbox.Title = toolboxTitle
	drill := assetItem("cordless-drill-1", "tenant-home", "inventory-home", asset.KindItem, "toolbox-1")
	drillTitle, _ := asset.NewTitle("Cordless drill")
	drill.Title = drillTitle
	seedRealtimeVoiceLoopAsset(t, store, garage, "audit-garage")
	seedRealtimeVoiceLoopAsset(t, store, toolbox, "audit-toolbox")
	seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}
	if len(language.seenTools) != 1 || len(language.seenTools[0]) != 0 {
		t.Fatalf("expected only final model turn after server-selected reads, got %+v", language.seenTools)
	}
	if len(language.seenToolResults) != 1 || len(language.seenToolResults[0]) < 2 || language.seenToolResults[0][0].Name != RealtimeVoiceToolSearchAuthorizedAssets || language.seenToolResults[0][1].Name != RealtimeVoiceToolListAuthorizedAssets || !containsAll(language.seenToolResults[0][1].Content, "Toolbox", "Cordless drill") {
		t.Fatalf("expected contents list result with drill, got %+v", language.seenToolResults)
	}
	if tts.lastText != "The toolbox contains the cordless drill." {
		t.Fatalf("expected contents answer, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceDoesNotBroadListAfterPluralWhereNoMatch(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-tools",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "tools"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I could not find anything matching tools.",
				DisplayResponse: "I could not find anything matching tools.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where are my tools?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	garage := assetItem("garage-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	garageTitle, _ := asset.NewTitle("Garage")
	garage.Title = garageTitle
	drill := assetItem("cordless-drill-1", "tenant-home", "inventory-home", asset.KindItem, "garage-1")
	drillTitle, _ := asset.NewTitle("Cordless drill")
	drill.Title = drillTitle
	seedRealtimeVoiceLoopAsset(t, store, garage, "audit-garage")
	seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}
	if len(language.seenFinalOnly) != 2 || !language.seenFinalOnly[1] {
		t.Fatalf("expected final-only turn after no-match search, got %+v", language.seenFinalOnly)
	}
	if len(language.seenTools) != 2 || len(language.seenTools[1]) != 0 {
		t.Fatalf("expected no broad list tools on final-only turn, got %+v", language.seenTools)
	}
	if tts.lastText != "I could not find anything matching tools." {
		t.Fatalf("expected no-match fall-forward answer, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceContentsListPrefersNamedContainerOverOverlappingLocation(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "The TV box contains the remote.",
				DisplayResponse: "The TV box contains the remote.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "What's in the TV box?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	livingRoom := assetItem("living-room-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	livingRoomTitle, _ := asset.NewTitle("Living room")
	livingRoom.Title = livingRoomTitle
	tvBox := assetItem("tv-box-1", "tenant-home", "inventory-home", asset.KindContainer, "living-room-1")
	tvBoxTitle, _ := asset.NewTitle("TV box")
	tvBox.Title = tvBoxTitle
	remote := assetItem("remote-1", "tenant-home", "inventory-home", asset.KindItem, "tv-box-1")
	remoteTitle, _ := asset.NewTitle("Remote")
	remote.Title = remoteTitle
	seedRealtimeVoiceLoopAsset(t, store, livingRoom, "audit-living-room")
	seedRealtimeVoiceLoopAsset(t, store, tvBox, "audit-tv-box")
	seedRealtimeVoiceLoopAsset(t, store, remote, "audit-remote")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}
	if len(language.seenToolResults) != 1 || len(language.seenToolResults[0]) < 2 {
		t.Fatalf("expected server-selected contents list result, got %+v", language.seenToolResults)
	}
	result := language.seenToolResults[0][1]
	if result.Name != RealtimeVoiceToolListAuthorizedAssets || !containsAll(result.Content, `"parentTitle":"TV box"`, "Remote") {
		t.Fatalf("expected contents list to target TV box, got %+v", result)
	}
}

func TestRealtimeVoiceReadsDestinationBeforePlanningCreateInNamedParent(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-phone-charger",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Add a phone charger to the office.",
					"modelInterpretationSummary": "Create the phone charger inside the visible Office location.",
					"confirmationSummary":        "Add a phone charger to the office?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-phone-charger",
							"kind":    "create_asset",
							"summary": "Create phone charger in Office",
							"arguments": map[string]any{
								"title":         "phone charger",
								"kind":          "item",
								"parentAssetId": "office-1",
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
	seedRealtimeVoiceLoopAsset(t, store, office, "audit-office")

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
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}
	if len(language.seenToolResults) != 1 || len(language.seenToolResults[0]) != 1 || language.seenToolResults[0][0].Name != RealtimeVoiceToolSearchAuthorizedAssets || !containsAll(language.seenToolResults[0][0].Content, `"query":"office"`, "office-1") {
		t.Fatalf("expected server-selected Office destination read, got %+v", language.seenToolResults)
	}
	if proposed == nil || len(proposed.Commands) != 1 || proposed.Commands[0].ParentAssetID != "office-1" {
		t.Fatalf("expected phone charger proposal inside Office, got %+v", proposed)
	}
}

func seedRealtimeVoiceLoopAsset(t *testing.T, store interface {
	CreateAsset(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error
}, item asset.Asset, auditID string) {
	t.Helper()
	if err := store.CreateAsset(context.Background(), item, audit.Record{ID: audit.ID(auditID), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: item.ID.String(), OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed asset %s: %v", item.ID, err)
	}
}

func containsAll(text string, terms ...string) bool {
	for _, term := range terms {
		if !strings.Contains(text, term) {
			return false
		}
	}
	return true
}
