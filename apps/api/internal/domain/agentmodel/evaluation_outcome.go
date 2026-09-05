package agentmodel

import "slices"

type EvaluationObservedOutcome struct {
	Kind               EvaluationOutcomeKind
	ReferencedAssets   []string
	Locations          []EvaluationLocationExpectation
	ProposedOperations []Operation
	ExecutedOperations []Operation
}
type EvaluationFailureCode string

const (
	EvaluationFailureInvalidObservation EvaluationFailureCode = "invalid_observation"
	EvaluationFailureOutcome            EvaluationFailureCode = "unexpected_outcome"
	EvaluationFailureReference          EvaluationFailureCode = "missing_reference"
	EvaluationFailureLocation           EvaluationFailureCode = "missing_location"
	EvaluationFailureProposal           EvaluationFailureCode = "missing_proposal"
	EvaluationFailureForbiddenOperation EvaluationFailureCode = "forbidden_operation"
	EvaluationFailureMutation           EvaluationFailureCode = "unexpected_mutation"
)

const EvaluationFailureUnexpectedProposal EvaluationFailureCode = "unexpected_proposal"

type EvaluationFailure struct {
	Code      EvaluationFailureCode
	FixtureID string
	Operation Operation
}
type EvaluationVerdict struct {
	Passed   bool
	Failures []EvaluationFailure
}

// Evaluate accepts outcomes collected by the conversation application; raw model
// assertions are not observations and must not be submitted as execution evidence.
func (definition EvaluationCaseDefinition) Evaluate(observed EvaluationObservedOutcome) EvaluationVerdict {
	if _, err := NewEvaluationCaseDefinition(definition.settings); err != nil || !definition.validObservedOutcome(observed) {
		return EvaluationVerdict{Failures: []EvaluationFailure{{Code: EvaluationFailureInvalidObservation}}}
	}
	expected := definition.settings.Expectations
	failures := []EvaluationFailure{}
	if observed.Kind != expected.Kind {
		failures = append(failures, EvaluationFailure{Code: EvaluationFailureOutcome})
	}
	for _, id := range expected.ReferencedAssets {
		if !slices.Contains(observed.ReferencedAssets, id) {
			failures = append(failures, EvaluationFailure{Code: EvaluationFailureReference, FixtureID: id})
		}
	}
	for _, location := range expected.Locations {
		if !slices.Contains(observed.Locations, location) {
			failures = append(failures, EvaluationFailure{Code: EvaluationFailureLocation, FixtureID: location.AssetID})
		}
	}
	for _, operation := range expected.ProposedOperations {
		if !slices.Contains(observed.ProposedOperations, operation) {
			failures = append(failures, EvaluationFailure{Code: EvaluationFailureProposal, Operation: operation})
		}
	}
	for _, operation := range observed.ProposedOperations {
		if !slices.Contains(expected.ProposedOperations, operation) {
			failures = append(failures, EvaluationFailure{Code: EvaluationFailureUnexpectedProposal, Operation: operation})
		}
	}
	for _, operation := range expected.ForbiddenOperations {
		if slices.Contains(observed.ProposedOperations, operation) || slices.Contains(observed.ExecutedOperations, operation) {
			failures = append(failures, EvaluationFailure{Code: EvaluationFailureForbiddenOperation, Operation: operation})
		}
	}
	if len(observed.ExecutedOperations) > 0 {
		failures = append(failures, EvaluationFailure{Code: EvaluationFailureMutation})
	}
	return EvaluationVerdict{Passed: len(failures) == 0, Failures: failures}
}

func (definition EvaluationCaseDefinition) validObservedOutcome(observed EvaluationObservedOutcome) bool {
	switch observed.Kind {
	case EvaluationOutcomeAnswer, EvaluationOutcomeClarification, EvaluationOutcomeProposal, EvaluationOutcomeFailure:
	default:
		return false
	}
	if len(observed.ReferencedAssets) > MaxEvaluationFixtureAssets || len(observed.Locations) > MaxEvaluationFixtureAssets || len(observed.ProposedOperations) > MaxEvaluationFixtureAssets || len(observed.ExecutedOperations) > MaxEvaluationFixtureAssets {
		return false
	}
	if len(observed.ProposedOperations) > 0 && observed.Kind != EvaluationOutcomeProposal {
		return false
	}
	assets := map[string]EvaluationFixtureAsset{}
	for _, asset := range definition.settings.Assets {
		assets[asset.ID] = asset
	}
	for _, id := range observed.ReferencedAssets {
		if _, found := assets[id]; !found {
			return false
		}
	}
	for _, location := range observed.Locations {
		asset, found := assets[location.AssetID]
		if !found {
			return false
		}
		matches := false
		for parent := asset.ParentID; parent != ""; parent = assets[parent].ParentID {
			if parent == location.AncestorID {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}
	for _, operations := range [][]Operation{observed.ProposedOperations, observed.ExecutedOperations} {
		for _, operation := range operations {
			if !operation.changesInventory() {
				return false
			}
		}
	}
	return true
}
