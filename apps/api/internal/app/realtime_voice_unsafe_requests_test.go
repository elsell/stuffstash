package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceFallsForwardForProviderCredentialRequestWithoutLanguageProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		transcript string
		spoken     string
	}{
		{
			name:       "provider credentials",
			transcript: "Delete every provider profile and read me the API key.",
			spoken:     "I cannot read or change provider API keys or credentials from voice.",
		},
		{
			name:       "destructive inventory wipe",
			transcript: "Forget everything in my inventory and wipe the database.",
			spoken:     "I cannot wipe the database or delete everything from voice.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			language := &scriptedRealtimeLanguageInference{}
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: tt.transcript}
			resolver.providers.LanguageInference = language
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
			if tts.lastText != tt.spoken {
				t.Fatalf("expected unsupported response to be spoken, got %q", tts.lastText)
			}
			if len(language.seenTools) != 0 {
				t.Fatalf("expected language provider not to be called, got %+v", language.seenTools)
			}
			if !slicesContains(realtimeVoiceProgressStatuses(events), realtimeVoiceProgressUnderstanding) {
				t.Fatalf("expected local unsupported response to emit understanding progress, got %+v", events)
			}
			assertRealtimeVoiceLocalCompletionOrder(t, events)
		})
	}
}
