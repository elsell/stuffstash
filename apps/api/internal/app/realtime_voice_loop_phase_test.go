package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
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

func TestRealtimeVoiceAnswersMovementHistoryFromAssetAuditTool(t *testing.T) {
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
			ToolCalls: []ports.AgentToolCall{{
				ID:        "history-water-bottle",
				Name:      RealtimeVoiceToolListAssetAuditHistory,
				Arguments: map[string]any{"assetId": "water-bottle-1", "limit": float64(5)},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "You moved the water bottle from the Office to the Kitchen on June 28, 2026.",
				DisplayResponse: "You moved the water bottle from the Office to the Kitchen on June 28, 2026.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "When did I move my water bottle to the kitchen?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	kitchen := assetItem("kitchen-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	kitchenTitle, _ := asset.NewTitle("Kitchen")
	kitchen.Title = kitchenTitle
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "kitchen-1")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	seedRealtimeVoiceLoopAsset(t, store, office, "audit-office")
	seedRealtimeVoiceLoopAsset(t, store, kitchen, "audit-kitchen")
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")
	for index := 0; index < 25; index++ {
		if err := store.SaveAuditRecord(context.Background(), audit.Record{
			ID:          audit.ID("audit-unrelated-" + strings.Repeat("x", index%3) + string(rune('a'+index))),
			TenantID:    audit.TenantID("tenant-home"),
			InventoryID: audit.InventoryID("inventory-home"),
			PrincipalID: audit.PrincipalID("user-1"),
			Action:      audit.ActionAssetViewed,
			Source:      audit.SourceAPI,
			TargetType:  audit.TargetAsset,
			TargetID:    "office-1",
			OccurredAt:  time.Date(2026, 6, 27, 9, index, 0, 0, time.UTC),
			Metadata:    map[string]string{"asset_kind": "location"},
		}); err != nil {
			t.Fatalf("seed unrelated audit: %v", err)
		}
	}
	if err := store.SaveAuditRecord(context.Background(), audit.Record{
		ID:          audit.ID("audit-water-bottle-moved"),
		TenantID:    audit.TenantID("tenant-home"),
		InventoryID: audit.InventoryID("inventory-home"),
		PrincipalID: audit.PrincipalID("user-1"),
		Action:      audit.ActionAssetMoved,
		Source:      audit.SourceAPI,
		TargetType:  audit.TargetAsset,
		TargetID:    "water-bottle-1",
		OccurredAt:  time.Date(2026, 6, 28, 9, 30, 0, 0, time.UTC),
		Metadata: map[string]string{
			"asset_kind":      "item",
			"previous_parent": "office-1",
			"new_parent":      "kitchen-1",
			"operation_id":    "hidden-operation-id",
		},
	}); err != nil {
		t.Fatalf("seed move audit: %v", err)
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

	if len(language.seenToolResults) < 3 || len(language.seenToolResults[2]) < 2 {
		t.Fatalf("expected search and audit tool results before final answer, got %+v", language.seenToolResults)
	}
	history := language.seenToolResults[2][1]
	if history.Name != RealtimeVoiceToolListAssetAuditHistory {
		t.Fatalf("expected audit history tool result, got %+v", history)
	}
	for _, required := range []string{"asset.moved", "Office", "Kitchen", "2026-06-28T09:30:00Z"} {
		if !strings.Contains(history.Content, required) {
			t.Fatalf("expected audit history to include %q, got %s", required, history.Content)
		}
	}
	if strings.Contains(history.Content, "hidden-operation-id") {
		t.Fatalf("expected operation id to be redacted from audit tool result, got %s", history.Content)
	}
	if strings.Contains(history.Content, "audit_record.listed") || strings.Contains(history.Content, "asset.viewed") {
		t.Fatalf("expected audit history result to include only target asset history, got %s", history.Content)
	}
	auditRecords, err := store.ListInventoryAuditRecords(context.Background(), tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), ports.AuditRecordPageRequest{Limit: 200})
	if err != nil {
		t.Fatalf("list audit records after voice history query: %v", err)
	}
	auditHistoryReads := 0
	for _, record := range auditRecords {
		if record.Action == audit.ActionAuditRecordListed {
			auditHistoryReads++
		}
	}
	if auditHistoryReads != 1 {
		t.Fatalf("expected one audit-history read audit record, got %d in %+v", auditHistoryReads, auditRecords)
	}
	if tts.lastText != "You moved the water bottle from the Office to the Kitchen on June 28, 2026." {
		t.Fatalf("expected history answer to be spoken, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceAuditHistoryRejectsAssetNotVisibleInSession(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	_, _, err = application.executeRealtimeVoiceTool(context.Background(), session, "When did I move my water bottle?", nil, ports.AgentToolCall{
		ID:        "history-water-bottle",
		Name:      RealtimeVoiceToolListAssetAuditHistory,
		Arguments: map[string]any{"assetId": "water-bottle-1"},
	}, map[string]struct{}{})
	if err == nil {
		t.Fatalf("expected unseen asset audit history request to be rejected")
	}
}

func TestRealtimeVoiceAuditHistoryToolAddsOnlyOneAuditReadRecord(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "office-1")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	seedRealtimeVoiceLoopAsset(t, store, office, "audit-office")
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")
	if err := store.SaveAuditRecord(context.Background(), audit.Record{
		ID:          audit.ID("audit-water-bottle-moved"),
		TenantID:    audit.TenantID("tenant-home"),
		InventoryID: audit.InventoryID("inventory-home"),
		PrincipalID: audit.PrincipalID("user-1"),
		Action:      audit.ActionAssetMoved,
		Source:      audit.SourceAPI,
		TargetType:  audit.TargetAsset,
		TargetID:    "water-bottle-1",
		OccurredAt:  time.Date(2026, 6, 28, 9, 30, 0, 0, time.UTC),
		Metadata: map[string]string{
			"asset_kind":      "item",
			"previous_parent": "",
			"new_parent":      "office-1",
		},
	}); err != nil {
		t.Fatalf("seed move audit: %v", err)
	}
	before, err := store.ListInventoryAuditRecords(context.Background(), tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), ports.AuditRecordPageRequest{Limit: 200})
	if err != nil {
		t.Fatalf("list audit records before history query: %v", err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	result, _, err := application.executeRealtimeVoiceTool(context.Background(), session, "When did I move my water bottle?", nil, ports.AgentToolCall{
		ID:        "history-water-bottle",
		Name:      RealtimeVoiceToolListAssetAuditHistory,
		Arguments: map[string]any{"assetId": "water-bottle-1"},
	}, map[string]struct{}{"water-bottle-1": {}})
	if err != nil {
		t.Fatalf("execute audit history tool: %v", err)
	}
	if !strings.Contains(result.Content, "Office") || !strings.Contains(result.Content, "asset.moved") {
		t.Fatalf("expected safe history output, got %s", result.Content)
	}
	after, err := store.ListInventoryAuditRecords(context.Background(), tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), ports.AuditRecordPageRequest{Limit: 200})
	if err != nil {
		t.Fatalf("list audit records after history query: %v", err)
	}
	newRecords := auditRecordsAfterSnapshot(before, after)
	if len(newRecords) != 1 || newRecords[0].Action != audit.ActionAuditRecordListed {
		t.Fatalf("expected exactly one audit history read record, got %+v", newRecords)
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
	if query := realtimeVoiceSpecificLookupObjectQuery("Where is it? Follow-up answer: Water bottle."); query != "water bottle" {
		t.Fatalf("expected follow-up read query to use concrete answer, got %q", query)
	}
	if query := realtimeVoiceSpecificLookupObjectQuery("Can you find it? Follow-up answer: The blue passport."); query != "blue passport" {
		t.Fatalf("expected follow-up find query to use concrete answer, got %q", query)
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

func TestRealtimeVoiceRetriesSpecificSingularWhereMissWithTranscriptObjectWords(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-electric-drill",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "cordless electric drill"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "Your cordless drill is in the Garage.",
				DisplayResponse: "Your cordless drill is in the Garage.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is my cordless drill?"}
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
	var events []RealtimeVoiceEvent
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}
	if len(language.seenToolResults) != 2 || len(language.seenToolResults[1]) < 2 {
		t.Fatalf("expected model final turn to receive original miss and narrow retry, got %+v", language.seenToolResults)
	}
	retry := language.seenToolResults[1][1]
	if retry.Name != RealtimeVoiceToolSearchAuthorizedAssets || !containsAll(retry.Content, `"query":"cordless drill"`, "cordless-drill-1", "Garage") {
		t.Fatalf("expected narrow transcript retry to find cordless drill, got %+v", retry)
	}
	statuses := realtimeVoiceProgressStatuses(events)
	for _, expected := range []string{"understanding", "exploring", "answering"} {
		if !slicesContains(statuses, expected) {
			t.Fatalf("expected progress status %q in %+v", expected, statuses)
		}
	}
	if slicesContains(statuses, "thinking") {
		t.Fatalf("expected bounded phase statuses, got %+v", statuses)
	}
	if tts.lastText != "Your cordless drill is in the Garage." {
		t.Fatalf("expected grounded drill answer, got %q", tts.lastText)
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

func TestRealtimeVoiceDerivesEffectiveTranscriptForClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I understand the follow-up.",
				DisplayResponse: "I understand the follow-up.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Kitchen."}
	resolver.providers.LanguageInference = language
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:                    session,
		AudioChunks:                [][]byte{[]byte("audio")},
		ContinueAfterClarification: true,
		ConversationTurns: []ports.AgentConversationTurn{
			{Role: ports.AgentConversationRoleUser, Text: "Move my water bottle."},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Where should I move it?"},
		},
	}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenTranscripts) == 0 || language.seenTranscripts[0] != "Move my water bottle. Follow-up answer: Kitchen." {
		t.Fatalf("expected model to receive effective follow-up transcript, got %+v", language.seenTranscripts)
	}
	if len(events) == 0 || events[0].Type != RealtimeVoiceEventTranscriptFinal || events[0].Text != "Kitchen." {
		t.Fatalf("expected client transcript event to keep literal follow-up transcript, got %+v", events)
	}
}

func TestRealtimeVoiceDerivesEffectiveTranscriptForReadClarificationFollowUp(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I understand the read follow-up.",
				DisplayResponse: "I understand the read follow-up.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Water bottle."}
	resolver.providers.LanguageInference = language
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:                    session,
		AudioChunks:                [][]byte{[]byte("audio")},
		ContinueAfterClarification: true,
		ConversationTurns: []ports.AgentConversationTurn{
			{Role: ports.AgentConversationRoleUser, Text: "Where is it?"},
			{Role: ports.AgentConversationRoleAssistant, Kind: string(ports.StructuredAgentResponseKindClarification), Text: "Which item should I find?"},
		},
	}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenTranscripts) == 0 || language.seenTranscripts[0] != "Where is it? Follow-up answer: Water bottle." {
		t.Fatalf("expected model to receive effective read follow-up transcript, got %+v", language.seenTranscripts)
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

func auditRecordsAfterSnapshot(before []audit.Record, after []audit.Record) []audit.Record {
	seen := map[audit.ID]struct{}{}
	for _, record := range before {
		seen[record.ID] = struct{}{}
	}
	newRecords := []audit.Record{}
	for _, record := range after {
		if _, ok := seen[record.ID]; ok {
			continue
		}
		newRecords = append(newRecords, record)
	}
	return newRecords
}

func containsAll(text string, terms ...string) bool {
	for _, term := range terms {
		if !strings.Contains(text, term) {
			return false
		}
	}
	return true
}

func realtimeVoiceProgressStatuses(events []RealtimeVoiceEvent) []string {
	statuses := []string{}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentProgress {
			statuses = append(statuses, event.Status)
		}
	}
	return statuses
}

func slicesContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
