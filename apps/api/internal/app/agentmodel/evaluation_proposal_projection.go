package agentmodel

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (p *EvaluationProjector) Proposal(commands []ports.ActionPlanCommandRecord) (domain.EvaluationObservedOutcome, error) {
	if len(commands) == 0 || len(commands) > domain.MaxEvaluationFixtureAssets {
		return domain.EvaluationObservedOutcome{}, ErrInvalidEvaluationObservation
	}
	outcome := domain.EvaluationObservedOutcome{Kind: domain.EvaluationOutcomeProposal}
	for _, command := range commands {
		proposal, err := p.command(command)
		if err != nil {
			return domain.EvaluationObservedOutcome{}, err
		}
		outcome.Proposals = append(outcome.Proposals, proposal)
	}
	return p.validated(outcome)
}

func (p *EvaluationProjector) command(command ports.ActionPlanCommandRecord) (domain.EvaluationProposal, error) {
	args, err := evaluationCommandArguments(command.ArgumentsJSON)
	if err != nil {
		return domain.EvaluationProposal{}, err
	}
	proposal := domain.EvaluationProposal{}
	allowed := map[string]bool{}
	switch command.Kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		proposal.Operation = domain.OperationCreate
		allowed = map[string]bool{"title": true, "kind": true, "parentAssetId": true}
		proposal.NewTitle = strings.TrimSpace(args["title"])
		proposal.NewKind = domain.EvaluationFixtureKind(args["kind"])
		if command.Kind == actionplan.CommandKindCreateLocation {
			if proposal.NewKind != "" && proposal.NewKind != domain.EvaluationFixtureLocation {
				return domain.EvaluationProposal{}, ErrInvalidEvaluationObservation
			}
			proposal.NewKind = domain.EvaluationFixtureLocation
		} else if proposal.NewKind == "" {
			proposal.NewKind = domain.EvaluationFixtureItem
		}
	case actionplan.CommandKindMoveAsset:
		proposal.Operation = domain.OperationMove
		allowed = map[string]bool{"assetId": true, "parentAssetId": true}
	case actionplan.CommandKindArchiveAsset:
		proposal.Operation = domain.OperationArchive
		allowed = map[string]bool{"assetId": true}
	case actionplan.CommandKindRestoreAsset:
		proposal.Operation = domain.OperationRestore
		allowed = map[string]bool{"assetId": true}
	case actionplan.CommandKindCheckoutAsset, actionplan.CommandKindReturnAsset:
		proposal.Operation = domain.OperationCheckout
		if command.Kind == actionplan.CommandKindReturnAsset {
			proposal.Operation = domain.OperationReturn
		}
		proposal.Details = strings.TrimSpace(args["details"])
		allowed = map[string]bool{"assetId": true, "details": true}
	default:
		return domain.EvaluationProposal{}, ErrInvalidEvaluationObservation
	}
	for key := range args {
		if !allowed[key] {
			return domain.EvaluationProposal{}, ErrInvalidEvaluationObservation
		}
	}
	if proposal.Operation != domain.OperationCreate {
		var found bool
		proposal.TargetID, found = p.runtimeIDs[args["assetId"]]
		if !found {
			return domain.EvaluationProposal{}, ErrInvalidEvaluationObservation
		}
	}
	if parent := args["parentAssetId"]; parent != "" {
		var found bool
		proposal.DestinationID, found = p.runtimeIDs[parent]
		if !found {
			return domain.EvaluationProposal{}, ErrInvalidEvaluationObservation
		}
	}
	return proposal, nil
}

// Decode each property once: ordinary map unmarshalling discards duplicate keys
// and would hide conflicting semantic arguments from the exact-command judge.
func evaluationCommandArguments(raw []byte) (map[string]string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, ErrInvalidEvaluationObservation
	}
	args := map[string]string{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return nil, ErrInvalidEvaluationObservation
		}
		if _, exists := args[key]; exists {
			return nil, ErrInvalidEvaluationObservation
		}
		token, err = decoder.Token()
		value, ok := token.(string)
		if err != nil || !ok {
			return nil, ErrInvalidEvaluationObservation
		}
		args[key] = value
	}
	token, err = decoder.Token()
	if err != nil || token != json.Delim('}') {
		return nil, ErrInvalidEvaluationObservation
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, ErrInvalidEvaluationObservation
	}
	return args, nil
}
