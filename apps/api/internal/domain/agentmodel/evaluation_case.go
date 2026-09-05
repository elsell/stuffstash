package agentmodel

import (
	"errors"
	"slices"
	"strings"
)

var ErrInvalidEvaluationCase = errors.New("invalid evaluation case")

const MaxEvaluationFixtureAssets = 100
const MaxEvaluationFixtureDepth = 32

type EvaluationFixtureKind string

const (
	EvaluationFixtureItem      EvaluationFixtureKind = "item"
	EvaluationFixtureContainer EvaluationFixtureKind = "container"
	EvaluationFixtureLocation  EvaluationFixtureKind = "location"
)

type EvaluationOutcomeKind string

const (
	EvaluationOutcomeAnswer        EvaluationOutcomeKind = "answer"
	EvaluationOutcomeClarification EvaluationOutcomeKind = "clarification"
	EvaluationOutcomeProposal      EvaluationOutcomeKind = "proposal"
	EvaluationOutcomeFailure       EvaluationOutcomeKind = "failure"
)

type EvaluationFixtureAsset struct {
	ID          string
	Title       string
	Description string
	Kind        EvaluationFixtureKind
	ParentID    string
	TagNames    []string
}
type EvaluationLocationExpectation struct {
	AssetID    string
	AncestorID string
}
type EvaluationExpectations struct {
	Kind                EvaluationOutcomeKind
	ReferencedAssets    []string
	Locations           []EvaluationLocationExpectation
	ProposedOperations  []Operation
	ForbiddenOperations []Operation
}
type EvaluationCaseDefinitionInput struct {
	Title        string
	Utterance    string
	Assets       []EvaluationFixtureAsset
	Expectations EvaluationExpectations
}
type EvaluationCaseDefinition struct{ settings EvaluationCaseDefinitionInput }

func NewEvaluationCaseDefinition(input EvaluationCaseDefinitionInput) (EvaluationCaseDefinition, error) {
	input = cloneEvaluationCase(input)
	input.Title = strings.TrimSpace(input.Title)
	input.Utterance = strings.TrimSpace(input.Utterance)
	if !workflowTextWithin(input.Title, 160, false) || !workflowTextWithin(input.Utterance, 4000, false) || len(input.Assets) > MaxEvaluationFixtureAssets {
		return EvaluationCaseDefinition{}, ErrInvalidEvaluationCase
	}
	assets := map[string]EvaluationFixtureAsset{}
	for i := range input.Assets {
		value := &input.Assets[i]
		value.Title = strings.TrimSpace(value.Title)
		value.Description = strings.TrimSpace(value.Description)
		if !workflowIdentifierValid(value.ID) || !workflowTextWithin(value.Title, 160, false) || !workflowTextWithin(value.Description, 2000, true) || !validObservationTagNames(value.TagNames) {
			return EvaluationCaseDefinition{}, ErrInvalidEvaluationCase
		}
		switch value.Kind {
		case EvaluationFixtureItem, EvaluationFixtureContainer, EvaluationFixtureLocation:
		default:
			return EvaluationCaseDefinition{}, ErrInvalidEvaluationCase
		}
		if _, duplicate := assets[value.ID]; duplicate {
			return EvaluationCaseDefinition{}, ErrInvalidEvaluationCase
		}
		assets[value.ID] = *value
	}
	for _, value := range assets {
		seen := map[string]bool{value.ID: true}
		for parent := value.ParentID; parent != ""; {
			ancestor, found := assets[parent]
			if !found || ancestor.Kind == EvaluationFixtureItem || seen[parent] || len(seen) >= MaxEvaluationFixtureDepth {
				return EvaluationCaseDefinition{}, ErrInvalidEvaluationCase
			}
			seen[parent] = true
			parent = ancestor.ParentID
		}
	}
	if !validEvaluationExpectations(input.Expectations, assets) {
		return EvaluationCaseDefinition{}, ErrInvalidEvaluationCase
	}
	return EvaluationCaseDefinition{settings: input}, nil
}
func (value EvaluationCaseDefinition) Settings() EvaluationCaseDefinitionInput {
	return cloneEvaluationCase(value.settings)
}
func cloneEvaluationCase(value EvaluationCaseDefinitionInput) EvaluationCaseDefinitionInput {
	value.Assets = slices.Clone(value.Assets)
	for i := range value.Assets {
		value.Assets[i].TagNames = slices.Clone(value.Assets[i].TagNames)
	}
	value.Expectations.ReferencedAssets = slices.Clone(value.Expectations.ReferencedAssets)
	value.Expectations.Locations = slices.Clone(value.Expectations.Locations)
	value.Expectations.ProposedOperations = slices.Clone(value.Expectations.ProposedOperations)
	value.Expectations.ForbiddenOperations = slices.Clone(value.Expectations.ForbiddenOperations)
	return value
}
