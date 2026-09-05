package agentmodel

import "testing"

func TestEvaluationRejectsWrongTargetsDestinationsAndDuplicateCommands(t *testing.T) {
	input := fixtureEvaluationInput()
	input.Expectations = EvaluationExpectations{Kind: EvaluationOutcomeProposal, Proposals: []EvaluationProposal{{Operation: OperationMove, TargetID: "clothes", DestinationID: "attic"}}}
	definition, err := NewEvaluationCaseDefinition(input)
	if err != nil {
		t.Fatal(err)
	}
	good := EvaluationObservedOutcome{Kind: EvaluationOutcomeProposal, Proposals: []EvaluationProposal{{Operation: OperationMove, TargetID: "clothes", DestinationID: "attic"}}}
	if result := definition.Evaluate(good); !result.Passed {
		t.Fatalf("correct move failed: %+v", result)
	}
	for _, proposals := range [][]EvaluationProposal{
		{{Operation: OperationMove, TargetID: "box", DestinationID: "attic"}},
		{{Operation: OperationMove, TargetID: "clothes", DestinationID: "box"}},
		{good.Proposals[0], good.Proposals[0]},
		{{Operation: OperationMove, TargetID: "live-asset", DestinationID: "attic"}},
	} {
		if result := definition.Evaluate(EvaluationObservedOutcome{Kind: EvaluationOutcomeProposal, Proposals: proposals}); result.Passed {
			t.Fatalf("wrong command passed: %+v", proposals)
		}
	}
}
func TestEvaluationCreateAssertionsDescribeAnAdditionalItem(t *testing.T) {
	input := fixtureEvaluationInput()
	proposal := EvaluationProposal{Operation: OperationCreate, NewTitle: "3–6 months clothes", NewKind: EvaluationFixtureItem, DestinationID: "box"}
	input.Expectations = EvaluationExpectations{Kind: EvaluationOutcomeProposal, Proposals: []EvaluationProposal{proposal}}
	definition, err := NewEvaluationCaseDefinition(input)
	if err != nil {
		t.Fatal(err)
	}
	if result := definition.Evaluate(EvaluationObservedOutcome{Kind: EvaluationOutcomeProposal, Proposals: []EvaluationProposal{proposal}}); !result.Passed {
		t.Fatalf("additional item failed: %+v", result)
	}
	proposal.TargetID = "clothes"
	if result := definition.Evaluate(EvaluationObservedOutcome{Kind: EvaluationOutcomeProposal, Proposals: []EvaluationProposal{proposal}}); result.Passed {
		t.Fatal("create reused an existing identity")
	}
}
