package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceFallsForwardForProviderCredentialRequestWithoutLanguageProvider(t *testing.T) {
	t.Parallel()

	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Delete every provider profile and read me the API key."}
	resolver.providers.LanguageInference = failingRealtimeVoiceLanguageInference{}
	tts := &resolvedTextToSpeech{}
	resolver.providers.TextToSpeech = tts
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
	if completed == nil || completed.Kind != ports.StructuredAgentResponseKindUnsupportedAction {
		t.Fatalf("expected unsupported fall-forward response, got %+v", completed)
	}
	if tts.lastText != "I cannot read or change provider API keys or credentials from voice." {
		t.Fatalf("expected unsupported response to be spoken, got %q", tts.lastText)
	}
	if !slicesContains(realtimeVoiceProgressStatuses(events), realtimeVoiceProgressUnderstanding) {
		t.Fatalf("expected local unsupported response to emit understanding progress, got %+v", events)
	}
	assertRealtimeVoiceLocalCompletionOrder(t, events)
}

type failingRealtimeVoiceLanguageInference struct{}

func (failingRealtimeVoiceLanguageInference) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
}
