package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestValidateRealtimeVoiceFinalResponseRejectsUnsafeSpokenText(t *testing.T) {
	t.Parallel()

	for _, spoken := range []string{
		`{"tool":"search_authorized_assets","query":"water bottle"}`,
		`I found this: {"title":"Water bottle","location":"Office"}`,
		`search_authorized_assets({"query":"water bottle"})`,
		`Call list_authorized_assets with the water bottle query.`,
		`Reasoning: I used chain of thought to find it.`,
		`The raw prompt says to call list_authorized_assets.`,
		`Provider response: stack trace from Gemini.`,
		`Use assetId water-bottle-1 to find it next time.`,
		`Authorization: bearer abc/def==`,
		`apiKey: should-not-be-spoken`,
	} {
		err := validateRealtimeVoiceFinalResponse(ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  spoken,
			DisplayResponse: "Safe display text.",
		})
		if !errors.Is(err, ports.ErrInvalidProviderInput) {
			t.Fatalf("expected unsafe spoken response to be rejected for %q, got %v", spoken, err)
		}
	}
}

func TestValidateRealtimeVoiceFinalResponseRejectsUnsafeDisplayText(t *testing.T) {
	t.Parallel()

	err := validateRealtimeVoiceFinalResponse(ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindAnswer,
		SpokenResponse:  "Your water bottle is in the Office.",
		DisplayResponse: `{"assetId":"water-bottle-1","providerResponse":"raw output"}`,
	})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected unsafe display response to be rejected, got %v", err)
	}
}

func TestValidateRealtimeVoiceFinalResponseAllowsNaturalInventoryAnswer(t *testing.T) {
	t.Parallel()

	for _, spoken := range []string{
		"Your water bottle is in the Office.",
		"Password notebook: office drawer.",
		"Authorization form: filing cabinet.",
		"Token board game: closet.",
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

func TestRealtimeVoiceRecoversUnsafeFinalResponseBeforeMobileOrTTS(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{
		Final: &ports.StructuredAgentResponse{
			Kind:            ports.StructuredAgentResponseKindAnswer,
			SpokenResponse:  `I found this: {"assetId":"water-bottle-1","title":"Water bottle"}`,
			DisplayResponse: `Call list_authorized_assets with assetId water-bottle-1.`,
		},
	}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is my water bottle?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}

	const recovered = "I could not finish that voice request safely. Please try again with a little more detail."
	if tts.lastText != recovered {
		t.Fatalf("expected only recovered safe response to reach TTS, got %q", tts.lastText)
	}
	for _, event := range events {
		if event.Response != nil && (strings.Contains(event.Response.SpokenResponse, "assetId") || strings.Contains(event.Response.DisplayResponse, "list_authorized_assets")) {
			t.Fatalf("unsafe final response reached mobile event: %+v", event.Response)
		}
	}
}
