package agentmodel

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func projectionFixture(t *testing.T) domain.EvaluationCaseDefinition {
	t.Helper()
	definition, err := domain.NewEvaluationCaseDefinition(domain.EvaluationCaseDefinitionInput{
		Title: "Find baby clothes", Utterance: "Where are my baby clothes?",
		Assets: []domain.EvaluationFixtureAsset{
			{ID: "attic", Title: "Attic", Kind: domain.EvaluationFixtureLocation},
			{ID: "box", Title: "Baby box", Kind: domain.EvaluationFixtureContainer, ParentID: "attic"},
			{ID: "clothes", Title: "3 to 6 months", Kind: domain.EvaluationFixtureItem, ParentID: "box", TagNames: []string{"baby", "clothes"}},
		},
		Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []domain.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return definition
}
func projectionIDs() map[string]string {
	return map[string]string{"runtime-attic": "attic", "runtime-box": "box", "runtime-clothes": "clothes"}
}
func TestEvaluationResponseProjectionRequiresDisplayedEvidence(t *testing.T) {
	definition := projectionFixture(t)
	ids := projectionIDs()
	projector, err := NewEvaluationProjector(definition, ids)
	if err != nil {
		t.Fatal(err)
	}
	delete(ids, "runtime-clothes")
	response := ports.StructuredAgentResponse{Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: "3 to 6 months is in Baby box.", DisplayResponse: "3 to 6 months is in Baby box.", Artifacts: []ports.StructuredAgentResponseArtifact{
		{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: "runtime-clothes", Title: "3 to 6 months", AssetKind: "item", Context: "Baby box"},
		{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: "runtime-box", Title: "Baby box", AssetKind: "container", Context: "Attic"},
	}}
	outcome, err := projector.Response(response)
	if err != nil || !definition.Evaluate(outcome).Passed {
		t.Fatalf("grounded displayed location lost: %+v %v", outcome, err)
	}
	if len(outcome.Locations) != 1 {
		t.Fatalf("undisplayed attic inferred: %+v", outcome.Locations)
	}
	response.Artifacts = response.Artifacts[:1]
	outcome, err = projector.Response(response)
	if err != nil || definition.Evaluate(outcome).Passed {
		t.Fatalf("fixture graph supplied missing parent evidence: %+v %v", outcome, err)
	}
	response.Artifacts[0].AssetID = "outside-fixture"
	if _, err := projector.Response(response); err == nil {
		t.Fatal("external asset accepted")
	}
	response.Artifacts = nil
	outcome, err = projector.Response(response)
	if err != nil || definition.Evaluate(outcome).Passed {
		t.Fatal("prose alone satisfied fixture references")
	}
}
func TestEvaluationProjectionRejectsIncompleteOrAliasedFixtureMaps(t *testing.T) {
	for name, ids := range map[string]map[string]string{
		"missing":       {"runtime-clothes": "clothes"},
		"alias":         {"runtime-attic": "attic", "runtime-box": "box", "runtime-clothes": "box"},
		"outside":       {"runtime-attic": "attic", "runtime-box": "box", "runtime-clothes": "outside"},
		"empty runtime": {"runtime-attic": "attic", "runtime-box": "box", "": "clothes"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewEvaluationProjector(projectionFixture(t), ids); err == nil {
				t.Fatal("invalid identity map accepted")
			}
		})
	}
}

func TestEvaluationProjectionCannotTreatInvalidEvidenceAsExpectedFailure(t *testing.T) {
	settings := projectionFixture(t).Settings()
	settings.Expectations = domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeFailure}
	definition, err := domain.NewEvaluationCaseDefinition(settings)
	if err != nil {
		t.Fatal(err)
	}
	projector, err := NewEvaluationProjector(definition, projectionIDs())
	if err != nil {
		t.Fatal(err)
	}
	for _, response := range []ports.StructuredAgentResponse{
		{Kind: "unknown"},
		{Kind: ports.StructuredAgentResponseKindSafeFailure},
		{Kind: ports.StructuredAgentResponseKindSafeFailure, SpokenResponse: " ", DisplayResponse: "I could not finish."},
		{Kind: ports.StructuredAgentResponseKindSafeFailure, SpokenResponse: "I could not finish.", DisplayResponse: " \n"},
		{Kind: ports.StructuredAgentResponseKindSafeFailure, SpokenResponse: "I could not finish.", DisplayResponse: "3 to 6 months", Artifacts: []ports.StructuredAgentResponseArtifact{{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: "outside", Title: "3 to 6 months", AssetKind: "item"}}},
		{Kind: ports.StructuredAgentResponseKindAnswer, SpokenResponse: "No items found.", DisplayResponse: "No items found.", Artifacts: []ports.StructuredAgentResponseArtifact{{Type: ports.StructuredAgentResponseArtifactAssetReference, AssetID: "runtime-clothes", Title: "3 to 6 months", AssetKind: "item"}}},
	} {
		if _, err := projector.Response(response); err == nil {
			t.Fatal("malformed response could satisfy an expected failure")
		}
	}
	outcome, err := projector.Response(ports.StructuredAgentResponse{Kind: ports.StructuredAgentResponseKindSafeFailure, SpokenResponse: "I could not finish. Please try again.", DisplayResponse: "I could not finish. Please try again."})
	if err != nil || !definition.Evaluate(outcome).Passed {
		t.Fatalf("valid safe failure not preserved: %+v %v", outcome, err)
	}
}
func TestEvaluationProposalProjectionPreservesFullCommandSemantics(t *testing.T) {
	projector, err := NewEvaluationProjector(projectionFixture(t), projectionIDs())
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name     string
		kind     actionplan.CommandKind
		args     string
		expected domain.EvaluationProposal
	}{
		{"move", actionplan.CommandKindMoveAsset, `{"assetId":"runtime-clothes","parentAssetId":"runtime-box"}`, domain.EvaluationProposal{Operation: domain.OperationMove, TargetID: "clothes", DestinationID: "box"}},
		{"create", actionplan.CommandKindCreateAsset, `{"title":"New charger","parentAssetId":"runtime-box"}`, domain.EvaluationProposal{Operation: domain.OperationCreate, NewTitle: "New charger", NewKind: domain.EvaluationFixtureItem, DestinationID: "box"}},
		{"location", actionplan.CommandKindCreateLocation, `{"title":"Loft"}`, domain.EvaluationProposal{Operation: domain.OperationCreate, NewTitle: "Loft", NewKind: domain.EvaluationFixtureLocation}},
		{"checkout", actionplan.CommandKindCheckoutAsset, `{"assetId":"runtime-clothes","details":"For Sam"}`, domain.EvaluationProposal{Operation: domain.OperationCheckout, TargetID: "clothes", Details: "For Sam"}},
		{"return", actionplan.CommandKindReturnAsset, `{"assetId":"runtime-clothes","details":"Washed"}`, domain.EvaluationProposal{Operation: domain.OperationReturn, TargetID: "clothes", Details: "Washed"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			outcome, err := projector.Proposal([]ports.ActionPlanCommandRecord{{Kind: scenario.kind, ArgumentsJSON: []byte(scenario.args)}})
			if err != nil || len(outcome.Proposals) != 1 || outcome.Proposals[0] != scenario.expected {
				t.Fatalf("semantics lost: %+v %v", outcome, err)
			}
		})
	}
}
func TestEvaluationProposalProjectionCannotHideUnsupportedArguments(t *testing.T) {
	projector, err := NewEvaluationProjector(projectionFixture(t), projectionIDs())
	if err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string]string{
		"new parent": `{"assetId":"runtime-clothes","parentCommandId":"create-room"}`,
		"outside":    `{"assetId":"other-tenant-item","parentAssetId":"runtime-box"}`,
		"unknown":    `{"assetId":"runtime-clothes","parentAssetId":"runtime-box","secretField":"ignored"}`,
		"wrong type": `{"assetId":123,"parentAssetId":"runtime-box"}`,
		"null":       `{"assetId":null,"parentAssetId":"runtime-box"}`,
		"duplicate":  `{"assetId":"outside","assetId":"runtime-clothes","parentAssetId":"runtime-box"}`,
		"trailing":   `{"assetId":"runtime-clothes"} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := projector.Proposal([]ports.ActionPlanCommandRecord{{Kind: actionplan.CommandKindMoveAsset, ArgumentsJSON: []byte(args)}}); err == nil {
				t.Fatal("unsupported semantics accepted")
			}
		})
	}
	if _, err := projector.Proposal([]ports.ActionPlanCommandRecord{{Kind: "unknown", ArgumentsJSON: []byte(`{}`)}}); err == nil {
		t.Fatal("unknown command accepted")
	}
}
