package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestValidateRealtimeVoiceFinalResponseEnforcesSpokenBounds(t *testing.T) {
	t.Parallel()
	for _, spoken := range []string{"", " ", strings.Repeat("x", 501), strings.Repeat(" ", 501) + "Hello", string([]byte{0xff})} {
		err := validateRealtimeVoiceFinalResponse(ports.StructuredAgentResponse{Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: spoken, DisplayResponse: "Display text."})
		if !errors.Is(err, ports.ErrInvalidProviderInput) {
			t.Fatalf("invalid spoken envelope accepted: %v", err)
		}
	}
}
func TestValidateRealtimeVoiceFinalResponseEnforcesDisplayBounds(t *testing.T) {
	t.Parallel()
	for _, display := range []string{" ", strings.Repeat("x", 1001), string([]byte{0xff})} {
		err := validateRealtimeVoiceFinalResponse(ports.StructuredAgentResponse{Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: "Hello.", DisplayResponse: display})
		if !errors.Is(err, ports.ErrInvalidProviderInput) {
			t.Fatalf("invalid display envelope accepted: %v", err)
		}
	}
}

func TestValidateRealtimeVoiceFinalResponseAllowsNaturalInventoryAnswer(t *testing.T) {
	t.Parallel()

	for _, spoken := range []string{
		"Your water bottle is in the Office.",
		"Password notebook: office drawer.",
		"Authorization form: filing cabinet.",
		"Token board game: closet.",
		"Your Chain of Thought book is in the office.",
		"Your Provider Response notes are in the cabinet.",
	} {
		err := validateRealtimeVoiceFinalResponse(ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  spoken,
			DisplayResponse: spoken,
		})
		if err != nil {
			t.Fatalf("expected natural final response to pass validation for %q: %v", spoken, err)
		}
	}
}

func TestCompleteRealtimeVoiceResponseValidatesBeforeMobileOrTTS(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.TextToSpeech = tts
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.completeRealtimeVoiceResponse(context.Background(), session, ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindAnswer,
		SpokenResponse:  strings.Repeat("x", 501),
		DisplayResponse: "Display text.",
	}, nil, nil, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected invalid provider input, got %v", err)
	}
	if tts.lastText != "" {
		t.Fatalf("unsafe response reached TTS: %q", tts.lastText)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseStarted || event.Type == RealtimeVoiceEventAssistantResponseCompleted {
			t.Fatalf("unsafe response reached mobile event: %+v", event)
		}
	}
}

func TestRealtimeVoiceConversationDeliversModelAnswerToSpeechAndMobile(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.TextToSpeech = tts
	application := newRealtimeVoiceResolutionTestApp(t, resolver)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("conversation failed: %v", err)
	}
	if tts.lastText == "" {
		t.Fatal("model answer did not reach TTS")
	}
	completed := false
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted {
			completed = true
		}
	}
	if !completed {
		t.Fatal("model answer did not reach mobile")
	}

}
