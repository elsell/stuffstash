package agentmodel

import (
	"errors"
	"regexp"
	"strings"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrInvalidEvaluationObservation = errors.New("invalid evaluation observation")

// EvaluationProjector maps completed application outputs to isolated fixture
// identities. Provider assertions and retrieval candidates are not outcomes.
type EvaluationProjector struct {
	definition domain.EvaluationCaseDefinition
	runtimeIDs map[string]string
	fixtures   map[string]domain.EvaluationFixtureAsset
}

func NewEvaluationProjector(definition domain.EvaluationCaseDefinition, runtimeIDs map[string]string) (*EvaluationProjector, error) {
	validated, err := domain.NewEvaluationCaseDefinition(definition.Settings())
	if err != nil {
		return nil, ErrInvalidEvaluationObservation
	}
	fixtures := map[string]domain.EvaluationFixtureAsset{}
	for _, fixture := range validated.Settings().Assets {
		fixtures[fixture.ID] = fixture
	}
	if len(runtimeIDs) != len(fixtures) {
		return nil, ErrInvalidEvaluationObservation
	}
	copied := map[string]string{}
	seen := map[string]bool{}
	for runtimeID, fixtureID := range runtimeIDs {
		if _, found := fixtures[fixtureID]; !found || strings.TrimSpace(runtimeID) == "" || seen[fixtureID] {
			return nil, ErrInvalidEvaluationObservation
		}
		copied[runtimeID] = fixtureID
		seen[fixtureID] = true
	}
	return &EvaluationProjector{definition: validated, runtimeIDs: copied, fixtures: fixtures}, nil
}

func (p *EvaluationProjector) Response(response ports.StructuredAgentResponse) (domain.EvaluationObservedOutcome, error) {
	outcome := domain.EvaluationObservedOutcome{}
	if strings.TrimSpace(response.SpokenResponse) == "" || strings.TrimSpace(response.DisplayResponse) == "" {
		return outcome, ErrInvalidEvaluationObservation
	}
	switch response.Kind {
	case ports.StructuredAgentResponseKindAnswer:
		outcome.Kind = domain.EvaluationOutcomeAnswer
	case ports.StructuredAgentResponseKindClarification:
		outcome.Kind = domain.EvaluationOutcomeClarification
	case ports.StructuredAgentResponseKindSafeFailure, ports.StructuredAgentResponseKindUnsupportedAction:
		outcome.Kind = domain.EvaluationOutcomeFailure
	default:
		return outcome, ErrInvalidEvaluationObservation
	}
	displayed := map[string]ports.StructuredAgentResponseArtifact{}
	for _, artifact := range response.Artifacts {
		id, found := p.runtimeIDs[artifact.AssetID.String()]
		if !found {
			return domain.EvaluationObservedOutcome{}, ErrInvalidEvaluationObservation
		}
		fixture := p.fixtures[id]
		if _, duplicate := displayed[id]; duplicate || artifact.Type != ports.StructuredAgentResponseArtifactAssetReference || artifact.Title != fixture.Title || artifact.AssetKind.String() != string(fixture.Kind) || !evaluationDisplaysTitle(response.DisplayResponse, artifact.Title) {
			return domain.EvaluationObservedOutcome{}, ErrInvalidEvaluationObservation
		}
		displayed[id] = artifact
		outcome.ReferencedAssets = append(outcome.ReferencedAssets, id)
	}
	for _, id := range outcome.ReferencedAssets {
		fixture := p.fixtures[id]
		if fixture.ParentID == "" {
			continue
		}
		parent, shown := displayed[fixture.ParentID]
		if shown && displayed[id].Context == parent.Title {
			outcome.Locations = append(outcome.Locations, domain.EvaluationLocationExpectation{AssetID: id, AncestorID: fixture.ParentID})
		}
	}
	return p.validated(outcome)
}

func evaluationDisplaysTitle(text, title string) bool {
	expression := `(^|[^\pL\pN])` + regexp.QuoteMeta(title) + `($|[^\pL\pN])`
	matched, err := regexp.MatchString(expression, text)
	return err == nil && matched
}

func (p *EvaluationProjector) validated(outcome domain.EvaluationObservedOutcome) (domain.EvaluationObservedOutcome, error) {
	// A mismatched expectation is still a valid observation. Malformed evidence is
	// distinct so it cannot satisfy a case that intentionally expects failure.
	for _, failure := range p.definition.Evaluate(outcome).Failures {
		if failure.Code == domain.EvaluationFailureInvalidObservation {
			return domain.EvaluationObservedOutcome{}, ErrInvalidEvaluationObservation
		}
	}
	return outcome, nil
}
