package agentmodel

import (
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"strings"
	"testing"
)

func TestGroundedVoiceRendererLocatesClothesWithoutInventingFacts(t *testing.T) {
	brief := domain.GroundedVoiceResponseBrief{Kind: domain.ResponseBriefKindAnswer, Mode: domain.ResponseAnswerModeLocate, Operation: domain.OperationLocate, Subject: "baby clothes", Confidence: domain.ResponseConfidenceStrong, Findings: []domain.ResponseFinding{{FactKey: "clothes", Title: "3–6 months clothes", Kind: "item", ContainmentPath: []string{"Attic", "Blue box", "3–6 months clothes"}}}}
	text, err := RenderGroundedVoiceResponse(brief, 500)
	if err != nil || text != "3–6 months clothes is in Attic / Blue box." {
		t.Fatalf("grounded location: %q %v", text, err)
	}
	brief.Confidence = domain.ResponseConfidencePlausible
	text, err = RenderGroundedVoiceResponse(brief, 500)
	if err != nil || !strings.Contains(text, "Possible matches for baby clothes") {
		t.Fatalf("uncertainty lost: %q %v", text, err)
	}
	brief.Findings[0].ContainmentPath = nil
	text, err = RenderGroundedVoiceResponse(brief, 500)
	if err != nil || !strings.Contains(text, "no recorded location") || strings.Contains(text, "Attic") {
		t.Fatalf("invented location: %q %v", text, err)
	}
	brief.Findings[0].Title = strings.Repeat("界", 150)
	brief.Truncated = true
	text, err = RenderGroundedVoiceResponse(brief, 500)
	if err != nil || len(text) > 500 || !strings.Contains(text, "only part") {
		t.Fatalf("unbounded or undisclosed response: %q %v", text, err)
	}
	if _, err = RenderGroundedVoiceResponse(domain.GroundedVoiceResponseBrief{}, 500); err == nil {
		t.Fatal("invalid brief accepted")
	}
}

func TestGroundedNotFoundPreservesAbsenceWhenSubjectNeedsAbbreviation(t *testing.T) {
	brief := domain.GroundedVoiceResponseBrief{Kind: domain.ResponseBriefKindAnswer, Mode: domain.ResponseAnswerModeNotFound, Operation: domain.OperationLocate, Subject: strings.Repeat("界", 160), Confidence: domain.ResponseConfidenceAbsent}
	text, err := RenderGroundedVoiceResponse(brief, 500)
	if err != nil || len(text) > 500 || !strings.Contains(text, "couldn't find") || strings.Contains(text, "I found results") {
		t.Fatalf("absence reversed: %q %v", text, err)
	}
}
func TestGroundedClarificationDistinguishesSameNameItems(t *testing.T) {
	brief := domain.GroundedVoiceResponseBrief{Kind: domain.ResponseBriefKindClarification, Mode: domain.ResponseAnswerModeClarify, Operation: domain.OperationLocate, Subject: "drill", Confidence: domain.ResponseConfidenceAmbiguous, Findings: []domain.ResponseFinding{{FactKey: "one", Title: "Drill", Kind: "item", ContainmentPath: []string{"Garage", "Drill"}}, {FactKey: "two", Title: "Drill", Kind: "item", ContainmentPath: []string{"Office", "Drill"}}}}
	text, err := RenderGroundedVoiceResponse(brief, 500)
	if err != nil || !strings.Contains(text, "Which drill") || !strings.Contains(text, "Garage") || !strings.Contains(text, "Office") {
		t.Fatalf("indistinguishable choices: %q %v", text, err)
	}
}
