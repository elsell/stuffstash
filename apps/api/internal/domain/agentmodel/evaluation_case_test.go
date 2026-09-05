package agentmodel

import (
	"fmt"
	"strings"
	"testing"
)

func fixtureEvaluationInput() EvaluationCaseDefinitionInput {
	return EvaluationCaseDefinitionInput{Title: "Locate tagged baby clothes", Utterance: "Where are my baby clothes?", Assets: []EvaluationFixtureAsset{{ID: "attic", Title: "Attic", Kind: EvaluationFixtureLocation}, {ID: "box", Title: "Blue box", Kind: EvaluationFixtureContainer, ParentID: "attic"}, {ID: "clothes", Title: "3–6 months clothes", Kind: EvaluationFixtureItem, ParentID: "box", TagNames: []string{"Baby", "Clothes"}}}, Expectations: EvaluationExpectations{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "attic"}}, ForbiddenOperations: []Operation{OperationCreate}}}
}
func TestEvaluationCasePreservesIsolatedFixtureAndExpectations(t *testing.T) {
	input := fixtureEvaluationInput()
	definition, err := NewEvaluationCaseDefinition(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Assets[2].TagNames[0] = "Changed"
	input.Expectations.ReferencedAssets[0] = "missing"
	output := definition.Settings()
	if output.Assets[2].TagNames[0] != "Baby" || output.Expectations.ReferencedAssets[0] != "clothes" {
		t.Fatal("input mutated evaluation snapshot")
	}
	output.Assets[2].TagNames[0] = "Changed again"
	output.Expectations.Locations[0].AncestorID = "missing"
	if definition.Settings().Assets[2].TagNames[0] != "Baby" || definition.Settings().Expectations.Locations[0].AncestorID != "attic" {
		t.Fatal("output mutated evaluation snapshot")
	}
}
func TestEvaluationCaseRejectsInvalidFixtureGraphsAndExpectations(t *testing.T) {
	cases := []struct {
		name   string
		change func(*EvaluationCaseDefinitionInput)
	}{
		{"empty utterance", func(v *EvaluationCaseDefinitionInput) { v.Utterance = " " }},
		{"duplicate asset", func(v *EvaluationCaseDefinitionInput) { v.Assets[1].ID = "attic" }},
		{"item parent", func(v *EvaluationCaseDefinitionInput) { v.Assets[1].Kind = EvaluationFixtureItem }},
		{"missing parent", func(v *EvaluationCaseDefinitionInput) { v.Assets[2].ParentID = "live-asset" }},
		{"cycle", func(v *EvaluationCaseDefinitionInput) { v.Assets[0].ParentID = "clothes" }},
		{"unsupported kind", func(v *EvaluationCaseDefinitionInput) { v.Assets[2].Kind = "arbitrary" }},
		{"foreign expected asset", func(v *EvaluationCaseDefinitionInput) { v.Expectations.ReferencedAssets = []string{"live-asset"} }},
		{"wrong expected location", func(v *EvaluationCaseDefinitionInput) {
			v.Expectations.Locations[0] = EvaluationLocationExpectation{AssetID: "attic", AncestorID: "clothes"}
		}},
		{"unknown outcome", func(v *EvaluationCaseDefinitionInput) { v.Expectations.Kind = "arbitrary" }},
		{"proposal without operation", func(v *EvaluationCaseDefinitionInput) { v.Expectations.Kind = EvaluationOutcomeProposal }},
		{"answer with proposal", func(v *EvaluationCaseDefinitionInput) { v.Expectations.ProposedOperations = []Operation{OperationMove} }},
		{"contradictory operation", func(v *EvaluationCaseDefinitionInput) {
			v.Expectations.Kind = EvaluationOutcomeProposal
			v.Expectations.ProposedOperations = []Operation{OperationCreate}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := fixtureEvaluationInput()
			tc.change(&input)
			if _, err := NewEvaluationCaseDefinition(input); err == nil {
				t.Fatal("invalid evaluation case accepted")
			}
		})
	}
}

func TestEvaluationFixtureBoundaries(t *testing.T) {
	input := fixtureEvaluationInput()
	input.Expectations = EvaluationExpectations{Kind: EvaluationOutcomeAnswer}
	input.Assets = nil
	for i := 0; i < 32; i++ {
		parent := ""
		if i > 0 {
			parent = fmt.Sprintf("fixture-%d", i-1)
		}
		input.Assets = append(input.Assets, EvaluationFixtureAsset{ID: fmt.Sprintf("fixture-%d", i), Title: "Box", Kind: EvaluationFixtureContainer, ParentID: parent})
	}
	if _, err := NewEvaluationCaseDefinition(input); err != nil {
		t.Fatalf("valid depth rejected: %v", err)
	}
	input.Assets = append(input.Assets, EvaluationFixtureAsset{ID: "too-deep", Title: "Box", Kind: EvaluationFixtureContainer, ParentID: "fixture-31"})
	if _, err := NewEvaluationCaseDefinition(input); err == nil {
		t.Fatal("excessive depth accepted")
	}
	input = fixtureEvaluationInput()
	input.Utterance = strings.Repeat("x", 4001)
	if _, err := NewEvaluationCaseDefinition(input); err == nil {
		t.Fatal("unbounded utterance accepted")
	}
	input = fixtureEvaluationInput()
	input.Assets[2].TagNames = []string{strings.Repeat("x", 81)}
	if _, err := NewEvaluationCaseDefinition(input); err == nil {
		t.Fatal("unbounded tag accepted")
	}
}
