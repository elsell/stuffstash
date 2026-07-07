package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceSkipsDuplicateToolCallButExecutesUnseenCallsInSameTurn(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{
				{
					ID:        "search-water-bottle",
					Name:      RealtimeVoiceToolSearchAuthorizedAssets,
					Arguments: map[string]any{"query": "water bottle", "limit": float64(5)},
				},
				{
					ID:        "search-water-bottle-again",
					Name:      RealtimeVoiceToolSearchAuthorizedAssets,
					Arguments: map[string]any{"query": "water bottle", "limit": float64(5)},
				},
				{
					ID:        "list-visible-items",
					Name:      RealtimeVoiceToolListAuthorizedAssets,
					Arguments: map[string]any{"kind": "item", "lifecycleState": "active", "limit": float64(5)},
				},
			},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I found your water bottle and listed the visible items.",
				DisplayResponse: "I found your water bottle and listed the visible items.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is my water bottle and what items are visible?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "office-1")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	flashlight := assetItem("flashlight-1", "tenant-home", "inventory-home", asset.KindItem, "office-1")
	flashlightTitle, _ := asset.NewTitle("Flashlight")
	flashlight.Title = flashlightTitle
	seedRealtimeVoiceLoopAsset(t, store, office, "audit-office")
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")
	seedRealtimeVoiceLoopAsset(t, store, flashlight, "audit-flashlight")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}

	if len(language.seenToolResults) < 2 {
		t.Fatalf("expected final turn to receive tool results, got %+v", language.seenToolResults)
	}
	finalTurnResults := language.seenToolResults[1]
	if !realtimeVoiceDuplicateTestToolResultContent(finalTurnResults, RealtimeVoiceToolSearchAuthorizedAssets, `"assetId":"water-bottle-1"`) {
		t.Fatalf("expected search result to be returned to final turn, got %+v", finalTurnResults)
	}
	if !realtimeVoiceDuplicateTestToolResultContent(finalTurnResults, RealtimeVoiceToolListAuthorizedAssets, `"assetId":"flashlight-1"`) {
		t.Fatalf("expected unseen list call to execute after duplicate, got %+v", finalTurnResults)
	}
	if !realtimeVoiceDuplicateTestToolResultContent(finalTurnResults, RealtimeVoiceToolSearchAuthorizedAssets, `"code":"duplicate_tool_request"`) {
		t.Fatalf("expected duplicate request repair result in final turn, got %+v", finalTurnResults)
	}
	if !realtimeVoiceDuplicateTestEventCode(events, RealtimeVoiceEventToolCallFailed, "duplicate_tool_request") {
		t.Fatalf("expected duplicate tool failure event, got %+v", events)
	}
	if !realtimeVoiceDuplicateTestToolCompleted(events, "list-visible-items") {
		t.Fatalf("expected unseen list call to complete after duplicate, got %+v", events)
	}
	if tts.lastText != "I found your water bottle and listed the visible items." {
		t.Fatalf("expected final response to be spoken, got %q", tts.lastText)
	}
}

func realtimeVoiceDuplicateTestToolResultContent(results []ports.AgentToolResult, name string, term string) bool {
	for _, result := range results {
		if result.Name == name && strings.Contains(result.Content, term) {
			return true
		}
	}
	return false
}

func realtimeVoiceDuplicateTestEventCode(events []RealtimeVoiceEvent, eventType string, code string) bool {
	for _, event := range events {
		if event.Type == eventType && event.Code == code {
			return true
		}
	}
	return false
}

func realtimeVoiceDuplicateTestToolCompleted(events []RealtimeVoiceEvent, toolCallID string) bool {
	for _, event := range events {
		if event.Type == RealtimeVoiceEventToolCallCompleted && event.ToolCallID == toolCallID {
			return true
		}
	}
	return false
}
