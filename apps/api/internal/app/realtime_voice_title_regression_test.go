package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Household titles are data, not instructions or internal diagnostics. Valid
// grounded answers must survive words and punctuation in those titles.
func TestRealtimeVoiceGroundedAnswerAcceptsOrdinaryAssetTitles(t *testing.T) {
	t.Parallel()
	for _, title := range []string{"High-resolution camera", "Candidate board game", "Where's Wally?"} {
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			brief := agentmodel.GroundedVoiceResponseBrief{
				Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeLocate,
				Operation: agentmodel.OperationLocate, Subject: title, Confidence: agentmodel.ResponseConfidenceStrong,
				Findings: []agentmodel.ResponseFinding{{
					FactKey: "finding.0", Title: title, Kind: "item", ContainmentPath: []string{"Office", title},
				}},
			}
			response := ports.VoiceResponseGenerationResult{
				SpokenResponse: title + " is in the Office.", DisplayResponse: title + " is in the Office.",
			}
			if err := validateRealtimeVoiceGeneratedResponse(brief, response); err != nil {
				t.Fatalf("valid grounded answer for ordinary title %q was rejected: %v", title, err)
			}
		})
	}
}
