package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceAmbiguousDestinationCompletesClarificationWithoutLanguageProvider(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle over there."}
	resolver.providers.LanguageInference = language
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var completed *ports.StructuredAgentResponse
	var events []RealtimeVoiceEvent
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted {
			completed = event.Response
		}
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if completed == nil || completed.Kind != ports.StructuredAgentResponseKindClarification || completed.SpokenResponse != "I need to know where to move it before I can prepare that move." {
		t.Fatalf("expected ambiguous destination clarification, got %+v", completed)
	}
	if language.callCount != 0 {
		t.Fatalf("expected language provider not to be called, got %d calls", language.callCount)
	}
	if !slicesContains(realtimeVoiceProgressStatuses(events), realtimeVoiceProgressUnderstanding) {
		t.Fatalf("expected local clarification to emit understanding progress, got %+v", events)
	}
	assertRealtimeVoiceLocalCompletionOrder(t, events)
}

func TestRealtimeVoiceAmbiguousDestinationDoesNotCatchNamedPlaces(t *testing.T) {
	t.Parallel()

	if _, ok := realtimeVoiceAmbiguousDestinationTranscriptResponse("Move my water bottle to the side yard."); ok {
		t.Fatalf("did not expect named side yard to be ambiguous")
	}
}
