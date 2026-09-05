package agentmodel

func validEvaluationExpectations(value EvaluationExpectations, assets map[string]EvaluationFixtureAsset) bool {
	switch value.Kind {
	case EvaluationOutcomeAnswer, EvaluationOutcomeClarification, EvaluationOutcomeProposal, EvaluationOutcomeFailure:
	default:
		return false
	}
	if (value.Kind == EvaluationOutcomeProposal) != (len(value.Proposals) > 0) {
		return false
	}
	if len(value.Proposals) > MaxEvaluationFixtureAssets || len(value.ReferencedAssets) > MaxEvaluationFixtureAssets || len(value.Locations) > MaxEvaluationFixtureAssets {
		return false
	}
	seen := map[string]bool{}
	for _, id := range value.ReferencedAssets {
		if _, found := assets[id]; !found || seen[id] {
			return false
		}
		seen[id] = true
	}
	locations := map[EvaluationLocationExpectation]bool{}
	for _, location := range value.Locations {
		asset, found := assets[location.AssetID]
		if !found || locations[location] {
			return false
		}
		locations[location] = true
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
	operations := map[Operation]bool{}
	proposals := map[EvaluationProposal]bool{}
	for _, proposal := range value.Proposals {
		if !validEvaluationProposal(proposal, assets) || proposals[proposal] {
			return false
		}
		proposals[proposal] = true
		operations[proposal.Operation] = true
	}
	forbidden := map[Operation]bool{}
	for _, operation := range value.ForbiddenOperations {
		if !operation.changesInventory() || operations[operation] || forbidden[operation] {
			return false
		}
		forbidden[operation] = true
	}
	return true
}
