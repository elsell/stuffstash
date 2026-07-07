package app

import (
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestValidateRealtimeVoiceFinalResponseRejectsUnsafeSpokenText(t *testing.T) {
	t.Parallel()

	for _, spoken := range []string{
		`{"tool":"search_authorized_assets","query":"water bottle"}`,
		`search_authorized_assets({"query":"water bottle"})`,
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

	err := validateRealtimeVoiceFinalResponse(ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindAnswer,
		SpokenResponse:  "Your water bottle is in the Office.",
		DisplayResponse: "Your water bottle is in the Office.",
	})
	if err != nil {
		t.Fatalf("expected natural final response to pass validation: %v", err)
	}
}
