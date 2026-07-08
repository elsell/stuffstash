package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoicePlannerOnlyTurnDoesNotAcceptFinalResponse(t *testing.T) {
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
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I checked out the drill.",
				DisplayResponse: "I checked out the drill.",
			},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindSafeFailure,
				SpokenResponse:  "I could not prepare that change safely.",
				DisplayResponse: "I could not prepare that change safely.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Check out the drill."}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	drillTitle, _ := asset.NewTitle("Drill")
	drill.Title = drillTitle
	if err := store.CreateAsset(context.Background(), drill, audit.Record{ID: audit.ID("audit-drill-planner-phase"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "drill-1", OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed drill: %v", err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:     session,
		AudioChunks: [][]byte{[]byte("audio")},
	}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}

	if len(language.seenPlanOnly) < 2 || !language.seenPlanOnly[1] {
		t.Fatalf("expected second turn to be planner-only, got %+v", language.seenPlanOnly)
	}
	if tts.lastText == "I checked out the drill." {
		t.Fatalf("planner-only final response must not be spoken")
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted &&
			event.Response != nil &&
			event.Response.SpokenResponse == "I checked out the drill." {
			t.Fatalf("planner-only final response must not complete the session, events=%+v", events)
		}
	}
}
