package agentmodel

import (
	"testing"
	"time"
)

func TestEvaluationCaseRevisionValidatesIdentityAndOwnsDefinition(t *testing.T) {
	settings := EvaluationCaseDefinitionInput{Title: "Locate", Utterance: "Where are my clothes?", Assets: []EvaluationFixtureAsset{{ID: "clothes", Title: "Clothes", Kind: EvaluationFixtureItem, TagNames: []string{"baby"}}}, Expectations: EvaluationExpectations{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}}
	definition, err := NewEvaluationCaseDefinition(settings)
	if err != nil {
		t.Fatal(err)
	}
	input := EvaluationCaseRevisionInput{ID: "revision-1", CaseID: "case-1", TenantID: "home", AuthorID: "owner", Number: 1, Definition: definition, CreatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)}
	revision, err := NewEvaluationCaseRevision(input)
	if err != nil {
		t.Fatal(err)
	}
	settings.Assets[0].TagNames[0] = "changed"
	copy := revision.Snapshot().Definition.Settings()
	copy.Assets[0].TagNames[0] = "changed again"
	if revision.Snapshot().Definition.Settings().Assets[0].TagNames[0] != "baby" {
		t.Fatal("immutable case changed")
	}
	for _, mutate := range []func(*EvaluationCaseRevisionInput){
		func(v *EvaluationCaseRevisionInput) { v.ID = "../revision" },
		func(v *EvaluationCaseRevisionInput) { v.CaseID = "" },
		func(v *EvaluationCaseRevisionInput) { v.TenantID = " " },
		func(v *EvaluationCaseRevisionInput) { v.AuthorID = "" },
		func(v *EvaluationCaseRevisionInput) { v.Number = 0 },
		func(v *EvaluationCaseRevisionInput) { v.CreatedAt = time.Time{} },
		func(v *EvaluationCaseRevisionInput) { v.Definition = EvaluationCaseDefinition{} },
	} {
		invalid := input
		mutate(&invalid)
		if _, err := NewEvaluationCaseRevision(invalid); err == nil {
			t.Fatal("invalid case revision accepted")
		}
	}
}
