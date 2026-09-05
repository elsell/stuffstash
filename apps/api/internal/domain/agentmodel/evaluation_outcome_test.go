package agentmodel

import "testing"

func TestEvaluationJudgesGroundedOutcomeInsteadOfProse(t *testing.T) {
	definition, err := NewEvaluationCaseDefinition(fixtureEvaluationInput())
	if err != nil {
		t.Fatal(err)
	}
	good := EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "attic"}}}
	verdict := definition.Evaluate(good)
	if !verdict.Passed || len(verdict.Failures) != 0 {
		t.Fatalf("grounded answer rejected: %+v", verdict)
	}
	cases := []struct {
		name   string
		change func(*EvaluationObservedOutcome)
		code   EvaluationFailureCode
	}{
		{"wrong kind", func(v *EvaluationObservedOutcome) { v.Kind = EvaluationOutcomeClarification }, EvaluationFailureOutcome},
		{"missing asset", func(v *EvaluationObservedOutcome) { v.ReferencedAssets = nil }, EvaluationFailureReference},
		{"missing location", func(v *EvaluationObservedOutcome) { v.Locations = nil }, EvaluationFailureLocation},
		{"forbidden create", func(v *EvaluationObservedOutcome) { v.ProposedOperations = []Operation{OperationCreate} }, EvaluationFailureForbiddenOperation},
		{"unapproved write", func(v *EvaluationObservedOutcome) { v.ExecutedOperations = []Operation{OperationMove} }, EvaluationFailureMutation},
		{"foreign asset", func(v *EvaluationObservedOutcome) { v.ReferencedAssets = []string{"live-inventory-id"} }, EvaluationFailureInvalidObservation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outcome := good
			tc.change(&outcome)
			verdict := definition.Evaluate(outcome)
			found := false
			for _, failure := range verdict.Failures {
				found = found || failure.Code == tc.code
			}
			if verdict.Passed || !found {
				t.Fatalf("bad verdict: %+v", verdict)
			}
		})
	}
}
func TestEvaluationRequiresProposedChangeWithoutExecutingIt(t *testing.T) {
	input := fixtureEvaluationInput()
	input.Expectations = EvaluationExpectations{Kind: EvaluationOutcomeProposal, ProposedOperations: []Operation{OperationMove}, ForbiddenOperations: []Operation{OperationCreate}}
	definition, err := NewEvaluationCaseDefinition(input)
	if err != nil {
		t.Fatal(err)
	}
	missing := definition.Evaluate(EvaluationObservedOutcome{Kind: EvaluationOutcomeProposal})
	if missing.Passed {
		t.Fatal("missing proposed move passed")
	}
	if verdict := definition.Evaluate(EvaluationObservedOutcome{Kind: EvaluationOutcomeProposal, ProposedOperations: []Operation{OperationMove}}); !verdict.Passed {
		t.Fatalf("valid proposal failed: %+v", verdict)
	}
	var empty EvaluationCaseDefinition
	if verdict := empty.Evaluate(EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer}); verdict.Passed {
		t.Fatal("zero definition passed")
	}
}
