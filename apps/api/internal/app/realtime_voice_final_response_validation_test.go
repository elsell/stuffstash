package app

import (
	"context"
	"errors"
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
		`Use toolCallId call-123 to debug the request.`,
		`The tenantId is tenant-home and inventoryId is inventory-home.`,
		`The parentAssetId should be kitchen-1.`,
		`The tenant_id is tenant-home and inventory_id is inventory-home.`,
		`The parent-asset-id should be kitchen-1.`,
		`The tool call id is call-123.`,
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
		SpokenResponse:  "Use assetId water-bottle-1.",
		DisplayResponse: "Call search_authorized_assets.",
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

func TestRealtimeVoiceResponseGenerationFailureDoesNotReachMobileOrTTS(t *testing.T) {
	t.Parallel()

	tts := &resolvedTextToSpeech{}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.ResponseGenerator = failingVoiceResponseGenerator{err: errors.New("generation unavailable")}
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
	if realtimeVoiceErrorCode(err) != realtimeVoiceFailureLanguageInference {
		t.Fatalf("expected language inference failure, got %v", err)
	}
	if tts.lastText != "" {
		t.Fatalf("failed generated response reached TTS: %q", tts.lastText)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseStarted || event.Type == RealtimeVoiceEventAssistantResponseCompleted {
			t.Fatalf("failed generated response reached mobile: %+v", event)
		}
	}
}

type failingVoiceResponseGenerator struct {
	err error
}

func (f failingVoiceResponseGenerator) GenerateResponse(context.Context, ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	return ports.VoiceResponseGenerationResult{}, f.err
}
