package agentmodel

import (
	"strings"
	"testing"
)

func TestEvaluationAssetTitlesFitProductionByteLimit(t *testing.T) {
	for _, scenario := range []struct {
		title string
		valid bool
	}{
		{strings.Repeat("x", 160), true},
		{strings.Repeat("家", 53) + "x", true},
		{strings.Repeat("家", 54), false},
		{strings.Repeat("x", 161), false},
		{string([]byte{0xff}), false},
	} {
		t.Run(scenario.title, func(t *testing.T) {
			fixture := fixtureEvaluationInput()
			fixture.Assets[2].Title = scenario.title
			_, err := NewEvaluationCaseDefinition(fixture)
			if (err == nil) != scenario.valid {
				t.Fatalf("fixture title accepted=%v, want %v", err == nil, scenario.valid)
			}
			fixture = fixtureEvaluationInput()
			fixture.Expectations = EvaluationExpectations{Kind: EvaluationOutcomeProposal, Proposals: []EvaluationProposal{{Operation: OperationCreate, NewTitle: scenario.title, NewKind: EvaluationFixtureItem}}}
			_, err = NewEvaluationCaseDefinition(fixture)
			if (err == nil) != scenario.valid {
				t.Fatalf("creation title accepted=%v, want %v", err == nil, scenario.valid)
			}
		})
	}
	fixture := fixtureEvaluationInput()
	fixture.Title = strings.Repeat("家", 160)
	if _, err := NewEvaluationCaseDefinition(fixture); err != nil {
		t.Fatal("case display title incorrectly used asset limit")
	}
}
