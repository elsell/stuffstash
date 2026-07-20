package agentmodel

import "testing"

func TestGroundedVoiceResponseBriefValidatesPresentationOnlyFacts(t *testing.T) {
	t.Parallel()
	valid := GroundedVoiceResponseBrief{
		Kind: ResponseBriefKindAnswer, Mode: ResponseAnswerModeLocate, Operation: OperationLocate,
		Subject: "tools", Confidence: ResponseConfidencePlausible,
		Findings: []ResponseFinding{{FactKey: "finding.0", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "Toolbox"}}},
	}
	if valid.Validate() != nil {
		t.Fatalf("expected valid grounded brief, got %+v", valid)
	}

	cases := []GroundedVoiceResponseBrief{
		{},
		{Kind: ResponseBriefKindAnswer, Mode: ResponseAnswerModeLocate, Operation: OperationLocate, Subject: "tools", Confidence: ResponseConfidencePlausible},
		{Kind: ResponseBriefKindClarification, Mode: ResponseAnswerModeLocate, Operation: OperationLocate, Subject: "tools", Confidence: ResponseConfidencePlausible, Findings: valid.Findings},
		{Kind: ResponseBriefKindAnswer, Mode: ResponseAnswerModeLocate, Operation: OperationLocate, Subject: "tools", Confidence: ResponseConfidence("maybe"), Findings: valid.Findings},
	}
	for _, input := range cases {
		if input.Validate() == nil {
			t.Fatalf("expected invalid brief: %+v", input)
		}
	}
}

func TestGroundedVoiceResponseBriefRejectsInternalIdentifiers(t *testing.T) {
	t.Parallel()
	brief := GroundedVoiceResponseBrief{
		Kind: ResponseBriefKindAnswer, Mode: ResponseAnswerModeLocate, Operation: OperationLocate,
		Subject: "tools", Confidence: ResponseConfidenceStrong,
		Findings: []ResponseFinding{{FactKey: "finding.0", Title: "Toolbox", Kind: "container", ContainmentPath: []string{"Garage", "asset-id: toolbox-1"}}},
	}
	if brief.Validate() == nil {
		t.Fatal("expected internal identifier-shaped presentation fact to be rejected")
	}
}

func TestGroundedVoiceResponseBriefRejectsUnboundedPresentationTextAndUnknownState(t *testing.T) {
	t.Parallel()
	brief := GroundedVoiceResponseBrief{
		Kind: ResponseBriefKindAnswer, Mode: ResponseAnswerModeDetail, Operation: OperationDetail,
		Subject: "item", Confidence: ResponseConfidenceStrong,
		Findings: []ResponseFinding{{FactKey: "finding.0", Title: string(make([]byte, 161)), Kind: "item"}},
	}
	if brief.Validate() == nil {
		t.Fatal("expected overlong presentation title to fail")
	}
	brief.Findings[0].Title = "Item"
	brief.Findings[0].CheckoutState = "somewhere"
	if brief.Validate() == nil {
		t.Fatal("expected unknown checkout state to fail")
	}
}
