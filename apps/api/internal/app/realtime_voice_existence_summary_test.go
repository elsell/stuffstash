package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestVoiceExistenceAnswerAllowsSpokenSummaryWithCompleteDisplay(t *testing.T) {
	t.Parallel()
	brief := chemicalExistenceBrief()
	display := "You have Isopropyl alcohol, Acetone, and Mineral spirits."
	for _, spoken := range []string{
		"Yes, you have chemicals, including Isopropyl alcohol and Acetone.",
		"Yes, you have chemicals in your inventory.",
	} {
		if err := validateRealtimeVoiceGeneratedResponse(brief, ports.VoiceResponseGenerationResult{SpokenResponse: spoken, DisplayResponse: display}); err != nil {
			t.Errorf("supported conversational summary rejected: %q: %v", spoken, err)
		}
	}
}

func TestVoiceExistenceSummaryRejectsContradictionAndUnrelatedAnswer(t *testing.T) {
	t.Parallel()
	brief := chemicalExistenceBrief()
	for _, spoken := range []string{
		"You don't have any chemicals.",
		"You have passports in your inventory.",
	} {
		if err := validateRealtimeVoiceGeneratedResponse(brief, ports.VoiceResponseGenerationResult{
			SpokenResponse: spoken, DisplayResponse: "You have Isopropyl alcohol, Acetone, and Mineral spirits.",
		}); err == nil {
			t.Errorf("unsupported summary accepted: %q", spoken)
		}
	}
}

func chemicalExistenceBrief() agentmodel.GroundedVoiceResponseBrief {
	return agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeExists, Operation: agentmodel.OperationExists,
		Subject: "chemicals", Confidence: agentmodel.ResponseConfidenceStrong,
		Findings: []agentmodel.ResponseFinding{
			{FactKey: "finding.0", Title: "Isopropyl alcohol", Kind: "item"},
			{FactKey: "finding.1", Title: "Acetone", Kind: "item"},
			{FactKey: "finding.2", Title: "Mineral spirits", Kind: "item"},
		},
	}
}
