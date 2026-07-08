package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
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
	seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill-planner-phase")

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

func TestRealtimeVoicePlannerOnlySafeFailureDoesNotClaimMutation(t *testing.T) {
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
				Kind:            ports.StructuredAgentResponseKindSafeFailure,
				SpokenResponse:  "I checked out the drill.",
				DisplayResponse: "I checked out the drill.",
			},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindSafeFailure,
				SpokenResponse:  "I could not prepare that checkout safely.",
				DisplayResponse: "I could not prepare that checkout safely.",
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
	seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill-planner-phase-claim")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:     session,
		AudioChunks: [][]byte{[]byte("audio")},
	}, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}

	if tts.lastText == "I checked out the drill." || strings.Contains(strings.ToLower(tts.lastText), "checked out") {
		t.Fatalf("planner-only safe failure must not claim mutation, got %q", tts.lastText)
	}
}
