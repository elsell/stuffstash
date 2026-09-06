package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
	"testing"
)

func TestRealtimeVoiceDiagnosticRedactionCoversHeaderBearerTokens(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		"Authorization: bearer abc/def==",
		"token: bearer abc/def==",
		"authorization=bearer abc/def==",
	} {
		redacted := safeRealtimeVoiceDiagnosticText(input, 500)
		if strings.Contains(redacted, "abc/def") || strings.Contains(strings.ToLower(redacted), "bearer ") || !strings.Contains(redacted, "[redacted") {
			t.Fatalf("expected bearer material to be redacted, input %q became %q", input, redacted)
		}
	}
}

func TestRealtimeVoiceTypedResponsePreservesTextToSpeechBoundaries(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name          string
		provider      ports.TextToSpeechProvider
		wantFailure   bool
		wantSafeError string
		wantChunks    int
	}{
		{name: "provider failure", provider: failingResolvedTextToSpeech{err: safeRealtimeVoiceDiagnosticFailure{safe: "provider_timeout"}}, wantFailure: true, wantSafeError: "provider_timeout"},
		{name: "malformed output", provider: malformedResolvedTextToSpeech{}, wantFailure: true, wantSafeError: "invalid_provider_output"},
		{name: "empty chunks compacted", provider: resolvedTextToSpeechWithChunks{chunks: [][]byte{[]byte("speech"), nil}}, wantChunks: 1},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated speech-boundary request"}
			resolver.providers.ConversationModel = &inventoryConversationModel{query: "Speech subject", answer: "The item is recorded in your inventory."}
			resolver.providers.TextToSpeech = testCase.provider
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("speech-subject", "Speech subject", asset.KindItem, ""), "audit-speech-subject")
			input := defaultRealtimeVoiceSessionInput()
			input.DeveloperDiagnostics = true
			session, err := application.StartRealtimeVoiceSession(context.Background(), input)
			if err != nil {
				t.Fatalf("start session: %v", err)
			}
			events := []RealtimeVoiceEvent{}
			err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
				events = append(events, event)
				return nil
			})
			if testCase.wantFailure {
				if err == nil || RealtimeVoiceSafeErrorCode(err) != realtimeVoiceFailureTextToSpeech {
					t.Fatalf("expected text-to-speech stage failure, got %v", err)
				}
				diagnostic := findRealtimeVoiceDiagnosticEvent(t, events, "Text-to-speech provider failed")
				if !strings.Contains(diagnostic.Detail, testCase.wantSafeError) || !strings.Contains(diagnostic.Detail, `"toolResultCount": 1`) {
					t.Fatalf("expected safe TTS diagnostic, got %s", diagnostic.Detail)
				}
				return
			}
			if err != nil {
				t.Fatalf("run voice entrypoint: %v", err)
			}
			audioChunks := []RealtimeVoiceEvent{}
			for _, event := range events {
				if event.Type == RealtimeVoiceEventTextToSpeechAudioChunk {
					audioChunks = append(audioChunks, event)
				}
			}
			if len(audioChunks) != testCase.wantChunks || !audioChunks[len(audioChunks)-1].FinalChunk {
				t.Fatalf("expected compacted final speech chunks, got %+v", audioChunks)
			}
		})
	}
}

func findRealtimeVoiceDiagnosticEvent(t *testing.T, events []RealtimeVoiceEvent, message string) RealtimeVoiceEvent {
	t.Helper()
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == message {
			return event
		}
	}
	t.Fatalf("expected diagnostic %q, got %+v", message, events)
	return RealtimeVoiceEvent{}
}

type safeRealtimeVoiceDiagnosticFailure struct{ safe string }

func (e safeRealtimeVoiceDiagnosticFailure) Error() string {
	return "provider failure with raw endpoint https://provider.invalid bearer should-not-leak"
}

func (e safeRealtimeVoiceDiagnosticFailure) SafeRealtimeVoiceDiagnostic() string { return e.safe }

type failingResolvedTextToSpeech struct{ err error }

func (f failingResolvedTextToSpeech) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{}, f.err
}

type malformedResolvedTextToSpeech struct{}

func (malformedResolvedTextToSpeech) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{nil}}, nil
}

type resolvedTextToSpeechWithChunks struct{ chunks [][]byte }

func (r resolvedTextToSpeechWithChunks) Synthesize(context.Context, ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: r.chunks}, nil
}
